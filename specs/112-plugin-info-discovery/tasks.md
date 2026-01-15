# Tasks: Plugin Info and DryRun Discovery

**Input**: Design documents from `/specs/112-plugin-info-discovery/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/ ✓

**Tests**: Per Constitution Principle II (Test-Driven Development), tests are MANDATORY and must be written BEFORE implementation. All code changes must maintain minimum 80% test coverage (95% for critical paths).

**Completeness**: Per Constitution Principle VI (Implementation Completeness), all tasks MUST be fully implemented. Stub functions, placeholders, and TODO comments are strictly forbidden.

**Documentation**: Per Constitution Principle IV, documentation (README, docs/) MUST be updated concurrently with implementation to prevent drift.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Project type**: Single (extending existing internal packages)
- Paths: `internal/`, `cmd/`, `pkg/` at repository root
- Tests: Adjacent to implementation files (`*_test.go`)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add new dependencies and ensure spec version is accessible

- [x] T001 Add `github.com/Masterminds/semver/v3` dependency via `go get github.com/Masterminds/semver/v3` and run `go mod tidy`
- [x] T002 Verify `pluginsdk.SpecVersion` constant exists in finfocus-spec v0.4.14 (read-only check)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Extend internal interfaces and adapter layer with new RPC methods

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T003 Extend `CostSourceClient` interface in `internal/proto/adapter.go` with `GetPluginInfo(ctx context.Context, in *pbc.Empty, opts ...grpc.CallOption) (*pbc.PluginInfo, error)` method
- [x] T004 Extend `CostSourceClient` interface in `internal/proto/adapter.go` with `DryRun(ctx context.Context, in *pbc.DryRunRequest, opts ...grpc.CallOption) (*pbc.DryRunResponse, error)` method
- [x] T005 Implement `GetPluginInfo` wrapper method in `internal/proto/adapter.go` on `grpcAdapter` struct to call the underlying gRPC client
- [x] T006 Implement `DryRun` wrapper method in `internal/proto/adapter.go` on `grpcAdapter` struct to call the underlying gRPC client
- [x] T007 [P] Define internal `PluginMetadata` struct in `internal/proto/types.go` with fields: Name, Version, SpecVersion, SupportedProviders, Metadata
- [x] T008 [P] Define internal `FieldMapping` struct in `internal/proto/types.go` with fields: FieldName, Status, Condition, ExpectedType
- [x] T009 [P] Add `--skip-version-check` global flag in `cmd/finfocus/root.go` with persistent flag binding

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Version Compatibility Check (Priority: P1) 🎯 MVP

**Goal**: Automatically verify plugin-core protocol version compatibility during initialization

**Independent Test**: Load a plugin with an incompatible `spec_version` and verify warning is logged but execution continues

### Tests for User Story 1 (MANDATORY - TDD Required) ⚠️

> **CONSTITUTION REQUIREMENT: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T010 [P] [US1] Unit test for semver compatibility logic in `internal/pluginhost/version_test.go`: TestCompareSpecVersions_Compatible, TestCompareSpecVersions_Incompatible, TestCompareSpecVersions_MajorMismatch
- [x] T011 [P] [US1] Unit test for GetPluginInfo handling in `internal/pluginhost/client_test.go`: TestGetPluginInfo_Success, TestGetPluginInfo_Unimplemented, TestGetPluginInfo_Timeout
- [x] T012 [P] [US1] Integration test for version check in `test/integration/plugin_version_test.go`: TestPluginInitialization_CompatibleVersion, TestPluginInitialization_IncompatibleVersion_Warning, TestPluginInitialization_LegacyPlugin_NoGetPluginInfo

### Implementation for User Story 1

- [x] T013 [US1] Implement `CompareSpecVersions(coreVersion, pluginVersion string) (CompatibilityResult, error)` function in `internal/pluginhost/version.go` using Masterminds/semver
- [x] T014 [US1] Define `CompatibilityResult` enum (Compatible, MinorMismatch, MajorMismatch, Invalid) in `internal/pluginhost/version.go`
- [x] T015 [US1] Implement gRPC status code checking for `codes.Unimplemented` in `internal/pluginhost/errors.go` with `IsUnimplementedError(err error) bool` helper
- [x] T016 [US1] Update `NewClient` in `internal/pluginhost/host.go` to call `GetPluginInfo` with 5-second timeout during initialization
- [x] T017 [US1] Update `NewClient` in `internal/pluginhost/host.go` to check compatibility result and log warning via zerolog if version mismatch detected
- [x] T018 [US1] Update `NewClient` in `internal/pluginhost/host.go` to respect `--skip-version-check` flag by reading from context or config
- [x] T019 [US1] Handle `codes.Unimplemented` error in `NewClient` for legacy plugins: log debug message and continue without storing metadata

**Checkpoint**: Version compatibility checking is fully functional and testable independently

---

## Phase 4: User Story 2 - Discover Plugin Capabilities (Priority: P2)

**Goal**: Provide `plugin inspect` command to discover FOCUS field mappings via DryRun RPC

**Independent Test**: Run `finfocus plugin inspect aws-public aws:ec2/instance:Instance` and verify field mapping table is displayed

### Tests for User Story 2 (MANDATORY - TDD Required) ⚠️

- [x] T020 [P] [US2] Unit test for DryRun wrapper in `internal/proto/adapter_test.go`: TestDryRun_Success, TestDryRun_Unimplemented, TestDryRun_InvalidResource
- [x] T021 [P] [US2] Unit test for inspect command logic in `internal/cli/plugin_inspect_test.go`: TestInspectCommand_TableOutput, TestInspectCommand_JSONOutput, TestInspectCommand_PluginNotFound
- [x] T022 [P] [US2] Integration test for inspect flow in `test/integration/plugin_inspect_test.go`: TestPluginInspect_WithRealPlugin, TestPluginInspect_LegacyPluginNoDryRun

### Implementation for User Story 2

- [x] T023 [US2] Create new `plugin_inspect.go` file in `internal/cli/` with Cobra command definition: `finfocus plugin inspect <plugin-name> <resource-type>`
- [x] T024 [US2] Add `--json` flag to inspect command in `internal/cli/plugin_inspect.go` for machine-readable output
- [x] T025 [US2] Add `--version` flag to inspect command in `internal/cli/plugin_inspect.go` for specifying plugin version
- [x] T026 [US2] Implement plugin discovery logic in `internal/cli/plugin_inspect.go` to find and launch plugin by name (reuse registry.FindPlugin)
- [x] T027 [US2] Implement DryRun call with 10-second timeout in `internal/cli/plugin_inspect.go` to fetch field mappings
- [x] T028 [US2] Implement table renderer for field mappings in `internal/cli/plugin_inspect.go` with columns: FIELD, STATUS, CONDITION
- [x] T029 [US2] Implement JSON output renderer for field mappings in `internal/cli/plugin_inspect.go` using `encoding/json`
- [x] T030 [US2] Handle `codes.Unimplemented` error in inspect command: display user-friendly message that plugin does not support capability discovery
- [x] T031 [US2] Register `inspectCmd` as subcommand of `pluginCmd` in `internal/cli/plugin.go`

**Checkpoint**: Plugin inspect command is fully functional and testable independently

---

## Phase 5: User Story 3 - Plugin Metadata Display (Priority: P3)

**Goal**: Enhance `plugin list` command to display version and spec version metadata

**Independent Test**: Run `finfocus plugin list` and verify VERSION and SPEC columns appear for all plugins

### Tests for User Story 3 (MANDATORY - TDD Required) ⚠️

- [x] T032 [P] [US3] Unit test for enhanced list output in `internal/cli/plugin_list_test.go`: TestListCommand_ShowsMetadata, TestListCommand_PluginTimeoutHidden, TestListCommand_LegacyPluginNoMetadata
- [x] T033 [P] [US3] Integration test for list metadata in `test/integration/plugin_list_test.go`: TestPluginList_WithMetadata

### Implementation for User Story 3

- [x] T034 [US3] Update `plugin list` command in `internal/cli/plugin_list.go` to call `GetPluginInfo` for each discovered plugin with 5-second timeout
- [x] T035 [US3] Update table renderer in `internal/cli/plugin_list.go` to add VERSION and SPEC columns to output
- [x] T036 [US3] Implement logic in `internal/cli/plugin_list.go` to omit plugins that timeout or fail `GetPluginInfo` call (per FR-007)
- [x] T037 [US3] Handle legacy plugins in `internal/cli/plugin_list.go`: show "N/A" for VERSION and SPEC if GetPluginInfo is unimplemented
- [x] T038 [US3] Add debug logging in `internal/cli/plugin_list.go` for plugin metadata fetch operations

**Checkpoint**: Plugin list command displays metadata and is testable independently

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, cleanup, and validation

- [x] T039 [P] Update `docs/reference/cli-commands.md` with new `plugin inspect` command documentation
- [x] T040 [P] Update `docs/reference/cli-commands.md` with updated `plugin list` output format
- [x] T041 [P] Update `README.md` if plugin management section exists with new capabilities
- [x] T042 [P] Add `--skip-version-check` flag documentation to `docs/reference/cli-commands.md`
- [x] T043 Run `make lint` and fix any linting errors across all modified files
- [x] T044 Run `make test` and ensure all tests pass with 80%+ coverage for new code
- [x] T045 Validate quickstart.md examples work correctly with implemented commands
- [x] T046 Run `make test-integration` to verify integration tests pass

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

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - No dependencies on US1
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - No dependencies on US1/US2

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Interface extensions before implementations
- Core logic before CLI integration
- Story complete before moving to next priority

### Parallel Opportunities

- T007, T008, T009 can run in parallel (different files)
- T010, T011, T012 can run in parallel (US1 tests)
- T020, T021, T022 can run in parallel (US2 tests)
- T032, T033 can run in parallel (US3 tests)
- T039, T040, T041, T042 can run in parallel (documentation)
- All user stories can run in parallel once Phase 2 is complete

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
Task: "Unit test for semver compatibility logic in internal/pluginhost/version_test.go"
Task: "Unit test for GetPluginInfo handling in internal/pluginhost/client_test.go"
Task: "Integration test for version check in test/integration/plugin_version_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (add semver dependency)
2. Complete Phase 2: Foundational (extend interfaces)
3. Complete Phase 3: User Story 1 (version compatibility)
4. **STOP and VALIDATE**: Test version checking independently
5. Deploy/demo if ready - core now warns on incompatible plugins

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Core warns on mismatches (MVP!)
3. Add User Story 2 → Test independently → Developers can inspect capabilities
4. Add User Story 3 → Test independently → Operators see full metadata
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (version check)
   - Developer B: User Story 2 (inspect command)
   - Developer C: User Story 3 (list enhancement)
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Timeouts: 5s for GetPluginInfo, 10s for DryRun (per research.md decisions)
