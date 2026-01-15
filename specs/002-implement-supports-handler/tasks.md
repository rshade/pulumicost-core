# Tasks: Implement Supports() gRPC Handler

**Input**: Design documents from `/specs/002-implement-supports-handler/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅

**Tests**: Per Constitution Principle II (Test-Driven Development), tests are MANDATORY and must be written BEFORE implementation. All code changes must maintain minimum 80% test coverage (95% for critical paths).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

## Path Conventions

- **Primary package**: `pkg/pluginsdk/` (existing SDK implementation)
- **Test files**: `pkg/pluginsdk/sdk_test.go` (co-located with implementation)

---

## Phase 1: Setup

**Purpose**: Validate environment and dependencies

- [x] T001 Verify finfocus-spec v0.1.0 dependency includes Supports proto definitions in go.mod
- [x] T002 Review existing pkg/pluginsdk/sdk.go to understand Server struct and Plugin interface patterns

**Checkpoint**: Environment validated and codebase patterns understood

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before user stories

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T003 [P] Define SupportsProvider interface in pkg/pluginsdk/sdk.go with Supports method signature
- [x] T004 [P] Define registry lookup function to find plugin by provider/region in pkg/pluginsdk/sdk.go

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Query Plugin Capabilities (Priority: P1) 🎯 MVP

**Goal**: Enable clients to query plugins that implement the Supports capability and receive accurate responses

**Independent Test**: Send gRPC request to Supports endpoint of a plugin implementing the capability; verify accurate supported/not-supported response

### Tests for User Story 1 (MANDATORY - TDD Required) ⚠️

> **CONSTITUTION REQUIREMENT: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T005 [P] [US1] Unit test for Supports when plugin implements SupportsProvider and returns supported in pkg/pluginsdk/sdk_test.go
- [x] T006 [P] [US1] Unit test for Supports when plugin implements SupportsProvider and returns not-supported with reason in pkg/pluginsdk/sdk_test.go
- [x] T007 [P] [US1] Unit test for Supports with invalid provider/region (not in registry) returns InvalidArgument in pkg/pluginsdk/sdk_test.go
- [x] T008 [P] [US1] Unit test for Supports when no plugin registered for provider/region returns InvalidArgument in pkg/pluginsdk/sdk_test.go
- [x] T009 [P] [US1] Unit test for Supports when plugin's Supports method returns error returns Internal status in pkg/pluginsdk/sdk_test.go

### Implementation for User Story 1

- [x] T010 [US1] Implement two-step validation: lookup plugin by provider/region from registry in pkg/pluginsdk/sdk.go
- [x] T011 [US1] Implement Supports method on Server struct with type assertion for SupportsProvider in pkg/pluginsdk/sdk.go
- [x] T012 [US1] Add error handling to return gRPC Internal status on plugin errors in pkg/pluginsdk/sdk.go
- [x] T013 [US1] Add logging for Supports operations and errors in pkg/pluginsdk/sdk.go

**Checkpoint**: User Story 1 fully functional - plugins implementing SupportsProvider work correctly

---

## Phase 4: User Story 2 - Handle Legacy Plugins Gracefully (Priority: P1)

**Goal**: Return default "not supported" response for plugins that don't implement the Supports capability

**Independent Test**: Send gRPC request to Supports endpoint of a plugin NOT implementing the capability; verify default response with reason

### Tests for User Story 2 (MANDATORY - TDD Required) ⚠️

- [x] T014 [P] [US2] Unit test for Supports when plugin does NOT implement SupportsProvider returns default response in pkg/pluginsdk/sdk_test.go
- [x] T015 [P] [US2] Unit test verifying default response includes reason explaining capability not implemented in pkg/pluginsdk/sdk_test.go

### Implementation for User Story 2

- [x] T016 [US2] Enhance Supports method to return default response when type assertion fails in pkg/pluginsdk/sdk.go
- [x] T017 [US2] Define standardized default reason message for unimplemented capability in pkg/pluginsdk/sdk.go

**Checkpoint**: Both user stories complete - Supports handler works for all plugin types

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Validation, documentation, and coverage improvements

- [x] T018 [P] Run make lint and fix any linting issues in pkg/pluginsdk/sdk.go
- [x] T019 [P] Run make test and verify all tests pass with minimum 80% coverage for pkg/pluginsdk
- [x] T020 [P] Add godoc comments for SupportsProvider interface and Supports method in pkg/pluginsdk/sdk.go
- [x] T021 Verify SC-005: Performance test that 99% of Supports queries complete within 50ms - **DEFERRED** to [#162](https://github.com/rshade/finfocus/issues/162)
- [x] T022 [P] Update docs/plugins/plugin-development.md with SupportsProvider interface documentation

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-4)**: All depend on Foundational phase completion
- **Polish (Phase 5)**: Depends on both user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational (Phase 2) - Shares implementation with US1 but independently testable

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Validation/lookup functions before handlers
- Handler implementation before error handling enhancements
- Story complete before moving to next

### Parallel Opportunities

- Foundational tasks T003, T004 can run in parallel (different functions)
- All tests for a user story marked [P] can run in parallel
- T005, T006, T007, T008, T009 can all be written in parallel
- T014, T015 can be written in parallel
- T018, T019, T020 in Polish phase can run in parallel

---

## Parallel Example: User Story 1 Tests

```bash
# Launch all tests for User Story 1 together:
Task: "Unit test for Supports when plugin implements SupportsProvider and returns supported"
Task: "Unit test for Supports when plugin implements SupportsProvider and returns not-supported"
Task: "Unit test for Supports with invalid provider/region returns InvalidArgument"
Task: "Unit test for Supports when no plugin registered returns InvalidArgument"
Task: "Unit test for Supports when plugin returns error returns Internal status"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Verify plugins implementing SupportsProvider work correctly

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Core capability works (MVP!)
3. Add User Story 2 → Test independently → Legacy plugin support works
4. Complete Polish → Full coverage and documentation
5. Each story adds value without breaking previous functionality

### Single Developer Strategy

Since all changes are in the same file (pkg/pluginsdk/sdk.go):

1. Complete Setup + Foundational sequentially
2. Write all US1 tests first (T005-T009)
3. Implement US1 (T010-T013)
4. Write US2 tests (T014-T015)
5. Implement US2 (T016-T017)
6. Complete Polish phase

---

## Notes

- [P] tasks = different test functions or independent code sections
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- All implementation in single file: pkg/pluginsdk/sdk.go

---

## Success Criteria Mapping

- **SC-001** (100% valid responses): Covered by T005, T006, T014, T015
- **SC-002** (Exact match for implementing plugins): Covered by T005, T006
- **SC-003** (Standardized default response): Covered by T014, T015, T017
- **SC-004** (No regressions): Covered by T019 (make test)
- **SC-005** (50ms latency): Covered by T021

---

## Validation Architecture (from Clarifications)

The validation approach uses a **two-step process**:

1. **Registry Lookup**: Find plugin by provider/region from registry.json
2. **Plugin Query**: Delegate resource_type validation to the matched plugin's Supports method

This allows plugins to expand supported resource types independently without core PRs.
