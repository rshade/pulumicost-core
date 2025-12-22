# Tasks: Plugin Ecosystem Maturity

**Input**: Design documents from `/specs/016-plugin-ecosystem-maturity/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Per Constitution Principle II (Test-Driven Development), tests are
MANDATORY and must be written BEFORE implementation. All code changes must
maintain minimum 80% test coverage (95% for critical paths).

**Organization**: Tasks are grouped by user story to enable independent
implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- Go project with standard `internal/` structure
- Tests alongside implementation (`*_test.go`)
- E2E tests in `test/e2e/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Package initialization and shared type definitions

- [X] T001 Update package documentation in internal/conformance/doc.go
- [X] T002 [P] Create shared types (Status, Category, Verbosity, CommMode) in internal/conformance/types.go
- [X] T003 [P] Create test fixtures directory structure at test/e2e/aws/fixtures/

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story
can be implemented

**CRITICAL**: No user story work can begin until this phase is complete

- [X] T004 Implement ConformanceTestCase type in internal/conformance/types.go
- [X] T005 [P] Implement TestResult type in internal/conformance/types.go
- [X] T006 [P] Implement PluginUnderTest type in internal/conformance/types.go
- [X] T007 [P] Implement ConformanceSuiteConfig type in internal/conformance/types.go
- [X] T008 [P] Implement SuiteReport type with Summary in internal/conformance/types.go
- [X] T009 Create reference plugin implementation for testing in test/mocks/plugin/conformance.go
- [X] T010 Write unit tests for all types in internal/conformance/types_test.go

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Plugin Developer Validates Implementation (P1) MVP

**Goal**: Plugin developers can run conformance suite against their plugin
binary to verify protocol compliance

**Independent Test**: Run `pulumicost plugin conformance ./test-plugin` and
receive pass/fail report for each protocol requirement

### Tests for User Story 1 (MANDATORY - TDD Required)

- [X] T011 [P] [US1] Write test for Suite.Run() in internal/conformance/suite_test.go
- [X] T012 [P] [US1] Write test for TestRunner in internal/conformance/runner_test.go
- [X] T013 [P] [US1] Write test for protocol version checking in internal/conformance/version_test.go
- [X] T014 [P] [US1] Write test for table output in internal/conformance/reporter_test.go

### Implementation for User Story 1

- [X] T015 [US1] Implement version checking logic in internal/conformance/version.go (FR-008a)
- [X] T016 [US1] Implement TestRunner that executes individual tests in internal/conformance/runner.go
- [X] T017 [P] [US1] Implement protocol tests (Name RPC) in internal/conformance/protocol.go (FR-001, FR-004)
- [X] T018 [P] [US1] Implement context cancellation tests in internal/conformance/context.go (FR-002)
- [X] T019 [P] [US1] Implement error code tests in internal/conformance/cost.go (FR-003)
- [X] T020 [P] [US1] Implement timeout tests in internal/conformance/context.go (FR-005)
- [X] T021 [P] [US1] Implement batch handling tests in internal/conformance/batch.go (FR-006)
- [X] T022 [US1] Implement ConformanceSuite orchestrator in internal/conformance/suite.go (FR-001)
- [X] T023 [US1] Implement crash handling with plugin restart in internal/conformance/runner.go (FR-020)
- [X] T024 [US1] Implement table output reporter in internal/conformance/reporter.go (FR-007, FR-017)
- [X] T025 [US1] Add verbosity support to TestRunner in internal/conformance/runner.go (FR-019)
- [X] T026 [US1] Add TCP and stdio mode support in internal/conformance/suite.go (FR-008)
- [X] T027 [US1] Implement CLI command `plugin conformance` in internal/cli/plugin_conformance.go
- [X] T028 [US1] Write CLI command tests in internal/cli/plugin_conformance_test.go

**Checkpoint**: Plugin developers can validate their implementation via CLI

---

## Phase 4: User Story 2 - Core Developer Ensures Protocol Stability (P2)

**Goal**: Conformance tests integrated into CI to detect breaking protocol
changes automatically

**Independent Test**: Run conformance suite against reference plugin when
protocol definitions change; CI fails on breaking changes

**Dependencies**: T032-T033 require T024 (base reporter) from US1

### Tests for User Story 2 (MANDATORY - TDD Required)

- [ ] T029 [P] [US2] Write test for JUnit XML output in internal/conformance/reporter_test.go
- [ ] T030 [P] [US2] Write test for JSON output in internal/conformance/reporter_test.go
- [ ] T031 [P] [US2] Write test for category filtering in internal/conformance/suite_test.go

### Implementation for User Story 2

- [ ] T032 [US2] Implement JUnit XML reporter in internal/conformance/reporter.go (FR-016)
- [ ] T033 [US2] Implement JSON reporter in internal/conformance/reporter.go (FR-016)
- [ ] T034 [US2] Implement category filtering (protocol, performance, error) in internal/conformance/suite.go (FR-018)
- [ ] T035 [US2] Implement test name regex filtering in internal/conformance/suite.go (FR-018)
- [ ] T036 [US2] Add --output-file flag support to CLI in internal/cli/plugin_conformance.go
- [ ] T037 [US2] Create GitHub Actions workflow for conformance in .github/workflows/conformance.yml (FR-015)
- [ ] T038 [US2] Document CI integration in docs/plugins/conformance-ci.md

