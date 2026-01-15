# Research: Extended RecommendationActionType Enum Support

**Feature**: 108-action-type-enum
**Date**: 2025-12-29

## Research Questions

### RQ-1: Upstream Enum Values Availability

**Question**: Are all 11 action types available in the current finfocus-spec
dependency?

**Finding**: Yes. finfocus-spec v0.4.11 (current dependency) includes all 11
`RecommendationActionType` values via the `AllRecommendationActionTypes()`
helper function in `sdk/go/proto/finfocus/v1/action_types.go`.

**Evidence**:

```go
// From finfocus-spec v0.4.11 action_types.go
var allRecommendationActionTypes = []RecommendationActionType{
    RecommendationActionType_RECOMMENDATION_ACTION_TYPE_UNSPECIFIED,
    RecommendationActionType_RECOMMENDATION_ACTION_TYPE_RIGHTSIZE,
    RecommendationActionType_RECOMMENDATION_ACTION_TYPE_TERMINATE,
    RecommendationActionType_RECOMMENDATION_ACTION_TYPE_PURCHASE_COMMITMENT,
    RecommendationActionType_RECOMMENDATION_ACTION_TYPE_ADJUST_REQUESTS,
    RecommendationActionType_RECOMMENDATION_ACTION_TYPE_MODIFY,
    RecommendationActionType_RECOMMENDATION_ACTION_TYPE_DELETE_UNUSED,
    RecommendationActionType_RECOMMENDATION_ACTION_TYPE_MIGRATE,
    RecommendationActionType_RECOMMENDATION_ACTION_TYPE_CONSOLIDATE,
    RecommendationActionType_RECOMMENDATION_ACTION_TYPE_SCHEDULE,
    RecommendationActionType_RECOMMENDATION_ACTION_TYPE_REFACTOR,
    RecommendationActionType_RECOMMENDATION_ACTION_TYPE_OTHER,
}
```

**Decision**: Use the existing `pbc.AllRecommendationActionTypes()` and
`pbc.IsValidRecommendationActionType()` functions from pluginsdk for validation.

### RQ-2: Current Recommendation Struct in Core

**Question**: How does finfocus-core currently represent recommendations?

**Finding**: The `internal/proto/adapter.go` defines a `Recommendation` struct
with an `ActionType string` field. This is a string representation, not a typed
enum.

**Evidence**:

```go
// From internal/proto/adapter.go:374-398
type Recommendation struct {
    ID          string
    Category    string
    ActionType  string  // String, e.g., "RIGHTSIZE", "TERMINATE"
    Description string
    ResourceID  string
    Source      string
    Impact      *RecommendationImpact
    Metadata    map[string]string
}
```

**Decision**: Create an action type mapping utility that:

1. Converts proto enum values to human-readable labels for TUI
2. Validates string values in filter expressions against valid enum names
3. Provides case-insensitive matching for filter parsing

### RQ-3: Filter Expression Pattern

**Question**: What filter parsing patterns exist in the codebase?

**Finding**: The `cost actual` command uses `--filter` flag with `key=value`
syntax. Filter values are parsed with `strings.Split` and validated against
known types.

**Evidence**:

```go
// From internal/cli/cost_actual.go:17
const (
    filterKeyValueParts = 2   // For "key=value" pairs
)
```

**Decision**: Follow the same pattern for action type filtering:

- Filter syntax: `action=MIGRATE,RIGHTSIZE` (comma-separated)
- Case-insensitive matching: `action=migrate` matches `MIGRATE`
- Validation: reject unknown types with error listing valid options

### RQ-4: TUI Label Display Pattern

**Question**: How does the TUI currently display typed values?

**Finding**: The `internal/tui/` package uses Lip Gloss for styling. Labels are
typically title-cased for display.

**Evidence**: TUI components use `text/transform` or direct title-casing for
human-readable labels.

**Decision**: Create a mapping function:

<!-- markdownlint-disable MD060 -->

