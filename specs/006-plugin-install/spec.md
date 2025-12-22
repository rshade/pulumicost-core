# Feature Specification: Plugin Install/Update/Remove System

**Feature Branch**: `006-plugin-install`
**Created**: 2025-11-23
**Status**: Complete
**GitHub Issue**: #163

## Clarifications

### Session 2025-11-23

- Q: What retry parameters for network failures? → A: Standard: 3 retries, base 1 second (max ~7 sec total)
- Q: Plugin version retention on update? → A: Keep only latest version, replace old on update

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Install Plugin from Registry (Priority: P1)

As a user, I want to install a well-known plugin by name so I can quickly add cost sources without manual binary management.

**Why this priority**: This is the core value proposition - simplifying plugin installation from a single command. Without this, users must manually download and place binaries.

**Independent Test**: Can be fully tested by running `pulumicost plugin install kubecost` and verifying the plugin binary exists in the correct directory and is executable.

**Acceptance Scenarios**:

1. **Given** the registry contains a "kubecost" plugin, **When** user runs `pulumicost plugin install kubecost`, **Then** the latest version is downloaded, extracted to `~/.pulumicost/plugins/kubecost/<version>/`, and marked executable
2. **Given** the registry contains a "kubecost" plugin with version v1.0.0, **When** user runs `pulumicost plugin install kubecost@v1.0.0`, **Then** that specific version is installed
3. **Given** a plugin is already installed, **When** user runs install without --force, **Then** system reports plugin already exists and skips installation
4. **Given** a plugin is already installed, **When** user runs install with --force, **Then** system reinstalls the plugin

---

### User Story 2 - Install Plugin from GitHub URL (Priority: P2)

As a user, I want to install a plugin from a GitHub URL so I can use third-party or private plugins not in the official registry.

**Why this priority**: Enables ecosystem growth beyond official plugins and supports private/internal plugins for organizations.

**Independent Test**: Can be tested by running `pulumicost plugin install github.com/owner/repo` and verifying the plugin is installed.

**Acceptance Scenarios**:

1. **Given** a valid GitHub repository with releases, **When** user runs `pulumicost plugin install github.com/owner/repo`, **Then** the latest release is downloaded and installed
2. **Given** a valid GitHub repository, **When** user runs `pulumicost plugin install github.com/owner/repo@v2.0.0`, **Then** that specific version is installed
3. **Given** a GitHub repository without releases, **When** user attempts to install, **Then** system displays a clear error message

---

### User Story 3 - Config Persistence and Auto-Install (Priority: P3)

As a user, I want installed plugins saved to my config so teammates get the same plugins when they run PulumiCost.

**Why this priority**: Enables reproducible environments across teams and machines, critical for CI/CD and collaboration.

**Independent Test**: Can be tested by installing a plugin, checking config.yaml for the entry, then deleting the plugin and running pulumicost to verify auto-install.

**Acceptance Scenarios**:

1. **Given** user installs a plugin, **When** installation completes, **Then** plugin entry is added to `~/.pulumicost/config.yaml`
2. **Given** user installs with --no-save flag, **When** installation completes, **Then** plugin is NOT added to config
3. **Given** config.yaml lists a plugin that is not installed, **When** pulumicost starts, **Then** missing plugin is automatically downloaded and installed
4. **Given** config.yaml specifies a version, **When** auto-install runs, **Then** exactly that version is installed

---

### User Story 4 - Update Plugins (Priority: P4)

As a user, I want to update plugins to latest versions easily so I can get bug fixes and new features.

**Why this priority**: Important for maintenance but not critical for initial adoption.

**Independent Test**: Can be tested by installing an older version, running update command, and verifying newer version is installed.

**Acceptance Scenarios**:

1. **Given** kubecost v1.0.0 is installed and v2.0.0 exists, **When** user runs `pulumicost plugin update kubecost`, **Then** v2.0.0 is installed
2. **Given** multiple plugins are installed, **When** user runs `pulumicost plugin update --all`, **Then** all plugins are updated to latest
3. **Given** --dry-run flag is used, **When** update runs, **Then** system shows what would be updated without making changes

---

### User Story 5 - Remove Plugins (Priority: P5)

As a user, I want to remove plugins I no longer need to clean up my environment.

**Why this priority**: Housekeeping feature, less critical than install/update.

**Independent Test**: Can be tested by installing a plugin, running remove command, and verifying files are deleted and config is updated.

**Acceptance Scenarios**:

1. **Given** kubecost is installed, **When** user runs `pulumicost plugin remove kubecost`, **Then** plugin directory is deleted and entry removed from config
2. **Given** multiple versions installed, **When** user runs `pulumicost plugin remove kubecost --all-versions`, **Then** all versions are removed
3. **Given** --keep-config flag is used, **When** remove runs, **Then** files are deleted but config entry remains

---

### Edge Cases

- What happens when network is unavailable during download? System retries with exponential backoff and displays clear error after max retries.
- What happens when GitHub API rate limit is hit? System displays rate limit error and suggests using GITHUB_TOKEN.
- What happens when archive is corrupted? System validates extraction and re-downloads on failure.
- What happens when plugin directory is not writable? System displays permission denied error with specific path.
- What happens when plugin has dependencies on another plugin? System auto-installs required dependencies from registry.
- What happens with circular dependencies between plugins? System detects cycle and reports error without infinite loop.
- What happens when installed version constraint conflicts with dependency requirements? System reports conflict and suggests resolution.

