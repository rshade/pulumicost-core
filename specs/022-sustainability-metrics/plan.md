# Implementation Plan - Integrate Sustainability Metrics

**Feature**: Integrate Sustainability Metrics into Engine & TUI
**Status**: In Progress
**Branch**: `022-sustainability-metrics`

## Technical Context

**Objective**: Update `pulumicost-core` to ingest, aggregate, and display sustainability metrics (specifically Carbon Footprint) from plugins.

**Architecture**:
- **Source**: Plugins return `ImpactMetrics` in `GetProjectedCostResponse` (Protobuf).
- **Ingest**: `internal/proto/adapter.go` maps Proto -> Internal Structs.
- **Storage**: `internal/engine/types.go` stores metrics in `CostResult`.
- **Display**: `internal/tui` renders the "CO₂" column if data exists.
- **Input**: New `--utilization` flag passed via CLI.

**Constraints**:
- Must handle mixed units (normalize to base).
- Graceful degradation (hide columns if empty).
- Backward compatible.

## Constitution Check

- [x] **Plugin-First**: Relies entirely on plugin data; no hardcoded emission factors in Core.
- [x] **Test-Driven**: Plan includes unit tests for aggregation and normalization logic.
- [x] **Cross-Platform**: Uses standard Go libraries.
- [x] **Docs-as-Code**: Quickstart and Contracts created.

## Phase 0: Research & Discovery

- [x] **Research**: Confirmed `aws-public` plugin support and Proto definitions.
- [x] **Design**: `data-model.md` and `contracts/cli-interface.md` created.
- [x] **Prototype**: `grpcurl` output validation (from spec).

## Phase 1: Design & Documentation

- [x] **Data Model**: `specs/022-sustainability-metrics/design/data-model.md`
- [x] **Contracts**: `specs/022-sustainability-metrics/design/contracts/cli-interface.md`
- [x] **Quickstart**: `specs/022-sustainability-metrics/design/quickstart.md`

## Phase 2: Implementation Steps

### Step 1: Core Data Structures (Engine)
- Modify `internal/engine/types.go` to add `ImpactMetrics` to `CostResult`.
- Add `ImpactMetric` struct and `MetricKind` constants.
- Add unit tests for struct serialization.

### Step 2: Proto Adapter (Ingest)
- Update `internal/proto/adapter.go` to map `pulumicost_v1.ImpactMetric` to `engine.ImpactMetric`.
- Unit test: Mock plugin response -> `ToCostResult` -> Verify fields.

### Step 3: Aggregation Logic (Engine)
- Update `internal/engine/engine.go` (or `project.go`) to aggregate metrics.
- Implement `NormalizeMetric(val, unit)` helper.
- Unit test: Summing mixed units (kg + g).

### Step 4: CLI Flag & Context (CLI)
- Add `--utilization` flag to `internal/cli`.
- Pass flag value into `AnalyzeRequest` context or specific request field.
- **Note**: If Proto doesn't support utilization field yet, we might need to handle it via generic config map if available, or just implement the flag to be ready for future proto updates. *Decision*: Implement flag, pass if possible, log warning if not supported by current proto context.

### Step 5: TUI Rendering (Presentation)
- Update `internal/tui/render.go` to inspect results for metrics.
- Add "CO₂" column logic (hide if empty).
- Implement `FormatCarbon(grams)` formatter (g/kg/t).

### Step 6: JSON Output (Presentation)
- Verify `json:"impactMetrics"` tag works as expected (should be automatic from Step 1).
- Add integration test for JSON output.

### Step 7: Documentation & Cleanup
- Update main `README.md` or `user-guide.md` with new feature.
- Verify all linting/tests pass.

## Verification Plan

### Automated Tests
- **Unit**: Normalization logic, Aggregation logic.
- **Integration**: Mock plugin returns metrics, Core outputs correct JSON.

### Manual Verification
- Run against `aws-public` plugin (if available locally).
- Check Table output for "CO₂".
- Check JSON output for `impactMetrics`.