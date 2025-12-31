# Feature Specification: Cost Recommendations Command Enhancement

**Feature Branch**: `109-cost-recommendations`
**Created**: 2025-12-30
**Status**: Draft
**Input**: User description: "Add recommendations command for FinOps optimization display with interactive TUI, summary views, and enhanced filtering"
**GitHub Issue**: #216

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Quick Overview of Recommendations (Priority: P1)

A DevOps engineer wants to quickly see the most impactful cost optimization recommendations without reading through every single recommendation. They run the recommendations command and see a summary with the top 5 recommendations sorted by potential savings.

**Why this priority**: This is the most common use case - users need quick visibility into the biggest cost-saving opportunities without information overload. The summary view provides immediate value.

**Independent Test**: Can be fully tested by running `pulumicost cost recommendations --pulumi-json plan.json` and verifying the output shows a summary section followed by top 5 recommendations sorted by savings descending.

**Acceptance Scenarios**:

1. **Given** a Pulumi plan with 15 resources and plugins returning 20 recommendations, **When** the user runs `pulumicost cost recommendations --pulumi-json plan.json`, **Then** they see a summary showing total recommendations count, total potential savings, and the top 5 recommendations sorted by estimated savings (highest first).

2. **Given** a Pulumi plan with 2 resources and plugins returning 3 recommendations, **When** the user runs the recommendations command, **Then** they see all 3 recommendations (since fewer than 5) without a "more recommendations" message.

3. **Given** plugins return 0 recommendations, **When** the user runs the recommendations command, **Then** they see a summary indicating 0 recommendations with $0 potential savings.

---

### User Story 2 - View All Recommendations in Detail (Priority: P2)

A platform engineer reviewing infrastructure costs needs to see all recommendations with complete details (current state, recommended changes, reasoning) to make informed decisions about which optimizations to pursue.

**Why this priority**: Detailed information is essential for decision-making but secondary to getting a quick overview. Users will use this after the summary identifies areas of interest.

**Independent Test**: Can be fully tested by running `pulumicost cost recommendations --pulumi-json plan.json --verbose` and verifying all recommendations display with full details.

**Acceptance Scenarios**:

1. **Given** plugins return 10 recommendations, **When** the user runs `pulumicost cost recommendations --pulumi-json plan.json --verbose`, **Then** all 10 recommendations display with resource ID, action type, description, estimated savings, and source plugin.

2. **Given** a recommendation with detailed description text, **When** displayed in verbose mode, **Then** the full description is shown without truncation.

---

### User Story 3 - Filter Recommendations by Action Type (Priority: P2)

A FinOps practitioner wants to focus only on rightsizing opportunities across the infrastructure. They filter recommendations to show only RIGHTSIZE action types.

**Why this priority**: Filtering enables focused analysis and is essential for teams with specific optimization workflows (e.g., rightsizing campaigns).

**Independent Test**: Can be fully tested by running `pulumicost cost recommendations --pulumi-json plan.json --filter "action=RIGHTSIZE"` and verifying only RIGHTSIZE recommendations appear.

**Acceptance Scenarios**:

1. **Given** plugins return recommendations with various action types (RIGHTSIZE, TERMINATE, MIGRATE), **When** the user runs with `--filter "action=RIGHTSIZE"`, **Then** only RIGHTSIZE recommendations appear in the output.

2. **Given** the user filters by multiple action types, **When** running with `--filter "action=RIGHTSIZE,TERMINATE"`, **Then** recommendations matching either action type appear.

3. **Given** the user specifies an invalid action type, **When** running with `--filter "action=INVALID_TYPE"`, **Then** an error message lists valid action types.

---

### User Story 4 - Interactive Exploration of Recommendations (Priority: P3)

A cloud architect using PulumiCost interactively in their terminal wants to navigate through recommendations using keyboard controls, view details of specific recommendations, and filter dynamically without re-running the command.

**Why this priority**: Interactive TUI enhances user experience for terminal users but requires more implementation complexity and is additive to the core functionality.

**Independent Test**: Can be fully tested by running the command in an interactive terminal and verifying keyboard navigation, detail view on Enter, and filter on "/" key.

**Acceptance Scenarios**:

1. **Given** the user runs the recommendations command in a TTY terminal, **When** the command executes, **Then** an interactive table appears with keyboard navigation (up/down arrows).

2. **Given** the user is viewing the interactive table, **When** they press Enter on a recommendation, **Then** a detail view shows the full recommendation information.

3. **Given** the user is viewing the detail view, **When** they press Escape, **Then** they return to the list view.

4. **Given** the user is viewing the interactive table, **When** they press "/" and type a filter term, **Then** the table filters to show matching recommendations.

5. **Given** the command runs in a non-TTY environment (CI/CD pipeline), **When** the command executes, **Then** plain text output is displayed without interactive features.

---

### User Story 5 - Machine-Readable Output for Automation (Priority: P2)

A DevOps team wants to integrate recommendations into their CI/CD pipeline to track optimization opportunities over time and generate reports. They need JSON output for programmatic processing.

**Why this priority**: JSON output enables integration with external tools and dashboards, which is critical for enterprise adoption.

**Independent Test**: Can be fully tested by running `pulumicost cost recommendations --pulumi-json plan.json --output json` and validating the JSON structure.

