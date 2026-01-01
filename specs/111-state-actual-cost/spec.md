# Feature Specification: State-Based Actual Cost Estimation

**Feature Branch**: `111-state-actual-cost`
**Created**: 2025-12-31
**Status**: Draft
**Input**: User description: "State-based actual cost estimation for cost actual
command with confidence levels and test reliability fixes"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Estimate Actual Costs from Pulumi State (Priority: P1)

A DevOps engineer wants to understand how much their deployed infrastructure has
cost since deployment, but their organization uses the `aws-public` plugin which
lacks billing API access. They have access to Pulumi state (via
`pulumi stack export`) and want runtime-based cost estimates.

**Why this priority**: This is the core value proposition - enabling actual cost
visibility without cloud billing API access. Users with public pricing plugins
(like `aws-public`) currently have no way to get actual cost estimates.

**Independent Test**: Can be fully tested by exporting Pulumi state, running
`cost actual --pulumi-state state.json`, and verifying estimated costs appear
based on resource runtime.

**Acceptance Scenarios**:

1. **Given** a Pulumi state file with resources containing `Created` timestamps,
   **When** I run `pulumicost cost actual --pulumi-state state.json`,
   **Then** I see estimated costs calculated as `hourly_rate × runtime_hours` for
   each resource.

2. **Given** a Pulumi state file without specifying `--from` date,
   **When** I run `pulumicost cost actual --pulumi-state state.json`,
   **Then** the command auto-detects `--from` as the earliest `Created` timestamp.

3. **Given** a plugin returns actual billing data for some resources but not others,
   **When** I run `cost actual --pulumi-state state.json`,
   **Then** billing data is used where available and state-based estimates fill gaps.

---

### User Story 2 - Understand Estimate Confidence Levels (Priority: P2)

A FinOps analyst reviews cost reports and needs to distinguish between accurate
billing data, reliable runtime estimates, and potentially inaccurate estimates
for imported resources.

**Why this priority**: Transparency about data quality is essential for financial
decisions. Without confidence indicators, users may make decisions based on
inaccurate estimates.

**Independent Test**: Can be tested by running `cost actual --estimate-confidence`
and verifying that each resource shows HIGH/MEDIUM/LOW confidence with appropriate
conditions.

**Acceptance Scenarios**:

1. **Given** a resource with actual billing data from a FinOps plugin,
   **When** I run `cost actual --estimate-confidence`,
   **Then** that resource shows `HIGH` confidence.

2. **Given** a resource created by Pulumi (External=false) with state-based estimate,
   **When** I run `cost actual --estimate-confidence`,
   **Then** that resource shows `MEDIUM` confidence.

3. **Given** an imported resource (External=true) with state-based estimate,
   **When** I run `cost actual --estimate-confidence`,
   **Then** that resource shows `LOW` confidence with a note about import timing.

---

### User Story 3 - Reliable CI Pipeline Execution (Priority: P1)

A platform engineer manages the CI/CD pipeline and needs nightly E2E tests and
Windows tests to pass reliably. Flaky tests waste time investigating false failures.

**Why this priority**: CI reliability is foundational - broken tests block all
development. This is marked P1 because it's a prerequisite for deploying new features.

**Independent Test**: Can be tested by running `make test-e2e` and `make test`
on Windows, verifying all tests pass consistently across 3+ consecutive runs.

**Acceptance Scenarios**:

1. **Given** the current E2E test suite with proper timeouts,
   **When** I run tests on Linux and Windows,
   **Then** all tests complete within their timeout limits without hanging.

2. **Given** Azure/GCP resources in a Pulumi plan,
   **When** `AWS_REGION` is set in the environment,
   **Then** Azure/GCP resources do NOT inherit the AWS region value.

3. **Given** cost results with map-based breakdowns,
   **When** rendered as table or JSON output,
   **Then** the output is deterministic (sorted keys) across runs.

---

### User Story 4 - Cross-Provider Cost Aggregation (Priority: P3)

A cloud architect manages multi-cloud infrastructure and wants to see daily/monthly
cost trends across AWS, Azure, and GCP with estimated actual costs.

**Why this priority**: Cross-provider analysis is valuable but builds on P1/P2
functionality. Users need basic actual cost estimation working first.

**Independent Test**: Can be tested by running `cost actual --pulumi-state state.json
--group-by daily` and verifying aggregated costs appear by date and provider.

**Acceptance Scenarios**:

1. **Given** state-based cost estimates for multiple providers,
   **When** I run `cost actual --pulumi-state state.json --group-by daily`,
   **Then** I see a time-series table with columns for each provider.

---

### Edge Cases

- **Resources without timestamps**: Pulumi state from before v3.60.0 lacks
  `Created`/`Modified` fields. These resources are skipped with a warning note.

- **Imported resources**: `pulumi import` sets `Created` to import time, not
  cloud creation time. These are marked with LOW confidence and a disclaimer.

- **Stopped/restarted resources**: Runtime assumes 100% uptime since creation.
  This is documented as a known limitation.

- **Mixed plugin results**: Some resources may have billing data (HIGH confidence)
  while others require estimation. Each resource shows its own confidence level.

