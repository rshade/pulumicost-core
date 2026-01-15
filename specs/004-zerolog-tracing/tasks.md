# Tasks: Zerolog Distributed Tracing

**Input**: Design documents from `/specs/004-zerolog-tracing/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md,
quickstart.md

**Tests**: Per Constitution Principle II (Test-Driven Development), tests are MANDATORY and must be
written BEFORE implementation. All code changes must maintain minimum 80% test coverage (95% for
critical paths).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of
each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

- **Go Project**: `internal/` for packages, `cmd/` for entry points, `test/` for tests
- **Logging Package**: `internal/logging/` (primary changes)
- **Integration Tests**: `test/integration/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization, dependency management, and replace existing unused slog implementation

- [x] T001 Add github.com/oklog/ulid/v2 dependency to go.mod
- [x] T002 Remove unused slog implementation in internal/logging/logger.go
- [x] T003 Remove unused slog tests in internal/logging/logger_test.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core logging infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Tests for Foundational Phase (MANDATORY - TDD Required) ⚠️

> **CONSTITUTION REQUIREMENT: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T004 [P] Unit test for TracingHook trace ID injection in internal/logging/zerolog_test.go
- [x] T005 [P] Unit test for NewLogger factory function in internal/logging/zerolog_test.go
- [x] T006 [P] Unit test for ULID trace ID generation in internal/logging/zerolog_test.go
- [x] T007 [P] Unit test for context helpers (FromContext, ContextWithTraceID) in internal/logging/zerolog_test.go
- [x] T008 [P] Unit test for log level parsing in internal/logging/zerolog_test.go
- [x] T009 [P] Unit test for sensitive data protection patterns in internal/logging/zerolog_test.go

### Implementation for Foundational Phase

- [x] T010 Create TracingHook struct implementing zerolog.Hook in internal/logging/zerolog.go
- [x] T011 Implement NewLogger factory function with format/level configuration in internal/logging/zerolog.go
- [x] T012 Implement GenerateTraceID using oklog/ulid/v2 in internal/logging/zerolog.go
- [x] T013 Implement GetOrGenerateTraceID (checks env, context, generates new) in internal/logging/zerolog.go
- [x] T014 Implement ContextWithTraceID and FromContext helpers in internal/logging/zerolog.go
- [x] T015 Implement parseLevel function mapping string to zerolog.Level in internal/logging/zerolog.go
- [x] T016 Implement createWriter function for JSON/console output in internal/logging/zerolog.go
- [x] T017 Implement isSensitiveKey function with blocklist patterns in internal/logging/zerolog.go
- [x] T018 Implement SafeStr helper for redacting sensitive values in internal/logging/zerolog.go
- [x] T019 Define traceIDKey context key type in internal/logging/zerolog.go

**Checkpoint**: Foundation ready - logging infrastructure complete, all unit tests passing

---

## Phase 3: User Story 1 - Debugging Failed Cost Calculations (Priority: P1) 🎯 MVP

**Goal**: Enable users to run any CLI command with `--debug` flag and see complete decision flow from
command start to finish, including plugin lookup attempts, fallback decisions, and duration.

**Independent Test**: Run `finfocus cost projected --debug --pulumi-json examples/plans/aws-simple-plan.json`
and verify structured log output shows command start, resource ingestion, plugin lookups, cost
calculations, and command completion with duration.

### Tests for User Story 1 (MANDATORY - TDD Required) ⚠️

- [x] T020 [P] [US1] Unit test for --debug flag enabling DEBUG level in internal/cli/root_test.go
- [x] T021 [P] [US1] Unit test for trace ID generation at command start in internal/cli/root_test.go
- [x] T022 [P] [US1] Unit test for command duration logging in internal/cli/cost_projected_test.go
- [x] T023 [P] [US1] Integration test for full debug output flow in test/integration/logging_test.go

### Implementation for User Story 1

