# Quickstart: Error Aggregation in Proto Adapter

**Date**: 2025-11-24
**Feature**: 003-proto-error-aggregation

## Overview

This feature adds error aggregation to the proto adapter, replacing silent failures with comprehensive error tracking.

---

## Implementation Steps

### Step 1: Add zerolog Dependency

```bash
go get github.com/rs/zerolog
```

### Step 2: Create New Types

Add to `internal/proto/adapter.go`:

```go
type ErrorDetail struct {
    ResourceType string
    ResourceID   string
    PluginName   string
    Error        error
    Timestamp    time.Time
}

type CostResultWithErrors struct {
    Results []engine.CostResult
    Errors  []ErrorDetail
}

func (c *CostResultWithErrors) HasErrors() bool {
    return len(c.Errors) > 0
}

func (c *CostResultWithErrors) ErrorSummary() string {
    // Implementation with truncation after 5 errors
}
```

### Step 3: Update Adapter Methods

Modify `GetProjectedCost` and `GetActualCost` to:
1. Return `*CostResultWithErrors`
2. Track errors in `Errors` slice
3. Add placeholder results for failures

### Step 4: Update Engine Methods

Modify engine to:
1. Return `*CostResultWithErrors`
2. Aggregate errors from all plugin calls
3. Log errors with zerolog

### Step 5: Update CLI Commands

Modify CLI to:
1. Handle new return type
2. Display errors inline in Notes column
3. Show error summary after table

---

## Testing Commands

```bash
# Run all tests
make test

# Run specific adapter tests
go test -v ./internal/proto/...

# Run with race detection
go test -race ./...

# Check coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

---

## Manual Testing

```bash
# Build
make build

# Test with example plan
./bin/pulumicost cost projected --pulumi-json examples/plans/aws-simple-plan.json

# Expected output with errors:
# COST SUMMARY
# ============
# Resource Type      | ID         | Monthly | Notes
# aws:ec2:Instance   | i-123      | 50.00   |
# aws:rds:Instance   | db-456     | 0.00    | ERROR: connection refused
#
# ⚠️ 1 resource(s) failed:
#   - aws:rds:Instance (db-456): plugin call failed: connection refused
```

---

## Key Files to Modify

1. `internal/proto/adapter.go` - New types + updated methods
2. `internal/proto/adapter_test.go` - Error aggregation tests
3. `internal/engine/engine.go` - Return type changes
4. `internal/engine/engine_test.go` - Updated tests
5. `internal/cli/cost_projected.go` - Display errors
6. `internal/cli/cost_actual.go` - Display errors

---

## Acceptance Criteria Checklist

- [ ] ErrorDetail and CostResultWithErrors types created
- [ ] GetProjectedCost tracks errors
- [ ] GetActualCost tracks errors
- [ ] Placeholder results for failed resources
- [ ] ErrorSummary() truncates after 5 errors
- [ ] Engine logs errors with zerolog
- [ ] CLI displays inline errors + summary
- [ ] Tests achieve 80%+ coverage
- [ ] All tests pass: `make test`
- [ ] Lint passes: `make lint`

---

## Common Issues

### Import Cycle

If you get import cycles, keep `ErrorDetail` and `CostResultWithErrors` in the proto package (not engine).

### zerolog Not Found

Run `go mod tidy` after adding the dependency.

### Test Mocking

Use interfaces for pluginhost.Client to enable mocking in tests.
