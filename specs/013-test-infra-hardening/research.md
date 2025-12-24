# Research Report - Test Infrastructure Hardening

**Date**: 2025-12-02
**Context**: Hardening test infrastructure with fuzzing, cross-platform support, and performance benchmarks.

## 1. Go Native Fuzzing Strategy

**Decision**: Use Go native fuzzing (`go test -fuzz`) integrated into both standard PR workflows (smoke tests) and nightly schedules (deep fuzzing).

**Rationale**:
-   **Standard Library**: No external dependencies required (Go 1.24+, project minimum).
-   **Integration**: Seamlessly integrates with existing `go test` infrastructure.
-   **Corpus Management**: Automatically manages interesting inputs in `testdata/fuzz`.

**Implementation Details**:
-   **Smoke Tests (PRs)**: Run with `-fuzztime=30s` to catch low-hanging fruit without blocking CI.
-   **Deep Fuzzing (Nightly)**: Run without `-fuzztime` (or with a long timeout like 6h) to explore deeper paths.
-   **Caching**: Cache `testdata/fuzz` in GitHub Actions to preserve the corpus between runs, improving efficiency over time.

**Alternatives Considered**:
-   *go-fuzz*: External tool, requires separate build steps and widely considered superseded by native fuzzing for most Go use cases.
-   *Google OSS-Fuzz*: Powerful but higher setup complexity; better suited for later maturity stages.

## 2. Performance Data Generation

**Decision**: Build a custom generator using standard Go structs and `math/rand`, potentially leveraging `gofakeit` for realistic leaf-node values (strings, IPs).

**Rationale**:
-   **Structure Control**: We need precise control over resource nesting depth and dependency graph complexity (DAGs) to simulate realistic Terraform/Pulumi state files.
-   **Performance**: Custom generator avoids overhead of generic reflection-based libraries for high-volume data generation.
-   **Simplicity**: A simple recursive generator with a `depth` parameter meets our "moderately complex" requirement without external heavy-lifting libraries.

**Design Pattern**:
-   **Two-Pass Generation**:
    1.  Generate pool of resource IDs.
    2.  Generate resource bodies and assign dependencies from the pool to ensure valid graphs.
-   **Parameters**: `ResourceCount`, `NestingDepth`, `DependencyFactor`.

**Alternatives Considered**:
-   *Terraform CLI*: Generating real state files via Terraform is too slow and complex for unit/benchmark tests.
-   *Raw JSON Templates*: Hard to scale to 100K resources dynamically.

## 3. Cross-Platform CI Strategy

**Decision**: Use GitHub Actions Matrix (`os: [ubuntu-latest, windows-latest, macos-latest]`) triggered only on `schedule` (Nightly) and `release`.

**Rationale**:
-   **Cost/Time Balance**: Windows/macOS runners are significantly more expensive and slower than Linux. Running them on every PR is wasteful for a CLI tool primarily developed on *nix.
-   **Risk Mitigation**: Nightly runs catch platform regressions within 24 hours, which is an acceptable feedback loop for this project's maturity.
-   **Configuration**: Use `fail-fast: true` to stop the matrix if any job fails, saving credits.

**Implementation Details**:
-   **Workflow**: `nightly.yml`.
-   **Triggers**: `cron: '0 0 * * *'` and `types: [published]`.
-   **Steps**: Use `if: runner.os == 'Windows'` for any platform-specific setup (though Go is mostly portable).

**Alternatives Considered**:
-   *Run on all PRs*: Rejected due to cost and velocity impact.
-   *Self-hosted runners*: Rejected due to maintenance overhead.