## Requirements *(mandatory)*

### Functional Requirements

**Plugin Installation:**

- **FR-001**: System MUST install plugins from the embedded well-known registry by name
- **FR-002**: System MUST install plugins from GitHub repository URLs
- **FR-003**: System MUST support version pinning with @version syntax (e.g., `kubecost@v1.0.0`)
- **FR-004**: System MUST download platform-specific binaries based on OS and architecture
- **FR-005**: System MUST extract tar.gz archives on Linux/macOS and zip archives on Windows
- **FR-006**: System MUST set executable permissions on installed binaries (Unix systems)
- **FR-007**: System MUST validate that installed binary is executable before marking success
- **FR-008**: System MUST support --force flag to reinstall existing versions
- **FR-009**: System MUST support --no-save flag to skip config file updates
- **FR-010**: System MUST support --plugin-dir flag to specify custom installation directory

**Version Management:**

- **FR-011**: System MUST query GitHub Releases API to discover available versions
- **FR-012**: System MUST default to latest release when no version specified
- **FR-013**: System MUST support semantic version constraints for dependencies (>=, <, ~, ^)
- **FR-014**: System MUST resolve and install plugin dependencies automatically
- **FR-015**: System MUST detect and report circular dependencies

**Configuration:**

- **FR-016**: System MUST persist installed plugins to config.yaml with name, URL, and version
- **FR-017**: System MUST check for missing configured plugins on startup
- **FR-018**: System MUST auto-install missing plugins matching config specifications
- **FR-019**: System MUST support updating config when plugins are installed/removed

**Plugin Updates:**

- **FR-020**: System MUST update individual plugins to latest or specified version, replacing the previous version (not retaining multiple versions)
- **FR-021**: System MUST support --all flag to update all installed plugins
- **FR-022**: System MUST support --dry-run flag to preview updates without changes
- **FR-023**: System MUST update config.yaml with new version after successful update

**Plugin Removal:**

- **FR-024**: System MUST remove plugin binaries and associated files
- **FR-025**: System MUST remove plugin entry from config.yaml (unless --keep-config)
- **FR-026**: System MUST support --all-versions flag to remove all installed versions
- **FR-027**: System MUST warn if other plugins depend on the plugin being removed

**Registry:**

- **FR-028**: System MUST embed registry.json file with well-known plugin metadata
- **FR-029**: System MUST validate registry entries have required fields (name, repository, etc.)
- **FR-030**: System MUST support plugin capabilities and supported providers metadata

**Authentication:**

- **FR-031**: System MUST use GITHUB_TOKEN environment variable for API authentication when available
- **FR-032**: System MUST fall back to unauthenticated API calls (with rate limits)
- **FR-033**: System MUST support reading token from `gh auth token` if available

**Error Handling:**

- **FR-034**: System MUST display clear error messages with actionable recovery steps
- **FR-035**: System MUST retry network failures with exponential backoff (3 retries, base 1 second delay, max ~7 seconds total)
- **FR-036**: System MUST validate downloaded archives before extraction

### Key Entities

- **Plugin Entry**: Represents a plugin in config.yaml - name, GitHub URL, pinned version
- **Registry Entry**: Represents a plugin in registry.json - name, description, repository, author, license, capabilities, supported providers, security level
- **Plugin Manifest**: Represents installed plugin metadata - binary path, version, dependencies, requirements
- **Version Constraint**: Represents dependency version requirements - operator (>=, <, ~, ^), version number

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can install a plugin from registry in under 30 seconds on standard broadband connection
- **SC-002**: Users can install a plugin from GitHub URL in under 30 seconds on standard broadband connection
- **SC-003**: Plugin installation success rate is above 95% when network is available and repository exists
- **SC-004**: Config file correctly persists all installed plugins and is readable by subsequent sessions
- **SC-005**: Auto-install on startup correctly installs all missing configured plugins
- **SC-006**: Update command correctly identifies and installs newer versions for all outdated plugins
- **SC-007**: Remove command successfully deletes all plugin files and config entries
- **SC-008**: Dependency resolution correctly identifies and installs all required dependencies
- **SC-009**: Cross-platform support works correctly on Linux, macOS, and Windows
- **SC-010**: All error scenarios display user-friendly messages with recovery suggestions
- **SC-011**: Test coverage on new plugin installation code is above 80%

## Assumptions

- Plugins follow GoReleaser v2 archive naming convention: `{name}_{version}_{os}_{arch}.{format}`
- Plugin releases are published as GitHub Releases with downloadable assets
- Plugins contain a `plugin.manifest.json` with metadata and dependency information
- The registry.json is embedded in the binary and updated with new releases
- Users have write access to `~/.pulumicost/` directory
- GitHub API is accessible (directly or via proxy)
- Default rate limit of 60 requests/hour for unauthenticated users is acceptable for typical usage

## Out of Scope

- Checksum/signature verification (future security enhancement)
- Plugin sandboxing (security isolation)
- Private registry support (enterprise feature)
- Proxy/mirror support
- Offline installation from local archives
