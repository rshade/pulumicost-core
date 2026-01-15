# Feature Specification: Plugin Info and DryRun Discovery

**Feature Branch**: `112-plugin-info-discovery`  
**Created**: 2026-01-10  
**Status**: Draft  
**Input**: User description: "1. Plugin Ecosystem Maturity: * GetPluginInfo Implementation: Marked as Unblocked (previously blocked on spec #029). * DryRun Discovery: Added a new task to implement plugin field mapping discovery using the new DryRun RPC."

## Clarifications

### Session 2026-01-10

- Q: Version compatibility enforcement strategy? → A: Best effort: Core ignores unknown fields and attempts to operate on a subset of the spec.
- Q: Mechanism for bypassing compatibility checks? → A: CLI flag (e.g., `--skip-version-check`): Manual override for advanced users.
- Q: Output format for `finfocus plugin inspect`? → A: Combined: Table by default, JSON via `--json` flag.
- Q: Handling timeout/errors in `plugin list`? → A: Hide the plugin: Only show plugins that successfully respond.
- Q: Timeout for `DryRun` capability discovery? → A: 10 seconds: Provides a buffer for more complex capability discovery.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Version Compatibility Check (Priority: P1)

As a CLI user, I want the system to automatically verify that installed plugins are compatible with the core system's protocol version so that I don't encounter unexpected errors during cost calculations.

**Why this priority**: Essential for system stability and preventing protocol mismatches that lead to runtime crashes.

**Independent Test**: Can be fully tested by loading a plugin with an incompatible `spec_version` and verifying the core system detects and warns the user.

**Acceptance Scenarios**:

1. **Given** a plugin is being initialized, **When** the core calls `GetPluginInfo`, **Then** it receives and validates the `spec_version` against supported ranges.
2. **Given** a plugin does not implement `GetPluginInfo` (legacy), **When** initialized, **Then** the core system logs a warning but allows execution to continue for backward compatibility.

---

### User Story 2 - Discover Plugin Capabilities (Priority: P2)

As a developer, I want to see what FOCUS fields a plugin supports for specific resource types without running a full cost estimate, so that I can understand data coverage for my infrastructure.

**Why this priority**: Improves developer experience and transparency of what data is being provided by plugins.

**Independent Test**: Can be tested by running a discovery command for a resource type (e.g., `aws:ec2/instance:Instance`) and verifying the list of supported FOCUS fields is displayed.

**Acceptance Scenarios**:

1. **Given** a plugin supports `DryRun` discovery, **When** a user queries for a resource type, **Then** the system returns a list of FOCUS fields marked as SUPPORTED, UNSUPPORTED, or CONDITIONAL.

---

### User Story 3 - Plugin Metadata Display (Priority: P3)

As an operator, I want to see detailed information (name, version, spec version) about all installed plugins when listing them, so that I can manage my plugin inventory effectively.

**Why this priority**: Provides inventory visibility and helps in troubleshooting environment-specific issues.

**Independent Test**: Can be tested by running `finfocus plugin list` and verifying all metadata fields from `GetPluginInfo` are displayed in the output.

**Acceptance Scenarios**:

1. **Given** multiple plugins are installed, **When** `plugin list` is executed, **Then** the output includes name, version, and protocol spec version for each plugin.

### Edge Cases

- What happens when a plugin returns malformed SemVer for its version?
- How does the system handle a `GetPluginInfo` call that times out?
- What happens if a plugin claims to support a field in `DryRun` but fails to provide it in actual cost calls?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Core system MUST call `GetPluginInfo` during plugin initialization with a 5-second timeout.
- **FR-002**: Core system MUST validate that `name`, `version`, and `spec_version` are present in the `GetPluginInfo` response.
- **FR-003**: System MUST provide a CLI command (e.g., `finfocus plugin inspect`) that utilizes the `DryRun` RPC to display field mappings, supporting both human-readable table output and machine-readable JSON via a `--json` flag.
- **FR-004**: System MUST handle "Unimplemented" errors for `GetPluginInfo` and `DryRun` gracefully for legacy plugins.
- **FR-005**: System MUST log a warning if a plugin's `spec_version` is different from the core's supported version, but continue operation using a "best-effort" approach for known fields.
- **FR-006**: System MUST provide a `--skip-version-check` global flag to bypass initialization compatibility warnings and allow execution with unknown spec versions.
- **FR-007**: The `plugin list` command MUST omit plugins that fail to respond to the `GetPluginInfo` metadata request within the 5-second timeout to ensure only active, valid plugins are listed.

### Key Entities *(include if feature involves data)*

- **PluginMetadata**: Contains name, version, spec_version, and supported providers.
- **FieldMapping**: Represents a FOCUS field and its support status (SUPPORTED, UNSUPPORTED, CONDITIONAL, DYNAMIC).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: System successfully identifies 100% of plugins with incompatible spec versions during initialization.
- **SC-002**: Discovery queries for field mappings return results in under 200ms for locally running plugins.
- **SC-003**: The `plugin list` command successfully displays metadata for all plugins implementing the `GetPluginInfo` RPC.
- **SC-004**: 100% of legacy plugins continue to function without crashing when new RPCs are invoked.