- **Empty state file**: If state has no custom resources, command returns early
  with an informative message.

- **State without state file**: If `--pulumi-state` is not provided, existing
  plugin-only behavior continues (no regression).

## Requirements *(mandatory)*

### Functional Requirements

#### Phase 1: Bug Fixes (Prerequisites)

- **FR-001**: System MUST scope AWS environment variable fallback (`AWS_REGION`,
  `AWS_DEFAULT_REGION`) to only apply when processing AWS resources.

- **FR-002**: System MUST render map-based output (breakdown, sustainability metrics)
  in deterministic sorted key order.

- **FR-003**: System MUST return typed errors from `plugin_validate.go` instead of
  calling `os.Exit(1)` directly.

- **FR-004**: E2E tests MUST use `exec.CommandContext` with configurable timeouts
  (default 30 seconds).

- **FR-005**: Conformance tests MUST use test context for gRPC calls instead of
  `context.Background()`.

- **FR-006**: Conformance context timeout test MUST use 10ms timeout (not 1µs).

- **FR-007**: E2E test main runner MUST separate stdout and stderr capture.

#### Phase 2: State-Based Actual Cost

- **FR-008**: CLI MUST accept `--pulumi-state <path>` flag on `cost actual` command.

- **FR-009**: When `--pulumi-state` is provided without `--from`, CLI MUST auto-detect
  the earliest `Created` timestamp from state resources.

- **FR-010**: Engine MUST calculate estimated actual costs as
  `hourly_rate × runtime_hours` where runtime is `now - created`.

- **FR-011**: Engine MUST try plugin `GetActualCost` first, then fall back to
  state-based estimation when plugin returns no data.

- **FR-012**: Output MUST include notes for imported resources (External=true)
  warning about timestamp accuracy.

- **FR-013**: Resources without `Created` timestamp MUST be skipped with a warning
  note: "No timestamp available - cannot estimate runtime".

#### Phase 3: Estimate Confidence

- **FR-014**: CLI MUST accept `--estimate-confidence` flag on `cost actual` command.

- **FR-015**: Confidence level MUST be calculated as:
  - HIGH: Real billing data from plugin (TotalCost > 0 from plugin)
  - MEDIUM: Runtime-based estimate, External=false
  - LOW: Runtime-based estimate, External=true

- **FR-016**: When `--estimate-confidence` is enabled, table output MUST include
  a "CONFIDENCE" column.

- **FR-017**: When `--estimate-confidence` is enabled, JSON output MUST include
  a `confidence` field in each result.

- **FR-018**: `CostResult` struct MUST include a `Confidence` field.

### Key Entities

- **StackExport**: Parsed Pulumi state containing deployment resources with timestamps.
  Already exists in `internal/ingest/state.go`.

- **CostResult**: Extended with `Confidence` field (string: "HIGH", "MEDIUM", "LOW",
  or empty).

- **ResourceDescriptor**: Cloud resource with type, provider, and properties including
  injected Pulumi metadata (`pulumi:created`, `pulumi:external`).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can obtain actual cost estimates for 100% of state resources
  that have `Created` timestamps, even without billing API access.

- **SC-002**: Nightly CI pipeline (E2E tests, Windows tests) achieves 100% pass
  rate over 5 consecutive runs after fixes are deployed.

- **SC-003**: Cost output is byte-identical across 10 consecutive runs with the
  same input (deterministic ordering).

- **SC-004**: State-based cost estimation adds less than 100ms latency for stacks
  with up to 100 resources.

- **SC-005**: Users can distinguish data source quality (billing vs estimate) for
  100% of cost results when using `--estimate-confidence` flag.

- **SC-006**: Azure and GCP resources no longer inherit AWS region configuration
  from environment variables.

## Assumptions

- Pulumi state files are generated from Pulumi CLI v3.60.0 or later (for timestamp
  support). Older states will have resources skipped with warnings.

- Estimated costs assume 100% uptime from creation to now. Stopped/started instances
  are not tracked.

- The `aws-public` and similar public pricing plugins support `GetProjectedCost`
  for hourly rate extraction.

- Users understand that state-based estimates are approximations, not billing-accurate
  figures.

## Dependencies

- **Pulumi CLI v3.60.0+**: Required for `Created`/`Modified` timestamps in state.

- **Existing infrastructure**: `internal/ingest/state.go` already provides state
  parsing, `MapStateResource()`, and `HasTimestamps()`.

- **Plugin SDK**: Hourly rates come from `GetProjectedCost` results.

## Out of Scope

- Cloud API queries for real resource launch times (would require credentials).
- Caching of state parsing results.
- Webhook/email notifications for cost alerts.
- Handling resources created before Pulumi v3.60.0 (documented limitation).
- Real-time cost monitoring or streaming updates.

## References

- Research Document: `GetActualCost.md`
- Existing State Parser: `internal/ingest/state.go`
- [GitHub Actions Run #20610627111](https://github.com/rshade/pulumicost-core/actions/runs/20610627111)
- [Pulumi v3.60.0 Release Notes](https://github.com/pulumi/pulumi/discussions/12529)
- Issues: #380, #333, #378, #324, #323
