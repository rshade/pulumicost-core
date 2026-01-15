# Tasks: Integrate Sustainability Metrics into Engine & TUI

**Status**: In Progress
**Feature**: Integrate Sustainability Metrics
**Branch**: `022-sustainability-metrics`

## Phase 1: Setup
*(Project structure and dependencies)*

- [x] T001 Verify project dependencies and proto definitions availability via `go mod` or local check in `go.mod`
- [x] T002 Update `internal/engine/types.go` to include `SustainabilityMetric` struct
- [x] T003 Update `CostResult` struct in `internal/engine/types.go` to include `Sustainability map[string]SustainabilityMetric` field

## Phase 2: Foundational
*(Blocking prerequisites for all user stories: Ingestion & Aggregation)*

- [x] T004 Create unit test `internal/engine/types_test.go` to verify serialization and defaults for new `SustainabilityMetric` structs
- [x] T005 [P] Create unit test `internal/proto/adapter_test.go` to mock plugin response with metrics and verify mapping
- [x] T006 Implement update in `internal/proto/adapter.go` to map `finfocus_v1.ImpactMetric` to `engine.SustainabilityMetric`
- [x] T007 [P] Create unit test `internal/engine/aggregation_test.go` to test summing metrics across resources
- [x] T008 Implement metric normalization logic in `internal/engine/engine.go`
- [x] T009 Implement metric aggregation logic in `internal/engine/engine.go` to sum metrics by kind across all resources

## Phase 3: User Story 1 - View Carbon Footprint in CLI Table (P1)
*(Goal: Users can see CO2 column in the table output)*

- [x] T010 [US1] Create unit test for sustainability rendering in `internal/engine/project_test.go`
- [x] T011 [US1] Implement formatting helpers for sustainability metrics in `internal/engine/project.go`
- [x] T012 [US1] Update `internal/engine/project.go` to include Sustainability Summary in output
- [x] T013 [US1] Implement dynamic column logic in `internal/engine/project.go` to show CO2 data
- [x] T014 [US1] Manual verification: Run `finfocus` against a stack using `aws-public` plugin (if available) to verify table output

## Phase 4: User Story 2 - Programmatic Access via JSON Output (P1)
*(Goal: Users can get metrics via --json flag)*

- [x] T015 [US2] Verify `finfocus --json` asserts `sustainability` field presence and structure
- [x] T016 [US2] Verify JSON output correctness

## Phase 5: User Story 3 - Adjust Utilization Assumptions (P2)
*(Goal: Users can change utilization rate via CLI flag)*

- [x] T017 [US3] Add `--utilization` float64 flag to `internal/cli/cost_projected.go` (using cobra)
- [x] T018 [US3] Update `AnalyzeRequest` context propagation in `internal/engine` to include utilization value
- [x] T019 [US3] Manual verification: Run `finfocus --utilization 0.8` vs `0.2`

## Phase 6: Polish & Cross-Cutting Concerns

- [x] T020 Update `user-guide.md` to document the new "CO₂" column and `--utilization` flag
- [x] T021 Run `make lint` and ensure no new linting errors
- [x] T022 Run `make test` to ensure full regression suite passes