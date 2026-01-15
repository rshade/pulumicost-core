# Quickstart: Pre-Flight Request Validation

**Feature**: 107-preflight-validation
**Date**: 2025-12-29

## Overview

Pre-flight validation catches malformed requests before they reach plugins, providing actionable error messages that help users fix issues quickly.

## What Changes

### Before (without pre-flight validation)

```bash
$ finfocus cost projected --pulumi-json plan.json

# Generic error from plugin (after waiting for gRPC call)
ERROR: rpc error: code = InvalidArgument desc = invalid request
```

### After (with pre-flight validation)

```bash
$ finfocus cost projected --pulumi-json plan.json

# Clear validation error with guidance (no plugin call made)
VALIDATION: SKU is empty: use mapping.ExtractAWSSKU() or mapping.ExtractSKU() to extract from resource properties

# Resource still appears in output with $0.00 cost
Resource                  | Monthly Cost | Notes
aws:ec2:Instance         | $0.00        | VALIDATION: SKU is empty...
```

## How to Fix Common Validation Errors

### "provider is empty"

**Cause**: The Pulumi plan doesn't include provider information for the resource.

**Fix**: Ensure your Pulumi program uses explicit provider configuration:

```typescript
// Pulumi TypeScript
const instance = new aws.ec2.Instance("my-instance", {
    instanceType: "t3.micro",
    // ...
});
```

### "SKU is empty"

**Cause**: The resource doesn't have instance type/size information.

**Fix**: Ensure your resource includes sizing properties:

```typescript
// EC2 Instance - instanceType is required
const instance = new aws.ec2.Instance("my-instance", {
    instanceType: "t3.micro",  // ← Required for cost calculation
    ami: "ami-12345678",
});

// RDS Instance - instanceClass is required
const db = new aws.rds.Instance("my-db", {
    instanceClass: "db.t3.micro",  // ← Required for cost calculation
    engine: "postgres",
});
```

### "region is empty"

**Cause**: The resource doesn't include region information.

**Fix**: Either include region in resource properties or set environment variables:

```bash
# Option 1: Environment variable
export AWS_REGION=us-east-1
finfocus cost projected --pulumi-json plan.json

# Option 2: Ensure Pulumi plan includes availabilityZone
```

### "end time must be after start time"

**Cause**: The `--end` flag is before the `--start` flag for actual cost queries.

**Fix**: Correct the time range:

```bash
# Wrong
finfocus cost actual --start 2025-01-15 --end 2025-01-01

# Correct
finfocus cost actual --start 2025-01-01 --end 2025-01-15
```

## Debug Logging

Enable debug logging to see validation failures with full context:

```bash
$ finfocus cost projected --debug --pulumi-json plan.json

# Log output includes:
# {"level":"warn","resource_type":"aws:ec2:Instance","error":"SKU is empty...","trace_id":"...","message":"pre-flight validation failed"}
```

## Output Format

Validation errors appear in the Notes column with a `VALIDATION:` prefix to distinguish them from plugin errors (which use `ERROR:`):

| Prefix | Source | Example |
|--------|--------|---------|
| `VALIDATION:` | Pre-flight check | Missing SKU, empty provider |
| `ERROR:` | Plugin call | Connection refused, timeout |

## Testing Validation Behavior

To verify validation is working correctly:

```bash
# Create a minimal plan with missing data
cat > test-plan.json << 'EOF'
{
  "steps": [
    {
      "op": "create",
      "urn": "urn:pulumi:dev::test::aws:ec2:Instance::my-instance",
      "newState": {
        "type": "aws:ec2:Instance",
        "inputs": {}
      }
    }
  ]
}
EOF

# Run with validation
finfocus cost projected --pulumi-json test-plan.json

# Expected: VALIDATION error for missing instanceType (SKU)
```
