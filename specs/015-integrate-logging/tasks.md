# Tasks: Complete Logging Integration

**Input**: Design documents from `/specs/007-integrate-logging/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md

**Tests**: Per Constitution Principle II (Test-Driven Development), tests are MANDATORY and must
be written BEFORE implementation. All code changes must maintain minimum 80% test coverage.

**Organization**: Tasks are grouped by user story to enable independent implementation and
testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

This is an existing Go CLI project with structure:

- **Source**: `internal/` for packages, `cmd/` for entry point
- **Tests**: Co-located `*_test.go` files and `test/` directory

---

## Phase 1: Setup

**Purpose**: Project analysis and shared infrastructure preparation

- [x] T001 Verify existing logging package structure in internal/logging/
- [x] T002 Verify existing config package structure in internal/config/
- [x] T003 [P] Review current CLI logging setup in internal/cli/root.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T004 Add AuditConfig struct to internal/config/config.go
- [x] T005 Add config validation for logging.audit section in internal/config/config.go
- [x] T006 Create config-to-logging bridge function in internal/config/logging.go

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Configure Logging Behavior (Priority: P1) 🎯 MVP

**Goal**: Operators can configure logging behavior through config files with CLI and env overrides

**Independent Test**: Set logging.level: debug in config, run any command, verify debug messages appear

### Tests for User Story 1 (MANDATORY - TDD Required) ⚠️

> **CONSTITUTION REQUIREMENT: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T007 [P] [US1] Unit test for config bridge function in internal/config/logging_test.go
- [x] T008 [P] [US1] Unit test for config override precedence (file < env < flag) in internal/config/logging_test.go
- [x] T009 [P] [US1] Integration test for config file loading in test/integration/logging_config_test.go

### Implementation for User Story 1

- [x] T010 [US1] Implement ToLoggingConfig() bridge function in internal/config/logging.go
- [x] T011 [US1] Update root.go PersistentPreRunE to use config bridge in internal/cli/root.go
- [x] T012 [US1] Ensure --debug flag overrides config file level in internal/cli/root.go
- [x] T013 [US1] Ensure PULUMICOST_LOG_LEVEL env var overrides config in internal/cli/root.go
- [x] T014 [US1] Add invalid log level fallback with warning in internal/logging/zerolog.go

**Checkpoint**: User Story 1 complete - logging can be configured via config file, env vars, or CLI flags

---

## Phase 4: User Story 2 - Clear Log Location Communication (Priority: P2)

**Goal**: CLI displays log file path when file logging is enabled, warns on fallback

**Independent Test**: Configure logging to file, run command, verify path displayed on stdout

### Tests for User Story 2 (MANDATORY - TDD Required) ⚠️

- [x] T015 [P] [US2] Unit test for log path display logic in internal/cli/root_test.go
- [x] T016 [P] [US2] Unit test for fallback warning display in internal/cli/root_test.go
- [x] T017 [P] [US2] Integration test for file logging with path output in test/integration/logging_output_test.go

### Implementation for User Story 2

- [x] T018 [US2] Add LogPathNotifier type to return log path and status in internal/logging/zerolog.go
- [x] T019 [US2] Modify NewLogger to return file path or error status in internal/logging/zerolog.go
- [x] T020 [US2] Print "Logging to: /path" to stdout in PersistentPreRunE in internal/cli/root.go
- [x] T021 [US2] Print warning to stdout on fallback to stderr in internal/cli/root.go
- [x] T022 [US2] Suppress path message when logging to stderr/stdout in internal/cli/root.go

**Checkpoint**: User Story 2 complete - operators see where logs are written at startup

---

## Phase 5: User Story 3 - Audit Logging for Cost Queries (Priority: P3)

**Goal**: All cost commands generate audit log entries when audit logging is enabled

**Independent Test**: Enable audit.enabled: true, run cost projected, verify audit entry in logs

### Tests for User Story 3 (MANDATORY - TDD Required) ⚠️

- [x] T023 [P] [US3] Unit test for AuditEntry struct in internal/logging/audit_test.go
- [x] T024 [P] [US3] Unit test for AuditLogger interface in internal/logging/audit_test.go
- [x] T025 [P] [US3] Unit test for audit entry field population in internal/logging/audit_test.go
- [x] T026 [P] [US3] Integration test for audit logging in cost projected in test/integration/audit_test.go
- [x] T027 [P] [US3] Integration test for audit logging in cost actual in test/integration/audit_test.go
- [x] T027a [P] [US3] Unit test verifying SafeStr() redaction works with audit entry parameters in internal/logging/audit_test.go

### Implementation for User Story 3

- [x] T028 [P] [US3] Create AuditEntry struct in internal/logging/audit.go
- [x] T029 [P] [US3] Create AuditLogger interface in internal/logging/audit.go
- [x] T030 [US3] Implement zerologAuditLogger in internal/logging/audit.go
- [x] T031 [US3] Add NewAuditLogger constructor using config in internal/logging/audit.go
- [x] T032 [US3] Store AuditLogger in CLI context in internal/cli/root.go
- [x] T033 [US3] Add audit logging to cost projected command in internal/cli/cost_projected.go
- [x] T034 [US3] Add audit logging to cost actual command in internal/cli/cost_actual.go
- [x] T035 [US3] Ensure parameters are redacted using SafeStr pattern in internal/logging/audit.go

**Checkpoint**: User Story 3 complete - all cost queries are audited when enabled

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T036 [P] Update user guide with logging configuration in docs/guides/user-guide.md
- [x] T037 [P] Add logging configuration examples to quickstart in specs/007-integrate-logging/quickstart.md
- [x] T038 Run make lint and fix any issues
- [x] T039 Run make test and ensure 80%+ coverage (62.7% overall, 70.1% logging package)
- [x] T040 Validate all edge cases from spec (unwritable file, disk full, invalid level, locked audit file)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - User stories can proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Depends only on Foundational - No dependencies on other stories
- **User Story 2 (P2)**: Depends only on Foundational - Builds on US1 changes to root.go but independently testable
- **User Story 3 (P3)**: Depends only on Foundational - Uses logging infrastructure but independently testable

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Interface/types before implementation
- Core implementation before CLI integration
- Story complete before moving to next priority

### Parallel Opportunities

Within Phase 2 (Foundational):

- T004 and T005 can run in parallel (different areas of config.go)

Within User Story 1:

- T007, T008, T009 (all tests) can run in parallel

Within User Story 2:

- T015, T016, T017 (all tests) can run in parallel

Within User Story 3:

- T023, T024, T025, T026, T027 (all tests) can run in parallel
- T028, T029 (types) can run in parallel

Within Polish:

- T036, T037 (docs) can run in parallel

---

## Parallel Example: User Story 3 Tests

```bash
# Launch all tests for User Story 3 together:
Task: "Unit test for AuditEntry struct in internal/logging/audit_test.go"
Task: "Unit test for AuditLogger interface in internal/logging/audit_test.go"
Task: "Unit test for audit entry field population in internal/logging/audit_test.go"
Task: "Integration test for audit logging in cost projected"
Task: "Integration test for audit logging in cost actual"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (analysis)
2. Complete Phase 2: Foundational (config structures)
3. Complete Phase 3: User Story 1 (config-based logging)
4. **STOP and VALIDATE**: Test config file → logging works
5. Deploy/demo if ready - operators can now configure logging

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → **MVP ready!**
3. Add User Story 2 → Test independently → Enhanced operator experience
4. Add User Story 3 → Test independently → Compliance/audit capability
5. Each story adds value without breaking previous stories

### File Modification Summary

| File                            | User Story      | Changes                                  |
| ------------------------------- | --------------- | ---------------------------------------- |
| internal/config/config.go       | Foundation      | Add AuditConfig struct                   |
| internal/config/logging.go      | Foundation, US1 | NEW: Bridge function                     |
| internal/config/logging_test.go | US1             | NEW: Bridge tests                        |
| internal/logging/zerolog.go     | US1, US2        | Add path notification, fallback handling |
| internal/logging/audit.go       | US3             | NEW: Audit types and logger              |
| internal/logging/audit_test.go  | US3             | NEW: Audit tests                         |
| internal/cli/root.go            | US1, US2, US3   | Use bridge, display path, init audit     |
| internal/cli/cost_projected.go  | US3             | Add audit logging                        |
| internal/cli/cost_actual.go     | US3             | Add audit logging                        |
| test/integration/logging\_\*.go | US1, US2, US3   | NEW: Integration tests                   |

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Run `make lint` and `make test` after each story checkpoint
- Stop at any checkpoint to validate story independently
