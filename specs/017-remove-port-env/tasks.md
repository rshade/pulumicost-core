# Tasks: Remove PORT Environment Variable

**Input**: Design documents from `/specs/017-remove-port-env/`
**Prerequisites**: plan.md (complete), spec.md (complete), research.md (complete)

**Tests**: Per Constitution Principle II (Test-Driven Development), tests are MANDATORY and must be written BEFORE implementation. All code changes must maintain minimum 80% test coverage (95% for critical paths).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Project structure**: `internal/pluginhost/` at repository root
- Tests in same package (white-box testing pattern)

---

## Phase 1: Setup (Verification)

**Purpose**: Verify external dependency and prepare for implementation

- [x] T001 Verify pulumicost-spec#129 is merged (external dependency check)
- [x] T002 [P] Update pulumicost-spec dependency if needed via `go get github.com/rshade/pulumicost-spec@latest`
- [x] T003 [P] Verify pluginsdk.Serve() handles --port flag by checking SDK documentation

---

## Phase 2: Foundational (Test Updates First - TDD)

**Purpose**: Update tests BEFORE implementation per TDD requirement

**CRITICAL**: Tests must be updated to expect new behavior BEFORE code changes

- [x] T004 Update `createEnvCheckingScript()` to verify PORT is NOT set in `internal/pluginhost/process_test.go`
- [x] T005 Update `TestProcessLauncher_StartPluginEnvironment` assertions to verify PORT is NOT set (validates FR-002, FR-005) in `internal/pluginhost/process_test.go`
- [x] T006 [P] Add test for DEBUG logging when PORT detected in user environment in `internal/pluginhost/process_test.go`
- [x] T007 [P] Add test for guidance logging when plugin fails to bind in `internal/pluginhost/process_test.go`
- [x] T008 Run tests - confirm they FAIL (expected, validates TDD approach)

**Checkpoint**: Tests updated and failing - ready for implementation

---

## Phase 3: User Story 1 - Plugin Developer Using Standard Port Flag (Priority: P1)

**Goal**: Remove PORT env var, ensure --port flag is sole authoritative mechanism

**Independent Test**: Launch plugin with --port flag, verify it binds to correct port without PORT env var interference

### Implementation for User Story 1

- [x] T009 [US1] Remove `envPortFallback` constant (line 43) in `internal/pluginhost/process.go`
- [x] T010 [US1] Remove PORT env var line (line 374) from `startPlugin()` in `internal/pluginhost/process.go`
- [x] T011 [US1] Add DEBUG logging for PORT detection in `startPlugin()` in `internal/pluginhost/process.go`
- [x] T012 [US1] Update comments (lines 368-372) to remove PORT references in `internal/pluginhost/process.go`
- [x] T013 [US1] Run `go test ./internal/pluginhost/...` - verify tests pass

**Checkpoint**: User Story 1 complete - PORT removed, --port is authoritative

---

## Phase 4: User Story 2 - PULUMICOST_PLUGIN_PORT Fallback (Priority: P2)

**Goal**: Verify PULUMICOST_PLUGIN_PORT is still set for backward compatibility/debugging

**Independent Test**: Launch plugin, verify PULUMICOST_PLUGIN_PORT matches --port value while PORT is not set

**Note**: T015 extends T005's test changes (from Phase 2) to verify PULUMICOST_PLUGIN_PORT is set alongside the PORT-not-set assertion. This is verification, not duplication.

### Implementation for User Story 2

- [x] T014 [US2] Verify PULUMICOST_PLUGIN_PORT line (line 375) is preserved in `internal/pluginhost/process.go`
- [x] T015 [US2] Extend test assertions in `TestProcessLauncher_StartPluginEnvironment` to also verify PULUMICOST_PLUGIN_PORT is set in `internal/pluginhost/process_test.go`
- [x] T016 [US2] Run `go test ./internal/pluginhost/...` - verify tests pass

**Checkpoint**: User Story 2 complete - PULUMICOST_PLUGIN_PORT verified

---

## Phase 5: User Story 3 - Multi-Plugin Cost Calculation (Priority: P3)

**Goal**: Verify multi-plugin scenarios work with unique ports via --port flag

**Independent Test**: Launch two plugins in parallel, verify each receives unique port

### Implementation for User Story 3

- [x] T017 [US3] Add guidance logging in `startOnce()` error path (line 219) when plugin fails to bind in `internal/pluginhost/process.go`
- [x] T018 [US3] Verify existing `TestProcessLauncher_ConcurrentPortAllocation` still passes (validates FR-004: unique ports) in `internal/pluginhost/process_test.go`
- [x] T019 [US3] Run integration tests `go test ./test/integration/plugin/...` - verify tests pass

**Checkpoint**: User Story 3 complete - multi-plugin scenarios verified

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, validation, and cleanup

- [x] T020 [P] Update `internal/pluginhost/CLAUDE.md` to document new PORT behavior
- [x] T021 [P] Update main `CLAUDE.md` Environment Variables section if needed
- [x] T022 Run `make lint` - verify no linting errors
- [x] T023 Run `make test` - verify all tests pass
- [x] T024 Run `make test-integration` - verify integration tests pass
- [x] T025 Verify code coverage meets 80% minimum for `internal/pluginhost/`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - must complete first (external dependency check)
- **Foundational (Phase 2)**: Depends on Setup - update tests before implementation (TDD)
- **User Story 1 (Phase 3)**: Depends on Foundational - core implementation
- **User Story 2 (Phase 4)**: Depends on User Story 1 - verification of PULUMICOST_PLUGIN_PORT
- **User Story 3 (Phase 5)**: Depends on User Story 1 - multi-plugin verification
- **Polish (Phase 6)**: Depends on all user stories

### User Story Dependencies

- **User Story 1 (P1)**: Core change - all other stories depend on this
- **User Story 2 (P2)**: Verifies PULUMICOST_PLUGIN_PORT still works after US1
- **User Story 3 (P3)**: Verifies multi-plugin works after US1

### Within Each Phase

- Tests MUST be written and FAIL before implementation (Phase 2)
- Implementation follows test updates (Phase 3+)
- Verification after each user story

### Parallel Opportunities

- T002, T003 can run in parallel (Setup verification)
- T006, T007 can run in parallel (adding new tests)
- T020, T021 can run in parallel (documentation updates)

---

## Parallel Example: Foundational Phase

```bash
# Launch new test additions together:
Task: "Add test for DEBUG logging when PORT detected" (T006)
Task: "Add test for guidance logging when plugin fails" (T007)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup verification
2. Complete Phase 2: Update tests (TDD)
3. Complete Phase 3: User Story 1 (remove PORT)
4. **STOP and VALIDATE**: Run `make test` to verify

### Incremental Delivery

1. Setup + Foundational → Tests ready
2. User Story 1 → Core change complete → `make test`
3. User Story 2 → Backward compat verified → `make test`
4. User Story 3 → Multi-plugin verified → `make test`
5. Polish → Documentation and final validation

---

## Notes

- This is a focused refactoring: ~20 lines of production code, ~100 lines of test updates
- External dependency on pulumicost-spec#129 MUST be verified before starting
- All changes are in `internal/pluginhost/` package
- No new files needed - updates to existing files only
- Commit after each user story for clear git history
