# Data Model: Error Aggregation in Proto Adapter

**Date**: 2025-11-24
**Feature**: 003-proto-error-aggregation

## Entities

### ErrorDetail

Captures information about a single failed resource cost calculation.

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| ResourceType | string | Resource type identifier (e.g., "aws:ec2:Instance") | Required, non-empty |
| ResourceID | string | Resource instance identifier | Required, non-empty |
| PluginName | string | Name of the plugin that failed | Required, defaults to "unknown" |
| Error | error | The underlying error from plugin call | Required |
| Timestamp | time.Time | When the error occurred | Required, set automatically |

**Location**: `internal/proto/adapter.go`

---

### CostResultWithErrors

Container type that holds both successful cost results and error details.

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| Results | []engine.CostResult | Successfully calculated cost results | May be empty |
| Errors | []ErrorDetail | Failed resource calculations | May be empty |

**Methods**:

| Method | Signature | Description |
|--------|-----------|-------------|
| HasErrors | `() bool` | Returns true if len(Errors) > 0 |
| ErrorSummary | `() string` | Returns human-readable error summary, truncates after 5 |

**Location**: `internal/proto/adapter.go`

---

## Relationships

```text
┌─────────────────────┐
│ CostResultWithErrors│
├─────────────────────┤
│ Results             │───────►[ ]engine.CostResult
│ Errors              │───────►[ ]ErrorDetail
└─────────────────────┘
         │
         │ returned by
         ▼
┌─────────────────────┐
│ Adapter             │
├─────────────────────┤
│ GetProjectedCost()  │
│ GetActualCost()     │
└─────────────────────┘
         │
         │ called by
         ▼
┌─────────────────────┐
│ Engine              │
├─────────────────────┤
│ GetProjectedCost()  │───────► *CostResultWithErrors
│ GetActualCost()     │───────► *CostResultWithErrors
└─────────────────────┘
         │
         │ displayed by
         ▼
┌─────────────────────┐
│ CLI Commands        │
├─────────────────────┤
│ cost projected      │
│ cost actual         │
└─────────────────────┘
```

---

## State Transitions

The `CostResultWithErrors` type is immutable after creation. No state transitions.

**Result Population States**:

1. **Empty**: Both Results and Errors are empty (no resources processed)
2. **Success Only**: Results populated, Errors empty (all calculations succeeded)
3. **Partial Success**: Both Results and Errors populated (some failed)
4. **All Failed**: Results contains placeholders with error notes, Errors fully populated

---

## Placeholder CostResult for Errors

When a resource calculation fails, a placeholder is added to Results:

| Field | Value |
|-------|-------|
| ResourceType | Original resource type |
| ResourceID | Original resource ID |
| Adapter | Plugin name |
| Currency | "USD" (default) |
| Monthly/TotalCost | 0 |
| Hourly | 0 |
| Notes | "ERROR: {error message}" |
| StartDate/EndDate | Original date range (for actual cost) |

This ensures:
- Result count matches resource count
- CLI table shows all resources
- Users see which specific resources failed

---

## Existing Types (Modified)

### engine.CostResult (unchanged structure)

Already contains `Notes` field which will hold error prefix for failed resources.

### Engine Methods (signature change)

**Before**:
```go
func (e *Engine) GetProjectedCost(ctx, resources) ([]CostResult, error)
func (e *Engine) GetActualCost(ctx, resources, from, to) ([]CostResult, error)
```

**After**:
```go
func (e *Engine) GetProjectedCost(ctx, resources) (*CostResultWithErrors, error)
func (e *Engine) GetActualCost(ctx, resources, from, to) (*CostResultWithErrors, error)
```

---

## Validation Rules

1. **ErrorDetail.ResourceType**: Must be non-empty string
2. **ErrorDetail.ResourceID**: Must be non-empty string
3. **ErrorDetail.Error**: Must be non-nil error
4. **ErrorDetail.Timestamp**: Set to time.Now() at creation
5. **CostResultWithErrors**: May have both empty slices (valid empty result)
6. **ErrorSummary**: Truncates after 5 errors with count of remaining
