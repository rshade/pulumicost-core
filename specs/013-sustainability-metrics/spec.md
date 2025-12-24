# Feature Specification: Integrate Sustainability Metrics into Engine & TUI

**Feature Branch**: `013-sustainability-metrics`
**Created**: 2025-12-21
**Status**: Draft
**Input**: User description (Issue #302 context)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - View Carbon Footprint in CLI Table (Priority: P1)

As a DevOps engineer or FinOps practitioner, I want to see carbon footprint estimates for my resources in the CLI table so I can understand the environmental impact of my infrastructure alongside cost.

**Why this priority**: This is the primary user-facing manifestation of the feature, directly addressing the objective of displaying sustainability metrics.

**Independent Test**: Can be tested by running `pulumicost` against a stack using a plugin that returns carbon metrics (e.g., `aws-public`) and verifying the output table.

**Acceptance Scenarios**:

1. **Given** a stack with resources (e.g., `aws:ec2/instance:Instance`) supported by the plugin, **When** I run `pulumicost`, **Then** the output table includes a "CO₂" column showing values formatted with appropriate units (e.g., "2.76 kg").
2. **Given** a resource that does not have carbon data, **When** I run `pulumicost`, **Then** the "CO₂" cell for that resource is empty.
3. **Given** no resources in the stack provide carbon data, **When** I run `pulumicost`, **Then** the "CO₂" column is entirely hidden to avoid clutter.

---

### User Story 2 - Programmatic Access via JSON Output (Priority: P1)

As a platform engineer building dashboards, I want to access detailed impact metrics (carbon, energy, water) in the JSON output so I can integrate this data into external reporting tools.

**Why this priority**: Essential for integration and automation; ensures data is accessible beyond the TUI.

**Independent Test**: Run `pulumicost --json` and validate the structure of the output using `jq` or similar tools.

**Acceptance Scenarios**:

1. **Given** a stack with resources having metrics, **When** I run `pulumicost --json`, **Then** the JSON output for each resource includes an `impactMetrics` field containing objects with `kind`, `value`, and `unit`.
2. **Given** a resource with no metrics, **When** I run `pulumicost --json`, **Then** the `impactMetrics` field is omitted or empty for that resource.

---

### User Story 3 - Adjust Utilization Assumptions (Priority: P2)

As a user performing "what-if" analysis, I want to specify an assumed utilization rate so I can see how it affects the estimated carbon footprint (and potentially cost).

**Why this priority**: Carbon emission estimates for cloud resources often depend heavily on utilization assumptions (e.g., CPU load).

**Independent Test**: Run `pulumicost --utilization 0.8` vs `--utilization 0.2` and observe changes in the output metrics.

**Acceptance Scenarios**:

1. **Given** I run `pulumicost --utilization 0.8`, **Then** the application uses this utilization rate for its calculations (passing it to plugins or using it in estimation logic).
2. **Given** I run `pulumicost` without the flag, **Then** the application defaults to a standard utilization rate (e.g., 0.5 or plugin default).

### Edge Cases

- **Unknown Metric Kind**: Plugin returns a metric type that the Core Engine does not recognize (e.g., from a newer plugin version). System should log it (debug) but not crash.
- **Mixed Units**: Different resources return metrics in different units (e.g., `gCO2e` vs `kgCO2e`). System must normalize them before aggregation.
- **Huge/Tiny Values**: Values requiring scientific notation or different scales (e.g., tonnes vs milligrams). Formatters must handle this gracefully.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST ingest sustainability metrics (Carbon, Energy, Water) provided by cost source plugins.
- **FR-002**: System MUST aggregate metrics by kind (Carbon, Energy, Water) for the entire project/stack.
- **FR-003**: System MUST display per-resource sustainability metrics (specifically Carbon) in the CLI table view when data is available.
- **FR-004**: System MUST hide metric columns in the table view if no resources in the current result set provide data for that metric.
- **FR-005**: System MUST normalize metric units (e.g., convert all weights to grams, energy to kWh) for consistent aggregation and sorting.
- **FR-006**: System MUST output detailed sustainability metrics data in the JSON output (`--json`) matching the logical structure.
- **FR-007**: System MUST accept a `--utilization` flag (float, 0.0-1.0) to override default utilization assumptions.
- **FR-008**: System MUST format metric values in the TUI using human-readable units (e.g., "1.5 kg", "500 g", "2.1 t") based on magnitude.

### Constraints & Assumptions

- **Backward Compatibility**: If a plugin returns no metrics, the user interface must remain unchanged (no empty columns).
- **Utilization Handling**: It is assumed that plugins (like `aws-public`) are capable of receiving or handling the utilization factor, or that the calculation logic allows for this adjustment.
- **Performance**: Metric aggregation should happen in a single pass to maintain O(N) performance.
- **Units**: We assume standard metric prefixes and units (g, kg, t, kWh, MWh, etc.) are used by plugins.

### Key Entities

- **Impact Metric**: A data structure containing the metric kind (e.g., Carbon), numeric value, and unit string.
- **Cost Result**: The existing result structure, updated to hold a list of Impact Metrics.
- **Metric Kind**: Categorization distinguishing between Carbon Footprint, Energy Consumption, and Water Usage.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can view Carbon Footprint (formatted in g/kg/t) for supported resources in the CLI table.
- **SC-002**: Aggregated totals for Carbon and Energy are accurate (sum of normalized resource values) and displayed in the summary.
- **SC-003**: JSON output includes valid metric data for resources that provide it.
- **SC-004**: Running `pulumicost` against a stack with no metric-supporting plugins produces identical visual output to the previous version (no empty columns).