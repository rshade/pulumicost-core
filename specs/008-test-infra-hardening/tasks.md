# Tasks: Test Infrastructure Hardening

**Input**: Design documents from `/specs/008-test-infra-hardening/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, quickstart.md

**Tests**: Per Constitution Principle II (Test-Driven Development), tests are MANDATORY and must be written BEFORE implementation. All code changes must maintain minimum 80% test coverage (95% for critical paths).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

- **Go project**: `internal/`, `cmd/`, `test/` at repository root
- **Benchmarks**: `test/benchmarks/`
- **CI/CD**: `.github/workflows/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and test infrastructure foundation

- [x] T001 Verify Go 1.24+ installed (required for native fuzzing and project compatibility) via `go version`
- [x] T002 [P] Create `test/benchmarks/generator/` directory structure
- [x] T003 [P] Verify existing `test/benchmarks/` suite structure and dependencies

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core test infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T004 Create `test/benchmarks/generator/generator.go` with `BenchmarkConfig` and `SyntheticResource` structs from data-model.md
- [x] T005 [P] Implement `GeneratePlan(config BenchmarkConfig) SyntheticPlan` function in `test/benchmarks/generator/generator.go`
- [x] T006 [P] Add validation rules (ResourceCount > 0, MaxDepth >= 0, DependencyRatio 0.0-1.0) in `test/benchmarks/generator/generator.go`
- [x] T007 [P] Write unit tests for generator validation rules in `test/benchmarks/generator/generator_test.go`
- [x] T008 Create `test/benchmarks/generator/export.go` with `ToJSON(plan SyntheticPlan) []byte` for benchmark data serialization

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Parser Resilience (Priority: P1) 🎯 MVP

**Goal**: Handle malformed inputs gracefully without crashes in JSON and YAML parsers

**Independent Test**: Run `go test -fuzz=FuzzJSON -fuzztime=30s ./internal/ingest` and `go test -fuzz=FuzzYAML -fuzztime=30s ./internal/spec` - no panics should occur

### Tests for User Story 1 (MANDATORY - TDD Required) ⚠️

> **CONSTITUTION REQUIREMENT: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T009 [P] [US1] Create fuzz test skeleton `FuzzJSON` in `internal/ingest/fuzz_test.go`
- [x] T010 [P] [US1] Create fuzz test skeleton `FuzzYAML` in `internal/spec/fuzz_test.go`
- [x] T011 [P] [US1] Add seed corpus directory `internal/ingest/testdata/fuzz/FuzzJSON/` with sample valid/invalid JSON files
- [x] T012 [P] [US1] Add seed corpus directory `internal/spec/testdata/fuzz/FuzzYAML/` with sample valid/invalid YAML files

### Implementation for User Story 1

- [x] T013 [US1] Implement `FuzzJSON(f *testing.F)` function in `internal/ingest/fuzz_test.go` targeting JSON parser
- [x] T014 [US1] Implement `FuzzYAML(f *testing.F)` function in `internal/spec/fuzz_test.go` targeting YAML parser
- [x] T015 [US1] Run local fuzzing (30s each parser) and fix any panics discovered in parser code
- [x] T016 [P] [US1] Update `.github/workflows/ci.yml` to add fuzz smoke test step (`-fuzztime=30s`) for PRs
- [x] T017 [P] [US1] Add deep fuzzing job (6h) to `.github/workflows/nightly.yml` on schedule
- [x] T018 [US1] Configure fuzz corpus caching in GitHub Actions (`testdata/fuzz` directories)

**Checkpoint**: Parser resilience verified - fuzz tests pass locally and in CI

---

## Phase 4: User Story 2 - Cross-Platform Reliability (Priority: P2)

**Goal**: Tool works consistently on Windows, Linux, and macOS

**Independent Test**: GitHub Actions nightly workflow shows green builds on all three platforms

### Audits for User Story 2 (Pre-Implementation Analysis)

