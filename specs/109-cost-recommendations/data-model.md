# Data Model: Cost Recommendations Command Enhancement

**Feature Branch**: `109-cost-recommendations`
**Date**: 2025-12-30
**Status**: Complete

## Entity Overview

This feature enhances the existing recommendations command without introducing new persistent entities. All data models are in-memory view models for TUI rendering.

## Existing Entities (Read-Only)

### engine.Recommendation

**Location**: `internal/engine/types.go`

```go
type Recommendation struct {
    ResourceID       string  `json:"resourceId,omitempty"`
    Type             string  `json:"type"`
    Description      string  `json:"description"`
    EstimatedSavings float64 `json:"estimatedSavings,omitempty"`
    Currency         string  `json:"currency,omitempty"`
}
```

**Fields**:

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| ResourceID | string | Resource identifier | Optional, max 256 chars |
| Type | string | Action type (e.g., "RIGHTSIZE") | Required, must be valid action type |
| Description | string | Human-readable recommendation | Required, non-empty |
| EstimatedSavings | float64 | Monthly savings estimate | >= 0 |
| Currency | string | ISO 4217 currency code | Default "USD" |

### engine.RecommendationsResult

**Location**: `internal/engine/types.go`

```go
type RecommendationsResult struct {
    Recommendations []Recommendation
    Errors          []RecommendationError
    TotalSavings    float64
    Currency        string
}
```

**Methods**:

- `HasErrors() bool`: Returns true if any errors occurred
- `ErrorSummary() string`: Human-readable error summary

## New View Models

### RecommendationsViewModel

**Location**: `internal/tui/recommendations_model.go` (NEW)

```go
type RecommendationsViewModel struct {
    // View state
    state            ViewState
    allRecommendations []engine.Recommendation  // Source of truth
    recommendations  []engine.Recommendation    // Filtered/sorted

    // Interactive components
    table     table.Model
    textInput textinput.Model
    selected  int

    // Display configuration
    width      int
    height     int
    sortBy     RecommendationSortField
    showFilter bool
    verbose    bool

    // Loading state
    loading  *LoadingState
    fetchCmd tea.Cmd

    // Aggregated data
    summary  *RecommendationsSummary

    // Error handling
    errors []engine.RecommendationError
    err    error
}
```

**State Machine**:

```text
┌─────────────┐
│  Loading    │
└──────┬──────┘
       │ loadingCompleteMsg
       ▼
┌─────────────┐   Enter    ┌─────────────┐
│    List     │───────────▶│   Detail    │
└──────┬──────┘            └──────┬──────┘
       │ q/ctrl+c                 │ Esc
       ▼                          │
┌─────────────┐                   │
│  Quitting   │◀──────────────────┘
└─────────────┘                   (q/ctrl+c)

Error state accessible from any state on fatal error
```

**Validation Rules**:

| Field | Rule |
|-------|------|
| state | Must be valid ViewState enum value |
| selected | 0 <= selected < len(recommendations) when in Detail state |
| sortBy | Must be valid RecommendationSortField enum value |

### RecommendationsSummary

**Location**: `internal/tui/recommendations_model.go` (NEW)

```go
type RecommendationsSummary struct {
    TotalCount       int
    TotalSavings     float64
    Currency         string
    CountByAction    map[string]int
    SavingsByAction  map[string]float64
    TopRecommendations []engine.Recommendation  // Top 5 by savings
}
```

**Purpose**: Aggregated statistics for summary display mode.

**Computation**:

```go
func NewRecommendationsSummary(recs []engine.Recommendation) *RecommendationsSummary {
    summary := &RecommendationsSummary{
        TotalCount:      len(recs),
        CountByAction:   make(map[string]int),
        SavingsByAction: make(map[string]float64),
    }

    for _, rec := range recs {
        summary.TotalSavings += rec.EstimatedSavings
        summary.CountByAction[rec.Type]++
        summary.SavingsByAction[rec.Type] += rec.EstimatedSavings
        if summary.Currency == "" {
            summary.Currency = rec.Currency
        }
    }

    // Sort by savings descending, take top 5
    sorted := sortBySavings(recs)
    if len(sorted) > 5 {
        summary.TopRecommendations = sorted[:5]
    } else {
        summary.TopRecommendations = sorted
    }

    return summary
}
```

