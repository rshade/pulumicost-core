# Data Model: Extended RecommendationActionType Enum Support

**Feature**: 108-action-type-enum
**Date**: 2025-12-29

## Overview

This feature adds utility functions for working with the
`RecommendationActionType` enum from pulumicost-spec. No new entities are
created - the proto enum already exists in the dependency.

## Entities

### RecommendationActionType (from pulumicost-spec)

**Source**: `github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1`

**Type**: Proto3 enum

**Values**:

<!-- markdownlint-disable MD013 MD060 -->

| Enum Value | Int | Short Name | Label |
| ---------- | --- | ---------- | ----- |
| `RECOMMENDATION_ACTION_TYPE_UNSPECIFIED` | 0 | UNSPECIFIED | Unspecified |
| `RECOMMENDATION_ACTION_TYPE_RIGHTSIZE` | 1 | RIGHTSIZE | Rightsize |
| `RECOMMENDATION_ACTION_TYPE_TERMINATE` | 2 | TERMINATE | Terminate |
| `RECOMMENDATION_ACTION_TYPE_PURCHASE_COMMITMENT` | 3 | PURCHASE_COMMITMENT | Purchase Commitment |
| `RECOMMENDATION_ACTION_TYPE_ADJUST_REQUESTS` | 4 | ADJUST_REQUESTS | Adjust Requests |
| `RECOMMENDATION_ACTION_TYPE_MODIFY` | 5 | MODIFY | Modify |
| `RECOMMENDATION_ACTION_TYPE_DELETE_UNUSED` | 6 | DELETE_UNUSED | Delete Unused |
| `RECOMMENDATION_ACTION_TYPE_MIGRATE` | 7 | MIGRATE | Migrate |
| `RECOMMENDATION_ACTION_TYPE_CONSOLIDATE` | 8 | CONSOLIDATE | Consolidate |
| `RECOMMENDATION_ACTION_TYPE_SCHEDULE` | 9 | SCHEDULE | Schedule |
| `RECOMMENDATION_ACTION_TYPE_REFACTOR` | 10 | REFACTOR | Refactor |
| `RECOMMENDATION_ACTION_TYPE_OTHER` | 11 | OTHER | Other |

<!-- markdownlint-enable MD013 MD060 -->

### ActionTypeInfo (new utility type)

**Location**: `internal/proto/action_types.go`

**Purpose**: Provides Core-side utilities for action type handling

```go
// ActionTypeInfo holds display and filtering metadata for an action type.
type ActionTypeInfo struct {
    // ProtoValue is the proto enum value.
    ProtoValue pbc.RecommendationActionType

    // ShortName is the filter-friendly name (e.g., "RIGHTSIZE", "MIGRATE").
    ShortName string

    // DisplayLabel is the human-readable label for TUI display.
    DisplayLabel string
}
```

**Relationships**:

- Maps 1:1 with `pbc.RecommendationActionType` enum values
- Used by filter parser for validation
- Used by TUI renderer for display labels

## Functions

### ActionTypeLabel

**Purpose**: Get human-readable label for a proto enum value

**Signature**:

```go
func ActionTypeLabel(at pbc.RecommendationActionType) string
```

**Behavior**:

- Returns display label (e.g., "Migrate" for `MIGRATE`)
- Returns "Unknown (N)" for unrecognized values where N is the integer value

### ParseActionType

**Purpose**: Parse a filter string into proto enum value(s)

**Signature**:

```go
func ParseActionType(s string) (pbc.RecommendationActionType, error)
```

**Behavior**:

- Case-insensitive matching
- Returns error for unknown type names listing valid options
- Accepts both short names ("MIGRATE") and full names
  ("RECOMMENDATION_ACTION_TYPE_MIGRATE")

### ParseActionTypeFilter

**Purpose**: Parse comma-separated action types for filter expressions

**Signature**:

```go
func ParseActionTypeFilter(filter string) ([]pbc.RecommendationActionType, error)
```

**Behavior**:

- Splits on comma
- Trims whitespace
- Validates each value
- Returns slice of parsed types
- Returns error if any type is invalid

### ValidActionTypes

**Purpose**: Get list of valid action type names for help text

**Signature**:

```go
func ValidActionTypes() []string
```

**Behavior**:

- Returns short names suitable for filter expressions
- Excludes UNSPECIFIED (not user-selectable)

## Validation Rules

1. **Filter Value**: Must match a known action type short name
2. **Case Sensitivity**: Matching is case-insensitive
3. **Unknown Values**: Log warning but don't error during display
4. **UNSPECIFIED**: Not valid as a filter value (excluded from validation)

## State Transitions

N/A - Action types are immutable enum values with no state.
