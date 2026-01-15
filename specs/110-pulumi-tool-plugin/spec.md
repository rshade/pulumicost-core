# Feature Specification: Pulumi Tool Plugin Integration

**Feature Branch**: `110-pulumi-tool-plugin`
**Created**: 2025-12-30
**Status**: Draft
**Input**: User description from prompt

## Clarifications

### Session 2025-12-30

- Q: How should the tool detect it is running in plugin mode (especially on Windows)? → A: **Hybrid Approach**: Detect if binary name matches `pulumi-tool-cost` (case-insensitive, ignoring `.exe`) OR if `FINFOCUS_PLUGIN_MODE=true` is set.
- Q: Where specifically should configuration be stored when PULUMI_HOME is used? → A: **Subdirectory**: Store config in `$PULUMI_HOME/finfocus/`.
- Q: What should happen if the tool is in plugin mode but `PULUMI_HOME` is missing? → A: **Silent Fallback**: Use default standalone configuration paths (e.g., XDG or `~/.finfocus`) while maintaining plugin-style help text.
- Q: How should exit codes be handled when running as a plugin? → A: **Pulumi-First**: Use specific exit codes if expected by Pulumi (e.g., for schema/invocation errors), otherwise fallback to standard POSIX (0 for success, 1 for general error).
- Q: What version should the plugin report? → A: **Sync with Core**: The plugin version MUST match the current `finfocus` core version.

## User Scenarios & Testing

### User Story 1 - Invoke via Pulumi CLI (Priority: P1)

As a Pulumi user, I want to execute `finfocus` commands using `pulumi plugin run tool cost` so that I can integrate cost estimation directly into my Pulumi workflow without needing a separate binary in my path.

**Why this priority**: This is the core functionality of the integration.

**Independent Test**: Can be tested by installing the binary in the expected Pulumi plugin directory and running `pulumi plugin run tool cost -- help`.

**Acceptance Scenarios**:

1. **Given** the binary is installed in `~/.pulumi/plugins/tool-cost-vX.Y.Z/`, **When** I run `pulumi plugin run tool cost -- help`, **Then** the help text is displayed.
2. **Given** valid credentials and a plan file, **When** I run `pulumi plugin run tool cost -- cost projected --pulumi-json plan.json`, **Then** the cost estimate is returned.

---

### User Story 2 - Context-Aware Help Text (Priority: P2)

As a user, when I run the tool as a plugin, I want the help text and usage examples to reflect the `pulumi plugin run tool cost` syntax instead of `finfocus` so that I can copy-paste examples correctly.

**Why this priority**: Essential for user experience (UX) to avoid confusion.

**Independent Test**: Run the binary directly vs. via Pulumi plugin command and compare help output.

**Acceptance Scenarios**:

1. **Given** I am running as a plugin, **When** I run the help command, **Then** the usage line shows `Use: "pulumi plugin run tool cost"`.
2. **Given** I am running the binary directly as `finfocus`, **When** I run help, **Then** the usage line remains `Use: "finfocus"`.

---

### User Story 3 - Configuration Path Compliance (Priority: P2)

As a user, I want the tool to respect `PULUMI_HOME` for configuration files when running as a plugin so that my configuration is centralized and my home directory remains clean.

**Why this priority**: Important for "good citizenship" in the Pulumi ecosystem and filesystem hygiene.

**Independent Test**: Set `PULUMI_HOME`, run the tool, and verify where it looks for/creates config.

**Acceptance Scenarios**:

1. **Given** `PULUMI_HOME` is set, **When** the tool runs, **Then** it looks for configuration in `$PULUMI_HOME` (or a subdirectory) before default paths.
2. **Given** `PULUMI_HOME` is NOT set, **When** the tool runs, **Then** it falls back to standard XDG paths or `~/.finfocus`.

### Edge Cases

- **Missing PULUMI_HOME**: If the tool is in plugin mode but `PULUMI_HOME` is not set, it MUST silently fallback to default standalone configuration paths (XDG or `~/.finfocus`) but keep using plugin-style help text (`pulumi plugin run tool cost`).
- **Invalid PULUMI_HOME**: If `PULUMI_HOME` points to a non-existent or inaccessible directory, the tool should fallback to standard paths for read operations and provide a clear error for write operations.
- **Direct Invocation**: If the binary `pulumi-tool-cost` is run directly by a user in a terminal (not via Pulumi), it operates in plugin mode (showing plugin help text) but otherwise functions normally.

## Requirements

### Functional Requirements

- **FR-001**: The application MUST detect its invocation mode via **EITHER**:
    - The binary name matches `pulumi-tool-cost` (case-insensitive, ignoring `.exe` extension).
    - The environment variable `FINFOCUS_PLUGIN_MODE` is set to `true` (or `1`).
- **FR-002**: If either condition in FR-001 is met, the application MUST auto-configure for plugin mode.
- **FR-003**: When in plugin mode, the CLI root command `Use` string MUST be updated to `pulumi plugin run tool cost`.
- **FR-004**: When in plugin mode, CLI help text and examples MUST reference `pulumi plugin run tool cost`.
- **FR-005**: The application MUST strictly respect `FINFOCUS_` prefix for its own specific environment variables (e.g. API keys for pricing providers) to avoid collision with Pulumi core variables.
- **FR-006**: When in plugin mode, the application MUST prioritize `$PULUMI_HOME/finfocus/` as the configuration directory if `$PULUMI_HOME` is set.
- **FR-007**: The build system MUST support compiling the binary with the name `pulumi-tool-cost`.
- **FR-008**: When in plugin mode, the application MUST use Pulumi-compatible exit codes for known error states, falling back to standard POSIX codes for general failures.

### Key Entities

- **Plugin Context**: Information injected by the Pulumi CLI (RPC address, API URL, Access Token, Home Dir).

## Success Criteria

### Measurable Outcomes

- **SC-001**: `pulumi plugin ls` successfully lists `cost` as a tool plugin with a version matching the `finfocus` core binary.
- **SC-002**: `pulumi plugin run tool cost -- help` returns exit code 0 and displays the adapted help text.
- **SC-003**: Configuration files can be successfully read from a custom `PULUMI_HOME` directory.
- **SC-004**: The binary functions identically to the standalone `finfocus` regarding core logic (cost calculation), passing all existing regression tests.