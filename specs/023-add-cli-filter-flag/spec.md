# Feature Specification: CLI Filter Flag Support

**Feature Branch**: `023-add-cli-filter-flag`
**Created**: 2025-12-23
**Status**: Draft
**Input**: User description: "The failing job is caused by CLI integration tests that use the `--filter` flag, which is not recognized. Symptoms: `TestActualCost_FilterByTag` and `TestActualCost_FilterByTagAndType` fail with error: `unknown flag: --filter`. Proposed Solution: Ensure the `--filter` flag is defined and accepted by the CLI command."

## Clarifications

### Session 2025-12-23

- Q: How should the CLI handle multiple `--filter` flags (e.g., `--filter "A" --filter "B"`)? → A: Cumulative (AND logic); track explicit OR logic (`--filter-type`) as a future issue.
- Q: Should the CLI perform validation on the filter string format? → A: Pass-through; CLI passes strings to the core, which handles validation and errors.
- Q: Should the `--filter` flag be global or command-specific? → A: Specific to relevant commands (e.g., `actual-cost`); open future tickets for other commands as needed.

## User Scenarios & Testing

### User Story 1 - Filter Costs by Tag or Type (Priority: P1)

As a user, I want to filter cost results by tag or resource type using the `--filter` flag so that I can analyze specific subsets of my infrastructure costs without seeing irrelevant data.

**Why this priority**: High. This functionality is currently causing integration test failures and is a core requirement for granular cost analysis.

**Independent Test**: Can be tested by running the CLI with `--filter "tag:Environment=dev"` and verifying that only matching resources are displayed (and by verifying the integration tests pass).

**Acceptance Scenarios**:

1. **Given** the CLI is installed, **When** I run `finfocus actual-cost --filter "tag:Environment=dev"`, **Then** the output should only show costs associated with resources having the "Environment=dev" tag.
2. **Given** the CLI is installed, **When** I run `finfocus actual-cost --filter "type:aws:s3/bucket"`, **Then** the output should only show costs for S3 buckets.
3. **Given** the CLI is installed, **When** I run `finfocus actual-cost --filter "tag:A" --filter "tag:B"`, **Then** the output should only show resources that match BOTH tag A AND tag B.
4. **Given** the integration test suite, **When** I run `TestActualCost_FilterByTag`, **Then** the test should pass without "unknown flag" errors.

### Edge Cases

- What happens when an empty filter string is provided? (Expected: Should probably be ignored or show all results, or error depending on existing logic).
- What happens when the filter string is malformed? (Expected: CLI should handle gracefully, potentially returning an error message from the core logic, but the *flag* itself should still be accepted).
- What happens when an unknown flag is passed? (Expected: Cobra's default behavior, but `--filter` specifically must NOT trigger this).

## Requirements

### Functional Requirements

- **FR-001**: The CLI MUST accept a string argument named `--filter` on cost-related commands (starting with `actual-cost`).
- **FR-002**: The value provided to `--filter` MUST be passed to the underlying cost estimation/querying logic.
- **FR-003**: The CLI help output for supported commands MUST list the `--filter` flag with a description (e.g., "Filter costs by tag or type").
- **FR-004**: The implementation MUST resolve the `unknown flag: --filter` error currently blocking the integration tests.
- **FR-005**: The CLI MUST accept multiple instances of the `--filter` flag, combining them with AND logic (intersection of results).
- **FR-006**: The CLI MUST NOT perform its own validation on the content of the filter string, relying instead on the core engine for validation and error reporting.

### Future Work

- **FW-001**: Support configurable filter logic (e.g., OR instead of AND) via a new flag like `--filter-type=ENUM[AND, OR]`.
- **FW-002**: Add `--filter` support to other commands (e.g., `plan`) as requirements emerge.

### Assumptions

- The underlying cost estimation engine already supports the filtering logic; this feature focuses on exposing that capability via the CLI.
- The format of the filter string (e.g., "tag:Key=Value") is already defined by the core logic.

## Success Criteria

### Measurable Outcomes

- **SC-001**: The integration test suite for cost filtering passes successfully without flag parsing errors.
- **SC-002**: Running the supported CLI commands with `--help` displays the `--filter` option.
- **SC-003**: Users can execute queries with the `--filter` flag without receiving syntax errors from the CLI parser.