**Checkpoint**: Conformance tests run in CI on every commit

---

## Phase 5: User Story 3 - QA Engineer Validates Real-World Cost Data (P3)

**Goal**: E2E tests validate plugins against real cloud provider APIs with
expected cost tolerances

**Independent Test**: Run E2E tests against test AWS account; verify costs
within 5% tolerance of actual billing

### Tests for User Story 3 (MANDATORY - TDD Required)

- [ ] T039 [P] [US3] Write test for E2ETestConfig in test/e2e/config_test.go
- [ ] T040 [P] [US3] Write test for credential skip logic in test/e2e/skip_test.go

### Implementation for User Story 3

- [ ] T041 [US3] Implement E2ETestConfig type in test/e2e/config.go (FR-009, FR-010)
- [ ] T042 [US3] Implement credential loading from environment in test/e2e/credentials.go (FR-010)
- [ ] T043 [US3] Implement graceful skip when credentials missing in test/e2e/skip.go (FR-011)
- [ ] T044 [US3] Create AWS E2E test framework in test/e2e/aws/cost_test.go (FR-009)
- [ ] T045 [US3] Create expected cost fixtures in test/e2e/aws/fixtures/expected_costs.json (FR-012)
- [ ] T046 [US3] Implement cost tolerance comparison (5%) in test/e2e/aws/validation.go (FR-013)
- [ ] T047 [US3] Document test account setup in docs/plugins/e2e-setup.md (FR-014)

**Checkpoint**: QA can run E2E tests against AWS test account

---

## Phase 6: User Story 4 - Enterprise Admin Certifies Plugin Compatibility (P4)

**Goal**: Enterprise admins can generate certification reports for third-party
plugins before production deployment

**Independent Test**: Run certification command; receive certification report
with compatibility status and any issues

### Tests for User Story 4 (MANDATORY - TDD Required)

- [ ] T048 [P] [US4] Write test for certification report generation in internal/conformance/certification_test.go
- [ ] T049 [P] [US4] Write test for CLI certification command in internal/cli/plugin_certify_test.go

### Implementation for User Story 4

- [ ] T050 [US4] Implement CertificationReport type in internal/conformance/certification.go
- [ ] T051 [US4] Implement certification logic (all tests must pass) in internal/conformance/certification.go
- [ ] T052 [US4] Implement certification report formatter in internal/conformance/certification.go
- [ ] T053 [US4] Implement CLI command `plugin certify` in internal/cli/plugin_certify.go
- [ ] T054 [US4] Document certification process in docs/plugins/certification.md

**Checkpoint**: Enterprise admins can certify plugins before deployment

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T055 [P] Update quickstart guide with actual CLI examples in specs/016-plugin-ecosystem-maturity/quickstart.md
- [ ] T056 [P] Add docstrings to all exported functions per Constitution (80% coverage)
- [ ] T057 Run `make lint` and fix any issues
- [ ] T058 Run `make test` and ensure 80% coverage
- [ ] T059 [P] Create plugin developer documentation in docs/plugins/plugin-development.md
- [ ] T060 [P] Add troubleshooting guide to docs/plugins/troubleshooting.md

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - User stories can proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 -> P2 -> P3 -> P4)
- **Polish (Phase 7)**: Depends on desired user stories being complete

### User Story Dependencies

- **US1 (P1)**: Can start after Foundational - No dependencies on other stories
- **US2 (P2)**: Tests (T029-T031) can start after Foundational; Implementation (T032-T038) requires T024 from US1
- **US3 (P3)**: Can start after Foundational - Independent of US1/US2
- **US4 (P4)**: Depends on US1 conformance suite being complete

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Types before implementations
- Core logic before CLI integration
- Implementation before documentation

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (T005-T008)
- Once Foundational completes, US1, US2, US3 can start in parallel
- US4 must wait for US1 core completion
- All test files marked [P] can run in parallel within their story
- All conformance test implementations (T017-T021) can run in parallel

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
Task: "Write test for Suite.Run() in internal/conformance/suite_test.go"
Task: "Write test for TestRunner in internal/conformance/runner_test.go"
Task: "Write test for protocol version checking in internal/conformance/version_test.go"
Task: "Write test for table output in internal/conformance/reporter_test.go"

# Launch all conformance test implementations together:
Task: "Implement protocol tests in internal/conformance/protocol.go"
Task: "Implement context cancellation tests in internal/conformance/context.go"
Task: "Implement error code tests in internal/conformance/cost.go"
Task: "Implement timeout tests in internal/conformance/context.go"
Task: "Implement batch handling tests in internal/conformance/batch.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Run conformance against reference plugin
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational -> Foundation ready
2. Add US1 -> Test independently -> Plugin developers can validate (MVP!)
3. Add US2 -> Test independently -> CI integration works
4. Add US3 -> Test independently -> E2E testing available
5. Add US4 -> Test independently -> Certification available
6. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (P1)
   - Developer B: User Story 3 (P3) - independent
3. After US1 core:
   - Developer A: User Story 2 (P2)
   - Developer C: User Story 4 (P4)
4. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Run `make lint` and `make test` before claiming any task complete
