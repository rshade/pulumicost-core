# Tasks: Pre-Flight Request Validation

**Input**: Design documents from `/specs/107-preflight-validation/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md

**Tests**: Per Constitution Principle II (Test-Driven Development), tests are MANDATORY and
must be written BEFORE implementation. All code changes must maintain minimum 80% test coverage
(95% for critical paths).

**Completeness**: Per Constitution Principle VI (Implementation Completeness), all tasks MUST be
fully implemented. Stub functions, placeholders, and TODO comments are strictly forbidden.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add required import for pluginsdk validation

- [x] T001 [US1] Add pluginsdk import to `internal/proto/adapter.go` alongside existing mapping import

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: No foundational work needed - existing infrastructure supports this feature

**Note**: The following already exist and require no changes:

- `logging.FromContext(ctx)` for structured logging with trace_id
- `CostResult.Notes` field for validation error messages
- `CostResultWithErrors.Errors` for ErrorDetail tracking
- pluginsdk v0.4.11 already in go.mod

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Clear Error Messages Before Plugin Calls (Priority: P1) 🎯 MVP

**Goal**: Provide actionable validation errors for GetProjectedCost requests before sending to plugins

**Independent Test**: Run `finfocus cost projected --pulumi-json plan.json` with incomplete
plan and verify VALIDATION: prefix in notes

### Tests for User Story 1 (MANDATORY - TDD Required) ⚠️

> **CONSTITUTION REQUIREMENT: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T002 [P] [US1] Unit test `TestGetProjectedCost_ValidationFailure_EmptyProvider` in `internal/proto/adapter_test.go`
- [x] T003 [P] [US1] Unit test `TestGetProjectedCost_ValidationFailure_EmptySKU` in `internal/proto/adapter_test.go`
- [x] T004 [P] [US1] Unit test `TestGetProjectedCost_ValidationFailure_EmptyRegion` in `internal/proto/adapter_test.go`
- [x] T005 [P] [US1] Unit test `TestGetProjectedCost_ValidationFailure_MixedValidInvalid` in `internal/proto/adapter_test.go`

### Implementation for User Story 1

- [x] T006 [US1] Add pre-flight validation call after request construction in `GetProjectedCostWithErrors()` at `internal/proto/adapter.go:68-162`
- [x] T007 [US1] Add WARN-level logging with resource_type and trace_id context for validation failures
- [x] T008 [US1] Add placeholder CostResult with `Notes: "VALIDATION: <error>"` on validation failure
- [x] T009 [US1] Verify tests pass with `go test -v ./internal/proto/... -run TestGetProjectedCost_ValidationFailure`

**Checkpoint**: User Story 1 complete - GetProjectedCost validation works with actionable error messages

---

## Phase 4: User Story 2 - Consistent Validation for Actual Costs (Priority: P2)

**Goal**: Apply same validation pattern to GetActualCost requests for consistency

**Independent Test**: Test actual cost validation with empty resourceId or invalid time range

### Tests for User Story 2 (MANDATORY - TDD Required) ⚠️

- [x] T010 [P] [US2] Unit test `TestGetActualCost_ValidationFailure_EmptyResourceID` in `internal/proto/adapter_test.go`
- [x] T011 [P] [US2] Unit test `TestGetActualCost_ValidationFailure_InvalidTimeRange` in `internal/proto/adapter_test.go`

### Implementation for User Story 2

- [x] T012 [US2] Add pre-flight validation call in `GetActualCostWithErrors()` at `internal/proto/adapter.go:164-267`
- [x] T013 [US2] Add WARN-level logging with resource_id and trace_id context for validation failures
- [x] T014 [US2] Add placeholder CostResult with VALIDATION note on validation failure
- [x] T015 [US2] Verify tests pass with `go test -v ./internal/proto/... -run TestGetActualCost_ValidationFailure`

**Checkpoint**: User Story 2 complete - GetActualCost validation works with same patterns as US1

---

## Phase 5: User Story 3 - Debug Logging for Validation Failures (Priority: P3)

**Goal**: Ensure validation failures include full context for troubleshooting

**Note**: This story is integrated into US1 and US2 implementation (T007, T013). Verify logging works correctly.

### Verification for User Story 3

- [x] T016 [US3] Verify WARN-level log includes resource_type (projected) or resource_id (actual), trace_id, and error message
- [x] T017 [US3] Manual test: Verified log output format matches quickstart.md examples (see test output)

**Checkpoint**: User Story 3 complete - Logging provides full context for troubleshooting

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Quality gates and documentation

- [x] T018 Run `make lint` and fix any issues
- [x] T019 Run `make test` and verify 80%+ coverage for modified code
- [x] T020 [P] Update `CLAUDE.md` with validation pattern documentation under "Key Patterns" section
- [x] T021 Run quickstart.md validation: unit tests verify validation behavior as documented

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - import statement only
- **Foundational (Phase 2)**: N/A - using existing infrastructure
- **User Story 1 (Phase 3)**: Depends on Phase 1 (import)
- **User Story 2 (Phase 4)**: Can proceed in parallel with US1 after Phase 1, but may share test patterns
- **User Story 3 (Phase 5)**: Verification only - logging integrated into US1/US2
- **Polish (Phase 6)**: Depends on all user stories complete

### Within Each User Story

- Tests (T002-T005, T010-T011) MUST be written and FAIL before implementation
- Implementation (T006-T008, T012-T014) follows tests
- Verification (T009, T015) confirms tests pass

### Parallel Opportunities

- T002, T003, T004, T005 can all run in parallel (different test functions, same file)
- T010, T011 can run in parallel
- US1 and US2 can be developed in parallel after Phase 1

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (add import)
2. Write tests T002-T005 (should fail)
3. Implement T006-T008
4. Run T009 verification
5. **STOP and VALIDATE**: Test US1 independently with sample plan

### Full Implementation

1. Complete US1 as MVP
2. Add US2 tests (T010-T011) and implementation (T012-T014)
3. Verify US3 logging (T016-T017)
4. Run quality gates (T018-T021)

---

## Notes

- [P] tasks = different test functions, no dependencies
- [Story] label maps task to specific user story for traceability
- Tests use existing mock client patterns from adapter_test.go
- Validation errors use "VALIDATION:" prefix per design decision D4 in plan.md
- No new files created - all changes to existing adapter.go and adapter_test.go