- [x] T024 [US1] Add --debug persistent flag to root command in internal/cli/root.go
- [x] T025 [US1] Initialize logger in PersistentPreRunE with config in internal/cli/root.go
- [x] T026 [US1] Generate trace ID at command start and store in context in internal/cli/root.go
- [x] T027 [US1] Add command start logging with trace_id, component=cli in internal/cli/root.go
- [x] T028 [US1] Add cost_projected command logging (start, duration) in internal/cli/cost_projected.go
- [x] T029 [US1] Add cost_actual command logging (start, duration) in internal/cli/cost_actual.go
- [x] T030 [US1] Add resource ingestion logging in internal/ingest/pulumi_plan.go
- [x] T031 [US1] Add plugin lookup logging with resource_type in internal/registry/registry.go
- [x] T032 [US1] Add cost calculation logging with duration_ms in internal/engine/engine.go
- [x] T033 [US1] Add fallback decision logging (WARN level) when plugin returns no price in internal/engine/engine.go
- [x] T034 [US1] Add spec loading logging in internal/spec/loader.go
- [x] T035 [US1] Initialize main logger at startup in cmd/finfocus/main.go

**Checkpoint**: User Story 1 complete - debug mode shows full decision flow, independently testable

---

## Phase 4: User Story 2 - Correlating Logs Across Plugin Boundaries (Priority: P2)

**Goal**: Enable trace ID propagation from CLI through gRPC to plugins, allowing end-to-end request
tracing across process boundaries.

**Independent Test**: Run a command with an installed plugin, capture logs from both core and plugin,
verify same trace ID appears in both log streams.

### Tests for User Story 2 (MANDATORY - TDD Required) ⚠️

- [x] T036 [P] [US2] Unit test for gRPC unary client interceptor in internal/pluginhost/grpc_test.go
- [x] T037 [P] [US2] Unit test for trace ID metadata injection in internal/pluginhost/grpc_test.go
- [x] T038 [P] [US2] Integration test for trace propagation to mock plugin in test/integration/trace_propagation_test.go

### Implementation for User Story 2

- [x] T039 [US2] Create TraceInterceptor function returning grpc.UnaryClientInterceptor in internal/pluginhost/grpc.go
- [x] T040 [US2] Define TraceIDMetadataKey constant as "x-finfocus-trace-id" in internal/pluginhost/grpc.go
- [x] T041 [US2] Apply interceptor when creating gRPC connection in internal/pluginhost/process.go
- [x] T042 [US2] Add plugin connection lifecycle logging (connect, disconnect, errors) in internal/pluginhost/process.go
- [x] T043 [US2] Add gRPC call logging with method name in internal/pluginhost/grpc.go

**Checkpoint**: User Story 2 complete - trace IDs propagate to plugins, independently testable

---

## Phase 5: User Story 3 - Configuring Log Verbosity (Priority: P3)

**Goal**: Allow operators to configure log level and format via config file, environment variables,
or CLI flags with proper precedence (CLI > env > config > default).

**Independent Test**: Set `FINFOCUS_LOG_LEVEL=error` and config file level to debug, verify only
ERROR logs appear (environment takes precedence over config file).

### Tests for User Story 3 (MANDATORY - TDD Required) ⚠️

- [x] T044 [P] [US3] Unit test for configuration precedence (CLI > env > config) in internal/config/config_test.go
- [x] T045 [P] [US3] Unit test for FINFOCUS_LOG_LEVEL environment variable in internal/config/config_test.go
- [x] T046 [P] [US3] Unit test for FINFOCUS_LOG_FORMAT environment variable in internal/config/config_test.go
- [x] T047 [P] [US3] Unit test for invalid log level fallback to INFO in internal/config/config_test.go

### Implementation for User Story 3

- [x] T048 [US3] Extend LoggingConfig struct with all fields (level, format, output, file, caller, stack_trace) in internal/config/config.go
- [x] T049 [US3] Implement resolveLogLevel function checking CLI > env > config > default in internal/config/config.go
- [x] T050 [US3] Implement resolveLogFormat function checking env > config > default in internal/config/config.go
- [x] T051 [US3] Add FINFOCUS_LOG_LEVEL environment variable support in internal/config/config.go
- [x] T052 [US3] Add FINFOCUS_LOG_FORMAT environment variable support in internal/config/config.go
- [x] T053 [US3] Add file output support with fallback to stderr in internal/logging/zerolog.go
- [x] T054 [US3] Add invalid log level warning and fallback in internal/logging/zerolog.go
- [x] T055 [US3] Wire config resolution to logger initialization in internal/cli/root.go (call resolveLogLevel/resolveLogFormat, pass to NewLogger)

**Checkpoint**: User Story 3 complete - configuration precedence works correctly, independently testable

---

## Phase 6: User Story 4 - Injecting External Trace IDs (Priority: P4)

**Goal**: Allow enterprise users to inject their pipeline's trace ID via FINFOCUS_TRACE_ID environment
variable for correlation with broader observability systems.

