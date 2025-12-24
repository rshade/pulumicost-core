# Tasks: CLI Filter Flag Support

**Feature Branch**: `023-add-cli-filter-flag`
**Spec**: [specs/023-add-cli-filter-flag/spec.md](../spec.md)
**Plan**: [specs/023-add-cli-filter-flag/plan.md](../plan.md)

## Phase 1: Setup

**Goal**: Ensure the environment is ready and dependencies are clear.

- [X] T001 Verify project compiles and runs with 'make build' and 'make run'
- [X] T002 Run integration tests to confirm failure in test/integration/cli/filter_test.go

## Phase 2: User Story 1 (Filter Costs by Tag or Type)

**Goal**: Add the `--filter` flag to the `actual` cost command to fix integration tests and enable granular cost analysis.

**Independent Test**:
- Integration tests in `test/integration/cli/filter_test.go` pass.
- `pulumicost cost actual --help` displays the `--filter` flag.
- Manual execution with `--filter` properly filters results.

**Implementation**:

- [X] T003 [US1] Add 'filter' field as '[]string' to 'costActualParams' struct in internal/cli/cost_actual.go
- [X] T004 [US1] Register '--filter' flag using 'StringArrayVar' in 'NewCostActualCmd' function in internal/cli/cost_actual.go
- [X] T005 [US1] Implement iterative application of filters using 'engine.FilterResources' in 'executeCostActual' function in internal/cli/cost_actual.go
- [X] T006 [US1] Run 'test/integration/cli/filter_test.go' to verify fix
- [X] T007 [US1] Verify help output includes new flag by running 'pulumicost cost actual --help'

## Phase 3: Polish

**Goal**: Ensure code quality and consistency.

- [X] T008 Run 'make lint' to ensure code quality standards

## Dependencies

1.  **US1** is the primary story and has no dependencies on other stories.
2.  **T003**, **T004**, and **T005** must be done sequentially as they modify the same file and function chain.

## Implementation Strategy

1.  **Baseline**: Confirm the test failure to establish the "Red" state of TDD (T002).
2.  **Implementation**: Modify the CLI command definition to accept the flag and wire it to the existing engine logic (T003-T005).
3.  **Verification**: Confirm the test pass ("Green" state) (T006).
4.  **Polish**: Ensure linting passes (T008).
