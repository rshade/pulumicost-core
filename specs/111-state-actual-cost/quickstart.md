# Quickstart: State-Based Actual Cost Estimation

**Feature**: 111-state-actual-cost
**Date**: 2025-12-31

## Overview

This guide covers using the new `--pulumi-state` flag on the `cost actual`
command to estimate actual costs from Pulumi state files without requiring
cloud billing API access.

## Prerequisites

- FinFocus CLI installed
- A Pulumi stack deployed with Pulumi CLI v3.60.0+ (for timestamp support)
- Exported state file from `pulumi stack export`

## Basic Usage

### Export Pulumi State

```bash
# Export state from your deployed stack
pulumi stack export > state.json
```

### Estimate Actual Costs

```bash
# Calculate costs since deployment
finfocus cost actual --pulumi-state state.json
```

The command will:

1. Load resources from the state file
2. Auto-detect the earliest `Created` timestamp as the start date
3. Calculate runtime for each resource (`now - created`)
4. Multiply runtime by hourly rate to estimate costs

### Example Output

```text
RESOURCE               TYPE                        COST        NOTES
web-server             aws:ec2/instance:Instance   $142.35     Runtime: 30d 4h
database               aws:rds/instance:Instance   $876.24     Runtime: 30d 4h
cache                  aws:elasticache:Cluster     $234.12     Runtime: 15d 2h
─────────────────────────────────────────────────────────────────────────
TOTAL                                              $1,252.71
```

## Understanding Confidence Levels

Use `--estimate-confidence` to see how reliable each cost estimate is:

```bash
finfocus cost actual --pulumi-state state.json --estimate-confidence
```

### Confidence Levels

| Level    | Meaning                                              |
| -------- | ---------------------------------------------------- |
| `HIGH`   | Real billing data from FinOps plugin (e.g., Kubecost)|
| `MEDIUM` | Runtime estimate for Pulumi-created resources        |
| `LOW`    | Runtime estimate for imported resources              |

### Example with Confidence

```text
RESOURCE       TYPE                        COST      CONFIDENCE  NOTES
web-server     aws:ec2/instance:Instance   $142.35   MEDIUM      Runtime: 30d
imported-lb    aws:elb:LoadBalancer        $45.00    LOW         Imported resource
─────────────────────────────────────────────────────────────────────────────
```

**Important**: Imported resources show `LOW` confidence because the `Created`
timestamp reflects when the resource was imported to Pulumi, not when it was
actually created in the cloud.

## Common Scenarios

### Specify Date Range

Override auto-detection with explicit dates:

```bash
finfocus cost actual --pulumi-state state.json \
    --from 2025-01-01 --to 2025-01-31
```

### JSON Output

Get structured output for automation:

```bash
finfocus cost actual --pulumi-state state.json --output json
```

```json
{
  "resources": [
    {
      "resourceId": "web-server",
      "resourceType": "aws:ec2/instance:Instance",
      "totalCost": 142.35,
      "currency": "USD",
      "confidence": "MEDIUM",
      "notes": "Runtime: 30d 4h"
    }
  ],
  "summary": {
    "totalCost": 1252.71,
    "currency": "USD"
  }
}
```

### Daily Cost Breakdown

See costs grouped by day:

```bash
finfocus cost actual --pulumi-state state.json --group-by daily
```

### Multi-Provider Stacks

For stacks with AWS, Azure, and GCP resources:

```bash
finfocus cost actual --pulumi-state state.json --group-by provider
```

## Edge Cases

### Resources Without Timestamps

If your state was created before Pulumi v3.60.0, resources won't have
timestamps. These are skipped with a warning:

```text
WARNING: Skipping 3 resources without timestamps (pre-v3.60.0 state)
```

### Mixed Plugin/Estimate Results

If you have a FinOps plugin (like Kubecost) that provides billing data for
some resources, the command automatically uses:

- Real billing data where available (HIGH confidence)
- State-based estimates for everything else (MEDIUM/LOW confidence)

## Fallback Behavior

If `--pulumi-state` is not provided, the command behaves exactly as before:

```bash
# Original behavior - requires billing API access via plugin
finfocus cost actual --pulumi-json plan.json --from 2025-01-01
```

## Troubleshooting

### "No timestamp available" Warning

Your Pulumi state file doesn't contain `Created` timestamps. Upgrade to
Pulumi CLI v3.60.0+ and redeploy or update resources.

### "Imported resource" Notes

Resources added via `pulumi import` have `Created` set to the import time.
Cost estimates for these resources may be inaccurate. The `LOW` confidence
indicator helps identify these.

### Performance

For stacks with 100+ resources, processing should complete in under 100ms.
If you experience slowness, ensure you're using a local state file rather
than fetching from a remote backend.