**Independent Test**: Set `FINFOCUS_TRACE_ID=external-trace-12345` and run any command, verify all
log entries use "external-trace-12345" as trace_id value.

### Tests for User Story 4 (MANDATORY - TDD Required) ⚠️

- [x] T056 [P] [US4] Unit test for FINFOCUS_TRACE_ID environment variable override in internal/logging/zerolog_test.go
- [x] T057 [P] [US4] Unit test for external trace ID appearing in all log entries in internal/logging/zerolog_test.go
- [x] T058 [P] [US4] Integration test validating external trace ID flow in test/integration/trace_propagation_test.go

### Implementation for User Story 4

- [x] T059 [US4] Check FINFOCUS_TRACE_ID before generating new trace ID in internal/logging/zerolog.go

**Checkpoint**: User Story 4 complete - external trace IDs properly injected, independently testable

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories, documentation, and quality assurance

- [x] T060 [P] Add component sub-logger initialization for cli package in internal/cli/root.go
- [x] T061 [P] Add component sub-logger initialization for engine package in internal/engine/engine.go
- [x] T062 [P] Add component sub-logger initialization for registry package in internal/registry/registry.go
- [x] T063 [P] Add component sub-logger initialization for pluginhost package in internal/pluginhost/process.go
- [x] T064 [P] Add component sub-logger initialization for ingest package in internal/ingest/pulumi_plan.go
- [x] T065 [P] Add component sub-logger initialization for spec package in internal/spec/loader.go
- [x] T066 [P] Add component sub-logger initialization for config package in internal/config/config.go
- [x] T067 Update user documentation with --debug flag examples in docs/guides/user-guide.md
- [x] T068 Update developer documentation with logging patterns in docs/guides/developer-guide.md
- [x] T069 Run make lint and fix any issues
- [x] T070 Run make test and ensure 80% coverage minimum
- [x] T071 Run quickstart.md validation scenarios
- [x] T072 Update CLAUDE.md with zerolog logging patterns

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - User stories can proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3 → P4)
- **Polish (Phase 7)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Integrates with US1 logging infrastructure
  but independently testable
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - Uses logger factory from US1 but
  independently testable
- **User Story 4 (P4)**: Can start after Foundational (Phase 2) - Uses trace ID generation from US1
  but independently testable

### Within Each User Story

- Tests (if included) MUST be written and FAIL before implementation
- Infrastructure tasks before feature tasks
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks can run sequentially (small scope)
- All Foundational tests marked [P] can run in parallel
- All Foundational implementation tasks can run sequentially (same file: zerolog.go)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- All tests for a user story marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members
- All component sub-logger tasks in Phase 7 marked [P] can run in parallel

---

## Parallel Example: Foundational Phase

```bash
# Launch all foundational tests together (parallel - different test functions):
Task: T004 "Unit test for TracingHook trace ID injection in internal/logging/zerolog_test.go"
Task: T005 "Unit test for NewLogger factory function in internal/logging/zerolog_test.go"
Task: T006 "Unit test for ULID trace ID generation in internal/logging/zerolog_test.go"
Task: T007 "Unit test for context helpers in internal/logging/zerolog_test.go"
Task: T008 "Unit test for log level parsing in internal/logging/zerolog_test.go"
Task: T009 "Unit test for sensitive data protection in internal/logging/zerolog_test.go"
```

---

## Parallel Example: User Story 1

```bash
# Launch all US1 tests together (parallel - different test files):
Task: T020 "Unit test for --debug flag in internal/cli/root_test.go"
Task: T021 "Unit test for trace ID generation at command start in internal/cli/root_test.go"
Task: T022 "Unit test for command duration logging in internal/cli/cost_projected_test.go"
Task: T023 "Integration test for full debug output flow in test/integration/logging_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test `finfocus cost projected --debug` independently
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP! Users can debug commands)
3. Add User Story 2 → Test independently → Deploy/Demo (Plugin correlation enabled)
4. Add User Story 3 → Test independently → Deploy/Demo (Configuration flexibility)
5. Add User Story 4 → Test independently → Deploy/Demo (Enterprise integration)
6. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (CLI --debug flag)
   - Developer B: User Story 2 (gRPC trace propagation)
   - Developer C: User Story 3 (Configuration)
   - Developer D: User Story 4 (External trace IDs)
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- internal/logging/zerolog.go is the primary new file (most foundational tasks there)
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
