# Feature Specification: Project Rename to FinFocus

**Feature Branch**: `113-rebrand-to-finfocus`  
**Created**: 2026-01-14  
**Status**: Draft  
**Input**: User description: "Project Rename to FinFocus (Engineering Plan)"

## Clarifications

### Session 2026-01-14
- Q: How should JSON/YAML output root keys be handled during the rebrand? → A: Rename keys to `finfocus` (breaking change).
- Q: How should legacy `FINFOCUS_` environment variables be handled? → A: Support them only via explicit `FINFOCUS_COMPAT=1` toggle.
- Q: How should the user be prompted to use the recommended 'fin' alias? → A: Show a persistent reminder on every run until suppressed.
- Q: How should the configuration migration be triggered? → A: Automatically prompt the user on first run if old directory detected.
- Q: How should legacy 'finfocus-plugin-*' binaries be handled? → A: Allow discovery via a legacy toggle (e.g., FINFOCUS_LOG_LEGACY=1).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Fresh Installation & Branding (Priority: P1)

A new user installs the tool and interacts with it for the first time, seeing only "FinFocus" branding.

**Why this priority**: Establishing the new brand identity is the primary goal of this change.

**Independent Test**: Install binary, run help, verify output.

**Acceptance Scenarios**:

1. **Given** a clean environment (no config), **When** user runs `finfocus --help`, **Then** output displays "FinFocus" and no references to "FinFocus".
2. **Given** a clean environment, **When** user runs `finfocus version`, **Then** it reports the correct version under the FinFocus name.
3. **Given** the binary is installed, **When** user checks the filename, **Then** it is `finfocus` (not `finfocus`).

---

### User Story 2 - Migration from FinFocus (Priority: P1)

An existing user with `~/.finfocus` configuration upgrades to `finfocus`.

**Why this priority**: Ensuring smooth transition for existing user base to prevent data/config loss.

**Independent Test**: Create dummy `~/.finfocus`, run `finfocus`, verify `~/.finfocus` creation.

**Acceptance Scenarios**:

1. **Given** `~/.finfocus` exists and `~/.finfocus` does not, **When** user runs `finfocus` in an interactive terminal, **Then** system prompts to migrate configuration and state.
2. **Given** user accepts migration, **When** migration completes, **Then** `~/.finfocus` contains copied config/state and `~/.finfocus` is preserved (not deleted).
3. **Given** non-interactive environment (CI), **When** `finfocus` runs with old config present, **Then** it logs a warning about migration but proceeds if possible (or fails gracefully if config is strictly required).

---

### User Story 3 - Environment Variable Configuration (Priority: P2)

A user configures the tool using environment variables for CI/CD or secrets.

**Why this priority**: Enterprise users rely heavily on env vars for authentication and configuration.

**Independent Test**: Set `FINFOCUS_LOG_LEVEL=debug`, run tool, check logs.

**Acceptance Scenarios**:

1. **Given** `FINFOCUS_LOG_LEVEL=debug` is set, **When** user runs `finfocus`, **Then** debug logs are emitted.
2. **Given** `FINFOCUS_LOG_LEVEL` is set (legacy), **When** user runs `finfocus`, **Then** it is ignored (or warns and ignored), favoring `FINFOCUS_` prefix.

---

### User Story 4 - Plugin Discovery (Priority: P2)

A user installs and uses plugins with the new naming convention.

**Why this priority**: The ecosystem depends on plugins for cloud provider support.

**Independent Test**: Mock a plugin executable `finfocus-plugin-test`.

**Acceptance Scenarios**:

1. **Given** a binary `finfocus-plugin-myplugin` in `~/.finfocus/plugins`, **When** user runs `finfocus plugin list`, **Then** "myplugin" is listed.
2. **Given** a legacy binary `finfocus-plugin-old` in the path, **When** user runs `finfocus plugin list`, **Then** it is NOT discovered (unless explicit backward compat is requested, which is not currently planned).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI binary MUST be named `finfocus`.
- **FR-002**: The root command and help text MUST display "FinFocus" as the application name.
- **FR-003**: The system MUST use `~/.finfocus` as the default configuration and data directory.
- **FR-004**: The system MUST support environment variables with the `FINFOCUS_` prefix (replacing `FINFOCUS_`).
  - `FINFOCUS_HOME`
  - `FINFOCUS_LOG_LEVEL`
  - `FINFOCUS_LOG_FORMAT`
  - `FINFOCUS_LOG_FILE`
  - `FINFOCUS_TRACE_ID`
  - `FINFOCUS_OUTPUT_FORMAT`
  - `FINFOCUS_CONFIG_STRICT`
  - `FINFOCUS_PLUGIN_PORT`
  - `FINFOCUS_CONCURRENCY_MULTIPLIER`
- **FR-012**: The system MUST ignore legacy `FINFOCUS_` variables unless `FINFOCUS_COMPAT=1` environment variable is set.
- **FR-013**: The system MUST display a reminder to add `alias fin=finfocus` to the user's shell profile on every execution, until suppressed via configuration.
- **FR-005**: The system MUST detect `~/.finfocus` on startup and, if `~/.finfocus` is missing, automatically prompt the user to migrate configuration and state.
- FR-006: The migration mechanism MUST copy data to the new location, leaving the old directory intact.
- **FR-007**: The system MUST discover plugins named `finfocus-plugin-<name>`.
- **FR-014**: The system MUST support discovery of legacy `finfocus-plugin-<name>` binaries only when `FINFOCUS_LOG_LEGACY=1` (or similar toggle) is set.
- **FR-008**: The system MUST look for plugins in `~/.finfocus/plugins`.
- **FR-009**: The Go module path MUST be updated to `github.com/rshade/finfocus`.
- **FR-010**: The Pulumi Analyzer integration MUST respond to the executable name `pulumi-analyzer-finfocus`.
- **FR-011**: JSON and YAML output formats MUST use `finfocus` as the root key for results (replacing `finfocus`).

### Key Entities

- **Configuration**: User settings stored in `config.yaml` or `config.json` within the home directory.
- **Plugin**: An external executable complying with the `CostSourceService` (formerly `FinFocusPlugin`) protocol.
- **State**: Cached pricing data or history stored in the home directory.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `finfocus --version` executes successfully and returns a valid version string.
- **SC-002**: End-to-end cost estimation (e.g., `finfocus cost`) functions correctly using the new `~/.finfocus` directory.
- **SC-003**: Migration of a standard `~/.finfocus` setup (config + 1 plugin) to `~/.finfocus` completes without error.
- **SC-004**: All standard CLI commands (`cost`, `plugin`, etc.) function identically to their `finfocus` counterparts.