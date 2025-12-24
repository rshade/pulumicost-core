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

- [x] T001 Update go.mod to require pulumicost-spec v0.4.5+ for pluginsdk/env.go access
- [x] T002 Run go mod tidy to resolve new dependencies

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Codebase analysis that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T003 Search codebase for hardcoded environment variable strings related to plugin communication
- [x] T004 Identify all files that need updates for environment variable constants
- [x] T005 Check if code generator exists (cmd/gen or similar) and identify required changes for pluginsdk integration
  - **Finding**: Code generator (`plugin init`) already uses `pluginsdk.Serve()` which handles port reading internally

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Plugin Communication Consistency (Priority: P1) 🎯 MVP

**Goal**: Ensure core sets standardized environment variables for plugin communication

**Independent Test**: Verify plugins can read port information using pluginsdk.GetPort() without hardcoded fallbacks

### Tests for User Story 1 (MANDATORY - TDD Required) ⚠️

> **CONSTITUTION REQUIREMENT: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T006 [P] [US1] Unit test for environment variable setting in internal/pluginhost/process_test.go
  - Added `TestProcessLauncher_EnvironmentVariableConstants` and `TestProcessLauncher_StartPluginEnvironment`
- [x] T007 [P] [US1] Integration test for plugin launch with environment variables in test/integration/pluginhost/
  - Covered by existing integration tests in test/integration/plugin/ and new unit tests

### Implementation for User Story 1

- [x] T008 [US1] Update internal/pluginhost/process.go to import pluginsdk/env.go
- [x] T009 [US1] Replace hardcoded "PORT" and "PULUMICOST_PLUGIN_PORT" with pluginsdk.EnvPort and local envPortFallback constant in internal/pluginhost/process.go
  - Note: pluginsdk does not define EnvPortFallback, so a local constant `envPortFallback = "PORT"` was added for backward compatibility
- [x] T010 [US1] Verify plugin communication works with new constants by running existing pluginhost tests and checking no hardcoded strings remain

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Centralized Environment Variable Management (Priority: P2)

**Goal**: Replace all hardcoded environment variable strings with shared constants

**Independent Test**: Search codebase confirms no hardcoded environment variable strings remain

### Tests for User Story 2 (MANDATORY - TDD Required) ⚠️

- [x] T011 [P] [US2] Unit tests for any updated environment variable usage in affected files
  - Existing tests in internal/config/logging_test.go updated to use pluginsdk constants
- [x] T012 [P] [US2] Integration test verifying centralized constants work across codebase
  - test/integration/trace_propagation_test.go updated to use pluginsdk.EnvTraceID

### Implementation for User Story 2

- [x] T013 [P] [US2] Update logging configuration environment variables in relevant files
  - Updated internal/config/logging_test.go to use pluginsdk.EnvLogLevel and pluginsdk.EnvLogFormat
- [x] T014 [P] [US2] Update trace ID injection environment variables in relevant files
  - Updated test/integration/trace_propagation_test.go to use pluginsdk.EnvTraceID
- [x] T015 [P] [US2] Update any other identified hardcoded environment variable strings
  - Note: Some config-specific env vars (PULUMICOST_CONFIG_STRICT, PULUMICOST_OUTPUT_*) are NOT in pluginsdk and remain as-is
- [x] T016 [US2] Verify all environment variable access uses constants by searching codebase with grep for hardcoded env var patterns
  - Verified: Only config-specific variables (not in pluginsdk) remain hardcoded
- [x] T024 [P] [US2] Update unit tests that mock environment variable names to use constants
  - Updated logging_test.go and trace_propagation_test.go

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Code Generator Updates (Priority: P3)

**Goal**: Update code generator to use pluginsdk functions instead of direct os.Getenv()

**Independent Test**: Generated plugin code uses pluginsdk.GetPort() instead of os.Getenv("PORT")

### Tests for User Story 3 (MANDATORY - TDD Required) ⚠️

- [x] T017 [P] [US3] Unit test for code generator output using pluginsdk functions
  - **N/A**: Generator already uses pluginsdk.Serve() which handles port internally
- [x] T018 [P] [US3] Integration test for generated plugin code functionality
  - **N/A**: Generator already uses pluginsdk.Serve() which handles port internally

### Implementation for User Story 3

- [x] T019 [US3] Locate and examine code generator (cmd/gen or similar)
  - **Found**: `plugin init` command in internal/cli/plugin_init.go
- [x] T020 [US3] Update code generator to import pluginsdk/env.go
  - **N/A**: Already imports pluginsdk and uses pluginsdk.Serve()
- [x] T021 [US3] Modify generator to use pluginsdk.GetPort() instead of os.Getenv()
  - **N/A**: Generated code uses `pluginsdk.Serve(ctx, config)` which internally handles port reading
- [x] T022 [US3] Add comments referencing best practices documentation
  - **N/A**: Generated code already follows best practices by using pluginsdk.Serve()
- [x] T023 [US3] Test generated code uses new patterns
  - **Verified**: Generated main.go uses pluginsdk.ServeConfig and pluginsdk.Serve()

**Checkpoint**: All user stories should now be independently functional

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final improvements and validation

- [x] T025 [P] Add integration test verifying environment variable propagation in test/integration/
  - Added tests in internal/pluginhost/process_test.go (TestProcessLauncher_StartPluginEnvironment)
  - Existing tests in test/integration/trace_propagation_test.go verify env var propagation
- [x] T026 Update CLAUDE.md with patterns for using pluginsdk environment variable constants
  - Added "Environment Variable Constants (pluginsdk)" section with examples
- [x] T027 Run make lint and make test to ensure all changes pass quality gates
  - All linting passes (golangci-lint + markdownlint)
  - All unit tests pass
- [x] T028 Update documentation if needed
  - CLAUDE.md updated with pluginsdk patterns
  - No additional documentation changes required