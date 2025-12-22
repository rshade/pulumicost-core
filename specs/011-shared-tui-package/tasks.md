# Tasks: Create shared TUI package with Bubble Tea/Lip Gloss components

**Input**: Design documents from `/specs/011-shared-tui-package/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Per Constitution Principle II (Test-Driven Development), tests are MANDATORY and must be written BEFORE implementation. All code changes must maintain minimum 80% test coverage (95% for critical paths).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Internal package**: `internal/tui/` at repository root
- Tests in same package: `internal/tui/*_test.go`

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Create internal/tui/ directory structure
- [x] T002 [P] Add Bubble Tea and Lip Gloss dependencies to go.mod
- [x] T003 [P] Configure Go module and import paths
- [x] T004 [P] Bump pulumicost-spec dependency to v0.4.4
- [x] T005 [P] Define color constants in internal/tui/colors.go
- [x] T006 [P] Define icon constants in internal/tui/components.go
- [x] T007 [P] Define OutputMode type and constants in internal/tui/detect.go
- [x] T008 [P] Define ProgressBar struct in internal/tui/progress.go

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Consistent CLI Styling (Priority: P1) 🎯 MVP

**Goal**: Provide consistent visual styling across all CLI commands

**Independent Test**: Run multiple CLI commands and verify identical color schemes and styling patterns

### Tests for User Story 1 (MANDATORY - TDD Required) ⚠️

> **CONSTITUTION REQUIREMENT: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T009 [P] [US1] Unit tests for color constants in internal/tui/colors_test.go
- [x] T010 [P] [US1] Unit tests for style definitions in internal/tui/styles_test.go

### Implementation for User Story 1

- [x] T011 [US1] Implement Lip Gloss style definitions in internal/tui/styles.go
- [x] T012 [US1] Add usage documentation in code comments for styles.go

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Reusable UI Components (Priority: P2)

**Goal**: Provide reusable TUI components for progress bars, status indicators, and formatting

**Independent Test**: Import package and use components like progress bars and status renderers in test code

### Tests for User Story 2 (MANDATORY - TDD Required) ⚠️

- [x] T013 [P] [US2] Unit tests for progress bar rendering in internal/tui/progress_test.go
- [x] T014 [P] [US2] Unit tests for status rendering functions in internal/tui/components_test.go
- [x] T015 [P] [US2] Unit tests for money formatting utilities in internal/tui/render_test.go

### Implementation for User Story 2

- [x] T016 [US2] Implement progress bar Render method in internal/tui/progress.go
- [x] T017 [US2] Implement status rendering functions in internal/tui/components.go
- [x] T018 [US2] Implement delta rendering function in internal/tui/components.go
- [x] T019 [US2] Implement priority rendering function in internal/tui/components.go
- [x] T020 [US2] Implement money formatting utilities in internal/tui/render.go
- [x] T021 [US2] Implement percentage formatting utility in internal/tui/render.go
- [x] T022 [US2] Add usage documentation in code comments for components.go and render.go

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - TTY Detection and Fallbacks (Priority: P3)

**Goal**: Automatically adapt output based on terminal capabilities and environment

**Independent Test**: Run commands with different terminal configurations and verify appropriate output modes

### Tests for User Story 3 (MANDATORY - TDD Required) ⚠️

- [x] T023 [P] [US3] Unit tests for output mode detection in internal/tui/detect_test.go
- [x] T024 [P] [US3] Unit tests for TTY utilities in internal/tui/detect_test.go
- [x] T025 [US3] Implement DetectOutputMode function in internal/tui/detect.go
- [x] T026 [US3] Implement IsTTY utility function in internal/tui/detect.go
- [x] T027 [US3] Implement TerminalWidth utility function in internal/tui/detect.go
- [x] T028 [US3] Add usage documentation in code comments for detect.go

**Checkpoint**: All user stories should now be independently functional

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T029 [P] Run go mod tidy and verify dependencies
- [x] T030 [P] Run golangci-lint and fix any issues
- [x] T031 [P] Run tests and verify 80%+ coverage (95%+ for critical paths)
- [x] T032 [P] Update quickstart.md with any additional examples
- [x] T033 [P] Add package-level documentation in internal/tui/doc.go
- [x] T034 [P] Validate all acceptance criteria from spec.md
- [x] T035 [P] Validate SC-001: Verify all package files compile successfully
- [x] T036 [P] Validate SC-003: Test TTY detection in multiple environments (TTY, no-TTY, CI, TERM=dumb)
- [x] T037 [P] Validate SC-005: Import package in test CLI command and verify functionality

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - No dependencies on other stories

### Within Each User Story

- Tests (if included) MUST be written and FAIL before implementation
- Core types/constants before dependent functions
- Implementation before documentation
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- All tests for a user story marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members

---

## Parallel Example: User Story 2

```bash
# Launch all tests for User Story 2 together:
Task: "Unit tests for progress bar rendering in internal/tui/progress_test.go"
Task: "Unit tests for status rendering functions in internal/tui/components_test.go"
Task: "Unit tests for money formatting utilities in internal/tui/render_test.go"

# Launch implementation tasks that can run in parallel:
Task: "Implement progress bar Render method in internal/tui/progress.go"
Task: "Implement status rendering functions in internal/tui/components.go"
Task: "Implement money formatting utilities in internal/tui/render.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo
4. Add User Story 3 → Test independently → Deploy/Demo
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1
   - Developer B: User Story 2
   - Developer C: User Story 3
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence