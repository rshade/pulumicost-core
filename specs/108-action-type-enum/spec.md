# Feature Specification: Extended RecommendationActionType Enum Support

**Feature Branch**: `108-action-type-enum`
**Created**: 2025-12-29
**Status**: Draft
**Input**: "Support extended RecommendationActionType enum from finfocus-spec"
**GitHub Issue**: #298

## Overview

Update finfocus-core components to recognize and properly handle the 5 new
`RecommendationActionType` enum values (MIGRATE, CONSOLIDATE, SCHEDULE,
REFACTOR, OTHER) that were added to finfocus-spec v0.4.9.

## Context

The finfocus-spec v0.4.9 release (2025-12-18) extended the
`RecommendationActionType` enum from 6 to 11 values via PR #173. The current
finfocus-core (v0.4.11) already has this dependency but internal components
(filter parser, TUI display, CLI help text) only recognize the original 6 types:

**Original Types (already supported)**:

- RIGHTSIZE - Resize resources to match actual usage
- TERMINATE - Stop or delete idle resources
- PURCHASE_COMMITMENT - Reserved instances/savings plans
- ADJUST_REQUESTS - Kubernetes request/limit tuning
- MODIFY - General configuration changes
- DELETE_UNUSED - Remove orphaned/unused resources

**New Types (to be supported)**:

- MIGRATE - Move workloads to different regions/zones/SKUs
- CONSOLIDATE - Combine multiple resources into fewer, larger ones
- SCHEDULE - Start/stop resources on schedule (dev/test)
- REFACTOR - Architectural changes (e.g., move to serverless)
- OTHER - Catch-all for provider-specific recommendations

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Filter Recommendations by New Action Types (Priority: P1)

As a cloud infrastructure operator, I want to filter cost recommendations by the
new action types (MIGRATE, CONSOLIDATE, SCHEDULE, REFACTOR, OTHER), so that I
can focus on specific categories of optimization opportunities.

**Why this priority**: Filtering is the primary way users narrow down large sets
of recommendations. Without parser support for new types, users cannot filter by
these categories at all.

**Independent Test**: Can be fully tested by using the filter flag with new
action type values and verifying only matching recommendations are displayed.

**Acceptance Scenarios**:

1. **Given** a set of recommendations including MIGRATE and RIGHTSIZE types,
   **When** I run `finfocus cost recommendations --filter "action=MIGRATE"`,
   **Then** only MIGRATE recommendations are shown.

2. **Given** a set of recommendations with various action types,
   **When** I run `finfocus cost recommendations --filter "action=SCHEDULE,REFACTOR"`,
   **Then** only SCHEDULE and REFACTOR recommendations are shown.

3. **Given** a filter with the value "action=INVALID_TYPE",
   **When** I run the command,
   **Then** an error message lists all valid action types including the 5 new ones.

---

### User Story 2 - Display New Action Types in TUI (Priority: P2)

As a cloud infrastructure operator viewing recommendations in the TUI, I want
new action types to display with appropriate labels and visual indicators, so
that I can quickly identify the type of optimization without referring to
documentation.

**Why this priority**: Users who enable TUI mode expect consistent visual
treatment of all data. Displaying "UNKNOWN" or raw enum values for new types
would be confusing.

**Independent Test**: Can be fully tested by returning recommendations with new
action types from a mock plugin and verifying the TUI renders them with proper
labels.

**Acceptance Scenarios**:

1. **Given** a recommendation with action type MIGRATE,
   **When** I view it in TUI mode,
   **Then** I see a "Migrate" label (not "MIGRATE" or "Unknown").

2. **Given** recommendations with all 11 action types,
   **When** I view them in TUI mode,
   **Then** each type displays with an appropriate human-readable label.

3. **Given** a recommendation with action type OTHER,
   **When** I view it in TUI mode,
   **Then** I see "Other" label to indicate provider-specific recommendation.

---

### User Story 3 - JSON Output Includes New Action Types (Priority: P3)

As an automation engineer consuming recommendation output programmatically, I
want new action types to be correctly serialized in JSON output, so that my
scripts can process all recommendation types consistently.

**Why this priority**: JSON output is critical for automation but follows the
same data flow as TUI - if action types are correctly parsed, JSON serialization
follows naturally.

**Independent Test**: Can be fully tested by requesting JSON output for
recommendations with new action types and validating the structure.

**Acceptance Scenarios**:

1. **Given** recommendations with CONSOLIDATE and SCHEDULE types,
   **When** I request output as `--output json`,
   **Then** the `action_type` field contains the correct string values.

