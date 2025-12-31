# Research: Cost Recommendations Command Enhancement

**Feature Branch**: `109-cost-recommendations`
**Date**: 2025-12-30
**Status**: Complete

## Overview

This document captures research findings for enhancing the `pulumicost cost recommendations` command with summary/verbose modes, interactive TUI, and improved output formatting.

## Existing Implementation Analysis

### Current Command Structure

**File**: `internal/cli/cost_recommendations.go` (374 lines)

The command already exists with:

- `--pulumi-json` (required): Path to Pulumi preview JSON
- `--adapter`: Restrict to specific plugin
- `--output`: Format selection (table, json, ndjson)
- `--filter`: Action type filtering (e.g., `action=RIGHTSIZE,TERMINATE`)

**Key Functions**:

- `executeCostRecommendations()`: Main orchestration
- `applyActionTypeFilter()`: Filter by action type
- `RenderRecommendationsOutput()`: Route to appropriate renderer
- `renderRecommendationsTable()`: Table output with tabwriter
- `renderRecommendationsJSON()`: JSON output
- `renderRecommendationsNDJSON()`: Streaming NDJSON

### Engine Types

**File**: `internal/engine/types.go`

```go
type Recommendation struct {
    ResourceID       string  `json:"resourceId,omitempty"`
    Type             string  `json:"type"`
    Description      string  `json:"description"`
    EstimatedSavings float64 `json:"estimatedSavings,omitempty"`
    Currency         string  `json:"currency,omitempty"`
}

type RecommendationsResult struct {
    Recommendations []Recommendation
    Errors          []RecommendationError
    TotalSavings    float64
    Currency        string
}
```

### TUI Package Components

**File**: `internal/tui/cost_model.go` (398 lines)

Pattern to follow for `RecommendationsViewModel`:

```go
type ViewState int
const (
    ViewStateLoading ViewState = iota
    ViewStateList
    ViewStateDetail
    ViewStateQuitting
    ViewStateError
)

type CostViewModel struct {
    state      ViewState
    allResults []engine.CostResult
    results    []engine.CostResult  // Filtered/sorted
    table      table.Model
    textInput  textinput.Model
    loading    *LoadingState
    // ...
}
```

**Keyboard Constants** (reuse these):

```go
const (
    keyEsc   = "esc"
    keyEnter = "enter"
    keyQuit  = "q"
    keyCtrlC = "ctrl+c"
    keySlash = "/"
    keyS     = "s"
)
```

**File**: `internal/tui/detect.go`

```go
type OutputMode int
const (
    OutputModePlain OutputMode = iota
    OutputModeStyled
    OutputModeInteractive
)

func DetectOutputMode(forceColor, noColor, plain bool) OutputMode
func IsTTY() bool
func TerminalWidth() int
```

**File**: `internal/tui/cost_view.go`

Reusable components:

- `RenderCostSummary()`: Styled summary box
- `RenderDetailView()`: Detail view with breakdowns
- `RenderLoading()`: Loading spinner
- `NewResultTable()`: Table creation pattern

### Proto Utilities

**File**: `internal/proto/action_types.go`

```go
func ActionTypeLabelFromString(actionType string) string
func ParseActionTypeFilter(filter string) ([]pbc.RecommendationActionType, error)
func MatchesActionType(recType string, types []pbc.RecommendationActionType) bool
func ValidActionTypes() []string
```

## Design Decisions

### D1: TUI Model Architecture

**Decision**: Create `RecommendationsViewModel` following `CostViewModel` pattern

**Rationale**:

- Proven pattern already in codebase
- Same state machine works for recommendations
- Reuses existing keyboard handling
- Maintains consistency across TUI components

**Alternatives Considered**:

- Generic ViewModel with type parameters - Rejected: Go generics would add complexity without significant benefit
- Embed CostViewModel - Rejected: Different data types require different rendering

### D2: Summary Mode Implementation

**Decision**: Show top 5 recommendations by savings with aggregate statistics

**Rationale**:

- 5 recommendations fit in standard 80x24 terminal
- Sorting by savings surfaces most impactful optimizations
- Users can use `--verbose` for full list
- Matches common UX patterns in cost tools

**Alternatives Considered**:

- Top 3 recommendations - Rejected: Too limited for useful overview
- Top 10 recommendations - Rejected: Clutters default view

### D3: Output Mode Routing

**Decision**: Machine-readable formats (`--output json|ndjson`) bypass TTY detection

**Rationale**:

- JSON/NDJSON are for automation, not human viewing
- Prevents accidental TUI in scripts that capture output
- Existing behavior is preserved

**Alternatives Considered**:

- Always check TTY even for JSON - Rejected: No benefit, adds complexity
- Separate `--interactive` flag - Rejected: Auto-detection is better UX

### D4: Filter UI in TUI

**Decision**: Use "/" key to activate filter input (vim-style)

**Rationale**:

- Consistent with existing `CostViewModel`
- Familiar to vim/less users
- Single keystroke to activate

**Alternatives Considered**:

- "f" key for filter - Rejected: Less intuitive, conflicts with potential navigation
- Always-visible filter bar - Rejected: Takes screen space

### D5: Sorting Strategy

**Decision**: Default sort by estimated savings (descending)

**Rationale**:

- Shows highest-impact recommendations first
- Matches user expectation for cost optimization tools
- "s" key cycles through sort options

**Sort Fields**:

1. Savings (descending) - Default
2. Resource ID (ascending)
3. Action Type (ascending)

## Technology Choices

### Bubble Tea Model Pattern

Using Elm-architecture pattern with:

- `Init()`: Initialize model and return initial command
- `Update(msg)`: Handle messages and return updated model + command
- `View()`: Render current state to string

### Lip Gloss Styling

Reuse existing styles from `internal/tui/styles.go`:

- `HeaderStyle`: Section headers
- `LabelStyle`: Field labels
- `ValueStyle`: Field values
- `BoxStyle`: Content boxes
- `TableHeaderStyle`: Table headers
- `TableSelectedStyle`: Selected row

### Table Component

Using `github.com/charmbracelet/bubbles/table` with:

- Fixed column widths for consistent layout
- Focused state for keyboard navigation
- Custom styles matching existing TUI

## Performance Considerations

### Large Result Sets

For 1000+ recommendations:

- Table uses virtual scrolling (built-in)
- Filtering is O(n) but instant for typical sizes
- Sorting uses Go's built-in sort (O(n log n))

### Startup Time

- Plugin queries are the bottleneck
- Loading spinner provides feedback
- No additional processing overhead for TUI mode

## Compatibility

### Terminal Support

- Full TUI: Modern terminals with ANSI support
- Styled output: Terminals with color support
- Plain text: All terminals, pipes, CI environments

### Backward Compatibility

All existing functionality preserved:

- `--pulumi-json` remains required
- `--output table|json|ndjson` works unchanged
- `--filter "action=TYPE"` works unchanged
- `--adapter` works unchanged

## Testing Strategy

### Unit Tests

1. Model state transitions
2. Filter logic
3. Sort logic
4. View rendering (snapshot tests)

### Integration Tests

1. CLI → TUI flow
2. Output format routing
3. TTY detection behavior

### Manual Testing

1. Interactive navigation
2. Keyboard responsiveness (<100ms)
3. Cross-platform terminals
