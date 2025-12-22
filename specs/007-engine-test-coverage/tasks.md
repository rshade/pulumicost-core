# Tasks: Engine Test Coverage Completion

**Input**: Design documents from `/specs/007-engine-test-coverage/`
**Prerequisites**: plan.md, spec.md, research.md, quickstart.md

**Tests**: This feature IS about tests. All tasks create or extend test files.
Tests must follow anti-slop constraints from spec.md.

**Organization**: Tasks grouped by user story for independent implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

Following existing project structure:

- Unit tests: `test/unit/engine/`
- Benchmarks: `test/benchmarks/`
- Internal tests: `internal/engine/` (colocated)

---

## Phase 1: Setup (Analysis & Baseline)

**Purpose**: Establish current coverage baseline and identify specific gaps

- [ ] T001 Run coverage analysis and document baseline in coverage-baseline.txt
- [ ] T002 [P] Identify uncovered lines in tryStoragePricing (internal/engine/engine.go:845-856)
- [ ] T003 [P] Identify uncovered lines in getDefaultMonthlyByType (internal/engine/engine.go:879-891)
- [ ] T004 [P] Identify uncovered lines in parseFloatValue (internal/engine/engine.go:914-928)
- [ ] T005 [P] Identify uncovered lines in distributeDailyCosts (internal/engine/engine.go:1490-1509)

---

## Phase 2: Foundational (Test File Setup)

**Purpose**: Create test file structure following existing conventions

**CRITICAL**: No test implementation until files are created

- [ ] T006 Create internal/engine/pricing_test.go with package declaration and imports (colocated)
- [ ] T007 [P] Create internal/engine/pricing_test.go includes conversion tests (colocated with pricing)
- [ ] T008 [P] Create internal/engine/distribution_test.go with package declaration and imports (colocated)

**Checkpoint**: Test files exist with proper structure (implemented as colocated tests)

---

## Phase 3: User Story 1 - Core Calculation Reliability (Priority: P1)

**Goal**: Achieve 80% engine coverage with meaningful error type tests

**Independent Test**: `go test ./internal/engine/... -coverprofile=coverage.out && go tool cover -func=coverage.out | grep total`

### Implementation for User Story 1

- [ ] T009 [US1] Add TestParseFloatValue table-driven test in internal/engine/pricing_test.go
- [ ] T010 [US1] Add TestGetDefaultMonthlyByType table-driven test in internal/engine/pricing_test.go
- [ ] T011 [US1] Add TestTryStoragePricing tests in internal/engine/pricing_test.go
- [ ] T012 [P] [US1] Add ErrEmptyResults test case in internal/engine/types_test.go (TestErrorTypes)
- [ ] T013 [P] [US1] Add ErrMixedCurrencies test case in internal/engine/types_test.go (TestErrorTypes)
- [ ] T014 [P] [US1] Add ErrInvalidDateRange test case in internal/engine/types_test.go (TestErrorTypes)
- [ ] T015 [P] [US1] Add ErrInvalidGroupBy test case in internal/engine/types_test.go (TestErrorTypes)
- [ ] T016 [P] [US1] Add ErrNoCostData test case in internal/engine/types_test.go (TestErrorTypes)
- [ ] T017 [P] [US1] Add TestAggregation_ZeroCostsNoDivideByZero in internal/engine/aggregation_test.go
- [ ] T018 [P] [US1] Add TestAggregation_SingleResultUnchanged in internal/engine/aggregation_test.go
- [ ] T019 [US1] Run coverage check and verify >=80% for internal/engine/ (85.1% achieved)

**Checkpoint**: Coverage target met with all error types tested ✅

---

## Phase 4: User Story 2 - Edge Case Handling (Priority: P2)

**Goal**: Cover edge cases without adding slop tests

**Independent Test**: `go test ./test/unit/engine/... -run TestEdgeCase -v`

### Implementation for User Story 2

- [ ] T020 [US2] Add TestDistributeDailyCosts_DailyGrouping in internal/engine/distribution_test.go
- [ ] T021 [P] [US2] Add TestDistributeDailyCosts_MonthlyGrouping in internal/engine/distribution_test.go
- [ ] T022 [P] [US2] Add TestDistributeDailyCosts_CrossMonthBoundary in internal/engine/distribution_test.go
- [ ] T023 [P] [US2] Add TestDistributeDailyCosts_EmptyDailyCosts in internal/engine/distribution_test.go
- [ ] T024 [US2] Add TestCreateCrossProviderAggregation_EmptyCurrencyDefaultsToUSD in internal/engine/aggregation_test.go
- [ ] T025 [P] [US2] Add TestEdgeCase_NilPropertiesNoNilPointerPanic in internal/engine/aggregation_test.go
- [ ] T026 [P] [US2] Add TestEdgeCase_UnknownProviderReturnsUnknown in internal/engine/aggregation_test.go
- [ ] T027 [P] [US2] Add TestEdgeCase_LargeValuesNoOverflow in internal/engine/aggregation_test.go

**Checkpoint**: All edge cases from spec covered with meaningful tests ✅

---

## Phase 5: User Story 3 - Performance at Scale (Priority: P3)

**Goal**: Extend benchmarks to 1K, 10K, 100K resource scale

**Independent Test**: `go test ./test/benchmarks/... -bench=. -benchmem`

### Implementation for User Story 3

