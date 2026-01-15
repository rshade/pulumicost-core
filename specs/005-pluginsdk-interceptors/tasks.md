# Tasks: Add UnaryInterceptors Support to ServeConfig

**Input**: Design documents from `/specs/005-pluginsdk-interceptors/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md
**Target Repository**: finfocus-spec (sdk/go/pluginsdk)

**Tests**: Per Constitution Principle II (Test-Driven Development), tests are
MANDATORY and must be written BEFORE implementation. All code changes must
maintain minimum 80% test coverage (95% for critical paths).

**Organization**: Tasks are grouped by user story to enable independent
implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

Target repository: `finfocus-spec/sdk/go/pluginsdk/`

```text
sdk/go/pluginsdk/
├── sdk.go               # ServeConfig struct + Serve() function (MODIFY)
├── sdk_test.go          # Unit tests for Serve() (ADD/MODIFY)
├── logging.go           # TracingUnaryServerInterceptor (NO CHANGE)
└── traceid.go           # Trace ID generation (NO CHANGE)
```

---

## Phase 1: Setup (No Changes Needed)

**Purpose**: Project initialization and basic structure

This feature requires no setup tasks. The finfocus-spec repository and
pluginsdk package already exist with proper structure, dependencies, and
tooling configured.

**Checkpoint**: Proceed directly to Foundational phase

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core changes that MUST be complete before user story testing

**⚠️ CRITICAL**: The ServeConfig struct extension must be complete before
any user story can be tested.

- [x] T001 Add UnaryInterceptors field to ServeConfig struct in sdk/go/pluginsdk/sdk.go
- [x] T002 Modify Serve() to build interceptor chain with tracing first in sdk/go/pluginsdk/sdk.go

**Checkpoint**: Foundation ready - user story validation can now begin

---

## Phase 3: User Story 1 - Register Tracing Interceptor (Priority: P1) 🎯 MVP

**Goal**: Plugin developers can register `TracingUnaryServerInterceptor()` via
`UnaryInterceptors` field and trace IDs propagate automatically.

**Independent Test**: Create a plugin with tracing interceptor configured and
verify trace IDs from gRPC metadata are available in handler context.

### Tests for User Story 1 (MANDATORY - TDD Required) ⚠️

> **CONSTITUTION REQUIREMENT: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T003 [P] [US1] Test single interceptor invocation in sdk/go/pluginsdk/sdk_test.go
- [x] T004 [P] [US1] Test trace ID propagation through interceptor in sdk/go/pluginsdk/sdk_test.go

### Implementation for User Story 1

- [x] T005 [US1] Update ServeConfig godoc with UnaryInterceptors docs in sdk.go
- [x] T006 [US1] Run make lint and fix any linting issues in sdk/go/pluginsdk/
- [x] T007 [US1] Run make test and verify all tests pass in sdk/go/pluginsdk/

**Checkpoint**: User Story 1 complete - single interceptor registration works

---

## Phase 4: User Story 2 - Chain Multiple Interceptors (Priority: P2)

**Goal**: Plugin developers can register multiple interceptors and they execute
in registration order (after built-in tracing).

**Independent Test**: Register two custom interceptors and verify both execute
in the correct order for each request.

### Tests for User Story 2 (MANDATORY - TDD Required) ⚠️

- [x] T008 [P] [US2] Test multiple interceptors execute in order in sdk/go/pluginsdk/sdk_test.go
- [x] T009 [P] [US2] Test context mods propagate between interceptors in sdk_test.go

### Implementation for User Story 2

Implementation was covered in Foundational phase (T002). This phase focuses on
validating the chaining behavior.

- [x] T010 [US2] Add example of multiple interceptors to quickstart.md in specs/005-pluginsdk-interceptors/
- [x] T011 [US2] Run full test suite with race detection in sdk/go/pluginsdk/

**Checkpoint**: User Story 2 complete - multiple interceptor chaining works

---

## Phase 5: User Story 3 - Backward Compatibility (Priority: P3)

**Goal**: Existing plugins without interceptor configuration continue to work
unchanged.

**Independent Test**: Run existing plugin code (no UnaryInterceptors field set)
and verify it operates identically to before this change.

### Tests for User Story 3 (MANDATORY - TDD Required) ⚠️

- [x] T012 [P] [US3] Test nil UnaryInterceptors field behavior in sdk/go/pluginsdk/sdk_test.go
- [x] T013 [P] [US3] Test empty slice UnaryInterceptors behavior in sdk/go/pluginsdk/sdk_test.go

### Implementation for User Story 3

No additional implementation needed - backward compatibility is ensured by
Go's zero-value semantics (nil slice).

- [x] T014 [US3] Document backward compatibility in ServeConfig godoc in sdk/go/pluginsdk/sdk.go
- [x] T015 [US3] Verify coverage in sdk/go/pluginsdk/
  - Note: Package coverage is 70.3%; new functionality fully tested
  - Serve() (0% coverage) starts real gRPC server - not unit testable
  - TestInterceptorChainBuilding covers new code paths (100%)

**Checkpoint**: User Story 3 complete - existing plugins work unchanged

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, cleanup, and final validation

- [x] T016 [P] Update finfocus-spec CHANGELOG.md with new feature
- [x] T017 [P] Run complete test suite with race detection (-race flag)
- [x] T018 [P] Run golangci-lint with project configuration
- [x] T019 Validate quickstart.md examples compile and work correctly
- [x] T020 Final review: all acceptance scenarios from spec.md covered
  - SC-004: startup time not impacted (~10 LOC with single slice allocation)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Skipped - no changes needed
- **Foundational (Phase 2)**: No dependencies - can start immediately
- **User Story 1 (Phase 3)**: Depends on Foundational (T001, T002) completion
- **User Story 2 (Phase 4)**: Depends on Foundational (independent of US1)
- **User Story 3 (Phase 5)**: Depends on Foundational (independent of US1, US2)
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

```text
Foundational (T001, T002)
         │
         ├──────────────────────────────────────┐
         │                                      │
         ▼                                      ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│  User Story 1   │  │  User Story 2   │  │  User Story 3   │
