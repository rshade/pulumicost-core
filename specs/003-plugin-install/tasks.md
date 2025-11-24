# Tasks: Plugin Install/Update/Remove System

**Input**: Design documents from `/specs/001-plugin-install/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Per Constitution Principle II (Test-Driven Development), tests are MANDATORY and must be written BEFORE implementation. All code changes must maintain minimum 80% test coverage (95% for critical paths).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Project Structure**: `internal/` for Go packages, `registry/` for embedded data, `test/` for tests
- Based on plan.md: Single CLI application following existing pulumicost-core patterns

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Create registry directory at repository root: `registry/`
- [x] T002 Add semver dependency to go.mod: `github.com/Masterminds/semver/v3`
- [x] T003 [P] Create internal/registry package directory: `internal/registry/`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Tests for Foundational (MANDATORY - TDD Required) ⚠️

- [ ] T004 [P] Unit tests for registry embedding in `internal/registry/embed_test.go`
- [ ] T005 [P] Unit tests for GitHub API client in `internal/registry/github_test.go`
- [ ] T006 [P] Unit tests for archive extraction in `internal/registry/archive_test.go`
- [ ] T007 [P] Unit tests for version constraint parsing in `internal/registry/version_test.go`

### Implementation for Foundational

- [ ] T008 [P] Create registry.json with schema and initial plugins (kubecost, aws-public) in `registry/registry.json`
- [ ] T009 [P] Implement registry embedding and parsing in `internal/registry/embed.go`
- [ ] T010 [P] Implement RegistryEntry types and validation in `internal/registry/entry.go`
- [ ] T011 [P] Implement GitHub API client with retry logic and GITHUB_TOKEN/gh auth support in `internal/registry/github.go`
- [ ] T012 [P] Implement archive extraction (tar.gz/zip) in `internal/registry/archive.go`
- [ ] T013 [P] Implement semver constraint parsing in `internal/registry/version.go`
- [ ] T014 Add PluginConfig struct and plugins section to `internal/config/config.go`
- [ ] T015 Implement plugin config management in `internal/config/plugins.go`

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Install Plugin from Registry (Priority: P1) 🎯 MVP

**Goal**: Users can install well-known plugins by name from the embedded registry

**Independent Test**: Run `pulumicost plugin install kubecost` and verify binary exists at `~/.pulumicost/plugins/kubecost/<version>/`

### Tests for User Story 1 (MANDATORY - TDD Required) ⚠️

- [ ] T016 [P] [US1] Unit tests for plugin install command in `internal/cli/plugin_install_test.go`
- [ ] T017 [P] [US1] Integration test for registry install flow in `test/integration/plugin_install_test.go`

### Implementation for User Story 1

- [ ] T018 [P] [US1] Implement ParsePluginSpecifier (name@version parsing) in `internal/registry/entry.go`
- [ ] T019 [US1] Implement plugin install command with registry lookup in `internal/cli/plugin_install.go`
- [ ] T020 [US1] Add install subcommand to plugin command in `internal/cli/plugin.go`
- [ ] T021 [US1] Implement platform-specific binary detection in `internal/registry/github.go`
- [ ] T022 [US1] Implement download with progress reporting in `internal/registry/github.go`
- [ ] T023 [US1] Implement post-install validation (executable check) in `internal/registry/archive.go`
- [ ] T024 [US1] Add --force, --no-save, and --plugin-dir flags to install command in `internal/cli/plugin_install.go`
- [ ] T025 [US1] Persist installed plugin to config.yaml in `internal/cli/plugin_install.go`

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Install Plugin from GitHub URL (Priority: P2)

**Goal**: Users can install plugins from any GitHub repository URL

**Independent Test**: Run `pulumicost plugin install github.com/owner/repo` and verify plugin is installed

### Tests for User Story 2 (MANDATORY - TDD Required) ⚠️

- [ ] T026 [P] [US2] Unit tests for GitHub URL parsing in `internal/registry/entry_test.go`
- [ ] T027 [P] [US2] Integration test for URL install flow in `test/integration/plugin_install_test.go`

### Implementation for User Story 2

- [ ] T028 [P] [US2] Implement ParseGitHubURL (owner/repo extraction) in `internal/registry/entry.go`
- [ ] T029 [US2] Extend install command to handle GitHub URLs in `internal/cli/plugin_install.go`
- [ ] T030 [US2] Derive plugin name from repository name in `internal/cli/plugin_install.go`
- [ ] T031 [US2] Handle repositories without releases (error message) in `internal/registry/github.go`

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Config Persistence and Auto-Install (Priority: P3)

**Goal**: Installed plugins are saved to config and auto-installed on startup

**Independent Test**: Install plugin, delete binary, run pulumicost and verify auto-install

### Tests for User Story 3 (MANDATORY - TDD Required) ⚠️

- [ ] T032 [P] [US3] Unit tests for config persistence in `internal/config/plugins_test.go`
- [ ] T033 [P] [US3] Integration test for auto-install flow in `test/integration/plugin_autoinstall_test.go`

### Implementation for User Story 3

- [ ] T034 [US3] Implement GetMissingPlugins to compare config vs installed in `internal/config/plugins.go`
- [ ] T035 [US3] Implement auto-install on application startup in `cmd/pulumicost/main.go`
- [ ] T036 [US3] Add startup check for configured plugins in `internal/cli/root.go`
- [ ] T037 [US3] Display auto-install progress messages in `internal/cli/root.go`

**Checkpoint**: Config persistence and auto-install working independently

---

## Phase 6: User Story 4 - Update Plugins (Priority: P4)

**Goal**: Users can update installed plugins to latest or specific versions

**Independent Test**: Install old version, run update command, verify new version installed

### Tests for User Story 4 (MANDATORY - TDD Required) ⚠️

- [ ] T038 [P] [US4] Unit tests for update command in `internal/cli/plugin_update_test.go`
- [ ] T039 [P] [US4] Integration test for update flow in `test/integration/plugin_update_test.go`

### Implementation for User Story 4

- [ ] T040 [US4] Implement plugin update command in `internal/cli/plugin_update.go`
- [ ] T041 [US4] Add update subcommand to plugin command in `internal/cli/plugin.go`
- [ ] T042 [US4] Implement --all flag for bulk updates in `internal/cli/plugin_update.go`
- [ ] T043 [US4] Implement --dry-run flag for preview in `internal/cli/plugin_update.go`
- [ ] T044 [US4] Update config.yaml with new version after update in `internal/cli/plugin_update.go`
- [ ] T045 [US4] Replace old version directory with new version after successful update in `internal/cli/plugin_update.go`

**Checkpoint**: Update functionality working independently

---

## Phase 7: User Story 5 - Remove Plugins (Priority: P5)

**Goal**: Users can remove plugins they no longer need

**Independent Test**: Install plugin, run remove command, verify files deleted and config updated

### Tests for User Story 5 (MANDATORY - TDD Required) ⚠️

- [ ] T046 [P] [US5] Unit tests for remove command in `internal/cli/plugin_remove_test.go`
- [ ] T047 [P] [US5] Integration test for remove flow in `test/integration/plugin_remove_test.go`

### Implementation for User Story 5

- [ ] T048 [US5] Implement plugin remove command in `internal/cli/plugin_remove.go`
- [ ] T049 [US5] Add remove subcommand to plugin command in `internal/cli/plugin.go`
- [ ] T050 [US5] Implement --all-versions flag in `internal/cli/plugin_remove.go`
- [ ] T051 [US5] Implement --keep-config flag in `internal/cli/plugin_remove.go`
- [ ] T052 [US5] Remove plugin entry from config.yaml in `internal/cli/plugin_remove.go`

**Checkpoint**: Remove functionality working independently

---

## Phase 8: Dependency Resolution (Priority: P6)

**Goal**: Plugins can declare and resolve dependencies on other plugins

**Independent Test**: Install plugin with dependency, verify dependency auto-installed

### Tests for Dependency Resolution (MANDATORY - TDD Required) ⚠️

- [ ] T053 [P] [US6] Unit tests for dependency resolution in `internal/registry/dependency_test.go`
- [ ] T054 [P] [US6] Integration test for dependency install in `test/integration/plugin_dependency_test.go`

### Implementation for Dependency Resolution

- [ ] T055 [P] [US6] Implement dependency resolver with cycle detection in `internal/registry/dependency.go`
- [ ] T056 [US6] Parse dependencies from plugin.manifest.json in `internal/registry/dependency.go`
- [ ] T057 [US6] Integrate dependency resolution into install flow in `internal/cli/plugin_install.go`
- [ ] T058 [US6] Add dependency warning to remove command in `internal/cli/plugin_remove.go`

**Checkpoint**: Dependency resolution working

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T059 [P] Update user guide documentation in `docs/guides/user-guide.md`
- [ ] T060 [P] Add plugin management section to CLI reference in `docs/reference/cli.md`
- [ ] T061 Run make lint and fix all linting errors
- [ ] T062 Run make test and ensure 80%+ coverage on new code
- [ ] T063 [P] Add example plugin installation in quickstart documentation
- [ ] T064 Security review: validate zip-slip prevention in archive extraction
- [ ] T065 Performance validation: ensure install completes < 30 seconds

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-8)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3 → P4 → P5 → P6)
- **Polish (Phase 9)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational - Extends US1 install command
- **User Story 3 (P3)**: Can start after Foundational - Uses US1 install logic
- **User Story 4 (P4)**: Can start after Foundational - Reuses install logic from US1
- **User Story 5 (P5)**: Can start after Foundational - Independent removal logic
- **Dependency Resolution (P6)**: Can start after Foundational - Extends US1 install flow

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Foundation tasks (embedding, GitHub client, archive) before CLI commands
- Core implementation before flags and options
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tests (T004-T007) can run in parallel
- All Foundational implementation tasks (T008-T013) can run in parallel
- Once Foundational phase completes, all user story tests can start
- Within each story, tests marked [P] can run in parallel

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
Task: "Unit tests for plugin install command in internal/cli/plugin_install_test.go"
Task: "Integration test for registry install flow in test/integration/plugin_install_test.go"

# Launch foundational components together (Phase 2):
Task: "Implement registry embedding and parsing in internal/registry/embed.go"
Task: "Implement GitHub API client with retry logic in internal/registry/github.go"
Task: "Implement archive extraction (tar.gz/zip) in internal/registry/archive.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test `pulumicost plugin install kubecost`
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently (GitHub URLs)
4. Add User Story 3 → Test independently (Config persistence)
5. Add User Story 4 → Test independently (Updates)
6. Add User Story 5 → Test independently (Remove)
7. Add Dependency Resolution → Test independently
8. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (MVP)
   - Developer B: User Story 2 + 3
   - Developer C: User Story 4 + 5
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