- [ ] T028 [US3] Add BenchmarkEngine_GetProjectedCost_1K in test/benchmarks/engine_bench_test.go
- [ ] T029 [P] [US3] Add BenchmarkEngine_GetProjectedCost_10K in test/benchmarks/engine_bench_test.go
- [ ] T030 [P] [US3] Add BenchmarkEngine_GetProjectedCost_100K in test/benchmarks/engine_bench_test.go
- [ ] T031 [US3] Add BenchmarkEngine_GetActualCost_1K in test/benchmarks/engine_bench_test.go
- [ ] T032 [P] [US3] Add BenchmarkEngine_GetActualCost_10K in test/benchmarks/engine_bench_test.go
- [ ] T033 [P] [US3] Add BenchmarkEngine_CrossProviderAggregation_1K in test/benchmarks/engine_bench_test.go
- [ ] T034 [P] [US3] Add BenchmarkEngine_CrossProviderAggregation_10K in test/benchmarks/engine_bench_test.go
- [ ] T035 [US3] Document benchmark baselines in specs/007-engine-test-coverage/benchmark-baseline.md

**Checkpoint**: Benchmarks cover all scale tiers with documented baselines ✅

---

## Phase 6: User Story 4 - Integration Test Completeness (Priority: P2)

**Goal**: Verify engine components work together correctly

**Independent Test**: `go test ./test/integration/... -v`

### Implementation for User Story 4

- [ ] T036 [US4] Add TestIntegration_PluginFallbackToSpec in test/integration/plugin/plugin_communication_test.go
- [ ] T037 [P] [US4] Add TestIntegration_TagFilterProcessing in test/integration/plugin/plugin_communication_test.go
- [ ] T038 [P] [US4] Add TestIntegration_DailyCostsAggregation in test/integration/plugin/plugin_communication_test.go
- [ ] T039 [US4] Add TestIntegration_MultiPluginCostCollection in test/integration/plugin/plugin_communication_test.go

**Checkpoint**: Integration scenarios from acceptance criteria covered ✅

---

## Phase 7: Polish & Validation

**Purpose**: Final quality checks and documentation

- [ ] T040 Run make lint and fix any linting errors (0 issues)
- [ ] T041 Run make test and verify all tests pass with -race flag
- [ ] T042 [P] Verify test suite completes within 60 seconds (avg 7-10s)
- [ ] T043 [P] Run tests 5 times consecutively to check for flakiness (all passed)
- [ ] T044 Review each new test for anti-slop compliance (distinct purpose, clear assertions)
- [ ] T045 Update specs/007-engine-test-coverage/checklists/requirements.md with final status

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - establishes baseline
- **Foundational (Phase 2)**: Depends on Setup - creates file structure
- **User Stories (Phase 3-6)**: All depend on Foundational completion
  - US1 (P1): Core coverage - do first
  - US2 (P2): Edge cases - can parallel with US4
  - US3 (P3): Benchmarks - independent, lower priority
  - US4 (P2): Integration - can parallel with US2
- **Polish (Phase 7)**: Depends on all user stories complete

### User Story Dependencies

- **User Story 1 (P1)**: Must complete first - establishes coverage baseline
- **User Story 2 (P2)**: Can start after US1 - uses same test patterns
- **User Story 3 (P3)**: Independent - benchmarks don't affect unit tests
- **User Story 4 (P2)**: Can start after US1 - integration tests are separate

### Within Each User Story

- Setup tests (file structure) before implementation
- Table-driven tests grouped logically
- Run coverage check after each story completes

### Parallel Opportunities

**Phase 2 (Foundational)**:

- T006, T007, T008 can run in parallel (different files)

**User Story 1**:

- T012-T018 can run in parallel (error types + aggregation tests)

**User Story 2**:

- T021-T023 can run in parallel (distribution test cases)
- T025-T027 can run in parallel (edge case tests)

**User Story 3**:

- T029-T030 can run in parallel (projected cost benchmarks)
- T033-T034 can run in parallel (aggregation benchmarks)

**User Story 4**:

- T037-T038 can run in parallel (integration scenarios)

---

## Parallel Example: User Story 1 Error Tests

```bash
# Launch all error type tests in parallel:
Task: "Add ErrEmptyResults test case in test/unit/engine/errors_test.go"
Task: "Add ErrMixedCurrencies test case in test/unit/engine/errors_test.go"
Task: "Add ErrInvalidDateRange test case in test/unit/engine/errors_test.go"
Task: "Add ErrInvalidGroupBy test case in test/unit/engine/errors_test.go"
Task: "Add ErrNoCostData test case in test/unit/engine/errors_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (baseline analysis)
2. Complete Phase 2: Foundational (file structure)
3. Complete Phase 3: User Story 1 (core coverage)
4. **STOP and VALIDATE**: Run `go tool cover -func=coverage.out | grep total`
5. If >=80% achieved, MVP complete

### Incremental Delivery

1. Complete Setup + Foundational → Files ready
2. Add User Story 1 → Test coverage >=80% (MVP!)
3. Add User Story 2 → Edge cases covered
4. Add User Story 3 → Benchmarks established
5. Add User Story 4 → Integration verified
6. Polish → Quality gates passed

### Quality Gate Checkpoints

After each user story:

1. `go test ./internal/engine/... -race` passes
2. Coverage threshold maintained
3. No test flakiness
4. Anti-slop review passed

---

## Notes

- [P] tasks = different files or independent test cases
- [Story] label maps task to specific user story
- Each test MUST have comment explaining its distinct purpose
- Anti-slop: No redundant cases, no unused struct fields, no generic names
- Coverage target is secondary to test quality
- Avoid: slop tests, helper functions that add complexity, tests just for coverage
