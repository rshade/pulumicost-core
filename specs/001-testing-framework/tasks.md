# Implementation Tasks: Testing Framework and Strategy

**Feature**: Testing Framework and Strategy  
**Branch**: `001-testing-framework`  
**Spec**: [spec.md](./spec.md) | **Plan**: [plan.md](./plan.md)  
**GitHub Issue**: [#9](https://github.com/rshade/pulumicost-core/issues/9)

## Task Organization

### Conventions
- **[P]**: Can be executed in parallel with other [P] tasks in same phase
- **[US#]**: Maps to User Story # from spec.md for traceability
- **Dependencies**: Listed explicitly when task requires another task's completion
- **Tests**: Per Constitution Principle II (Test-Driven Development), tests are MANDATORY and must be written BEFORE implementation

### Directory Structure
- Unit tests: `test/unit/[package]/`
- Integration tests: `test/integration/`
- E2E tests: `test/e2e/`
- Fixtures: `test/fixtures/[category]/`
- Mocks: `test/mocks/[type]/`
- Benchmarks: `test/benchmarks/`

---

## Phase 1: Setup (Shared Infrastructure) ✅ COMPLETE

**Purpose**: Create test directory structure and configure dependencies

- [x] T001 Create test directory structure (test/unit/, test/integration/, test/e2e/, test/fixtures/, test/mocks/, test/benchmarks/)
- [x] T002 [P] Add github.com/sebdah/goldie/v2 dependency to go.mod for golden file testing
- [x] T003 [P] Add github.com/google/go-cmp dependency to go.mod for deep comparisons
- [x] T004 [P] Create test/README.md documenting test organization and conventions
- [x] T005 [P] Create test/fixtures/README.md documenting fixture organization

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Mock plugin infrastructure that enables all other testing

**⚠️ CRITICAL**: Mock plugin must be complete before integration/E2E tests can be written

- [ ] T006 Design mock plugin API (ConfigureResponse, SetError, Reset methods) in test/mocks/plugin/api.go
- [ ] T007 Implement mock plugin gRPC server in test/mocks/plugin/server.go
- [ ] T008 Implement mock plugin response configuration in test/mocks/plugin/config.go
- [ ] T009 Add mock plugin test helpers (NewMockPlugin, StartServer, StopServer) in test/mocks/plugin/helpers.go
- [ ] T010 [P] Create test fixture loader utility in test/fixtures/loader.go
- [ ] T011 [P] Document mock plugin usage examples in test/mocks/plugin/README.md

---

## Phase 3: User Story 1 - Unit Testing Foundation (Priority: P1) 🎯 MVP

**Goal**: Establish unit testing for all packages with 80% coverage minimum

### Tests for User Story 1 (MANDATORY - TDD Required) ⚠️

> **CONSTITUTION REQUIREMENT: Write these tests FIRST, ensure they FAIL before writing tested code**

#### Engine Package Tests (Critical Path - 95% coverage required)

- [ ] T012 [P] [US1] Unit test for cost calculation in test/unit/engine/engine_test.go
- [ ] T013 [P] [US1] Unit test for GetProjectedCost in test/unit/engine/projected_test.go
- [ ] T014 [P] [US1] Unit test for GetActualCost in test/unit/engine/actual_test.go
- [ ] T015 [P] [US1] Unit test for output rendering (table, JSON, NDJSON) in test/unit/engine/render_test.go
- [ ] T016 [P] [US1] Unit test for error handling paths in test/unit/engine/errors_test.go
- [ ] T017 [P] [US1] Unit test for cross-provider aggregation in test/unit/engine/aggregation_test.go

#### CLI Package Tests (Critical Path - 95% coverage required)

- [ ] T018 [P] [US1] Unit test for cost projected command in test/unit/cli/cost_projected_test.go
- [ ] T019 [P] [US1] Unit test for cost actual command in test/unit/cli/cost_actual_test.go
- [ ] T020 [P] [US1] Unit test for plugin commands in test/unit/cli/plugin_test.go
- [ ] T021 [P] [US1] Unit test for CLI flag parsing in test/unit/cli/flags_test.go

#### PluginHost Package Tests (Critical Path - 95% coverage required)

- [ ] T022 [P] [US1] Unit test for plugin discovery in test/unit/pluginhost/discovery_test.go
- [ ] T023 [P] [US1] Unit test for plugin lifecycle in test/unit/pluginhost/lifecycle_test.go
- [ ] T024 [P] [US1] Unit test for gRPC client management in test/unit/pluginhost/client_test.go

#### Registry Package Tests

- [ ] T025 [P] [US1] Unit test for plugin registry scanning in test/unit/registry/scan_test.go
- [ ] T026 [P] [US1] Unit test for manifest validation in test/unit/registry/manifest_test.go

#### Ingest Package Tests

- [ ] T027 [P] [US1] Unit test for Pulumi plan parsing in test/unit/ingest/plan_test.go
- [ ] T028 [P] [US1] Unit test for resource mapping in test/unit/ingest/mapper_test.go

#### Config Package Tests

- [ ] T029 [P] [US1] Unit test for configuration loading in test/unit/config/load_test.go
- [ ] T030 [P] [US1] Unit test for environment variable handling in test/unit/config/env_test.go

#### Spec Package Tests

- [ ] T031 [P] [US1] Unit test for YAML spec parsing in test/unit/spec/parse_test.go
- [ ] T032 [P] [US1] Unit test for spec loading in test/unit/spec/load_test.go

### Implementation for User Story 1

- [ ] T033 [US1] Run full test suite with coverage reporting: `go test -coverprofile=coverage.out ./...`
- [ ] T034 [US1] Generate coverage HTML report: `go tool cover -html=coverage.out`
- [ ] T035 [US1] Verify 80% overall coverage achieved
- [ ] T036 [US1] Verify 95% coverage on critical paths (CLI, engine, pluginhost)
- [ ] T037 [US1] Fix any coverage gaps to meet thresholds
- [ ] T038 [US1] Document coverage results in test/COVERAGE.md

---

## Phase 4: User Story 4 - Mock Plugin Enhancement (Priority: P2)

**Goal**: Expand mock plugin with error injection and performance testing

### Tests for User Story 4 (MANDATORY - TDD Required) ⚠️

- [ ] T039 [US4] Test mock plugin response configuration in test/mocks/plugin/config_test.go
- [ ] T040 [US4] Test error injection scenarios in test/mocks/plugin/errors_test.go
- [ ] T041 [US4] Test performance simulation in test/mocks/plugin/perf_test.go

### Implementation for User Story 4

- [ ] T042 [US4] Add 5 configurable response scenarios to mock plugin
- [ ] T043 [US4] Implement 3 error injection types (timeout, protocol error, invalid data)
- [ ] T044 [US4] Add latency simulation for performance testing
- [ ] T045 [US4] Add response validation helpers
- [ ] T046 [US4] Document all mock plugin capabilities in test/mocks/plugin/README.md
- [ ] T047 [US4] Create examples for each mock scenario

---

## Phase 5: User Story 2 - Integration Testing (Priority: P2)

**Goal**: Verify cross-component communication

**Dependencies**: Requires Phase 2 (Mock Plugin) completion

### Tests for User Story 2 (MANDATORY - TDD Required) ⚠️

- [ ] T048 [US2] Integration test for CLI → Engine workflow in test/integration/cli_workflow_test.go
- [ ] T049 [US2] Integration test for Engine → Plugin gRPC communication in test/integration/plugin_comm_test.go
- [ ] T050 [US2] Integration test for configuration loading across components in test/integration/config_loading_test.go
- [ ] T051 [US2] Integration test for error propagation in test/integration/errors_test.go
- [ ] T052 [US2] Integration test for output format generation in test/integration/output_test.go

### Implementation for User Story 2

- [ ] T053 [US2] Create integration test helper for launching mock plugin
- [ ] T054 [US2] Create integration test helper for CLI command execution
- [ ] T055 [US2] Verify all integration tests pass with mock plugin
- [ ] T056 [US2] Document integration testing patterns in test/integration/README.md

---

## Phase 6: User Story 6 - CI/CD Automation (Priority: P1) 🎯 MVP

**Goal**: Automate testing in CI with coverage enforcement

### Tests for User Story 6 (MANDATORY - TDD Required) ⚠️

- [ ] T057 [US6] Test coverage threshold validation script in scripts/check-coverage.sh
- [ ] T058 [US6] Test critical path coverage validation in scripts/check-critical-coverage.sh

### Implementation for User Story 6

- [ ] T059 [US6] Create coverage threshold validation script (scripts/check-coverage.sh)
- [ ] T060 [US6] Create critical path coverage validation script (scripts/check-critical-coverage.sh)
- [ ] T061 [US6] Update .github/workflows/ci.yml to run coverage validation
- [ ] T062 [US6] Add coverage report generation (Cobertura XML for PR comments)
- [ ] T063 [US6] Configure PR blocking on coverage < 80%
- [ ] T064 [US6] Configure PR blocking on test failures
- [ ] T065 [US6] Add coverage badge to README.md
- [ ] T066 [US6] Test full CI pipeline with intentional coverage drop
- [ ] T067 [US6] Verify PR is blocked when coverage threshold not met

---

## Phase 7: User Story 5 - Test Fixtures (Priority: P3)

**Goal**: Comprehensive test data coverage

### Tests for User Story 5 (MANDATORY - TDD Required) ⚠️

- [ ] T068 [US5] Test fixture loader in test/fixtures/loader_test.go
- [ ] T069 [US5] Test fixture validation in test/fixtures/validate_test.go

### Implementation for User Story 5

#### AWS Fixtures

- [ ] T070 [P] [US5] Create test/fixtures/plans/aws/simple.json (EC2, S3, RDS)
- [ ] T071 [P] [US5] Create test/fixtures/plans/aws/complex.json (Multi-AZ, AutoScaling)
- [ ] T072 [P] [US5] Create test/fixtures/plans/aws/lambda.json (Serverless)

#### Azure Fixtures

- [ ] T073 [P] [US5] Create test/fixtures/plans/azure/simple.json (VM, Storage, SQL)
- [ ] T074 [P] [US5] Create test/fixtures/plans/azure/complex.json (App Service, AKS)
- [ ] T075 [P] [US5] Create test/fixtures/plans/azure/functions.json (Azure Functions)

#### GCP Fixtures

- [ ] T076 [P] [US5] Create test/fixtures/plans/gcp/simple.json (Compute, Storage, SQL)
- [ ] T077 [P] [US5] Create test/fixtures/plans/gcp/complex.json (GKE, Cloud Run)
- [ ] T078 [P] [US5] Create test/fixtures/plans/gcp/functions.json (Cloud Functions)

---

## Phase 8: User Story 3 - End-to-End Testing (Priority: P3)

**Goal**: Complete workflow validation

**Dependencies**: Requires Phases 2, 5, 7 completion

### Tests for User Story 3 (MANDATORY - TDD Required) ⚠️

- [ ] T079 [US3] E2E test for projected cost workflow in test/e2e/projected_cost_test.go
- [ ] T080 [US3] E2E test for actual cost workflow in test/e2e/actual_cost_test.go
- [ ] T081 [US3] E2E test for golden file validation (table output) in test/e2e/output_table_test.go
- [ ] T082 [US3] E2E test for golden file validation (JSON output) in test/e2e/output_json_test.go
- [ ] T083 [US3] E2E test for golden file validation (NDJSON output) in test/e2e/output_ndjson_test.go
- [ ] T084 [US3] E2E test for error scenarios in test/e2e/errors_test.go
- [ ] T085 [US3] E2E test for AWS plan processing in test/e2e/aws_test.go
- [ ] T086 [US3] E2E test for Azure plan processing in test/e2e/azure_test.go
- [ ] T087 [US3] E2E test for GCP plan processing in test/e2e/gcp_test.go

### Implementation for User Story 3

- [ ] T088 [US3] Create golden files for all output formats in test/fixtures/golden/
- [ ] T089 [US3] Create E2E test helper for full workflow execution
- [ ] T090 [US3] Verify all E2E tests pass
- [ ] T091 [US3] Document E2E testing patterns in test/e2e/README.md

---

## Phase 9: Benchmarks (Performance Testing)

**Goal**: Performance regression detection

### Tests for Benchmarks (MANDATORY) ⚠️

- [ ] T092 [P] Benchmark for cost calculation in test/benchmarks/engine_bench_test.go
- [ ] T093 [P] Benchmark for JSON parsing in test/benchmarks/parse_bench_test.go
- [ ] T094 [P] Benchmark for plugin communication in test/benchmarks/plugin_bench_test.go
- [ ] T095 [P] Benchmark for large plan processing in test/benchmarks/scale_bench_test.go

### Implementation for Benchmarks

- [ ] T096 Run baseline benchmarks: `go test -bench=. -benchmem ./test/benchmarks/`
- [ ] T097 Store baseline results in test/benchmarks/baseline.txt
- [ ] T098 Create benchmark comparison script (scripts/compare-benchmarks.sh)
- [ ] T099 Add benchmark validation to CI (weekly or on perf-critical PRs)
- [ ] T100 Document benchmarking process in test/benchmarks/README.md

---

## Phase 10: Polish & Cross-Cutting Concerns

**Goal**: Documentation, cleanup, and final validation

### Documentation

- [ ] T101 [P] Create comprehensive testing guide in docs/testing/guide.md
- [ ] T102 [P] Create troubleshooting guide in docs/testing/troubleshooting.md
- [ ] T103 [P] Update CONTRIBUTING.md with testing requirements
- [ ] T104 [P] Add testing examples to docs/testing/examples/

### Quality Validation

- [ ] T105 Run full test suite with race detection: `go test -race ./...`
- [ ] T106 Verify zero race conditions reported
- [ ] T107 Run golangci-lint on all test code
- [ ] T108 Fix any linting issues in test code
- [ ] T109 Generate final coverage report
- [ ] T110 Verify all success criteria met (SC-001 through SC-012 from spec.md)

---

## MVP Scope (Phases 1-3, 6)

**Total MVP Tasks**: 48 tasks
**Estimated Time**: 2-3 weeks
**Coverage Target**: 80% overall, 95% critical paths

### MVP Includes:
- ✅ Phase 1: Setup (5 tasks) - COMPLETE
- Phase 2: Mock Plugin (6 tasks)
- Phase 3: Unit Tests (27 tasks)
- Phase 6: CI/CD Automation (9 tasks)

### Post-MVP (Optional Enhancement):
- Phase 4: Mock Plugin Enhancement (9 tasks)
- Phase 5: Integration Testing (9 tasks)
- Phase 7: Test Fixtures (11 tasks)
- Phase 8: E2E Testing (14 tasks)
- Phase 9: Benchmarks (9 tasks)
- Phase 10: Polish (11 tasks)

---

## Task Dependencies Graph

```
Phase 1 (Setup) → Phase 2 (Mock Plugin) → Phase 3 (Unit Tests)
                                        → Phase 4 (Mock Enhancement)
                                        → Phase 5 (Integration Tests)
Phase 1 → Phase 6 (CI/CD)
Phase 1 → Phase 7 (Fixtures) → Phase 8 (E2E Tests)
Phase 1 → Phase 9 (Benchmarks)
All Phases → Phase 10 (Polish)
```

---

## Progress Tracking

- Total Tasks: 110
- Completed: 5 (Phase 1)
- Remaining: 105
- Progress: 4.5%
