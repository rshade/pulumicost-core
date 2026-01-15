# Implementation Plan - Test Infrastructure Hardening

## 1. Technical Context

### Architecture Overview

This plan outlines the implementation of a hardened test infrastructure for FinFocus Core. The primary goal is to enhance system stability, platform compatibility, and scalability through a multi-layered testing strategy. This involves:

1.  **Fuzz Testing**: Integrating Go's native fuzzing capabilities (`testing.F` introduced in Go 1.18) into the `internal/ingest` and `internal/spec` packages to test JSON and YAML parsers against malformed inputs.
2.  **Cross-Platform Testing**: Leveraging GitHub Actions matrix builds to execute the test suite on Ubuntu, macOS, and Windows runners. This will be configured to run on Nightly builds and Release tags to optimize CI resource usage.
3.  **Performance Benchmarking**: Extending the existing `test/benchmarks` suite to include large-scale datasets (up to 100K resources). We will generate synthetic datasets of varying complexity (flat vs. nested) to simulate realistic enterprise workloads.
4.  **Negative Testing**: Systematically increasing code coverage for error handling paths. We will use `go test -cover` to measure coverage and target specific error conditions in `internal/config` and other critical paths.

### Key Decisions

-   **Fuzzing Strategy**: We will use native Go fuzzing (`go test -fuzz`) over external tools like `go-fuzz` because it is built into the toolchain and requires no external dependencies.
    -   *Constraint*: Fuzzing will run in a "smoke test" mode (short duration, e.g., 30s) during standard CI PR checks to ensure basic stability without blocking velocity. Deep fuzzing will be scheduled nightly.
-   **Cross-Platform Strategy**: We will not run full cross-platform tests on every PR to save costs and time. Instead, we rely on Linux CI for PRs and offload Windows/macOS validation to Nightly/Release workflows.
    -   *Risk*: Platform-specific bugs might be caught later (next morning), but this tradeoff is acceptable for a CLI tool primarily developed on *nix.
-   **Performance Metrics**: We define "acceptable performance" for 100K resources as < 5 minutes. This is a "stress test" limit. Standard usage (1K-10K resources) should be significantly faster (seconds).
-   **Validation Coverage**: We will use standard Go coverage tools to measure validation logic coverage, aiming for >85%.

### System Boundaries

-   **In-Scope**:
    -   `internal/ingest`: JSON parser resilience.
    -   `internal/spec`: YAML parser resilience.
    -   `test/benchmarks`: Performance test suite.
    -   `.github/workflows`: CI configuration updates.
    -   `internal/config`: Configuration validation logic.
-   **Out-of-Scope**:
    -   Changes to the core engine logic (unless bugs are found during testing).
    -   Plugin-specific testing (this is the responsibility of `finfocus-plugin` repos, though integration tests may touch them).
    -   UI/UX changes.

### Data Model Changes

-   No core data model changes are expected.
-   **Test Data**: We will introduce new schemas/generators for synthetic performance test data (JSON/YAML plans of varying sizes).

## 2. Constitution Check

| Principle | Compliance Plan |
|-----------|-----------------|
| **I. Plugin-First** | N/A - This feature focuses on core infrastructure testing, not plugin logic. |
| **II. TDD** | **Non-Negotiable**: We will write the test harness (fuzz targets, benchmarks, negative test cases) *before* modifying any code to fix discovered issues. Coverage metrics will be strictly monitored. |
| **III. Cross-Platform** | **Core Goal**: This feature *explicitly* implements the cross-platform verification requirement of the constitution. |
| **IV. Documentation** | We will update `CONTRIBUTING.md` or `docs/developer-guide.md` to include instructions on running fuzz tests and benchmarks. |
| **V. Protocol Stability** | N/A - No protocol changes. |

**Compliance Statement**: This plan directly supports Principles II and III by hardening the test infrastructure and ensuring cross-platform compatibility. It adheres to TDD by defining test cases (fuzz targets, benchmarks) as the primary deliverable.

### Constitution Deviation: Principle III (Cross-Platform)

