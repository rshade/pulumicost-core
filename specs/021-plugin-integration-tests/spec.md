# Feature Specification: Integration Tests for Plugin Management

**Feature Branch**: `021-plugin-integration-tests`
**Created**: 2025-12-17
**Status**: Draft
**Input**: User description (see prompt)

## Clarifications

### Session 2025-12-17
- Q: How should the system handle multiple simultaneous plugin management commands? -> A: Implement file locking; subsequent commands wait or fail with a busy error.
- Q: What is the expected response format for the mock registry? -> A: JSON metadata endpoint + direct binary download endpoint.
- Q: Is the GitHub URL security warning interactive? -> A: Purely informative (log message only).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Plugin Initialization Verification (Priority: P1)

As a plugin developer, I need to verify that the `plugin init` command creates a fully functional plugin project structure with correct metadata so that I can start developing a new plugin without manual setup errors.

**Why this priority**: Essential for the plugin ecosystem; broken scaffolding prevents new plugin creation.

**Independent Test**: Can be tested by running `plugin init` in a temp dir and verifying file existence/content.

**Acceptance Scenarios**:

1. **Given** a clean directory, **When** I run `plugin init my-plugin --author "Test" --providers aws`, **Then** a directory `my-plugin` is created with `main.go`, `manifest.yaml`, `go.mod`, and other required files.
2. **Given** a request for multiple providers, **When** I run `plugin init` with `--providers aws,azure`, **Then** the generated `manifest.yaml` lists both providers.
3. **Given** a custom output path, **When** I run `plugin init` with `--output-dir /tmp/custom`, **Then** the plugin is created in `/tmp/custom` instead of the current directory.
4. **Given** an existing directory, **When** I run `plugin init` with `--force`, **Then** the existing files are overwritten.
5. **Given** an invalid plugin name, **When** I run `plugin init "bad name!"`, **Then** the command fails with a clear error message.

---

### User Story 2 - Plugin Installation Verification (Priority: P1)

As a user, I need to verify that the `plugin install` command correctly downloads and installs plugins from a registry so that I can extend the core functionality.

**Why this priority**: Core functionality for end-users to acquire plugins.

**Independent Test**: Can be tested using a mock registry server and verifying file placement in the plugin directory.

**Acceptance Scenarios**:

1. **Given** a mock registry, **When** I run `plugin install kubecost`, **Then** the plugin artifact is downloaded and extracted to the correct plugin directory (`~/.pulumicost/plugins/kubecost`).
2. **Given** a specific version request, **When** I run `plugin install kubecost@v1.0.0`, **Then** that specific version is installed.
3. **Given** a GitHub URL, **When** I run `plugin install github.com/user/repo`, **Then** the tool logs a security warning and proceeds with installation automatically.
4. **Given** an installed plugin, **When** I run `plugin install` with `--force`, **Then** the plugin is reinstalled.
5. **Given** the `--no-save` flag, **When** I run `plugin install`, **Then** the plugin is installed but not added to the persistent configuration.

---

### User Story 3 - Plugin Update Verification (Priority: P2)

As a user, I need to verify that `plugin update` correctly upgrades installed plugins to newer versions so that I can access bug fixes and new features.

**Why this priority**: Important for maintaining plugin health and security.

**Independent Test**: Can be tested by installing an "old" version and running update against a mock registry with a "new" version.

**Acceptance Scenarios**:

1. **Given** an outdated plugin, **When** I run `plugin update <name>`, **Then** the plugin is replaced with the latest version from the registry.
2. **Given** a specific version, **When** I run `plugin update <name> --version v2.0.0`, **Then** the plugin is updated to that specific version.
3. **Given** the `--dry-run` flag, **When** I run `plugin update`, **Then** the command shows pending updates but does not modify files.
4. **Given** an already up-to-date plugin, **When** I run `plugin update`, **Then** the system reports no updates available.
5. **Given** a non-existent plugin, **When** I run `plugin update`, **Then** the command fails with an appropriate error.

---

### User Story 4 - Plugin Removal Verification (Priority: P2)

As a user, I need to verify that `plugin remove` cleanly deletes plugin files and configuration entries so that my system doesn't accumulate unused artifacts.

**Why this priority**: Necessary for system hygiene.

**Independent Test**: Can be tested by installing a plugin and then running remove, verifying file/config absence.

**Acceptance Scenarios**:

1. **Given** an installed plugin, **When** I run `plugin remove <name>`, **Then** the plugin directory is deleted and the entry is removed from config.
2. **Given** the `--keep-config` flag, **When** I run `plugin remove`, **Then** files are deleted but the config entry remains.
3. **Given** the alias `uninstall` or `rm`, **When** I run `plugin <alias> <name>`, **Then** it behaves exactly like `remove`.
4. **Given** a non-existent plugin, **When** I run `plugin remove`, **Then** the command fails with a "not found" error.

### Edge Cases

- **Network Failure**: System handles registry timeouts or connection errors gracefully during install/update.
- **Corrupt Artifacts**: System detects and rejects invalid or corrupt plugin archives.
- **Concurrent Operations**: System handles multiple commands using file locking (wait or fail with busy error).
- **Permission Issues**: System handles write permission errors in the plugin directory.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The test suite MUST verify that `plugin init` generates all required files for a valid plugin (Go module, manifest, main entry point).
- **FR-002**: The test suite MUST verify that `plugin install` correctly handles registry interactions, including version resolution and artifact download.
- **FR-003**: The test suite MUST verify that `plugin update` identifies available updates and performs upgrade operations correctly.
- **FR-004**: The test suite MUST verify that `plugin remove` cleans up both the filesystem and the configuration file.
- **FR-005**: The test suite MUST verify support for command aliases (e.g., `rm` for `remove`).
- **FR-006**: The test suite MUST execute in a sandboxed environment (using temporary directories) to prevent modification of the user's actual configuration or plugins.
- **FR-007**: The test suite MUST use a mock registry server serving JSON metadata and binary blobs to simulate network interactions, ensuring tests are deterministic and do not require internet access.
- **FR-008**: The test suite MUST verify error handling for common failure modes (invalid names, missing plugins, network errors).
- **FR-009**: The test suite MUST verify that concurrent operations are safely handled via file locking (either queuing or rejecting with a busy error).

### Key Entities

- **Plugin Manifest**: Metadata file defining plugin properties (name, version, supported providers).
- **Plugin Registry**: Remote source (mocked) serving plugin metadata and artifacts.
- **Plugin Artifact**: The compressed archive containing the plugin binary.
- **User Configuration**: Local configuration file tracking installed plugins.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Integration tests cover 100% of the defined acceptance scenarios for `init`, `install`, `update`, and `remove` commands.
- **SC-002**: All integration tests pass consistently in the CI environment (0% flakiness).
- **SC-003**: Tests execute without requiring external network connectivity (verified by blocking network or using only localhost mocks).
- **SC-004**: Test execution time for the full suite remains under 30 seconds (using mocked downloads).