**Acceptance Scenarios**:

1. **Given** plugins return recommendations, **When** the user runs with `--output json`, **Then** valid JSON output is produced with a summary object and recommendations array.

2. **Given** the user needs streaming output for large datasets, **When** running with `--output ndjson`, **Then** each recommendation is output as a separate JSON line.

---

### User Story 6 - View Loading Progress (Priority: P4)

A user with multiple plugins installed wants to see which plugins are being queried and their progress while waiting for recommendations to load.

**Why this priority**: Loading feedback improves perceived performance but is polish rather than core functionality.

**Independent Test**: Can be tested by running the command with slow plugins and verifying spinner/progress display.

**Acceptance Scenarios**:

1. **Given** multiple plugins are being queried, **When** the command is fetching recommendations, **Then** a loading spinner displays with status text.

2. **Given** a plugin returns results while others are still loading, **When** displayed, **Then** the user sees which plugins have completed and which are still pending.

---

### Edge Cases

- What happens when all plugins fail? Display an error summary with no recommendations.
- What happens when some plugins fail but others succeed? Display successful recommendations with an error summary section.
- What happens with mixed currencies across plugins? Display recommendations with their respective currencies, noting the currency in each row.
- What happens with very long resource IDs or descriptions? Truncate with ellipsis in table view, show full text in detail/verbose view.
- What happens with zero estimated savings? Display the recommendation but show "-" or "$0.00" in the savings column.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST display a summary section showing total recommendation count and total potential savings.
- **FR-002**: System MUST display the top 5 recommendations by estimated savings in default (non-verbose) mode.
- **FR-003**: System MUST support a `--verbose` flag that displays all recommendations with full details.
- **FR-004**: System MUST support filtering by action type via `--filter "action=TYPE"` syntax.
- **FR-005**: System MUST validate action type filters against the allowed values and provide helpful error messages.
- **FR-006**: System MUST aggregate recommendations from all available plugins.
- **FR-007**: System MUST gracefully handle plugins that fail or return errors, displaying an error summary.
- **FR-008**: System MUST support `--output json` and `--output ndjson` formats for machine-readable output.
- **FR-009**: System MUST detect TTY vs non-TTY environments and render appropriately (interactive vs plain text).
- **FR-010**: System MUST provide interactive table navigation when in TTY mode with Bubble Tea.
- **FR-011**: System MUST provide a detail view accessible via Enter key in interactive mode.
- **FR-012**: System MUST support keyboard filtering via "/" key in interactive mode.
- **FR-013**: System MUST display a loading spinner during plugin queries in interactive mode.
- **FR-014**: System MUST maintain backward compatibility with the existing `--pulumi-json` required flag.
- **FR-015**: System MUST support the existing `--adapter` flag to restrict to a specific plugin.

### Key Entities

- **Recommendation**: A cost optimization suggestion with resource ID, action type, description, estimated savings, and source plugin.
- **RecommendationSummary**: Aggregated statistics including total count, total savings, and breakdown by action type.
- **ActionType**: Enumeration of optimization actions (RIGHTSIZE, TERMINATE, PURCHASE_COMMITMENT, ADJUST_REQUESTS, MODIFY, DELETE_UNUSED, MIGRATE, CONSOLIDATE, SCHEDULE, REFACTOR, OTHER).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can view a recommendations summary in under 3 seconds after plugin responses complete.
- **SC-002**: Summary mode displays in 50% less vertical space compared to showing all recommendations.
- **SC-003**: 100% of valid action type filters produce accurate filtered results.
- **SC-004**: Interactive mode responds to keyboard input within 100ms.
- **SC-005**: JSON output passes schema validation for integration tools.
- **SC-006**: Command gracefully handles up to 1000 recommendations without performance degradation.
- **SC-007**: 90% of users can find their desired recommendation within 3 interactions (summary, filter, or navigation).
- **SC-008**: Error states clearly communicate what went wrong and what the user can do.
- **SC-009**: Test coverage for the recommendations command achieves 80% minimum.

## Assumptions

- The `internal/tui` package (#222) provides necessary components (spinner, table, styles).
- The `GetRecommendations` RPC is available in pulumicost-spec (rshade/pulumicost-spec#122).
- Plugins return recommendations with consistent structure (resource ID, type, description, savings, currency).
- The existing engine.Recommendation struct remains unchanged.
- Standard terminal dimensions of 80x24 characters minimum are assumed for interactive mode.

## Dependencies

- **Prerequisite**: #222 - Shared TUI package (already implemented in `internal/tui/`)
- **Prerequisite**: rshade/pulumicost-spec#122 - GetRecommendations RPC (already implemented)
- **Existing**: `cost_recommendations.go` - Basic command structure already exists

## Scope Boundaries

### In Scope

- Summary view with top 5 recommendations
- Verbose flag for all recommendations
- Action type filtering (existing functionality enhanced)
- Interactive Bubble Tea TUI with table navigation
- Detail view for individual recommendations
- Loading spinner during plugin queries
- JSON and NDJSON output formats

### Out of Scope

- Category filtering (cost, performance, security, reliability)
- Priority filtering (low, medium, high, critical)
- Savings threshold filtering (savings>100)
- Provider filtering
- Persistent recommendation history
- Recommendation tracking/dismissal
- Cost comparison before/after implementation