**Deviation**: Cross-platform CI verification will run on Nightly builds and Release tags only, not on every commit as mandated by Constitution Principle III.

**Justification**:

1. **Why the principle cannot be followed**: Windows and macOS GitHub Actions runners are significantly more expensive and slower than Linux runners. Running them on every PR would:
   - Increase CI costs by ~3x
   - Slow PR feedback loops by 5-10 minutes
   - Provide diminishing returns for a CLI tool primarily developed on Linux

2. **Simpler alternatives considered and rejected**:
   - *Full matrix on every commit*: Rejected due to cost/velocity impact
   - *Self-hosted runners*: Rejected due to maintenance overhead for a small project
   - *Linux-only CI*: Rejected as it provides no cross-platform verification

3. **Remediation plan**:
   - Nightly runs catch platform regressions within 24 hours (acceptable feedback loop)
   - Release tag runs ensure published binaries are verified
   - Future consideration: Adopt self-hosted runners when project scale justifies maintenance cost

**Risk Acceptance**: Platform-specific bugs may be caught next-morning rather than immediately. This is acceptable for the current project maturity.

## 3. Phase 0: Research & Validation

**Goal**: Resolve technical unknowns and validate approach.

-   [ ] **Research Fuzzing Implementation**: Verify best practices for Go native fuzzing in a CI environment (e.g., how to cache corpus, how to set time limits).
-   [ ] **Benchmark Data Generation**: Determine the best way to generate realistic, nested infrastructure plans programmatically (e.g., using a library or custom generator).
-   [ ] **CI/CD Matrix Configuration**: finalizing the GitHub Actions YAML configuration for nightly cross-platform builds.

## 4. Phase 1: Design & Contracts

**Goal**: Define test interfaces and data structures.

-   [ ] **Design Fuzz Targets**: Define the function signatures for `FuzzJSON` and `FuzzYAML` targets.
-   [ ] **Design Benchmark Suite**: Define the structure of the benchmark tests (`BenchmarkLargeScale`, `BenchmarkDeeplyNested`).
-   [ ] **Design Synthetic Data Generator**: Define the interface for the test data generator (e.g., `GeneratePlan(resourceCount int, depth int) -> Plan`).
-   [ ] **Update Agent Context**: Register the new testing capabilities in `GEMINI.md`.

## 5. Phase 2: Implementation Strategy

**Goal**: Execute the changes in logical steps.

### Step 1: Fuzz Testing Framework (P1)
-   Create `internal/ingest/fuzz_test.go` and `internal/spec/fuzz_test.go`.
-   Implement `FuzzJSON` and `FuzzYAML` targets.
-   Run initial fuzzing locally to seed the corpus.
-   Integrate into CI (short run) and add deep fuzzing job to consolidated `nightly.yml` (long run).

### Step 2: Cross-Platform CI (P2)
-   Create/Update `.github/workflows/nightly.yml` (or equivalent).
-   Configure build matrix: `os: [ubuntu-latest, windows-latest, macos-latest]`.
-   Ensure all tests pass on all platforms (fix any path separator or OS-specific issues found).

### Step 3: Performance Benchmarks (P3)
-   Implement `test/benchmarks/generator` package.
-   Create `test/benchmarks/scale_test.go`.
-   Implement benchmarks for 1K, 10K, 100K resources.
-   Run benchmarks and baseline performance.

### Step 4: Negative Testing & Validation (P4)
-   Audit `internal/config` for error paths.
-   Write table-driven tests for invalid configurations.
-   Measure coverage and add cases until >85% validation / >90% error path coverage is reached.

## 6. Verification Plan

**Automated Verification**:
-   **Fuzzing**: `go test -fuzz=Fuzz -fuzztime=30s ./...` passes in CI.
-   **Benchmarks**: `go test -bench=. ./test/benchmarks` completes successfully.
-   **Coverage**: `go test -coverprofile=coverage.out ./...` shows >85% for targeted packages.
-   **Cross-Platform**: GitHub Actions logs show green builds for Windows/macOS/Linux.

**Manual Verification**:
-   Run `make lint` and `make test` locally.
-   Review benchmark results to ensure 100K resources < 5 mins.
