# Feature Specification: Remove PORT Environment Variable from Plugin Spawning

**Feature Branch**: `104-remove-port-env`
**Created**: 2025-12-15
**Status**: Draft
**Input**: GitHub Issue #232 - refactor(pluginhost): Remove PORT env var, use only --port flag

## User Scenarios & Testing

### User Story 1 - Plugin Developer Using Standard Port Flag (Priority: P1)

A plugin developer creates a PulumiCost plugin that reads the port from the `--port` command-line flag. When the core launches the plugin, it receives the port correctly via the flag without any environment variable interference.

**Why this priority**: This is the primary use case - plugins should reliably receive their assigned port through a well-defined, non-conflicting mechanism (command-line flag).

**Independent Test**: Can be tested by launching a plugin with `--port=XXXXX` flag and verifying the plugin binds to that exact port.

**Acceptance Scenarios**:

1. **Given** a plugin binary that reads port from `--port` flag, **When** core launches the plugin, **Then** the plugin receives the correct port via `--port=XXXXX` argument
2. **Given** a user has `PORT=3000` in their environment, **When** core launches a plugin, **Then** the plugin does NOT see this conflicting PORT value (only sees its assigned port via flag)
3. **Given** multiple plugins are launched simultaneously, **When** each plugin starts, **Then** each receives a unique port via `--port` flag without conflicts

---

### User Story 2 - Plugin Developer Using PULUMICOST_PLUGIN_PORT Fallback (Priority: P2)

A plugin developer prefers to read the port from an environment variable for debugging or tooling integration. The plugin can read `PULUMICOST_PLUGIN_PORT` as a fallback/documentation mechanism while `--port` remains authoritative.

**Why this priority**: Backward compatibility and debugging use cases are important but secondary to the primary `--port` flag mechanism.

**Independent Test**: Can be tested by launching a plugin and verifying `PULUMICOST_PLUGIN_PORT` is set in the environment while `PORT` is not.

**Acceptance Scenarios**:

1. **Given** a plugin that reads `PULUMICOST_PLUGIN_PORT` env var, **When** core launches the plugin, **Then** the environment variable matches the `--port` flag value
2. **Given** a plugin expecting the legacy `PORT` env var, **When** core launches the plugin, **Then** `PORT` is NOT set by core (plugin should migrate to `--port` flag)

---

### User Story 3 - Multi-Plugin Cost Calculation (Priority: P3)

An operator runs a cost calculation that requires multiple plugins simultaneously (e.g., aws-public and aws-ce plugins). Each plugin receives its own unique port without any environment variable conflicts.

**Why this priority**: Multi-plugin scenarios are advanced use cases but must work correctly for comprehensive cost analysis.

**Independent Test**: Can be tested by launching two plugins in parallel and verifying each binds to its assigned unique port.

**Acceptance Scenarios**:

1. **Given** two plugins need to run simultaneously, **When** core launches both, **Then** each receives a unique port via `--port` flag and can bind successfully
2. **Given** a user's environment has `PORT` set to some value, **When** multiple plugins launch, **Then** neither plugin inherits or sees the user's `PORT` variable

---

### Edge Cases

- What happens when the user's shell has `PORT` set? The plugin should NOT see this value; port comes exclusively from `--port` flag.
- How does the system handle plugins that haven't migrated from `PORT` to `--port`? Core detects the timeout and logs guidance suggesting the plugin may need an update to support the `--port` flag. Migration to `--port` is required after pulumicost-spec#129 is completed.
- What if `--port` flag parsing fails in the plugin? The plugin will fail to bind, and core will timeout waiting for the plugin (existing behavior).

## Requirements

### Functional Requirements

- **FR-001**: Core MUST pass `--port=XXXXX` flag to all launched plugins as the authoritative port communication mechanism
- **FR-002**: Core MUST NOT set the `PORT` environment variable when launching plugins
- **FR-003**: Core MUST continue to set `PULUMICOST_PLUGIN_PORT` environment variable for backward compatibility and debugging
- **FR-004**: Core MUST ensure each plugin receives a unique port value when multiple plugins run simultaneously
- **FR-005**: Core MUST NOT allow inherited environment variables (like user's `PORT`) to interfere with plugin port communication
- **FR-006**: Existing test suite MUST be updated to verify `PORT` is no longer set
- **FR-007**: Core MUST log a guidance message when plugin fails to bind, suggesting the plugin may need an update to support `--port` flag
- **FR-008**: Core MUST log a DEBUG-level message when `PORT` is detected in the user's environment (visible only with `--debug` flag)

### Dependencies

- **External Dependency**: This change is blocked by pulumicost-spec#129 (Add --port flag parsing to pluginsdk.Serve()). Plugins MUST support `--port` flag before core removes `PORT` env var.

### Assumptions

- Plugin SDK (pluginsdk) in pulumicost-spec has been updated to support `--port` flag with highest priority (pulumicost-spec#129 complete)
- All actively maintained plugins have been or will be updated to use `pluginsdk.Serve()` with `--port` support
- The `PULUMICOST_PLUGIN_PORT` environment variable serves documentation/debugging purposes only; `--port` flag is authoritative

## Clarifications

### Session 2025-12-15

- Q: What should happen when a user runs an old plugin that only reads PORT env var? → A: Core detects timeout and logs guidance suggesting plugin may need update to support --port flag
- Q: Should core log when it detects PORT in user's environment? → A: Log DEBUG message when PORT is detected (visible only with --debug flag)

## Success Criteria

### Measurable Outcomes

- **SC-001**: All unit tests pass with `PORT` environment variable removed from plugin spawning code
- **SC-002**: Multi-plugin scenarios (two or more plugins running simultaneously) work correctly without port conflicts
- **SC-003**: Users with `PORT` set in their shell environment can run plugins without interference
- **SC-004**: Existing E2E tests continue to pass, confirming no regression in plugin communication
