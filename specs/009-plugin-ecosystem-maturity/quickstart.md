# Quickstart: Plugin Conformance Testing

**Feature**: 009-plugin-ecosystem-maturity
**Audience**: Plugin Developers

## Overview

This guide shows you how to validate your plugin against the PulumiCost
conformance test suite to ensure it correctly implements the gRPC protocol.

---

## Prerequisites

- PulumiCost CLI installed (`pulumicost --version` works)
- Your plugin binary compiled for the target platform
- Plugin implements the CostSourceService gRPC interface

---

## Quick Validation (5 minutes)

### Step 1: Locate Your Plugin

```bash
# Your plugin binary should be an executable file
ls -la ./my-plugin-binary
# -rwxr-xr-x  1 user  staff  12345678 Dec  2 10:00 my-plugin-binary
```

### Step 2: Run Basic Conformance Check

```bash
pulumicost plugin conformance ./my-plugin-binary
```

**Expected output** (all tests pass):

```text
CONFORMANCE TEST RESULTS
========================
Plugin: my-plugin v1.0.0 (protocol v1.0)
Mode:   TCP

TESTS
-----
✓ Name_ReturnsPluginIdentifier              [  50ms]
✓ Name_ReturnsProtocolVersion               [  45ms]
✓ GetProjectedCost_ValidResource            [ 120ms]
...

SUMMARY
-------
Total: 20 | Passed: 20 | Failed: 0 | Skipped: 0 | Duration: 3.2s
```

---

## Understanding Test Results

### Pass (✓)

Test passed. Your plugin correctly implements this protocol requirement.

### Fail (✗)

Test failed. The error message indicates what went wrong.

```text
✗ GetProjectedCost_InvalidResource          [ 110ms]
  Error: expected NotFound, got InvalidArgument
```

**Fix**: Update your plugin to return `codes.NotFound` for invalid resources.

### Skip (⊘)

Test skipped because a precondition wasn't met.

```text
⊘ GetActualCost_RequiresCredentials         [  --  ] (skipped)
```

**Common reasons**: Missing credentials, version mismatch, optional feature.

### Error (!)

Infrastructure error (plugin crash, timeout, connection lost).

```text
! GetProjectedCost_LargeBatch               [  --  ] (error)
  Error: plugin process exited unexpectedly (signal: killed)
```

**Fix**: Check plugin for memory issues or increase timeout.

---

## Common Fixes

### Protocol Version Mismatch

```text
Error: protocol version mismatch: got 0.9, want 1.0
```

**Fix**: Update your plugin to implement protocol version 1.0. See the
pulumicost-spec repository for the latest protocol definitions.

### Missing Name Response

```text
✗ Name_ReturnsPluginIdentifier
  Error: Name() returned empty string
```

**Fix**: Ensure your plugin's `Name()` RPC returns a non-empty identifier:

```go
func (s *Server) Name(ctx context.Context, req *pb.Empty) (*pb.NameResponse, error) {
    return &pb.NameResponse{
        Name:            "my-plugin",
        Version:         "1.0.0",
        ProtocolVersion: "1.0",
    }, nil
}
```

### Timeout Exceeded

```text
! GetProjectedCost_ValidResource            [10.0s] (error)
  Error: context deadline exceeded
```

**Fix**: Optimize your plugin or check for network/API delays. Default
timeout is 10 seconds per test.

### Wrong Error Code

```text
✗ GetProjectedCost_InvalidResource
  Error: expected NotFound, got OK with empty response
```

**Fix**: Return appropriate gRPC error codes:

```go
if resource == nil || resource.Type == "" {
    return nil, status.Error(codes.InvalidArgument, "resource type required")
}
if !s.SupportsResource(resource.Type) {
    return nil, status.Error(codes.NotFound, "unsupported resource type")
}
```

---

## Verbose Output for Debugging

When tests fail, use verbose mode to see request/response details:

```bash
pulumicost plugin conformance --verbosity verbose ./my-plugin-binary
```

**Output includes**:

```text
✗ GetProjectedCost_InvalidResource          [ 110ms]
  Request:
    resource_type: "unknown:invalid:resource"
    properties: {}
  Response:
    status: OK
    cost_per_month: 0
  Expected: gRPC status NotFound
  Actual:   gRPC status OK with empty response
```

---

## CI/CD Integration

### GitHub Actions

```yaml
- name: Run Conformance Tests
  run: |
    pulumicost plugin conformance \
      --output junit \
      --output-file conformance-report.xml \
      ./bin/my-plugin

- name: Upload Test Results
  uses: actions/upload-artifact@v4
  with:
    name: conformance-results
    path: conformance-report.xml
```

### JSON Output for Custom Processing

```bash
pulumicost plugin conformance --output json ./my-plugin > results.json
jq '.summary.failed' results.json  # Check failure count
```

---

## Test Categories

Filter tests by category to focus on specific areas:

```bash
# Protocol compliance only
pulumicost plugin conformance --category protocol ./my-plugin

# Error handling tests
pulumicost plugin conformance --category error ./my-plugin

# All categories (default)
pulumicost plugin conformance ./my-plugin
```

**Categories**:

| Category    | Tests                                          |
|-------------|------------------------------------------------|
| protocol    | Name, GetProjectedCost, GetActualCost basics   |
| error       | Error codes, invalid inputs, edge cases        |
| performance | Timeout behavior, batch handling               |
| context     | Context cancellation, deadline propagation     |

---

## Next Steps

1. **Fix all failing tests** before releasing your plugin
2. **Run conformance in CI** to catch regressions
3. **Test with verbose mode** when debugging issues
4. **Check pulumicost-spec** for protocol updates

For E2E testing with real cloud providers, see the E2E Testing Guide
(requires test account setup).