### RecommendationSortField

**Location**: `internal/tui/recommendations_model.go` (NEW)

```go
type RecommendationSortField int

const (
    SortBySavings RecommendationSortField = iota  // Default
    SortByResourceID
    SortByActionType
)

const numRecommendationSortFields = 3
```

### RecommendationRow

**Location**: `internal/tui/recommendations_view.go` (NEW)

```go
type RecommendationRow struct {
    ResourceID   string  // Truncated to 30 chars
    ActionType   string  // Human-readable label
    Description  string  // Truncated to 40 chars
    Savings      string  // Formatted: "$123.45 USD"
    HasSavings   bool    // For styling (green if savings > 0)
}
```

**Purpose**: Display-ready row for table rendering.

## JSON Output Structures

### Enhanced JSON Output

**Location**: `internal/cli/cost_recommendations.go` (MODIFY)

```go
type recommendationsJSONOutput struct {
    Summary         *jsonSummary        `json:"summary"`
    Recommendations []recommendationJSON `json:"recommendations"`
    TotalSavings    float64             `json:"total_savings"`
    Currency        string              `json:"currency"`
    Errors          []engine.RecommendationError `json:"errors,omitempty"`
}

type jsonSummary struct {
    TotalCount      int                `json:"total_count"`
    TotalSavings    float64            `json:"total_savings"`
    Currency        string             `json:"currency"`
    CountByAction   map[string]int     `json:"count_by_action_type"`
    SavingsByAction map[string]float64 `json:"savings_by_action_type"`
}
```

## Relationships

```text
┌─────────────────────────┐
│ RecommendationsViewModel│
│─────────────────────────│
│ state                   │
│ recommendations[]───────┼──────┐
│ summary ────────────────┼──┐   │
│ table                   │  │   │
│ errors[]                │  │   │
└─────────────────────────┘  │   │
                             │   │
         ┌───────────────────┘   │
         ▼                       │
┌─────────────────────────┐      │
│ RecommendationsSummary  │      │
│─────────────────────────│      │
│ TotalCount              │      │
│ TotalSavings            │      │
│ CountByAction{}         │      │
│ SavingsByAction{}       │      │
│ TopRecommendations[]────┼──────┤
└─────────────────────────┘      │
                                 │
                                 ▼
                    ┌─────────────────────────┐
                    │ engine.Recommendation   │
                    │─────────────────────────│
                    │ ResourceID              │
                    │ Type                    │
                    │ Description             │
                    │ EstimatedSavings        │
                    │ Currency                │
                    └─────────────────────────┘
```

## Action Type Enumeration

Valid action types (from `internal/proto/action_types.go`):

| Type | Description |
|------|-------------|
| RIGHTSIZE | Resize to more appropriate instance size |
| TERMINATE | Shut down unused resources |
| PURCHASE_COMMITMENT | Buy reserved instances or savings plans |
| ADJUST_REQUESTS | Adjust Kubernetes resource requests |
| MODIFY | General modification recommendation |
| DELETE_UNUSED | Delete orphaned/unused resources |
| MIGRATE | Move to different service/region |
| CONSOLIDATE | Combine multiple resources |
| SCHEDULE | Add scheduling (start/stop times) |
| REFACTOR | Architectural refactoring |
| OTHER | Uncategorized recommendation |

## Message Types (Bubble Tea)

### loadingCompleteMsg

```go
type loadingCompleteMsg struct {
    recommendations []engine.Recommendation
    errors          []engine.RecommendationError
    err             error
}
```

**Triggers**: Sent when async data fetch completes.

**Handling**:

- On success: Transition to List state, populate model
- On error: Transition to Error state, display error
