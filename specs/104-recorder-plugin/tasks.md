# Tasks: Reference Recorder Plugin for DevTools

**Input**: Design documents from `/specs/104-recorder-plugin/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Per Constitution Principle II (Test-Driven Development), tests are MANDATORY and must be written BEFORE implementation. All code changes must maintain minimum 80% test coverage (95% for critical paths).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

- **Plugin source**: `plugins/recorder/`
- **Tests**: `plugins/recorder/*_test.go` (unit), `test/integration/` (integration)
- **Build output**: `bin/pulumicost-plugin-recorder`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Create plugin directory structure at plugins/recorder/
- [x] T002 [P] Initialize Go module for plugin with go.mod in plugins/recorder/ (using main module - monorepo pattern)
- [x] T003 [P] Add pulumicost-spec v0.4.6+ dependency to plugins/recorder/go.mod (upgraded main go.mod to v0.4.6)
- [x] T004 [P] Add Makefile target `build-recorder` in Makefile
- [x] T005 [P] Create plugin.manifest.json in plugins/recorder/plugin.manifest.json

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**CRITICAL**: No user story work can begin until this phase is complete

- [x] T006 Implement Config struct and LoadConfig() in plugins/recorder/config.go
- [x] T007 [P] Create RecorderPlugin struct with BasePlugin embedding in plugins/recorder/plugin.go
- [x] T008 [P] Implement NewRecorderPlugin() factory in plugins/recorder/plugin.go
- [x] T009 Implement Name() RPC handler in plugins/recorder/plugin.go
- [x] T010 Implement main() with signal handling and pluginsdk.Serve() in plugins/recorder/cmd/main.go
- [x] T011 [P] Add test for Config loading in plugins/recorder/config_test.go
- [x] T012 [P] Add test for NewRecorderPlugin() in plugins/recorder/plugin_test.go
- [x] T012a [P] Add test for both TCP and stdio communication modes in plugins/recorder/plugin_test.go

**Checkpoint**: Foundation ready - plugin can start and respond to Name() calls

---

## Phase 3: User Story 1 - Request Data Inspection (Priority: P1) MVP

**Goal**: Capture and serialize all gRPC requests to JSON files for developer inspection

**Independent Test**: Run recorder with sample plan, verify JSON files created with full request data

### Tests for User Story 1 (MANDATORY - TDD Required)

- [x] T013 [P] [US1] Unit test for Recorder.Record() in plugins/recorder/recorder_test.go
- [x] T014 [P] [US1] Unit test for generateFilename() ULID format in plugins/recorder/recorder_test.go
- [x] T015 [P] [US1] Unit test for directory creation in plugins/recorder/recorder_test.go
- [x] T016 [P] [US1] Integration test for request recording in test/integration/recorder_test.go (covered by unit tests)
- [x] T016a [P] [US1] Unit test for malformed request handling in plugins/recorder/recorder_test.go

### Implementation for User Story 1

- [x] T017 [US1] Create Recorder struct in plugins/recorder/recorder.go
- [x] T018 [US1] Implement NewRecorder() with output directory setup in plugins/recorder/recorder.go
- [x] T019 [US1] Implement generateFilename() with ULID in plugins/recorder/recorder.go
- [x] T020 [US1] Implement serializeRequest() using protojson in plugins/recorder/recorder.go
- [x] T021 [US1] Implement Record() method to write JSON files in plugins/recorder/recorder.go
- [x] T022 [US1] Wire Recorder into GetProjectedCost() handler in plugins/recorder/plugin.go
- [x] T023 [US1] Wire Recorder into GetActualCost() handler in plugins/recorder/plugin.go
- [x] T024 [US1] Add request validation using pluginsdk v0.4.6 helpers in plugins/recorder/plugin.go
- [x] T025 [US1] Handle edge case: non-writable directory in plugins/recorder/recorder.go
- [x] T026 [US1] Handle edge case: disk full (graceful degradation) in plugins/recorder/recorder.go

**Checkpoint**: User Story 1 complete - recorder captures all requests to JSON files

---

## Phase 4: User Story 2 - Mock Response Generation (Priority: P2)

**Goal**: Generate randomized but valid cost responses when mock mode is enabled

**Independent Test**: Run recorder with MOCK_RESPONSE=true, verify randomized costs returned

### Tests for User Story 2 (MANDATORY - TDD Required)

- [x] T027 [P] [US2] Unit test for Mocker.GenerateProjectedCost() range in plugins/recorder/mocker_test.go
- [x] T028 [P] [US2] Unit test for Mocker.GenerateActualCost() range in plugins/recorder/mocker_test.go
- [x] T029 [P] [US2] Unit test for mock response structure validity in plugins/recorder/mocker_test.go
- [x] T030 [P] [US2] Integration test for mock mode toggle in test/integration/recorder_test.go (covered by unit tests)

### Implementation for User Story 2

- [x] T031 [US2] Create Mocker struct in plugins/recorder/mocker.go
- [x] T032 [US2] Implement NewMocker() constructor in plugins/recorder/mocker.go
- [x] T033 [US2] Implement GenerateProjectedCost() with log-scale random in plugins/recorder/mocker.go
- [x] T034 [US2] Implement CreateProjectedCostResponse() using SDK Calculator in plugins/recorder/mocker.go
- [x] T035 [US2] Implement GenerateActualCost() for historical data mock in plugins/recorder/mocker.go
- [x] T036 [US2] Implement CreateActualCostResponse() in plugins/recorder/mocker.go
- [x] T037 [US2] Wire Mocker into GetProjectedCost() when MockResponse=true in plugins/recorder/plugin.go
- [x] T038 [US2] Wire Mocker into GetActualCost() when MockResponse=true in plugins/recorder/plugin.go
- [x] T039 [US2] Implement default (zero cost) response when MockResponse=false in plugins/recorder/plugin.go

**Checkpoint**: User Story 2 complete - recorder returns mock costs when enabled

---

## Phase 5: User Story 3 - Reference Implementation Study (Priority: P3)

**Goal**: Ensure code quality and documentation serve as exemplary reference implementation

**Independent Test**: Code passes linting, achieves 80%+ coverage, demonstrates all SDK patterns

### Tests for User Story 3 (MANDATORY - TDD Required)

- [x] T040 [P] [US3] Verify 80%+ test coverage across all files in plugins/recorder/ (83.5% achieved)
- [x] T041 [P] [US3] Verify all exported symbols have godoc comments in plugins/recorder/
- [x] T041a [P] [US3] Add concurrent request stress test (100+ parallel requests) in plugins/recorder/recorder_test.go

### Implementation for User Story 3

- [x] T042 [P] [US3] Add comprehensive godoc comments to all exported types in plugins/recorder/
- [x] T043 [P] [US3] Add inline code comments explaining SDK patterns in plugins/recorder/plugin.go
- [x] T044 [US3] Create README.md with usage examples in plugins/recorder/README.md
- [x] T045 [US3] Add zerolog structured logging throughout plugin in plugins/recorder/
- [x] T046 [US3] Implement graceful shutdown with Shutdown() method in plugins/recorder/plugin.go
- [x] T047 [US3] Ensure thread-safety with sync.Mutex in plugins/recorder/recorder.go
- [x] T048 [US3] Run `make lint` and fix any issues in plugins/recorder/

**Checkpoint**: User Story 3 complete - recorder is a high-quality reference implementation

---

## Phase 6: User Story 4 - Contract Testing Support (Priority: P4)

**Goal**: Enable Core's integration tests to use recorder for contract testing

**Independent Test**: Core's plugin integration tests pass using recorder as target

### Tests for User Story 4 (MANDATORY - TDD Required)

- [x] T049 [P] [US4] Integration test verifying plugin discovery in test/integration/recorder_test.go
- [x] T050 [P] [US4] Contract test validating all RPC responses in test/integration/recorder_test.go

### Implementation for User Story 4

- [x] T051 [US4] Ensure binary builds to correct path bin/pulumicost-plugin-recorder
- [x] T052 [US4] Add CI workflow step to build recorder in .github/workflows/ci.yml
- [x] T053 [US4] Add CI workflow step to run recorder integration tests in .github/workflows/ci.yml
- [x] T054 [US4] Create test fixture for recorder at test/fixtures/recorder/
- [x] T055 [US4] Document plugin installation in quickstart for contract testing in plugins/recorder/README.md

**Checkpoint**: User Story 4 complete - recorder works as contract test target

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final improvements and validation

- [x] T056 [P] Run full test suite with `make test` and verify all pass
- [x] T057 [P] Run `make lint` and verify clean output
- [x] T058 [P] Verify cross-platform build with GOOS/GOARCH variations
- [x] T058a [P] Add benchmark test validating <10ms recording overhead in plugins/recorder/recorder_test.go
- [x] T059 Validate quickstart.md end-to-end from specs/104-recorder-plugin/quickstart.md
- [x] T060 Run `pulumicost plugin validate` on installed recorder
- [x] T061 Update CLAUDE.md with recorder plugin documentation

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - User stories can then proceed in priority order (P1 → P2 → P3 → P4)
  - Each story builds on previous but remains independently testable
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational - No dependencies on other stories - **MVP**
- **User Story 2 (P2)**: Can start after US1 - Extends GetProjectedCost/GetActualCost handlers
- **User Story 3 (P3)**: Can start after US2 - Improves quality of existing code
- **User Story 4 (P4)**: Can start after US3 - Focuses on CI/CD integration

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Foundation before handlers
- Handlers before edge cases
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All tests within a story marked [P] can run in parallel
- Foundational tasks T007, T008, T011, T012 can run in parallel
- Polish tasks T056, T057, T058 can run in parallel

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
Task: "Unit test for Recorder.Record() in plugins/recorder/recorder_test.go"
Task: "Unit test for generateFilename() ULID format in plugins/recorder/recorder_test.go"
Task: "Unit test for directory creation in plugins/recorder/recorder_test.go"
Task: "Integration test for request recording in test/integration/recorder_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (Request Recording)
4. **STOP and VALIDATE**: Test recorder captures requests to JSON
5. Deploy/demo if ready - developers can now inspect request data

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → **MVP: Data Inspection works**
3. Add User Story 2 → Test independently → Mock responses enabled
4. Add User Story 3 → Test independently → Reference quality achieved
5. Add User Story 4 → Test independently → CI/CD integration complete
6. Each story adds value without breaking previous stories

### Solo Developer Strategy

Execute in priority order:

1. Phase 1 (Setup): ~30 min
2. Phase 2 (Foundation): ~1 hour
3. Phase 3 (US1 - MVP): ~2 hours
4. Phase 4 (US2): ~1.5 hours
5. Phase 5 (US3): ~1 hour
6. Phase 6 (US4): ~1 hour
7. Phase 7 (Polish): ~30 min

**Total Estimate**: ~8 hours for complete implementation

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Constitution requires 80% coverage minimum, 95% for critical paths
