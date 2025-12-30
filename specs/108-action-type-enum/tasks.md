# Tasks: Extended RecommendationActionType Enum Support

**Input**: Design documents from `/specs/108-action-type-enum/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: Per Constitution Principle II (Test-Driven Development), tests are
MANDATORY and must be written BEFORE implementation. All code changes must
maintain minimum 80% test coverage.

**Completeness**: Per Constitution Principle VI (Implementation Completeness),
all tasks MUST be fully implemented. Stub functions, placeholders, and TODO
comments are strictly forbidden.

**Organization**: Tasks are grouped by user story to enable independent
implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

This is a Go CLI project with existing packages:

- `internal/proto/` - Protocol adapter and recommendation types
- `internal/cli/` - CLI commands
- `internal/tui/` - Terminal UI components
- `test/unit/` - Unit tests

---

## Phase 0: CLI Command (Priority: P0)

**Purpose**: Add `cost recommendations` CLI command using existing engine
infrastructure. This is a prerequisite for all filtering/display features.

### Tests for CLI Command (TDD Required)

- [x] T001 [P] [US4] Write tests for NewCostRecommendationsCmd() in
      internal/cli/cost_recommendations_test.go
- [x] T002 [P] [US4] Write tests for recommendations output rendering (table,
      JSON, NDJSON) in internal/cli/cost_recommendations_test.go
- [x] T003 [P] [US4] Write tests for --filter flag parsing with action types in
      internal/cli/cost_recommendations_test.go

### Implementation for CLI Command

- [x] T004 [US4] Create internal/cli/cost_recommendations.go with
      NewCostRecommendationsCmd() following cost_actual.go patterns
- [x] T005 [US4] Add --pulumi-json, --adapter, --output, --filter flags to
      command
- [x] T006 [US4] Implement executeCostRecommendations() calling
      Engine.GetRecommendationsForResources()
- [x] T007 [US4] Implement RenderRecommendationsOutput() for table/JSON/NDJSON
      formats in internal/cli/cost_recommendations.go (follows cost_actual.go pattern)
- [x] T008 [US4] Wire command into cost command group in internal/cli/root.go

**Checkpoint**: `cost recommendations` command works with table/JSON output

---

## Phase 1: Setup

**Purpose**: Verify dependencies and create new files

- [x] T009 [Setup] Verify pulumicost-spec v0.4.11 is present with extended
      action types in go.mod
- [x] T010 [Setup] Create new file internal/proto/action_types.go for action
      type utilities

---

## Phase 2: Foundational (Core Utilities)

**Purpose**: Create shared action type utilities used by ALL user stories

**CRITICAL**: User stories depend on these utilities being complete first

### Tests for Foundational Phase (TDD Required)

- [x] T011 [P] [Core] Write table-driven tests for ActionTypeLabel() covering
      all 11 types in internal/proto/action_types_test.go
- [x] T012 [P] [Core] Write table-driven tests for ParseActionType() with
      valid/invalid inputs in internal/proto/action_types_test.go
- [x] T013 [P] [Core] Write table-driven tests for ParseActionTypeFilter() with
      comma-separated values in internal/proto/action_types_test.go
- [x] T014 [P] [Core] Write tests for ValidActionTypes() excluding UNSPECIFIED
      in internal/proto/action_types_test.go
- [x] T015 [P] [Core] Write tests for unknown/future enum value handling
      (display as "Unknown (N)") in internal/proto/action_types_test.go

### Implementation for Foundational Phase

- [x] T016 [Core] Implement actionTypeLabels map with all 11 human-readable
      labels in internal/proto/action_types.go
- [x] T017 [Core] Implement ActionTypeLabel(at pbc.RecommendationActionType)
      string in internal/proto/action_types.go
- [x] T018 [Core] Implement ParseActionType(s string)
      (pbc.RecommendationActionType, error) with case-insensitive matching in
      internal/proto/action_types.go
- [x] T019 [Core] Implement ParseActionTypeFilter(filter string)
      ([]pbc.RecommendationActionType, error) in internal/proto/action_types.go
- [x] T020 [Core] Implement ValidActionTypes() []string excluding UNSPECIFIED in
      internal/proto/action_types.go
- [x] T021 [Core] Run tests and verify all pass with 80%+ coverage: go test
      -cover ./internal/proto/...

**Checkpoint**: Foundation ready - utilities for all user stories are complete

---

## Phase 3: User Story 1 - Filter Recommendations (Priority: P1)

**Goal**: Enable filtering recommendations by the new action types (MIGRATE,
CONSOLIDATE, SCHEDULE, REFACTOR, OTHER)

**Independent Test**: Filter flag accepts new action types and returns only
matching recommendations

### Tests for User Story 1 (TDD Required)

- [x] T022 [P] [US1] Write tests for action type filter parsing in CLI command
      in internal/cli/cost_recommendations_test.go
- [x] T023 [P] [US1] Write tests for invalid action type filter error message
      listing all 11 valid types in internal/cli/cost_recommendations_test.go

### Implementation for User Story 1

- [x] T024 [US1] Add MatchesActionType(rec Recommendation, types
      []pbc.RecommendationActionType) bool helper in internal/proto/action_types.go
- [x] T025 [US1] Integrate action type filter into cost recommendations command
      using ParseActionTypeFilter() from action_types.go
- [x] T026 [US1] Update CLI help text to list all 11 valid action types in
      internal/cli/cost_recommendations.go

**Checkpoint**: User Story 1 complete - users can filter by all action types

---

## Phase 4: User Story 2 - TUI Display Labels (Priority: P2)

**Goal**: Display human-readable labels for all action types in TUI mode

**Note**: The TUI package (`internal/tui/`) exists with shared components. This
phase adds action type formatting utilities that will be used by CLI table output
and any future recommendations TUI view.

**Independent Test**: Mock recommendations with new action types display proper
labels in TUI

### Tests for User Story 2 (TDD Required)

- [x] T027 [P] [US2] Write tests for TUI action type label rendering in
      internal/tui/components_test.go

### Implementation for User Story 2

- [x] T028 [US2] Add FormatActionType(actionType string) string function using
      ActionTypeLabel() in internal/tui/components.go
- [x] T029 [US2] Use FormatActionType() in CLI table output for recommendations
      in internal/cli/cost_recommendations.go
- [x] T030 [US2] Verify unknown action types display as "Unknown (N)" with
      warning log in internal/tui/components.go

**Checkpoint**: User Story 2 complete - TUI shows human-readable labels

---

## Phase 5: User Story 3 - JSON Output (Priority: P3)

**Goal**: JSON output correctly serializes all action types

**Independent Test**: JSON output contains correct action_type string values

### Tests for User Story 3 (TDD Required)

- [x] T031 [P] [US3] Write tests for JSON serialization of all 11 action types
      in internal/proto/adapter_test.go

### Implementation for User Story 3

- [x] T032 [US3] Verify Recommendation struct's ActionType string field is
      populated correctly from proto in internal/proto/adapter.go
- [x] T033 [US3] Add integration test validating JSON output format with new
      action types in internal/cli/cost_recommendations_test.go

**Checkpoint**: User Story 3 complete - JSON output includes all action types

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and documentation

- [x] T034 [P] Run full test suite: make test
- [x] T035 [P] Run linting: make lint
- [x] T036 [P] Verify 80% test coverage for new code: go test -cover
      ./internal/proto/... ./internal/cli/...
- [x] T037 Update CLAUDE.md with any new patterns or utilities discovered
- [x] T038 Validate quickstart.md examples work correctly with cost
      recommendations command

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 0 (CLI Command)**: No dependencies - can start immediately
- **Phase 1 (Setup)**: No dependencies - can run in parallel with Phase 0
- **Phase 2 (Foundational)**: Depends on Phase 1 - BLOCKS all user stories
- **Phase 3-5 (User Stories)**: Depend on Phase 0 (CLI) and Phase 2 (Foundational)
  - User stories can proceed in priority order (P1 → P2 → P3)
  - Or in parallel if team capacity allows
- **Phase 6 (Polish)**: Depends on all user stories being complete

### User Story Dependencies

- **US4 (CLI Command)**: Core command needed for all other stories
- **US1 (Filter)**: Needs US4 + ActionTypeLabel, ParseActionType, ValidActionTypes
- **US2 (TUI)**: Needs ActionTypeLabel for display
- **US3 (JSON)**: No additional utilities needed - uses existing struct

### Within Each Phase

- Tests MUST be written and FAIL before implementation (TDD)
- Tests can run in parallel (marked [P])
- Implementation follows test completion
- Verify tests pass before moving to next phase

### Parallel Opportunities

```text
Phase 0 - All tests (T001-T003) can run in parallel
Phase 0 - Implementation (T004-T008) is sequential (shared files)
Phase 2 - All tests (T011-T015) can run in parallel
Phase 2 - Implementation (T016-T020) is sequential (shared file)
Phase 3 - Tests (T022-T023) can run in parallel
Phase 4 - Tests (T027) independent
Phase 5 - Tests (T031) independent
Phase 6 - Validation tasks (T034-T036) can run in parallel
```

---

## Parallel Example: Phase 0 and Phase 2 Tests

```bash
# Launch Phase 0 CLI command tests:
Task: "Write tests for NewCostRecommendationsCmd() in cost_recommendations_test.go"
Task: "Write tests for recommendations output rendering in cost_recommendations_test.go"
Task: "Write tests for --filter flag parsing in cost_recommendations_test.go"

