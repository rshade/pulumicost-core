# Research: Pre-Flight Request Validation

**Feature**: 107-preflight-validation
**Date**: 2025-12-29

## Dependencies Resolution

### pluginsdk Validation Functions

**Decision**: Use `pluginsdk.ValidateProjectedCostRequest()` and `pluginsdk.ValidateActualCostRequest()` from finfocus-spec v0.4.11+.

**Rationale**:

- Functions are already designed for dual-use (Core pre-flight + Plugin defense-in-depth)
- Zero-allocation on happy path (<100ns target)
- Actionable error messages guide users to fix issues
- Already integrated in current go.mod (v0.4.11)

**Alternatives considered**:

- Custom validation in Core: Rejected - duplicates logic, creates drift risk
- No validation: Rejected - poor UX with cryptic plugin errors

### Validation Order (from pluginsdk documentation)

`ValidateProjectedCostRequest` checks (fail-fast):

1. Request nil check
2. Resource nil check
3. Provider empty check
4. ResourceType empty check
5. SKU empty check (with mapping helper guidance)
6. Region empty check (with mapping helper guidance)
7. Global utilization range check (if non-zero)
8. Resource-level utilization range check (if provided)

`ValidateActualCostRequest` checks (fail-fast):

1. Request nil check
2. ResourceId empty check
3. StartTime nil check
4. EndTime nil check
5. TimeRange validation (EndTime must be after StartTime)

### Logging Integration

**Decision**: Use `logging.FromContext(ctx)` for structured logging with trace_id.

**Rationale**:

- Already established pattern in codebase
- Trace ID correlation for debugging
- WARN level appropriate for user-fixable issues

**Alternatives considered**:

- Direct zerolog: Rejected - loses context/trace_id
- DEBUG level: Rejected - validation failures are noteworthy, not verbose debug

## Technical Research

### Import Path

```go
// Current import (for mapping)
import "github.com/rshade/finfocus-spec/sdk/go/pluginsdk/mapping"

// New import needed (for validation)
import "github.com/rshade/finfocus-spec/sdk/go/pluginsdk"
```

Both imports can coexist since `mapping` is a subpackage.

### Error Message Format

pluginsdk validation errors are structured as actionable messages:

- `"SKU is empty: use mapping.ExtractAWSSKU() or mapping.ExtractSKU() to extract from resource properties"`
- `"provider is empty"`
- `"resourceType is empty"`
- `"region is empty: use mapping.ExtractAWSRegion() or mapping.ExtractRegion() to extract from resource properties"`

These messages directly guide users to fix issues.

## No Outstanding Research Items

All dependencies are resolved. No NEEDS CLARIFICATION items remain.
