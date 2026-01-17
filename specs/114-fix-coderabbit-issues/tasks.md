# Actionable Tasks: CodeRabbit Issue Resolution

**Feature Branch**: `114-fix-coderabbit-issues`
**Status**: Pending

## Phase 1: Setup
**Goal**: Initialize shared constants package to support refactoring.

- [x] T001 Create `internal/constants` package and `env.go` file with `EnvAnalyzerMode` constant using value `"FINFOCUS_ANALYZER_MODE"`.

## Phase 2: Foundational
**Goal**: Eliminate duplicate constants to prevent circular dependencies.

- [x] T002 Update `internal/cli/analyzer_serve.go` to import `EnvAnalyzerMode` from `internal/constants`.
- [x] T003 Update `internal/pluginhost/process.go` to import `EnvAnalyzerMode` from `internal/constants`.

## Phase 3: User Story 1 - Reliable Error Handling & Logging (P1)
**Goal**: Ensure errors are properly logged and propagated in CLI and Plugin operations.
**Independent Test**: Simulate failures (closed connection, write error) and verify logs/exit codes.

- [x] T004 [US1] Refactor `defer client.Close()` in `internal/cli/plugin_inspect.go` to log errors at **debug level** (requires logging package).
- [x] T005 [US1] Update `renderTable` signature in `internal/cli/plugin_inspect.go` to return `error`.
- [x] T006 [US1] Propagate all write errors (fmt.Fprintf) in `renderTable` and its call sites in `internal/cli/plugin_inspect.go`.
- [x] T007 [US1] Log plugin launch failures (Debug level) in `internal/cli/plugin_list.go` before context cancellation.
- [x] T008 [US1] Explicitly handle or document `ErrServerStopped` in `internal/pluginhost/client_test.go` (remove empty block).

## Phase 4: User Story 2 - Code Consistency & Maintainability (P2)
**Goal**: Standardize paths, constants, and testing patterns.
**Independent Test**: Verify path resolution uses config dir; linting passes without warnings.

- [x] T009 [US2] Update `findPluginPath` in `internal/cli/plugin_inspect.go` to use `config.New().PluginDir`.
- [x] T010 [US2] Add `//nolint:funlen,gocognit` directive to `GetActualCostWithOptionsAndErrors` in `internal/engine/engine.go`.
- [x] T011 [US2] Add `omitempty` tags to `FieldMapping` struct in `internal/proto/types.go` and add GoDocs for status constants.
- [x] T012 [US2] Update `internal/proto/adapter_test.go` to use `require.Error(t, err)` for failure assertions.

## Phase 5: User Story 3 - Enhanced Documentation & Usability (P3)
**Goal**: Improve user-facing output and documentation.
**Independent Test**: Run `plugin inspect` to see new format; check `README.md` and logs.

- [x] T013 [P] [US3] Implement `String()` method for `CompatibilityResult` in `internal/pluginhost/version.go`.
- [x] T014 [P] [US3] Add unit test for `CompatibilityResult.String()` in `internal/pluginhost/version_test.go`.
- [x] T015 [P] [US3] Add example output (Table and JSON) for `plugin inspect` command to `README.md`.
- [x] T016 [P] [US3] Deduplicate stateless architecture description in `GEMINI.md`.

## Phase 6: Polish & Verification
**Goal**: Final quality checks.

- [x] T017 Run `make lint` to verify all new lint directives and code changes.
- [x] T018 Run `make test` to ensure no regressions in modified packages.

## Dependencies

1. **Setup & Foundational** (T001-T003) MUST complete first to resolve `EnvAnalyzerMode` duplication.
2. **Phase 3** (Error Handling) can proceed after T003.
3. **Phase 4** (Consistency) can run in parallel with Phase 3.
4. **Phase 5** (Docs/Usability) can run in parallel with Phase 3/4.
5. **Phase 6** requires all prior tasks.

## Parallel Execution Opportunities

- T013/T014 (Stringer) are independent of CLI/Engine changes.
- T015/T016 (Markdown docs) can be edited by a separate stream.
- T011/T012 (Proto/Tests) are isolated from CLI logic.

## Implementation Strategy

1. **Fix the foundation**: Solve the constant duplication first.
2. **Secure the CLI**: Fix the swallowed errors to prevent debugging headaches during later work.
3. **Standardize**: Apply path fixes and lint rules.
4. **Refine**: Add the nice-to-have documentation and stringers.