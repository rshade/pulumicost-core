# Tasks: Adopt pluginsdk/env.go for Environment Variable Handling

**Input**: Design documents from `/specs/014-pluginsdk-env-adoption/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Per Constitution Principle II (Test-Driven Development), tests are MANDATORY and must be written BEFORE implementation. All code changes must maintain minimum 80% test coverage (95% for critical paths).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Single project**: `src/`, `tests/` at repository root
- Paths shown below assume Go project structure - adjust based on plan.md structure

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and dependency updates

- [ ] T001 Update go.mod to require pulumicost-spec v0.4.5+ for pluginsdk/env.go access
- [ ] T002 Run go mod tidy to resolve new dependencies

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Codebase analysis that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T003 Search codebase for hardcoded environment variable strings related to plugin communication
- [ ] T004 Identify all files that need updates for environment variable constants
- [ ] T005 Check if code generator exists (cmd/gen or similar) and identify required changes for pluginsdk integration
  - **Finding**: Code generator (`plugin init`) already uses `pluginsdk.Serve()` which handles port reading internally

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Plugin Communication Consistency (Priority: P1) 🎯 MVP

**Goal**: Ensure core sets standardized environment variables for plugin communication

**Independent Test**: Verify plugins can read port information using pluginsdk.GetPort() without hardcoded fallbacks

### Tests for User Story 1 (MANDATORY - TDD Required) ⚠️

> **CONSTITUTION REQUIREMENT: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T006 [P] [US1] Unit test for environment variable setting in internal/pluginhost/process_test.go
  - Added `TestProcessLauncher_EnvironmentVariableConstants` and `TestProcessLauncher_StartPluginEnvironment`
- [ ] T007 [P] [US1] Integration test for plugin launch with environment variables in test/integration/pluginhost/
  - Covered by existing integration tests in test/integration/plugin/ and new unit tests

### Implementation for User Story 1

- [ ] T008 [US1] Update internal/pluginhost/process.go to import pluginsdk/env.go
- [ ] T009 [US1] Replace hardcoded "PORT" and "PULUMICOST_PLUGIN_PORT" with pluginsdk.EnvPort and local envPortFallback constant in internal/pluginhost/process.go
  - Note: pluginsdk does not define EnvPortFallback, so a local constant `envPortFallback = "PORT"` was added for backward compatibility
- [ ] T010 [US1] Verify plugin communication works with new constants by running existing pluginhost tests and checking no hardcoded strings remain

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Centralized Environment Variable Management (Priority: P2)

**Goal**: Replace all hardcoded environment variable strings with shared constants

**Independent Test**: Search codebase confirms no hardcoded environment variable strings remain

### Tests for User Story 2 (MANDATORY - TDD Required) ⚠️

- [ ] T011 [P] [US2] Unit tests for any updated environment variable usage in affected files
  - Existing tests in internal/config/logging_test.go updated to use pluginsdk constants
- [ ] T012 [P] [US2] Integration test verifying centralized constants work across codebase
  - test/integration/trace_propagation_test.go updated to use pluginsdk.EnvTraceID

### Implementation for User Story 2

- [ ] T013 [P] [US2] Update logging configuration environment variables in relevant files
  - Updated internal/config/logging_test.go to use pluginsdk.EnvLogLevel and pluginsdk.EnvLogFormat
- [ ] T014 [P] [US2] Update trace ID injection environment variables in relevant files
  - Updated test/integration/trace_propagation_test.go to use pluginsdk.EnvTraceID
- [ ] T015 [P] [US2] Update any other identified hardcoded environment variable strings
  - Note: Some config-specific env vars (PULUMICOST_CONFIG_STRICT, PULUMICOST_OUTPUT_*) are NOT in pluginsdk and remain as-is
- [ ] T016 [US2] Verify all environment variable access uses constants by searching codebase with grep for hardcoded env var patterns
  - Verified: Only config-specific variables (not in pluginsdk) remain hardcoded
- [ ] T024 [P] [US2] Update unit tests that mock environment variable names to use constants
  - Updated logging_test.go and trace_propagation_test.go

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Code Generator Updates (Priority: P3)

**Goal**: Update code generator to use pluginsdk functions instead of direct os.Getenv()

**Independent Test**: Generated plugin code uses pluginsdk.GetPort() instead of os.Getenv("PORT")

### Tests for User Story 3 (MANDATORY - TDD Required) ⚠️

- [ ] T017 [P] [US3] Unit test for code generator output using pluginsdk functions
  - **N/A**: Generator already uses pluginsdk.Serve() which handles port internally
- [ ] T018 [P] [US3] Integration test for generated plugin code functionality
  - **N/A**: Generator already uses pluginsdk.Serve() which handles port internally

### Implementation for User Story 3

- [ ] T019 [US3] Locate and examine code generator (cmd/gen or similar)
  - **Found**: `plugin init` command in internal/cli/plugin_init.go
- [ ] T020 [US3] Update code generator to import pluginsdk/env.go
  - **N/A**: Already imports pluginsdk and uses pluginsdk.Serve()
- [ ] T021 [US3] Modify generator to use pluginsdk.GetPort() instead of os.Getenv()
  - **N/A**: Generated code uses `pluginsdk.Serve(ctx, config)` which internally handles port reading
- [ ] T022 [US3] Add comments referencing best practices documentation
  - **N/A**: Generated code already follows best practices by using pluginsdk.Serve()
- [ ] T023 [US3] Test generated code uses new patterns
  - **Verified**: Generated main.go uses pluginsdk.ServeConfig and pluginsdk.Serve()

**Checkpoint**: All user stories should now be independently functional

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final improvements and validation

- [ ] T025 [P] Add integration test verifying environment variable propagation in test/integration/
  - Added tests in internal/pluginhost/process_test.go (TestProcessLauncher_StartPluginEnvironment)
  - Existing tests in test/integration/trace_propagation_test.go verify env var propagation
- [ ] T026 Update CLAUDE.md with patterns for using pluginsdk environment variable constants
  - Added "Environment Variable Constants (pluginsdk)" section with examples
- [ ] T027 Run make lint and make test to ensure all changes pass quality gates
  - All linting passes (golangci-lint + markdownlint)
  - All unit tests pass
- [ ] T028 Update documentation if needed
  - CLAUDE.md updated with pluginsdk patterns
  - No additional documentation changes required

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
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - May integrate with US1 but should be independently testable
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - May integrate with US1/US2 but should be independently testable

### Within Each User Story

- Tests (if included) MUST be written and FAIL before implementation
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- All tests for a user story marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
Task: "Unit test for environment variable setting in internal/pluginhost/process_test.go"
Task: "Integration test for plugin launch with environment variables in test/integration/pluginhost/"

# Launch implementation tasks sequentially:
Task: "Update internal/pluginhost/process.go to import pluginsdk/env.go"
Task: "Replace hardcoded PORT and PULUMICOST_PLUGIN_PORT with pluginsdk constants in internal/pluginhost/process.go"
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
