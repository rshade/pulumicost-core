# Research: Integrate Sustainability Metrics

**Feature**: Integrate Sustainability Metrics into Engine & TUI
**Status**: Complete
**Date**: 2025-12-21

## Executive Summary

The research phase for this feature is brief because the core dependencies (Protobuf definitions and Plugin capabilities) are already verified. The primary work is integration within the `pulumicost-core` codebase.

## Technical Decisions

### 1. Protobuf Integration
**Decision**: Use existing `ImpactMetrics` field in `pulumicost-spec` (v0.4.10+).
**Rationale**: The definitions already exist in the proto files:
```protobuf
message ImpactMetric {
  MetricKind kind = 1;
  double value = 2;
  string unit = 3;
}
```
**Verification**: Verified via `grpcurl` against `aws-public` plugin in the feature request.

### 2. TUI Rendering Library
**Decision**: Extend existing `bubbletea` / `lipgloss` usage in `internal/tui`.
**Rationale**: The codebase uses `charmbracelet/bubbletea` for the TUI. We will add a column to the table model.
**Approach**:
- Check if "Carbon" column is needed (iterate results).
- Use `lipgloss` for styling (e.g., green color for low carbon, or neutral).
- Use dynamic column generation logic in `internal/engine/project.go` or `render.go`.

### 3. Unit Normalization
**Decision**: Implement simple normalization logic in Go.
**Rationale**: Plugins might return `kg`, `g`, `t`. To aggregate correctly, we must normalize to a base unit (e.g., grams) for summation, then format for display.
**Base Units**:
- Carbon: Grams (gCO2e)
- Energy: Watt-hours (Wh) or kWh
- Water: Milliliters (mL) or Liters (L)

### 4. CLI Flags
**Decision**: Use `spf13/cobra` flags in `internal/cli`.
**Flag**: `--utilization` (float64).
**Propagation**: Pass this value via `AnalyzeRequest` or `GetProjectedCostRequest` context if the proto supports it.
**Note**: If the current proto `GetProjectedCostRequest` does not support utilization overrides, we may need to rely on the plugin's default or check if `CostSourceService` allows configuration.
*Correction*: The spec mentions "Assumed utilization rate". If the proto doesn't have a field for this, we might need to send it as `config` or similar. **However**, the feature spec implies the plugin *already* calculates this or we do it. The `aws-public` plugin likely uses a default. If we can't pass it to the plugin without a proto change, we might mark this as "Best Effort" or check if `inputs` map can carry it. *Assumption*: The plugin might respect a `utilization` input property if we inject it into the resource inputs.

## Unknowns Resolved

- **Plugin Support**: Confirmed `aws-public` branch `018-raw-pricing-embed` supports this.
- **Proto Definitions**: Confirmed available.
- **Formatting**: Will use standard Go `fmt.Sprintf` with logic for k/M/G prefixes.

## Plan Updates

No significant deviations from the initial feature request. The implementation plan remains focused on wiring these existing pieces together.
