# Tasks: Error Aggregation in Proto Adapter

**Input**: Design documents from `/specs/003-proto-error-aggregation/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Per Constitution Principle II (Test-Driven Development), tests are MANDATORY and must be written BEFORE implementation. All code changes must maintain minimum 80% test coverage (95% for critical paths).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Go project**: `internal/` packages at repository root
- **Tests**: Co-located with source files (*_test.go)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and dependency setup

- [x] T001 Add zerolog dependency with `go get github.com/rs/zerolog`
- [x] T002 Run `go mod tidy` to update go.sum

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core type tests that MUST pass before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Tests for Foundational Types (MANDATORY - TDD Required) ⚠️

> **CONSTITUTION REQUIREMENT: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T003 [P] Unit test for ErrorDetail struct creation in internal/proto/adapter_test.go
- [x] T004 [P] Unit test for CostResultWithErrors struct creation in internal/proto/adapter_test.go
- [x] T005 [P] Unit test for HasErrors() method in internal/proto/adapter_test.go
- [x] T006 [P] Unit test for ErrorSummary() output format in internal/proto/adapter_test.go

### Implementation for Foundational Types

- [x] T007 Create ErrorDetail struct in internal/proto/adapter.go
- [x] T008 Create CostResultWithErrors struct in internal/proto/adapter.go
- [x] T009 Implement HasErrors() method on CostResultWithErrors in internal/proto/adapter.go
- [x] T010 Implement ErrorSummary() method with truncation after 5 errors in internal/proto/adapter.go

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Error Reporting for Projected Costs (Priority: P1) 🎯 MVP

**Goal**: When a plugin fails to calculate projected cost for resources, the system returns results for successful resources and reports failed resources with error details.

**Independent Test**: Call GetProjectedCost with a mock client that fails for specific resources. Verify Results contains all resources (placeholders for failures) and Errors contains details for failed resources.

### Tests for User Story 1 (MANDATORY - TDD Required) ⚠️

> **CONSTITUTION REQUIREMENT: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T011 [P] [US1] Unit test for GetProjectedCost error tracking in internal/proto/adapter_test.go
- [x] T012 [P] [US1] Integration test for CLI error summary display after table in internal/cli/cost_projected_test.go
- [x] T013 [P] [US1] Unit test for Engine.GetProjectedCost return type in internal/engine/engine_test.go

### Implementation for User Story 1

- [x] T014 [US1] Update Adapter.GetProjectedCost to return *CostResultWithErrors in internal/proto/adapter.go
- [x] T015 [US1] Add error tracking loop in GetProjectedCost for plugin failures in internal/proto/adapter.go
- [x] T016 [US1] Add placeholder CostResult for failed resources with ERROR: prefix in Notes in internal/proto/adapter.go
- [x] T017 [US1] Update Engine.GetProjectedCost to return *CostResultWithErrors in internal/engine/engine.go
- [x] T018 [US1] Update CLI cost_projected command to handle CostResultWithErrors in internal/cli/cost_projected.go
- [x] T019 [US1] Add error summary display after table in cost_projected command in internal/cli/cost_projected.go

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Error Reporting for Actual Costs (Priority: P1)

**Goal**: Similar to projected costs, when a plugin fails to calculate actual cost for resources, the system provides results for successful ones and detailed error information for failures.

**Independent Test**: Call GetActualCost with a mock client that fails for specific resources. Verify Results contains all resources and Errors contains failure details.

### Tests for User Story 2 (MANDATORY - TDD Required) ⚠️

- [ ] T020 [P] [US2] Unit test for GetActualCost error tracking in internal/proto/adapter_test.go
- [ ] T021 [P] [US2] Integration test for CLI error summary display after table in internal/cli/cost_actual_test.go
- [ ] T022 [P] [US2] Unit test for Engine.GetActualCost return type in internal/engine/engine_test.go

### Implementation for User Story 2

- [ ] T023 [US2] Update Adapter.GetActualCost to return *CostResultWithErrors in internal/proto/adapter.go
- [ ] T024 [US2] Add error tracking loop in GetActualCost for plugin failures in internal/proto/adapter.go
- [ ] T025 [US2] Add placeholder CostResult for failed resources in GetActualCost in internal/proto/adapter.go
- [ ] T026 [US2] Update Engine.GetActualCost to return *CostResultWithErrors in internal/engine/engine.go
- [ ] T027 [US2] Update CLI cost_actual command to handle CostResultWithErrors in internal/cli/cost_actual.go
- [ ] T028 [US2] Add error summary display after table in cost_actual command in internal/cli/cost_actual.go

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Comprehensive Error Summary and Logging (Priority: P2)

**Goal**: Provide aggregated error summary that truncates long lists for readability and log detailed error information with structured logging for debugging.

**Independent Test**: Generate scenario with more than 5 errors and verify ErrorSummary() truncates output. Verify zerolog outputs contain expected structured fields.

### Tests for User Story 3 (MANDATORY - TDD Required) ⚠️

- [ ] T029 [P] [US3] Unit test for ErrorSummary with >5 errors truncation in internal/proto/adapter_test.go
- [ ] T030 [P] [US3] Unit test for ErrorSummary with exactly 5 errors in internal/proto/adapter_test.go
- [ ] T031 [P] [US3] Unit test for ErrorSummary with 0 errors in internal/proto/adapter_test.go

### Implementation for User Story 3

- [ ] T032 [US3] Add zerolog structured logging for errors in Engine.GetProjectedCost in internal/engine/engine.go
- [ ] T033 [US3] Add zerolog structured logging for errors in Engine.GetActualCost in internal/engine/engine.go
- [ ] T034 [US3] Verify ErrorSummary truncation logic handles edge cases in internal/proto/adapter.go

**Checkpoint**: All user stories should now be independently functional

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and documentation

- [ ] T035 Run `make test` and verify 80%+ coverage
- [ ] T036 Run `make lint` and fix any issues
- [ ] T037 Update CLAUDE.md with error aggregation documentation
- [ ] T038 Manual testing with example plans per quickstart.md
- [ ] T039 Verify all acceptance criteria from spec.md are met

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - US1 and US2 can proceed in parallel (if staffed)
  - US3 depends on US1 and US2 (needs error aggregation working)
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational (Phase 2) - Same priority as US1, can run in parallel
- **User Story 3 (P2)**: Depends on US1 and US2 - needs error aggregation working to test logging

### Within Each Phase

- Tests MUST be written and FAIL before implementation
- Adapter changes before engine changes
- Engine changes before CLI changes
- Core implementation before integration

### Parallel Opportunities

- T003, T004, T005, T006 (Foundational tests) can run in parallel
- T011, T012, T013 (US1 tests) can run in parallel
- T020, T021, T022 (US2 tests) can run in parallel
- T029, T030, T031 (US3 tests) can run in parallel
- US1 and US2 implementation can run in parallel after foundational phase

---

## Parallel Example: Foundational Phase

```bash
# Launch all tests for Foundational types together:
Task: "Unit test for ErrorDetail struct creation in internal/proto/adapter_test.go"
Task: "Unit test for CostResultWithErrors struct creation in internal/proto/adapter_test.go"
Task: "Unit test for HasErrors() method in internal/proto/adapter_test.go"
Task: "Unit test for ErrorSummary() output format in internal/proto/adapter_test.go"
```

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
Task: "Unit test for GetProjectedCost error tracking in internal/proto/adapter_test.go"
Task: "Integration test for CLI error summary display after table in internal/cli/cost_projected_test.go"
Task: "Unit test for Engine.GetProjectedCost return type in internal/engine/engine_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T002)
2. Complete Phase 2: Foundational (T003-T010)
3. Complete Phase 3: User Story 1 (T011-T019)
4. **STOP and VALIDATE**: Test GetProjectedCost error aggregation independently
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo
4. Add User Story 3 → Test independently → Deploy/Demo
5. Each story adds value without breaking previous stories

### Single Developer Strategy

1. Complete Setup (2 tasks)
2. Complete Foundational (8 tasks)
3. Complete US1 (9 tasks) → Validate
4. Complete US2 (9 tasks) → Validate
5. Complete US3 (6 tasks) → Validate
6. Complete Polish (5 tasks)

---

## Notes

- [P] tasks = different files/functions, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Total estimated effort: 2-3 hours (per original issue)