2. **Given** recommendations with all 11 action types,
   **When** I request output as `--output json`,
   **Then** all action types serialize to their expected string representation.

---

### User Story 4 - CLI Command for Recommendations (Priority: P0)

As a cloud infrastructure operator, I want a `cost recommendations` CLI command
that fetches and displays cost optimization recommendations from plugins, so that
I can review optimization opportunities independently of `pulumi preview`.

**Why this priority**: This is a prerequisite for all other user stories. Without
the CLI command, there's no way to filter, display, or serialize recommendations.
The engine infrastructure (`GetRecommendationsForResources()`) already exists.

**Independent Test**: Can be fully tested by running the command with a Pulumi
plan JSON and verifying recommendations are fetched from plugins and displayed.

**Acceptance Scenarios**:

1. **Given** a Pulumi plan JSON and an installed cost plugin,
   **When** I run `finfocus cost recommendations --pulumi-json plan.json`,
   **Then** I see a table of recommendations with ID, ActionType, Description,
   ResourceID, and EstimatedSavings columns.

2. **Given** a Pulumi plan JSON with resources that have recommendations,
   **When** I run `finfocus cost recommendations --output json`,
   **Then** I receive valid JSON with all recommendation fields.

3. **Given** no installed plugins that support recommendations,
   **When** I run `finfocus cost recommendations --pulumi-json plan.json`,
   **Then** I see an informative message that no recommendations are available.

4. **Given** a plan with multiple resources,
   **When** I run with `--filter "action=MIGRATE"`,
   **Then** only MIGRATE recommendations are shown.

---

### Edge Cases

- What happens when a plugin returns an unrecognized action type value?
  Display as "Unknown (value)" and log a warning; do not error.
- What happens when the protobuf enum has additional future values?
  Handle gracefully as "Unknown" - forward compatibility for enum extension.
- What happens when filtering by a mix of valid and invalid action types?
  Reject the entire filter with an error listing valid types.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST recognize all 11 RecommendationActionType values in
  filter expressions: RIGHTSIZE, TERMINATE, PURCHASE_COMMITMENT, ADJUST_REQUESTS,
  MODIFY, DELETE_UNUSED, MIGRATE, CONSOLIDATE, SCHEDULE, REFACTOR, OTHER.

- **FR-002**: System MUST provide human-readable labels for new action types in
  TUI display: "Migrate", "Consolidate", "Schedule", "Refactor", "Other".

- **FR-003**: System MUST include all 11 action types in CLI help text for the
  `--filter` flag.

- **FR-004**: System MUST serialize new action types correctly in JSON output
  using their string representation (e.g., "MIGRATE", "CONSOLIDATE").

- **FR-005**: System MUST handle unknown/future action type values gracefully by
  displaying "Unknown" and logging a warning (not erroring).

- **FR-006**: System MUST maintain backward compatibility - existing action types
  (RIGHTSIZE, TERMINATE, etc.) continue to work unchanged.

- **FR-007**: Filter parser MUST support case-insensitive matching for action
  types (e.g., "migrate" matches MIGRATE).

- **FR-008**: System MUST provide a `cost recommendations` CLI command that
  fetches recommendations from plugins via `Engine.GetRecommendationsForResources()`
  and displays them in table, JSON, or NDJSON format.

### Key Entities

- **RecommendationActionType**: Extended enum with 11 values representing
  different cost optimization action categories. Used for filtering, display,
  and serialization.

- **ActionTypeLabel**: Human-readable mapping from enum values to display
  labels for TUI rendering (e.g., MIGRATE -> "Migrate").

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can filter recommendations by all 11 action types without
  errors or unexpected behavior.

- **SC-002**: TUI displays appropriate labels for all 11 action types with no
  "Unknown" labels for valid types.

- **SC-003**: JSON output correctly serializes all 11 action types.

- **SC-004**: CLI help text includes all 11 action types as valid filter values.

- **SC-005**: Backward compatibility: existing filters using original 6 action
  types continue to work without modification.

- **SC-006**: Unit test coverage for new action type handling reaches 80%+.

- **SC-007**: `cost recommendations` command successfully fetches and displays
  recommendations from plugins using existing engine infrastructure.

## Assumptions

- Engine infrastructure for recommendations already exists:
  `Engine.GetRecommendationsForResources()`, `RecommendationsResult`,
  `Recommendation` types, and `CostSourceClient.GetRecommendations()`.

- Filter expression parsing follows patterns established in `cost actual` command.

- TUI components follow existing label/styling patterns in `internal/tui/`.

- The finfocus-spec v0.4.9+ dependency (already at v0.4.11) includes the
  extended enum values in generated Go code.
