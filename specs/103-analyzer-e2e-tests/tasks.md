# Tasks: Analyzer E2E Tests

**Feature**: Analyzer E2E Tests
**Status**: Complete
**Spec**: [spec.md](./spec.md)
**Plan**: [plan.md](./plan.md)
**Completed**: 2025-12-09

## Implementation Strategy

We will implement this feature by extending the existing E2E test suite in `test/e2e`.
The core strategy relies on treating the analyzer as a local plugin configured via `Pulumi.yaml`.
We will build the `pulumicost` binary once (handled by `TestMain` or Makefile) and then point the test fixture to this binary.

Implementation will follow a prioritized approach:

1. **Setup**: Establish the test fixture project.
2. **Foundations**: Create the test file and helper to configure the analyzer
   path.
3. **User Stories**: Implement tests for handshake, diagnostics, and summary in
   order.

## Dependencies

1. **US1** (Handshake) blocks all other stories.
2. **US2** (Diagnostics) and **US3** (Summary) depend on US1 but are mutually
   independent.
3. **US4** (Degradation) and **US5** (Portability) are independent validation
   tasks.

## Parallel Execution Opportunities

- **US2** and **US3** test functions can be implemented in parallel once the
  base test structure is ready.
- **US4** can be implemented alongside US2/US3.

---

## Phase 1: Setup

**Goal**: Initialize the test environment and fixture project.

- [x] T001 Create fixture directory `test/e2e/projects/analyzer`
- [x] T002 Create `test/e2e/projects/analyzer/Pulumi.yaml` with initial
      configuration (random resource)

---

## Phase 2: Foundational

**Goal**: Create the test structure and binary configuration logic.

- [x] T003 Create `test/e2e/analyzer_e2e_test.go` with initial package and
      imports
- [x] T003a Verify `go.mod` uses Pulumi SDK v3.210.0+ and ensure `TestMain`
      correctly builds the binary
- [x] T004 Implement `ensureAnalyzerConfig` helper in
      `test/e2e/analyzer_e2e_test.go` to inject absolute binary path into
      `Pulumi.yaml`

---

## Phase 3: Verify Analyzer Handshake Protocol (US1)

**Goal**: Ensure the analyzer plugin starts and communicates with Pulumi CLI.
**Priority**: P1

- [x] T005 [US1] Implement `TestAnalyzer_Handshake` in
      `test/e2e/analyzer_e2e_test.go` verifying process start
- [x] T006 [US1] Verify `pulumi preview` succeeds with the analyzer configured
      (exit code 0) in `test/e2e/analyzer_e2e_test.go`

---

## Phase 4: Verify Cost Diagnostics in Preview Output (US2)

**Goal**: Ensure per-resource cost estimates appear in the output.
**Priority**: P1

- [x] T007 [US2] Update `test/e2e/projects/analyzer/Pulumi.yaml` to include a
      priced resource (AWS EC2)
- [x] T008 [P] [US2] Define `AnalyzerDiagnostic` struct/helpers in
      `test/e2e/analyzer_e2e_test.go` for output validation
- [x] T009 [US2] Implement `TestAnalyzer_CostDiagnostics` in
      `test/e2e/analyzer_e2e_test.go` validating "Estimated Monthly Cost" string

---

## Phase 5: Verify Stack Cost Summary (US3)

**Goal**: Ensure total stack cost is summarized.
**Priority**: P2

- [x] T010 [P] [US3] Implement `TestAnalyzer_StackSummary` in
      `test/e2e/analyzer_e2e_test.go` validating "Total Estimated Monthly Cost"

---

## Phase 6: Verify Graceful Degradation (US4)

**Goal**: Ensure failures in cost calculation do not break the preview.
**Priority**: P2

- [x] T011 [US4] Implement `TestAnalyzer_GracefulDegradation` in
      `test/e2e/analyzer_e2e_test.go` using an invalid region or unsupported
      resource to verify non-blocking error diagnostics

---

## Phase 7: Verify Test Environment Portability (US5)

**Goal**: Ensure tests run correctly in CI/CD without hard cloud dependencies.
**Priority**: P3

- [x] T012 [P] [US5] Add `SkipIfNoPulumi` check to `test/e2e/analyzer_e2e_test.go`
- [x] T013 [US5] Verify tests pass in local environment using `make test-e2e`

---

## Phase 8: Polish & Cross-Cutting Concerns

**Goal**: Final code quality checks and cleanup.

- [x] T014 Run `go mod tidy` to ensure dependencies are clean
- [x] T015 Run `golangci-lint run test/e2e/...` to check for linting issues
- [x] T016 Update `test/README.md` with instructions for running the new
      analyzer E2E tests

---

## Additional Work Completed

During implementation, the following additional work was done:

### Bug Fixes

- **Duplicate diagnostics fix**: `AnalyzeStack()` was returning per-resource
  diagnostics in addition to `Analyze()`, causing duplicates. Fixed by:
  - Adding cost caching in `Analyze()` calls
  - Making `AnalyzeStack()` return only the summary diagnostic
  - Updating unit tests to reflect the correct Pulumi workflow

### Documentation

- Created `/docs/analyzer-integration.md` - User-facing integration guide
- Created `integration_testing_notes.md` - Technical testing notes with verified test results
- Added analyzer E2E tests to `.github/workflows/nightly.yml`

### Verified Working Configuration

The E2E tests use this exact setup (verified working 2025-12-09):

1. Binary named `pulumi-analyzer-policy-pulumicost`
2. `PulumiPolicy.yaml` with `runtime: pulumicost`
3. Policy pack directory added to PATH
4. `pulumi preview --policy-pack <path>` to activate
