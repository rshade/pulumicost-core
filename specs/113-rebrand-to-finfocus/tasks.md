# Tasks: Project Rename to FinFocus

**Feature Branch**: `113-rebrand-to-finfocus`
**Spec**: [specs/113-rebrand-to-finfocus/spec.md](spec.md)

## Implementation Strategy

This feature renames the core identity of the project.
- **Atomic Module Rename**: The module path update and directory move will be performed first as a foundational step.
- **Incremental Logic**: Migration and compatibility layers will be added sequentially.
- **Breaking Changes**: JSON output and env var prefixes will be updated, with compatibility toggles where specified.

## Dependencies

1. **Phase 1 (Setup)**: Blocks ALL subsequent phases. Module rename must be stable.
2. **Phase 2 (Foundational)**: Blocks US2, US3, US4. Basic CLI must work.
3. **Phase 3 (US1)**: Independent branding updates.
4. **Phase 4 (US2)**: Migration logic.
5. **Phase 5 (US3)**: Env var logic.
6. **Phase 6 (US4)**: Plugin logic.

## Parallel Execution

- **P1**: US1 (Branding text) and US3 (Env vars) can theoretically proceed in parallel after Phase 1/2 are complete, but US2 (Migration) establishes the directory structure they rely on, so sequential is safer here.
- **P2**: Docs updates (Phase 7) can happen alongside implementation.

---

## Phase 1: Setup & Module Rename (Foundational)

**Goal**: Successfully rename the Go module and binary entry point.

- [X] T001 Rename `go.mod` module path to `github.com/rshade/finfocus` in `go.mod`
- [X] T002 Update all internal imports from `github.com/rshade/finfocus` to `github.com/rshade/finfocus` (global replace)
- [X] T003 Rename directory `cmd/finfocus` to `cmd/finfocus`
- [X] T004 Update `Makefile` to build `finfocus` binary instead of `finfocus` in `Makefile`
- [X] T005 Update `.goreleaser.yaml` to output `finfocus` binary and update `ldflags` in `.goreleaser.yaml`
- [X] T006 Update `package.json` name to `finfocus` in `package.json`
- [X] T007 [P] Update `.github/workflows` to reference `finfocus` binary in `.github/workflows/`

---

## Phase 2: User Story 1 - Branding & CLI (P1)

**Goal**: The CLI reports itself as "FinFocus" in help text and version output.
**Test Criteria**: `finfocus --help` shows "FinFocus", `finfocus --version" works.

- [X] T008 [US1] Create CLI output tests in `cmd/finfocus/root_test.go` ensuring "FinFocus" appears in help/version
- [X] T009 [US1] Update `Use`, `Short`, and `Long` descriptions in `cmd/finfocus/root.go`
- [X] T010 [US1] Update TUI headers and titles to "FinFocus" in `internal/tui/view.go` (and related TUI files)
- [X] T011 [US1] Verify and update version output format in `cmd/finfocus/version.go`
- [X] T012 [US1] Rename Pulumi Analyzer executable detection logic in `cmd/finfocus/main.go` to support `pulumi-analyzer-finfocus`
- [X] T013 [US1] Add persistent "Did you mean alias fin?" reminder in `cmd/finfocus/root.go` (with suppression config)

---

## Phase 3: User Story 2 - Migration Strategy (P1)

**Goal**: Automatically migrate `~/.finfocus` to `~/.finfocus` on startup without data loss.
**Test Criteria**: Start with only `~/.finfocus`, end with both directories populated.

- [X] T014 [US2] Create `internal/migration` package directory
- [X] T015 [US2] Create integration test for migration flow in `internal/migration/integration_test.go`
- [X] T016 [US2] Implement `DetectLegacy` function to check for `~/.finfocus` in `internal/migration/migrator.go`
- [X] T017 [US2] Implement `SafeCopy` function to recursively copy config/state in `internal/migration/migrator.go`
- [X] T018 [US2] Add unit tests for `SafeCopy` using mock filesystem in `internal/migration/migrator_test.go`
- [X] T019 [US2] Integrate migration prompt and execution into `internal/config/loader.go` or `cmd/finfocus/root.go` startup sequence
- [X] T020 [US2] Update default configuration path constants to `~/.finfocus` in `internal/config/paths.go`

---

## Phase 4: User Story 3 - Environment Variables (P2)

**Goal**: Support `FINFOCUS_` prefix and legacy `FINFOCUS_` via toggle.
**Test Criteria**: `FINFOCUS_LOG_LEVEL=debug` works; `FINFOCUS_LOG_LEVEL` works only with compat toggle.

- [X] T021 [US3] Add tests for environment variable precedence and compatibility toggle in `internal/config/config_test.go`
- [X] T022 [US3] Update Viper configuration to use `FINFOCUS` env prefix in `internal/config/loader.go`
- [X] T023 [US3] Implement `LoadCompatEnv` function to read `FINFOCUS_` vars if `FINFOCUS_COMPAT=1` in `internal/config/compat.go`
- [X] T024 [US3] Update all explicit `os.Getenv` calls (if any remain outside Viper) to check `FINFOCUS_` vars in `internal/cli/`

---

## Phase 5: User Story 4 - Plugin Discovery (P2)

**Goal**: Discover plugins in `~/.finfocus/plugins` and support legacy names via toggle.
**Test Criteria**: Plugins found in new path; legacy names found only with toggle.

- [X] T025 [US4] Create plugin discovery tests in `internal/pluginhost/discovery_test.go` covering new path and legacy toggle
- [X] T026 [US4] Update default plugin directory path to `~/.finfocus/plugins` in `internal/pluginhost/discovery.go`
- [X] T027 [US4] Update plugin discovery logic to search for `finfocus-plugin-*` prefix in `internal/pluginhost/discovery.go`
- [X] T028 [US4] Implement legacy discovery logic for `finfocus-plugin-*` when `FINFOCUS_LOG_LEGACY=1` in `internal/pluginhost/discovery.go`
- [X] T029 [US4] Update `plugin install` command to target new directory in `internal/cli/plugin/install.go`

---

## Phase 6: Polish & Cross-Cutting (P3)

**Goal**: Finalize UX details and documentation.

- [X] T030 Update JSON/YAML output root keys to `finfocus` in `internal/cli/format/json.go` (and YAML)
- [X] T031 [P] Update `README.md` with new branding and installation instructions
- [X] T032 [P] Update `docs/` directory content to replace "FinFocus" with "FinFocus"
- [X] T033 Update `GEMINI.md` and `CLAUDE.md` context files with new project identity