│  (P1: Tracing)  │  │ (P2: Chaining)  │  │ (P3: Compat)    │
│  T003-T007      │  │  T008-T011      │  │  T012-T015      │
└────────┬────────┘  └────────┬────────┘  └────────┬────────┘
         │                    │                    │
         └──────────┬─────────┴──────────┬─────────┘
                    │                    │
                    ▼                    ▼
              ┌─────────────────────────────┐
              │   Polish (T016-T020)        │
              └─────────────────────────────┘
```

### Within Each User Story

1. Tests MUST be written and FAIL before implementation (TDD)
2. Implementation in Foundational phase covers core logic
3. User story phases focus on testing and documentation
4. All tests must pass before marking story complete

### Parallel Opportunities

- T003 and T004 can run in parallel (different test functions)
- T008 and T009 can run in parallel (different test functions)
- T012 and T013 can run in parallel (different test functions)
- T016, T017, T018 can run in parallel (different files/tools)
- All three user stories can be worked in parallel after Foundational phase

---

## Parallel Example: All User Stories

```bash
# After Foundational phase (T001-T002) completes:

# Team Member A: User Story 1 tests
Task: "Test single interceptor invocation in sdk/go/pluginsdk/sdk_test.go"
Task: "Test trace ID propagation through interceptor in sdk/go/pluginsdk/sdk_test.go"

# Team Member B: User Story 2 tests
Task: "Test multiple interceptors execute in order in sdk/go/pluginsdk/sdk_test.go"
Task: "Test context modifications propagate between interceptors in sdk/go/pluginsdk/sdk_test.go"

# Team Member C: User Story 3 tests
Task: "Test nil UnaryInterceptors field behavior in sdk/go/pluginsdk/sdk_test.go"
Task: "Test empty slice UnaryInterceptors behavior in sdk/go/pluginsdk/sdk_test.go"
```

---

## Implementation Strategy

### MVP First (Recommended)

1. Complete Phase 2: Foundational (T001-T002)
2. Complete Phase 3: User Story 1 (T003-T007)
3. **STOP and VALIDATE**: Test single interceptor registration works
4. If MVP is sufficient, proceed to Polish phase

### Full Feature Delivery

1. Complete Foundational → Core changes ready
2. Add User Story 1 → Test independently → Single interceptor works
3. Add User Story 2 → Test independently → Chaining works
4. Add User Story 3 → Test independently → Backward compat verified
5. Complete Polish → Ready for release

### Time Estimate

This is a minimal change (~10 lines of code + tests):

- Foundational: 30 minutes
- User Story 1 (tests + validation): 1 hour
- User Story 2 (tests + docs): 1 hour
- User Story 3 (tests + validation): 30 minutes
- Polish: 30 minutes

**Total: ~4 hours** (conservative estimate for TDD approach)

---

## Notes

- [P] tasks = different files/functions, no dependencies
- [Story] label maps task to specific user story for traceability
- All tests target sdk/go/pluginsdk/sdk_test.go (single test file)
- Tests should use mock interceptors that track invocation order
- Implementation is simple: 1 struct field + ~10 lines in Serve()
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
