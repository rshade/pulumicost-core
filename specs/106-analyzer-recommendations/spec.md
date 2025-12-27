# Feature Specification: Analyzer Recommendations Display

**Feature Branch**: `106-analyzer-recommendations`
**Created**: 2025-12-25
**Status**: Draft
**Input**: User description: "feat(analyzer): Add recommendations to analyzer
diagnostics. Update the Analyzer to display optimization recommendations
(e.g., 'Right-sizing: Switch to t3.small to save $5/mo') alongside cost
estimates in the pulumi preview output."

## Clarifications

### Session 2025-12-25

- Q: Should E2E Analyzer testing be explicitly required? → A: Yes, E2E
  Analyzer tests must be included in the implementation.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - View Recommendations During Preview (Priority: P1)

As a cloud infrastructure operator running `pulumi preview`, I want to see
cost optimization recommendations alongside my estimated costs, so that I can
identify savings opportunities before deploying changes.

**Why this priority**: This is the core value proposition - making
recommendations visible where users already see costs (during preview).
Without this, users must run separate commands to discover optimization
opportunities.

**Independent Test**: Can be fully tested by running `pulumi preview` with a
plugin that returns recommendations, and verifying the diagnostic output
includes both cost estimates and actionable recommendations.

**Acceptance Scenarios**:

1. **Given** a Pulumi project with AWS EC2 instances and a configured cost
   plugin that returns recommendations, **When** I run `pulumi preview`,
   **Then** I see cost diagnostics that include optimization recommendations
   (e.g., "Right-sizing: Switch to t3.small to save $15.00/mo").

2. **Given** a Pulumi project with resources that have no available
   recommendations, **When** I run `pulumi preview`, **Then** I see the
   standard cost diagnostic without a recommendations section.

3. **Given** a Pulumi project with resources that have multiple
   recommendations, **When** I run `pulumi preview`, **Then** I see all
   recommendations formatted in a readable list within the diagnostic message.

---

### User Story 2 - Aggregated Recommendations in Stack Summary (Priority: P2)

As a cloud infrastructure operator, I want to see a summary of total potential
savings from all recommendations in my stack, so that I can quickly assess
the overall optimization opportunity.

**Why this priority**: Aggregating savings provides a high-level view of
optimization potential, but individual resource recommendations (P1) are
more actionable.

**Independent Test**: Can be fully tested by running `pulumi preview` with
multiple resources that have recommendations, and verifying the stack summary
diagnostic shows aggregate savings.

**Acceptance Scenarios**:

1. **Given** a Pulumi stack with 5 resources and 3 of them have cost
   optimization recommendations totaling $45/month savings, **When** I run
   `pulumi preview`, **Then** the stack summary diagnostic shows
   "3 optimization recommendations with $45.00/mo potential savings"
   alongside the total estimated cost.

---

### User Story 3 - Graceful Handling of Missing Recommendations (Priority: P3)

As a cloud infrastructure operator, I want the preview to complete
successfully even when recommendation data is unavailable, so that I can
still see cost estimates without errors.

**Why this priority**: Error resilience is important but secondary to core
functionality. The analyzer already handles cost calculation failures
gracefully; recommendations follow the same pattern.

**Independent Test**: Can be fully tested by mocking a plugin that returns
costs but no recommendations, or by testing with resources that have no
recommendation data.

**Acceptance Scenarios**:

1. **Given** a plugin that returns cost data but no recommendations,
   **When** I run `pulumi preview`, **Then** I see cost diagnostics without
   recommendations and no errors.

2. **Given** a plugin that returns an error for recommendations but succeeds
   for costs, **When** I run `pulumi preview`, **Then** I see cost
   diagnostics with a note that recommendations were unavailable.

---

### Edge Cases

- What happens when a recommendation has no estimated savings?
  Display the recommendation text without savings information.
- What happens when recommendations have mixed currencies?
  Display each recommendation with its currency; aggregate savings only when
  currencies match (otherwise show "mixed currencies").
- What happens when a single resource has more than 3 recommendations?
  Display first 3 recommendations with "and X more" indicator.
- What happens when the plugin returns malformed recommendation data?
  Log a warning and skip the malformed recommendation; display valid ones.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST add a `Recommendations` field to
  `engine.CostResult` as a slice of `Recommendation` structs.
- **FR-002**: System MUST define a `Recommendation` struct with fields:
  `Type` (string, e.g., "Right-sizing"), `Description` (string, actionable
  text), `EstimatedSavings` (float64), and `Currency` (string).
- **FR-003**: System MUST update the engine to populate the `Recommendations`
  field from plugin responses when available.
- **FR-004**: System MUST update `CostToDiagnostic` to format recommendations
  as part of the diagnostic message when present.
- **FR-005**: System MUST update `StackSummaryDiagnostic` to include
  aggregate recommendation count and total potential savings.
- **FR-006**: System MUST format recommendations in a human-readable format
  (e.g., "Right-sizing: Switch to t3.small to save $15.00/mo").
- **FR-007**: System MUST gracefully handle resources with no recommendations
  by omitting the recommendations section from the diagnostic.
- **FR-008**: System MUST maintain ADVISORY enforcement level for all
  diagnostics (never block deployments).
- **FR-009**: System MUST handle recommendation display for resources with
  sustainability metrics (combine both in message).
- **FR-010**: Implementation MUST include E2E Analyzer tests that verify
  recommendations appear correctly in actual `pulumi preview` output
  (in `test/e2e/` directory).

### Key Entities

- **Recommendation**: Represents a single cost optimization suggestion with
  type, description, and optional savings estimate. Related to CostResult as
  an optional collection.
- **CostResult**: Extended to include a slice of Recommendation entities.
  Unchanged relationship with ResourceDescriptor.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users see cost optimization recommendations in `pulumi preview`
  output when recommendations are available from plugins.
- **SC-002**: All existing cost diagnostic functionality continues to work
  unchanged (no regressions).
- **SC-003**: Stack summary includes aggregate recommendation count and
  savings when recommendations exist.
- **SC-004**: Unit test coverage for new recommendation formatting code
  reaches 80% or higher.
- **SC-005**: E2E Analyzer tests in `test/e2e/` verify recommendations appear
  correctly in actual `pulumi preview` diagnostic output.
- **SC-006**: Preview completes successfully even when recommendation data
  is unavailable or malformed.

## Assumptions

- Plugins already support the `GetRecommendations` RPC via the pluginsdk
  v0.4.10 `RecommendationsProvider` interface.
- Recommendations are retrieved via a separate `GetRecommendations` RPC call
  (distinct from `GetProjectedCost`) and merged with cost results before
  diagnostic formatting.
- The existing `CostResult.Notes` field will NOT be used for recommendations;
  a dedicated field provides better structure.
- Display format uses text indicators for visual distinction in terminal
  output.
- Currency formatting follows existing patterns (e.g., "$15.00/mo").
