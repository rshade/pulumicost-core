# Data Model: State-Based Actual Cost Estimation

**Feature**: 111-state-actual-cost
**Date**: 2025-12-31

## Entity Overview

This feature extends existing entities rather than creating new ones. The data
model is minimal because the infrastructure already exists.

## Existing Entities (No Changes)

### StackExport

**Location**: `internal/ingest/state.go`
**Purpose**: Parsed Pulumi state from `pulumi stack export`

```go
type StackExport struct {
    Version    int                   `json:"version"`
    Deployment StackExportDeployment `json:"deployment"`
}
```

### StackExportResource

**Location**: `internal/ingest/state.go`
**Purpose**: Individual resource in Pulumi state

```go
type StackExportResource struct {
    URN      string                 `json:"urn"`
    Type     string                 `json:"type"`
    ID       string                 `json:"id,omitempty"`
    Custom   bool                   `json:"custom,omitempty"`
    External bool                   `json:"external,omitempty"`   // Key for confidence
    Provider string                 `json:"provider,omitempty"`
    Inputs   map[string]interface{} `json:"inputs,omitempty"`
    Outputs  map[string]interface{} `json:"outputs,omitempty"`
    Created  *time.Time             `json:"created,omitempty"`    // Key for runtime
    Modified *time.Time             `json:"modified,omitempty"`
}
```

### ResourceDescriptor

**Location**: `internal/engine/types.go`
**Purpose**: Cloud resource with properties for cost calculation

```go
type ResourceDescriptor struct {
    Type       string
    ID         string
    Provider   string
    Properties map[string]interface{}  // Contains pulumi:created, pulumi:external
}
```

## Modified Entity

### CostResult

**Location**: `internal/engine/types.go`
**Change**: Add `Confidence` field

```go
type CostResult struct {
    ResourceType   string                          `json:"resourceType"`
    ResourceID     string                          `json:"resourceId"`
    Adapter        string                          `json:"adapter"`
    Currency       string                          `json:"currency"`
    Monthly        float64                         `json:"monthly"`
    Hourly         float64                         `json:"hourly"`
    Notes          string                          `json:"notes"`
    Breakdown      map[string]float64              `json:"breakdown"`
    Sustainability map[string]SustainabilityMetric `json:"sustainability,omitempty"`
    Recommendations []Recommendation               `json:"recommendations,omitempty"`

    // Actual cost fields
    TotalCost  float64   `json:"totalCost,omitempty"`
    DailyCosts []float64 `json:"dailyCosts,omitempty"`
    CostPeriod string    `json:"costPeriod,omitempty"`
    StartDate  time.Time `json:"startDate,omitempty"`
    EndDate    time.Time `json:"endDate,omitempty"`
    Delta      float64   `json:"delta,omitempty"`

    // NEW: Confidence level for estimate transparency
    // Values: "HIGH", "MEDIUM", "LOW", or ""
    Confidence string    `json:"confidence,omitempty"`
}
```

**Validation Rules**:

- `Confidence` is optional (empty string when not applicable)
- When populated, MUST be one of: "HIGH", "MEDIUM", "LOW"
- JSON output includes field only when `--estimate-confidence` flag is used

## New Value Types

### ConfidenceLevel

**Location**: `internal/engine/confidence.go`
**Purpose**: Type-safe confidence level constants

```go
type ConfidenceLevel string

const (
    ConfidenceHigh   ConfidenceLevel = "HIGH"
    ConfidenceMedium ConfidenceLevel = "MEDIUM"
    ConfidenceLow    ConfidenceLevel = "LOW"
    ConfidenceNone   ConfidenceLevel = ""
)

func (c ConfidenceLevel) String() string {
    return string(c)
}

func (c ConfidenceLevel) IsValid() bool {
    switch c {
    case ConfidenceHigh, ConfidenceMedium, ConfidenceLow, ConfidenceNone:
        return true
    default:
        return false
    }
}
```

### StateCostInput

**Location**: `internal/engine/state_cost.go`
**Purpose**: Input for state-based cost calculation

```go
type StateCostInput struct {
    Resource    ResourceDescriptor
    HourlyRate  float64
    CreatedAt   time.Time
    IsExternal  bool
}
```

### StateCostResult

**Location**: `internal/engine/state_cost.go`
**Purpose**: Result of state-based cost calculation

```go
type StateCostResult struct {
    TotalCost     float64
    RuntimeHours  float64
    Confidence    ConfidenceLevel
    Notes         string
}
```

## State Transitions

### Confidence Level Assignment

```text
┌─────────────────────────────────────────────────────────────────┐
│                     Cost Data Source                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   Plugin GetActualCost returns data?                            │
│           │                                                      │
│           ├── YES ──► Confidence = HIGH                         │
│           │           Notes = "From billing API"                │
│           │                                                      │
│           └── NO ──► Use State-Based Estimation                 │
│                           │                                      │
│                           ├── Has Created timestamp?             │
│                           │       │                              │
│                           │       ├── NO ──► Skip resource       │
│                           │       │          Notes = "No timestamp"
│                           │       │                              │
│                           │       └── YES ──► Check External     │
│                           │               │                      │
│                           │               ├── External=true      │
│                           │               │   Confidence = LOW   │
│                           │               │   Notes = "Imported" │
│                           │               │                      │
│                           │               └── External=false     │
│                           │                   Confidence = MEDIUM│
│                           │                   Notes = "Runtime"  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Relationships

```text
StackExport
    │
    └── Deployment.Resources[] ──── StackExportResource
                                         │
                                         │ MapStateResource()
                                         ▼
                                    ResourceDescriptor
                                         │
                                         │ GetProjectedCost()
                                         ▼
                                    CostResult (Hourly rate)
                                         │
                                         │ CalculateStateCost()
                                         ▼
                                    CostResult (with TotalCost, Confidence)
```

## Data Volume Assumptions

Based on spec SC-004 and project patterns:

| Metric                    | Expected Range | Notes                         |
| ------------------------- | -------------- | ----------------------------- |
| Resources per stack       | 1-1000         | Most stacks have <100         |
| State file size           | 10KB-10MB      | Depends on resource count     |
| Properties per resource   | 5-50           | Pulumi injects metadata       |
| Cost calculations/request | 1-1000         | Matches resource count        |
| Processing time           | <100ms         | For 100 resources (SC-004)    |

## Index/Query Patterns

Not applicable - this is a CLI tool without persistent storage. All data is
processed in-memory from JSON files.

## Backwards Compatibility

- `Confidence` field has `omitempty` tag - absent when empty
- Existing JSON consumers will ignore the new field
- CLI output unchanged unless `--estimate-confidence` flag used
- No breaking changes to existing behavior
