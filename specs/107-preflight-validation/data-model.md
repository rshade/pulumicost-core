# Data Model: Pre-Flight Request Validation

**Feature**: 107-preflight-validation
**Date**: 2025-12-29

## Existing Types (No Changes Required)

This feature uses existing types - no new data model entities are introduced.

### CostResult (internal/proto/adapter.go:210-219)

```go
type CostResult struct {
    Currency       string
    MonthlyCost    float64
    HourlyCost     float64
    Notes          string                          // ← Validation errors go here
    CostBreakdown  map[string]float64
    Sustainability map[string]SustainabilityMetric
}
```

**Validation Error Pattern**: `Notes: "VALIDATION: <error message>"`

### ActualCostResult (internal/proto/adapter.go:240-247)

```go
type ActualCostResult struct {
    Currency       string
    TotalCost      float64
    CostBreakdown  map[string]float64
    Sustainability map[string]SustainabilityMetric
}
```

**Note**: ActualCostResult doesn't have a Notes field. Validation errors for actual cost are tracked through the existing ErrorDetail mechanism.

### ErrorDetail (internal/proto/adapter.go:21-28)

```go
type ErrorDetail struct {
    ResourceType string
    ResourceID   string
    PluginName   string
    Error        error                            // ← Wrapped validation error
    Timestamp    time.Time
}
```

**Validation Error Pattern**: `Error: fmt.Errorf("pre-flight validation failed: %w", validationErr)`

### CostResultWithErrors (internal/proto/adapter.go:30-34)

```go
type CostResultWithErrors struct {
    Results []*CostResult
    Errors  []ErrorDetail
}
```

**Behavior**: Validation errors are captured in `Errors` slice alongside plugin errors. The `HasErrors()` and `ErrorSummary()` methods work unchanged.

## Protobuf Types (from pluginsdk)

### pbc.GetProjectedCostRequest

```go
// Validated by pluginsdk.ValidateProjectedCostRequest()
type GetProjectedCostRequest struct {
    Resource *ResourceDescriptor
}

type ResourceDescriptor struct {
    Id           string
    Provider     string              // REQUIRED: must be non-empty
    ResourceType string              // REQUIRED: must be non-empty
    Sku          string              // REQUIRED: must be non-empty
    Region       string              // REQUIRED: must be non-empty
    Tags         map[string]string
    Utilization  float64             // OPTIONAL: if set, must be 0.0-1.0
}
```

### pbc.GetActualCostRequest

```go
// Validated by pluginsdk.ValidateActualCostRequest()
type GetActualCostRequest struct {
    ResourceId string               // REQUIRED: must be non-empty
    Start      *timestamppb.Timestamp // REQUIRED: must be non-nil
    End        *timestamppb.Timestamp // REQUIRED: must be non-nil, must be after Start
    Tags       map[string]string
}
```

## Validation Rules Summary

| Field | Rule | Error Message |
|-------|------|---------------|
| Provider | Non-empty | "provider is empty" |
| ResourceType | Non-empty | "resourceType is empty" |
| SKU | Non-empty | "SKU is empty: use mapping.ExtractAWSSKU()..." |
| Region | Non-empty | "region is empty: use mapping.ExtractAWSRegion()..." |
| Utilization | 0.0 ≤ x ≤ 1.0 | "utilization must be between 0.0 and 1.0" |
| ResourceId | Non-empty | "resourceId is empty" |
| Start | Non-nil | "start time is required" |
| End | Non-nil, > Start | "end time is required" / "end time must be after start time" |

## State Transitions

N/A - Validation is stateless. Each request is validated independently.

## Relationships

```text
GetProjectedCostRequest (internal) → pbc.GetProjectedCostRequest (proto) → Validation → gRPC call
                                      ↑
                                      └── resolveSKUAndRegion() extracts SKU/Region from Properties
```

The validation occurs after the internal request is converted to a protobuf request but before the gRPC call is made.
