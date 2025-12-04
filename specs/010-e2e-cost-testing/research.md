# Research: E2E Cost Testing

**Feature**: E2E Cost Testing using Pulumi Automation API
**Date**: 2025-12-03
**Status**: Completed

## Decision Log

### 1. Pulumi Automation API Strategy
**Context**: We need to programmatically deploy infrastructure to validate cost calculations.
**Decision**: Use `auto.UpsertStackInlineSource` with dynamic stack naming.
**Rationale**:
- Inline programs allow defining infrastructure code directly within the Go test file, keeping tests self-contained.
- Dynamic stack naming (e.g., `e2e-test-<ulid>`) ensures isolation for concurrent test runs.
- `UpsertStack` handles both creation and updates, simplifying the logic.

**Alternatives Considered**:
- **Shell Scripts**: Harder to manage error handling, cleanup, and cross-platform compatibility. Rejected per requirements.
- **Local Workspace**: Requires separate Pulumi projects on disk. Harder to manage in CI/ephemeral environments.

### 2. Cleanup Strategy
**Context**: AWS resources cost money. We must ensure cleanup even if tests panic or fail.
**Decision**: Use Go's `defer` mechanism combined with a custom `TestContext` cleanup handler.
**Rationale**:
- `defer` is the idiomatic Go way to ensure execution on function exit.
- A `TestContext` wrapper can track resources and enforce a final destroy call.
- We will also implement a "force destroy" timeout context to handle stuck deployments.

**Alternatives Considered**:
- **Cron Job Cleanup**: Too complex to set up and implies we expect leaks.
- **Manual Cleanup**: Not scalable or safe for CI.

### 3. Cost Validation Logic
**Context**: We need to compare calculated costs against expected values with a tolerance.
**Decision**: Implement a `CostValidator` struct that takes expected monthly costs and applies a ±5% tolerance.
**Rationale**:
- Floating point comparisons require tolerance.
- AWS pricing can vary slightly by region or due to free tier, so strict equality is brittle.
- Reusable validator logic allows easy addition of new resource types.

### 4. Integration with Go Test
**Context**: Needs to run with standard `go test`.
**Decision**: Use build tags `//go:build e2e` and a separate Makefile target `make test-e2e`.
**Rationale**:
- Prevents E2E tests (which take minutes) from slowing down unit tests (which take seconds).
- Standard Go tooling support.

### 5. AWS Isolation & Configuration
**Context**: Need to prevent interference between tests and prod, and configure regions.
**Decision**: Use a dedicated AWS account for tests. Configure region via Pulumi stack config, defaulting to `AWS_REGION` env var.
**Rationale**:
- dedicated account is the gold standard for safety.
- Flexible configuration supports local dev (env vars) and CI pipelines (config).

## Open Questions Resolved

- **Q**: How to handle AWS credentials?
  - **A**: Inherit from environment variables (`AWS_ACCESS_KEY_ID`, etc.), standard for Automation API.
- **Q**: Where to store state?
  - **A**: Use local backend for Pulumi state (`pulumi login --local`) to avoid needing a Pulumi Cloud account for CI tests.