# Data Model: Sustainability Metrics

**Context**: Defines the internal Go structures to support sustainability metrics.

## Entities

### ImpactMetric

Represents a single sustainability measurement (e.g., Carbon Footprint).

```go
type ImpactMetric struct {
    // Kind distinguishes the type of metric.
    // Maps to pulumicost-spec MetricKind.
    Kind string `json:"kind"` 

    // Value is the numeric magnitude.
    Value float64 `json:"value"`

    // Unit is the string representation of the unit (e.g., "gCO2e", "kWh").
    Unit string `json:"unit"`
}
```

### CostResult (Update)

The primary result structure for a resource's cost.

```go
type CostResult struct {
    // ... existing fields (Resource, CostComponents, etc.) ...

    // ImpactMetrics holds the list of sustainability metrics for this resource.
    ImpactMetrics []*ImpactMetric `json:"impactMetrics,omitempty"`
}
```

### CostComponent (Update - Optional)

If metrics are broken down by component (e.g., "Compute Carbon" vs "Storage Carbon"), we might add this to `CostComponent` too. **Decision**: For now, keep it at the Resource level (`CostResult`) as per spec.

## Enums

### MetricKind

String constants used in `ImpactMetric.Kind`.

```go
const (
    MetricKindCarbonFootprint    = "METRIC_KIND_CARBON_FOOTPRINT"
    MetricKindEnergyConsumption  = "METRIC_KIND_ENERGY_CONSUMPTION"
    MetricKindWaterUsage         = "METRIC_KIND_WATER_USAGE"
)
```

## JSON Schema Structure

The output JSON will reflect these structures:

```json
{
  "resource": "...",
  "cost": 10.5,
  "impactMetrics": [
    {
      "kind": "METRIC_KIND_CARBON_FOOTPRINT",
      "value": 1200.5,
      "unit": "gCO2e"
    }
  ]
}
```