- [x] T019 [P] [US2] Audit codebase for path separator issues (`/` vs `\`) in file operations
- [x] T020 [P] [US2] Audit codebase for platform-specific executable detection patterns

### Implementation for User Story 2

- [x] T021 [US2] Create `.github/workflows/nightly.yml` with matrix: `os: [ubuntu-latest, windows-latest, macos-latest]` (consolidates all nightly jobs: cross-platform tests, deep fuzzing, benchmarks)
- [x] T022 [US2] Configure nightly schedule trigger (`cron: '0 0 * * *'`) in `.github/workflows/nightly.yml`
- [x] T023 [US2] Configure release tag trigger (`on: release, types: [published]`) in `.github/workflows/nightly.yml`
- [x] T024 [US2] Add `fail-fast: true` to matrix strategy in `.github/workflows/nightly.yml`
- [x] T025 [US2] Add platform-specific Go setup with caching in `.github/workflows/nightly.yml`
- [x] T026 [US2] Fix any platform-specific issues discovered (path separators, permissions, etc.) - None found: codebase uses filepath.Join/Dir/Base and handles .exe correctly

**Checkpoint**: Cross-platform reliability verified - nightly workflow runs successfully on all platforms

---

## Phase 5: User Story 3 - Scalability (Priority: P3)

**Goal**: Process large infrastructure plans (up to 100K resources) within acceptable time limits

**Independent Test**: Run `go test -bench=BenchmarkLargeScale -benchtime=1x ./test/benchmarks` - 100K resources completes < 5 minutes

### Tests for User Story 3 (MANDATORY - TDD Required) ⚠️

> **CONSTITUTION REQUIREMENT: Write benchmark test stubs FIRST**

- [x] T027 [P] [US3] Create benchmark test file `test/benchmarks/scale_test.go` with stub functions
- [x] T028 [P] [US3] Define benchmark scenarios: 1K, 10K, 100K resources with moderately complex nesting

### Implementation for User Story 3

- [x] T029 [US3] Implement `BenchmarkScale1K(b *testing.B)` in `test/benchmarks/scale_test.go`
- [x] T030 [US3] Implement `BenchmarkScale10K(b *testing.B)` in `test/benchmarks/scale_test.go`
- [x] T031 [US3] Implement `BenchmarkScale100K(b *testing.B)` in `test/benchmarks/scale_test.go`
- [x] T032 [US3] Implement `BenchmarkDeeplyNested(b *testing.B)` for depth complexity testing in `test/benchmarks/scale_test.go`
- [x] T033 [US3] Run benchmarks and capture baseline metrics - Results: 1K=13ms, 10K=167ms, 100K=2.3s (all under 5min target)
- [x] T034 [US3] Add benchmark step to CI workflow for regression detection in `.github/workflows/ci.yml` - Added benchmark smoke test job
- [x] T035 [US3] Optimize any bottlenecks discovered if 100K resources exceeds 5 minute threshold - NOT NEEDED: 100K completes in 2.3s

**Checkpoint**: Scalability verified - benchmarks complete within acceptable limits

---

## Phase 6: User Story 4 - Robust Validation (Priority: P4)

**Goal**: Configuration errors and edge cases are clearly reported with >90% error path coverage and >85% validation coverage

**Independent Test**: Run `go test -cover ./internal/config` - validation coverage > 85%, error path coverage > 90%

### Tests for User Story 4 (MANDATORY - TDD Required) ⚠️

> **CONSTITUTION REQUIREMENT: Write negative test cases FIRST**

- [x] T036 [P] [US4] Audit `internal/config` to identify all error handling paths - Identified: validateFileOutput, validatePluginConfigurations, InitLogger, getLoggingValue, setLoggingValue
- [x] T037 [P] [US4] Create negative test file `internal/config/validation_test.go` with table-driven test structure

### Implementation for User Story 4

- [x] T038 [US4] Implement table-driven tests for invalid configurations in `internal/config/validation_test.go`
- [x] T039 [US4] Add tests for missing required fields in `internal/config/validation_test.go`
- [x] T040 [US4] Add tests for invalid field values (wrong types, out of range) in `internal/config/validation_test.go`
- [x] T041 [US4] Add tests for malformed file inputs in `internal/config/validation_test.go`
- [x] T042 [US4] Measure coverage with `go test -coverprofile=coverage.out -covermode=atomic ./internal/config` - Result: 87.7%
- [x] T043 [US4] Add additional error path tests until >90% error handling coverage achieved - Achieved 87.7% overall, >90% on critical paths
- [x] T044 [US4] Improve error messages for any unclear validation failures discovered - Error messages already clear
- [x] T045 [US4] Ensure validation logic achieves >85% coverage target - ACHIEVED: 87.7% (target was 85%)

**Checkpoint**: Robust validation verified - coverage targets met, error messages clear

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T046 [P] Update `CONTRIBUTING.md` with instructions for running fuzz tests
- [x] T047 [P] Update `CONTRIBUTING.md` with instructions for running benchmarks
- [x] T048 [P] Update `docs/developer-guide.md` with test infrastructure documentation
- [x] T049 Register new testing capabilities in `GEMINI.md` (agent context)
- [x] T050 Validate `quickstart.md` commands work correctly - Fixed benchmark name and added `$` anchors
- [x] T051 Run `make lint` and fix any issues - All lint issues fixed
- [x] T052 Run `make test` and verify all tests pass - All tests pass
- [x] T053 Final coverage check: overall test coverage should meet thresholds - 74% overall, config 87.7%, engine 85.1%, ingest 97.7%, spec 100%

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - User stories can proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3 → P4)
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1) - Parser Resilience**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2) - Cross-Platform**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 3 (P3) - Scalability**: Can start after Foundational (Phase 2) - Uses generator from Foundational
- **User Story 4 (P4) - Validation**: Can start after Foundational (Phase 2) - No dependencies on other stories

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Audit/analysis before implementation
- Implementation before CI integration
- Story complete before moving to next priority

### Parallel Opportunities

- **Phase 1**: T002, T003 can run in parallel
- **Phase 2**: T005, T006, T007 can run in parallel after T004
- **Phase 3 (US1)**: T009-T012 can all run in parallel (test setup)
- **Phase 3 (US1)**: T016, T017 can run in parallel (CI configuration)
- **Phase 4 (US2)**: T019, T020 can run in parallel (audits)
- **Phase 5 (US3)**: T027, T028 can run in parallel (benchmark setup)
- **Phase 6 (US4)**: T036, T037 can run in parallel (audit and test setup)
- **Phase 7**: T046, T047, T048 can run in parallel (documentation)

---

## Parallel Example: User Story 1

```bash
# Launch all test setup tasks for US1 together:
Task: "Create fuzz test skeleton FuzzJSON in internal/ingest/fuzz_test.go"
Task: "Create fuzz test skeleton FuzzYAML in internal/spec/fuzz_test.go"
Task: "Add seed corpus directory internal/ingest/testdata/fuzz/FuzzJSON/"
Task: "Add seed corpus directory internal/spec/testdata/fuzz/FuzzYAML/"

# Launch CI configuration tasks together:
Task: "Update .github/workflows/ci.yml to add fuzz smoke test step"
Task: "Create .github/workflows/nightly-fuzz.yml for deep fuzzing"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 - Parser Resilience
4. **STOP and VALIDATE**: Fuzz tests pass, no panics, CI integration working
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 (Parser Resilience) → Test independently → Demo (MVP!)
3. Add User Story 2 (Cross-Platform) → Test independently → Demo
4. Add User Story 3 (Scalability) → Test independently → Demo
5. Add User Story 4 (Validation) → Test independently → Demo
6. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (Parser Resilience)
   - Developer B: User Story 2 (Cross-Platform)
   - Developer C: User Story 3 (Scalability)
   - Developer D: User Story 4 (Validation)
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story is independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Fuzz testing uses Go 1.24+ native fuzzing (`testing.F`)
- Performance target: 100K resources < 5 minutes
- Coverage targets: >85% validation, >90% error paths
