# Feature Specification: Pre-Flight Request Validation

**Feature Branch**: `107-preflight-validation`
**Created**: 2025-12-29
**Status**: Draft
**Input**: GitHub Issue #233 - Adopt pluginsdk/validation.go for pre-flight request validation

## User Scenarios & Testing

### User Story 1 - Clear Error Messages Before Plugin Calls (Priority: P1)

As a developer using pulumicost to estimate infrastructure costs, I want to receive clear, actionable error messages when my resource definitions are incomplete, so that I can fix issues before waiting for plugin timeouts or cryptic gRPC errors.

**Why this priority**: This directly addresses the core user pain point - generic "InvalidArgument" errors from plugins provide no guidance on what's wrong. Pre-flight validation with actionable messages ("SKU empty - use mapping.ExtractAWSSKU()") helps developers fix issues quickly.

**Independent Test**: Can be fully tested by running `pulumicost cost projected` with an incomplete Pulumi plan (missing instance type) and verifying the error message includes specific guidance.

**Acceptance Scenarios**:

1. **Given** a Pulumi plan with an EC2 instance missing the `instanceType` property, **When** the user runs `pulumicost cost projected --pulumi-json plan.json`, **Then** the output shows a warning with guidance about the missing SKU and the resource is marked with $0.00 cost and an explanatory note.

2. **Given** a Pulumi plan with resources from multiple providers where some are valid and some are missing region, **When** the user runs projected cost calculation, **Then** valid resources return costs while invalid ones show validation errors with specific missing fields.

3. **Given** a resource with empty provider field, **When** pre-flight validation runs, **Then** the error message indicates "provider is empty" and processing continues with a placeholder result.

---

### User Story 2 - Consistent Validation Between Core and Plugins (Priority: P2)

As a plugin developer, I want Core to use the same validation logic that plugins use, so that validation behavior is predictable and plugins don't receive requests that would fail the same validation checks.

**Why this priority**: Consistency reduces debugging complexity - if Core validates using the same rules, plugins won't receive malformed requests that they would also reject, eliminating redundant error paths.

**Independent Test**: Can be tested by comparing validation results between Core pre-flight checks and plugin-side validation for the same malformed request.

**Acceptance Scenarios**:

1. **Given** the Core uses `pluginsdk.ValidateProjectedCostRequest()`, **When** a request would fail plugin-side validation, **Then** it fails Core pre-flight validation with the same error type.

2. **Given** a valid request passes Core pre-flight validation, **When** sent to a conformant plugin, **Then** it should not fail plugin-side validation.

---

### User Story 3 - Debug Logging for Validation Failures (Priority: P3)

As an operator troubleshooting cost calculation issues, I want validation failures to be logged with full context (resource type, trace ID), so that I can diagnose issues in production without enabling verbose debugging.

**Why this priority**: Observability is important but not critical for MVP - the feature works without detailed logging, but logging improves operability.

**Independent Test**: Can be tested by running with `--debug` flag and verifying validation failures appear in logs with resource context.

**Acceptance Scenarios**:

1. **Given** debug logging is enabled, **When** pre-flight validation fails for a resource, **Then** the log entry includes resource_type, trace_id, and the validation error.

2. **Given** normal logging level, **When** pre-flight validation fails, **Then** a WARN-level message appears with the resource type and error summary.

---

### Edge Cases

- What happens when all resources in a request fail validation? The command should still complete with all placeholder results showing validation errors.
- What happens when validation fails but the resource would have been skipped anyway (unsupported provider)? Validation failure takes precedence and is reported.
- What happens with partially populated properties (SKU present but region empty)? Each field is validated independently; only the first failure is reported (fail-fast).
- How are utilization range errors reported? If utilization is outside [0.0, 1.0], the error message includes the invalid value.

## Requirements

### Functional Requirements

- **FR-001**: System MUST call `pluginsdk.ValidateProjectedCostRequest()` for each resource before making the gRPC call to the plugin.
- **FR-002**: System MUST call `pluginsdk.ValidateActualCostRequest()` before making actual cost gRPC calls.
- **FR-003**: When validation fails, System MUST add a placeholder CostResult with MonthlyCost=0 and Notes containing the validation error.
- **FR-004**: When validation fails, System MUST log the failure at WARN level with resource context (type, trace_id, error).
- **FR-005**: When validation fails, System MUST continue processing remaining resources (fail-fast per resource, not per batch).
- **FR-006**: Validation errors MUST be included in the ErrorDetail slice returned by `GetProjectedCostWithErrors`.
- **FR-007**: System MUST NOT send requests to plugins when pre-flight validation fails.

### Key Entities

- **CostResultWithErrors**: Extended with ErrorDetail entries for validation failures (existing type, new error category).
- **ErrorDetail**: Captures validation failure context including resource type, error message, and timestamp (existing type, new error source).

## Success Criteria

### Measurable Outcomes

- **SC-001**: Users can identify and fix resource definition issues from the error message alone without needing to inspect debug logs.
- **SC-002**: Validation errors are distinguished from plugin errors in the output (different error prefix or note format).
- **SC-003**: Resources with validation errors appear in output with $0.00 cost and explanatory note rather than being silently dropped.
- **SC-004**: Processing time for malformed resources is reduced by avoiding plugin round-trips (pre-flight validation is faster than plugin call + error).

## Dependencies

- **pluginsdk v0.4.11+**: Already integrated - contains `ValidateProjectedCostRequest()` and `ValidateActualCostRequest()` functions.
- **zerolog logging**: Already integrated for structured logging with trace ID support.

## Assumptions

- The pluginsdk validation functions provide error messages that are sufficiently actionable for users (verified by reviewing function documentation).
- Existing error handling patterns in `GetProjectedCostWithErrors` and `GetActualCostWithErrors` support the addition of validation-specific errors.
- The fail-fast per-resource approach (continue processing remaining resources) is the desired behavior rather than batch failure.
- Performance impact of validation is negligible (<100ns per resource as documented in pluginsdk).

## Out of Scope

- Changing the pluginsdk validation logic itself (that's in pulumicost-spec).
- Adding validation for GetRecommendations requests (future enhancement if needed).
- Retry logic for validation failures (validation failures are deterministic, not transient).