| Proto Enum Value | Display Label |
| ---------------- | ------------- |
| `RECOMMENDATION_ACTION_TYPE_RIGHTSIZE` | Rightsize |
| `RECOMMENDATION_ACTION_TYPE_TERMINATE` | Terminate |
| `RECOMMENDATION_ACTION_TYPE_PURCHASE_COMMITMENT` | Purchase Commitment |
| `RECOMMENDATION_ACTION_TYPE_ADJUST_REQUESTS` | Adjust Requests |
| `RECOMMENDATION_ACTION_TYPE_MODIFY` | Modify |
| `RECOMMENDATION_ACTION_TYPE_DELETE_UNUSED` | Delete Unused |
| `RECOMMENDATION_ACTION_TYPE_MIGRATE` | Migrate |
| `RECOMMENDATION_ACTION_TYPE_CONSOLIDATE` | Consolidate |
| `RECOMMENDATION_ACTION_TYPE_SCHEDULE` | Schedule |
| `RECOMMENDATION_ACTION_TYPE_REFACTOR` | Refactor |
| `RECOMMENDATION_ACTION_TYPE_OTHER` | Other |

<!-- markdownlint-enable MD060 -->

### RQ-5: Unknown Enum Value Handling

**Question**: How should unknown/future enum values be handled?

**Finding**: Proto3 enums are forward-compatible. Unrecognized numeric values
deserialize as the raw integer. The `pbc.IsValidRecommendationActionType()`
function returns `false` for unknown values.

**Decision**:

1. For display: Show "Unknown (N)" where N is the raw integer value
2. For filtering: Only known string names are valid in filter expressions
3. Log warning when encountering unknown values
4. Never error on unknown values - allow operation to continue

### RQ-6: Existing Recommendations Infrastructure

**Question**: What recommendation infrastructure already exists in finfocus-core?

**Finding**: Substantial infrastructure exists for fetching and processing
recommendations from plugins. The CLI command just needs to wire this together.

**Evidence**:

```go
// From internal/proto/adapter.go - CostSourceClient interface
GetRecommendations(
    ctx context.Context,
    in *GetRecommendationsRequest,
    opts ...grpc.CallOption,
) (*GetRecommendationsResponse, error)

// From internal/engine/engine.go - Engine method
func (e *Engine) GetRecommendationsForResources(
    ctx context.Context,
    resources []ResourceDescriptor,
) (*RecommendationsResult, error)

// From internal/engine/types.go - Result types
type RecommendationsResult struct {
    Recommendations []Recommendation
    Errors          []RecommendationError
}
```

**Decision**: Implement `cost recommendations` CLI command following patterns from
`cost actual`:

1. Load Pulumi plan and map resources (reuse `loadAndMapResources()`)
2. Open plugin connections (reuse `openPlugins()`)
3. Call `engine.GetRecommendationsForResources()`
4. Render output (table/JSON/NDJSON)
5. Apply action type filter using new `ParseActionTypeFilter()` utility

## Alternatives Considered

### Alternative 1: Hardcode All 11 Types in Core

**Rejected because**: Creates maintenance burden - must update Core every time
spec adds new types.

**Chosen approach**: Use `pbc.AllRecommendationActionTypes()` from pluginsdk to
dynamically get valid types.

### Alternative 2: Use Proto Enum String Names Directly for Display

**Rejected because**: `RECOMMENDATION_ACTION_TYPE_RIGHTSIZE` is not user-
friendly.

**Chosen approach**: Create human-readable label mapping with title-cased,
space-separated words.

### Alternative 3: Case-Sensitive Filter Matching

**Rejected because**: Forces users to remember exact casing (`RIGHTSIZE` not
`rightsize`).

**Chosen approach**: Case-insensitive matching with `strings.EqualFold()`.

## Summary

<!-- markdownlint-disable MD060 -->

| Decision | Rationale |
| -------- | --------- |
| Use `pbc.AllRecommendationActionTypes()` | Dynamic enum list from spec SDK |
| String-to-label mapping table | Human-readable TUI display |
| Case-insensitive filter matching | Better UX for CLI |
| Log + display "Unknown" for unrecognized | Forward compatibility |
| Add `cost recommendations` command | Leverage existing engine infrastructure |

<!-- markdownlint-enable MD060 -->