# Launch Phase 2 foundational tests:
Task: "Write table-driven tests for ActionTypeLabel() in action_types_test.go"
Task: "Write table-driven tests for ParseActionType() in action_types_test.go"
Task: "Write table-driven tests for ParseActionTypeFilter() in action_types_test.go"
Task: "Write tests for ValidActionTypes() in action_types_test.go"
Task: "Write tests for unknown enum value handling in action_types_test.go"
```

---

## Implementation Strategy

### MVP First (CLI Command + Filtering)

1. Complete Phase 0: CLI Command (the command itself)
2. Complete Phase 1: Setup (verify dependencies)
3. Complete Phase 2: Foundational (core utilities with TDD)
4. Complete Phase 3: User Story 1 (filter support)
5. **STOP and VALIDATE**: Test `cost recommendations --filter "action=MIGRATE"`
6. Ship/demo MVP

### Incremental Delivery

1. Phase 0 (CLI Command) → Basic command works
2. Phase 1+2 (Setup + Foundational) → Action type utilities ready
3. Add User Story 1 (Filter) → Test independently → Deploy (MVP!)
4. Add User Story 2 (TUI Labels) → Test independently → Deploy
5. Add User Story 3 (JSON Output) → Test independently → Deploy
6. Each story adds value without breaking previous stories

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story (US1, US2, US3, US4)
- [Setup] = Phase 1 setup tasks
- [Core] = Phase 2 foundational utilities used by multiple stories
- US4 = CLI Command (P0 prerequisite for all other stories)
- All 11 action types must be covered: RIGHTSIZE, TERMINATE,
  PURCHASE_COMMITMENT, ADJUST_REQUESTS, MODIFY, DELETE_UNUSED, MIGRATE,
  CONSOLIDATE, SCHEDULE, REFACTOR, OTHER
- Use pbc.AllRecommendationActionTypes() from pluginsdk for dynamic validation
- Case-insensitive matching with strings.EqualFold() for filter parsing
- Display "Unknown (N)" for unrecognized values (forward compatibility)
- Commit after each task or logical group
- Total: 38 tasks across 7 phases (Phase 0-6)
