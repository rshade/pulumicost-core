# Feature Specification: Integration Tests for --filter Flag

**Feature Branch**: `105-cost-filter-tests`  
**Created**: Tue Dec 16 2025  
**Status**: Draft  
**Input**: User description: "Integration tests for --filter flag across cost commands"

## Clarifications

### Session 2025-12-16

- Q: What is the expected scale for test fixtures? → A: 10-20 resources
- Q: Is the filter case-sensitive? → A: Case-sensitive
- Q: What specific error messages should be expected for invalid syntax? → A: Invalid filter syntax: expected 'type=value' or 'provider=value'
- Q: Are there any performance requirements for filtering? → A: Performance should not degrade significantly compared to unfiltered queries
- Q: What are the key attributes of the TestPlan entity? → A: It should be an actual exported plan

## User Scenarios & Testing _(mandatory)_

<!--
  IMPORTANT: User stories should be PRIORITIZED as user journeys ordered by importance.
  Each user story/journey must be INDEPENDENTLY TESTABLE - meaning if you implement just ONE of them,
  you should still have a viable MVP (Minimum Viable Product) that delivers value.
-->

### User Story 1 - Projected Cost Filtering by Type and Provider (Priority: P1)

As a DevOps engineer, I want to filter projected cost outputs by resource type or provider so that I can isolate costs for specific infrastructure components (e.g., compute vs. storage) without seeing the entire plan.

**Why this priority**: The `--filter` flag is critical for usability in large stacks, and `cost projected` is the primary command for pre-deployment checks.

**Independent Test**: Can be tested by running `pulumicost cost projected` against a known plan with mixed resources and verifying the output contains only the requested types.

**Acceptance Scenarios**:

1. **Given** a plan with mixed AWS EC2, S3, and Azure resources, **When** I run `cost projected` with `--filter "type=aws:ec2/instance"`, **Then** the output contains _only_ EC2 instances.
2. **Given** a plan with mixed AWS and Azure resources, **When** I run `cost projected` with `--filter "provider=aws"`, **Then** the output contains _only_ AWS resources.
3. **Given** a plan with mixed resources, **When** I run `cost projected` with a substring filter `--filter "type=ec2"`, **Then** the output contains all resources with "ec2" in their type (instances, volumes, etc.).

---

### User Story 2 - Actual Cost Filtering by Tags (Priority: P1)

As a FinOps practitioner, I want to filter actual historical costs by tags (e.g., `env=prod`) so that I can generate reports for specific environments or cost centers.

**Why this priority**: Tag-based filtering is essential for cost allocation and is a documented feature of `cost actual`.

**Independent Test**: Can be tested by mocking the cost provider response and verifying the CLI passes the correct filters to the engine and displays filtered results.

**Acceptance Scenarios**:

1. **Given** historical cost data with tagged resources, **When** I run `cost actual` with `--filter "tag:env=prod"`, **Then** the output aggregates costs only for resources with the `env=prod` tag.
2. **Given** historical cost data, **When** I run `cost actual` with both a group-by and a filter (e.g., `--group-by "tag:env=prod" --filter "type=ec2"`), **Then** the output reflects the intersection of those criteria.

---

### User Story 3 - Robust Edge Case Handling (Priority: P2)

As a user, I want the CLI to handle no-match scenarios gracefully and reject invalid filter syntax so that I am not confused by empty or misleading outputs.

**Why this priority**: Ensures a polished user experience and prevents silent failures in automation scripts.

**Independent Test**: Can be tested by providing non-matching filters and malformed strings to the CLI.

**Acceptance Scenarios**:

1. **Given** a plan, **When** I provide a filter that matches nothing (e.g., `--filter "type=nonexistent"`), **Then** the command succeeds but returns an empty result set (not an error).
2. **Given** a plan, **When** I provide a filter with invalid syntax (e.g., `--filter "invalid string"`), **Then** the command fails with error message "Invalid filter syntax: expected 'type=value' or 'provider=value'".
3. **Given** a plan, **When** I filter using case-insensitive terms (if supported) or special characters, **Then** the behavior is consistent and predictable.

---

### User Story 4 - Consistent Output Formats (Priority: P2)

As a pipeline integrator, I want filtering to work consistently across all output formats (Table, JSON, NDJSON) so that I can parse results reliably in CI/CD workflows.

**Why this priority**: Users consume data in different ways; automation relies on JSON/NDJSON while humans use Table.

**Independent Test**: Run the same filter query with different `--output` flags and validate the structure.

**Acceptance Scenarios**:

1. **Given** a filter query, **When** I specify `--output json`, **Then** the result is a valid JSON array containing only matched items.
2. **Given** a filter query, **When** I specify `--output table`, **Then** the ASCII table displays only matched items.
3. **Given** a filter query, **When** I specify `--output ndjson`, **Then** the stream contains valid newline-delimited JSON objects for matched items.

---

### Edge Cases

- What happens when a filter string contains special characters like `/` or `:`? (Should handle them correctly as separators)
- How does the system handle conflicting filters if multiple are provided? (Current CLI implementation allows one flag instance; this spec focuses on the existing single-flag behavior)
- Case-sensitive filtering (e.g., `type=aws:ec2/instance` matches but `type=AWS:EC2/INSTANCE` does not).

## Requirements _(mandatory)_

### Functional Requirements

- **FR-001**: The test suite MUST verify that `cost projected` filters resources by exact type match.
- **FR-002**: The test suite MUST verify that `cost projected` filters resources by type substring match.
- **FR-003**: The test suite MUST verify that `cost projected` filters resources by provider name.
- **FR-004**: The test suite MUST verify that `cost actual` filters resources by tag key/value pairs.
- **FR-005**: The test suite MUST verify that queries returning zero matches produce an empty result set without returning a non-zero exit code.
- **FR-006**: The test suite MUST verify that invalid filter syntax strings result in a specific error message.
- **FR-007**: The test suite MUST verify that filtered results are correctly formatted when using `--output json`.
- **FR-008**: The test suite MUST verify that filtered results are correctly formatted when using `--output table`.

### Non-Functional Requirements

- **NFR-001**: Filter operations should not degrade significantly compared to unfiltered queries.

### Key Entities

- **TestPlan**: An actual exported Pulumi plan JSON file, containing 10-20 resources with a known distribution of resource types (AWS EC2, S3, Azure, GCP).
- **TestHelper**: The existing integration test helper framework used to execute CLI commands in a controlled environment.

## Success Criteria _(mandatory)_

### Measurable Outcomes

- **SC-001**: 100% of the defined acceptance criteria scenarios are covered by automated integration tests.
- **SC-002**: All new tests pass successfully in the standard test environment.
- **SC-003**: The integration test suite execution time does not increase by more than 10 seconds.
- **SC-004**: Test coverage for the filter functionality exceeds 80%.
