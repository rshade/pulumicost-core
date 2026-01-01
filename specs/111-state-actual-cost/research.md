# Research: State-Based Actual Cost Estimation

**Feature**: 111-state-actual-cost
**Date**: 2025-12-31

## Research Tasks

### 1. Existing State Parsing Infrastructure

**Question**: What state parsing capabilities already exist in the codebase?

**Findings**: Comprehensive state parsing already exists in
`internal/ingest/state.go`:

- `StackExport` struct parses `pulumi stack export` JSON output
- `StackExportResource` includes `Created`, `Modified`, `External` fields
- `LoadStackExport()` / `LoadStackExportWithContext()` for file loading
- `MapStateResource()` / `MapStateResources()` for conversion to
  `ResourceDescriptor`
- Property keys: `pulumi:created`, `pulumi:modified`, `pulumi:external`
- `GetCustomResources()` filters cloud resources (excludes component resources)
- `HasTimestamps()` checks if state contains timestamp data

**Decision**: Reuse existing infrastructure entirely. No new state parsing code
needed.

**Rationale**: The existing implementation is well-tested and follows project
patterns. It already injects Pulumi metadata into `ResourceDescriptor.Properties`.

**Alternatives Considered**:

- Create separate state parser for actual cost: Rejected (code duplication)
- Extend existing parser: Rejected (already has all needed functionality)

---

### 2. Hourly Rate Extraction Strategy

**Question**: How to obtain hourly rates for runtime-based cost calculation?

**Findings**: The engine already has `GetProjectedCost` which returns
`CostResult` with `Hourly` field. The existing pipeline:

1. Loads resources from plan/state
2. Queries plugins via `GetProjectedCost`
3. Falls back to local YAML specs if plugins unavailable
4. Returns `CostResult.Hourly` rate

**Decision**: Use existing projected cost pipeline to get hourly rates, then
multiply by runtime hours.

**Rationale**: Leverages existing plugin infrastructure and spec fallback.
Consistent with "plugin-first" architecture.

**Alternatives Considered**:

- Query `GetActualCost` and divide by time: Rejected (requires billing API)
- Hardcode rates in core: Rejected (violates plugin-first principle)
- New gRPC method `GetHourlyRate`: Rejected (unnecessary protocol change)

---

### 3. Confidence Level Determination Algorithm

**Question**: How to reliably determine confidence levels?

**Findings**: Three confidence sources identified:

| Level  | Condition                       | Detection Method              |
| ------ | ------------------------------- | ----------------------------- |
| HIGH   | Real billing data from plugin   | Plugin `GetActualCost` returns|
| MEDIUM | Runtime estimate, Pulumi-made   | `External=false` in state     |
| LOW    | Runtime estimate, imported      | `External=true` in state      |

**Decision**: Assign confidence based on data source (plugin vs estimate) and
resource origin (native vs imported).

**Rationale**: Clear, deterministic rules that users can understand. Imported
resources have `Created` set to import time, not cloud creation time.

**Alternatives Considered**:

- Add VERY_LOW for old state without timestamps: Rejected (resources skipped
  entirely with warning)
- Numeric confidence (0-100): Rejected (over-engineering for 3 distinct cases)

---

### 4. AWS Region Fallback Scoping

**Question**: How to fix AWS region fallback affecting Azure/GCP resources?

**Findings**: In `internal/proto/adapter.go` line 490-500, the fallback to
`AWS_REGION` and `AWS_DEFAULT_REGION` executes unconditionally after
provider-specific extraction:

```go
// Current (buggy):
if region == "" {
    if envReg := os.Getenv("AWS_REGION"); envReg != "" {
        region = envReg
    }
}
```

**Decision**: Wrap the fallback in a provider check:

```go
// Fixed:
if strings.EqualFold(provider, "aws") && region == "" {
    if envReg := os.Getenv("AWS_REGION"); envReg != "" {
        region = envReg
    }
}
```

**Rationale**: Simple, targeted fix. AWS env vars should only apply to AWS.

**Alternatives Considered**:

- Provider-specific env vars (AZURE_REGION, GCP_REGION): Rejected (scope creep)
- Remove env var fallback entirely: Rejected (breaks existing AWS workflows)

---

### 5. Deterministic Output Ordering

**Question**: Best practice for deterministic map iteration in Go?

**Findings**: Go maps have non-deterministic iteration order. Standard pattern:

```go
keys := make([]string, 0, len(m))
for k := range m {
    keys = append(keys, k)
}
sort.Strings(keys)
for _, k := range keys {
    // process m[k]
}
```

**Decision**: Apply sorted key iteration to all map-based output rendering in
`project.go` and `diagnostics.go`.

**Rationale**: Ensures byte-identical output across runs. Required for SC-003.

**Alternatives Considered**:

- Use `sync.Map`: Rejected (doesn't guarantee order either)
- Use `orderedmap` package: Rejected (adds dependency for simple fix)

---

### 6. E2E Test Timeout Strategy

**Question**: What timeout value for E2E tests on slow CI runners?

**Findings**: Current issues:

- `test/e2e/errors_test.go`: No timeout, can hang indefinitely
- `internal/conformance/context.go`: Uses 1µs timeout (too short, flaky)

Industry standard for E2E tests: 30-60 seconds per test.

**Decision**: Use 30 seconds default for E2E tests, 10ms for conformance
context timeout test (tests that context cancellation works, not real work).

**Rationale**: 30s is sufficient for Windows slow runners while fast enough to
catch genuine hangs. 10ms is enough to verify context cancellation without
being flaky.

**Alternatives Considered**:

- 60 seconds: Rejected (too slow for fast feedback)
- 5 seconds: Rejected (too aggressive for Windows/CI variability)

---

### 7. Plugin-First Fallback Behavior

**Question**: When should state-based estimation be used vs plugin data?

**Findings**: Current `cost actual` flow:

1. Load resources from Pulumi plan
2. Query plugins via `GetActualCost`
3. Return results (or error if no data)

Proposed flow with state:

1. Load resources from state file
2. Query plugins via `GetActualCost`
3. If plugin returns data: use it (HIGH confidence)
4. If plugin returns no data: calculate from state (MEDIUM/LOW confidence)

**Decision**: Try plugin first, fall back to state-based estimation per-resource.

**Rationale**: Respects plugin-first architecture. Uses best available data.
Mixed results (some plugin, some estimated) are correctly attributed.

**Alternatives Considered**:

- State-only mode (skip plugins): Rejected (misses real billing data)
- Plugin-only mode (ignore state): Rejected (current behavior, no improvement)

---

## Summary

All research tasks complete. No NEEDS CLARIFICATION items remain.

| Topic                     | Decision                                     |
| ------------------------- | -------------------------------------------- |
| State parsing             | Reuse existing `internal/ingest/state.go`    |
| Hourly rates              | Use `GetProjectedCost` from plugins/specs    |
| Confidence levels         | HIGH/MEDIUM/LOW based on source + External   |
| AWS region fix            | Wrap fallback in `provider == "aws"` check   |
| Deterministic output      | Sort map keys before iteration               |
| E2E timeouts              | 30s for tests, 10ms for conformance          |
| Plugin fallback           | Try plugin first, then state-based estimate  |
