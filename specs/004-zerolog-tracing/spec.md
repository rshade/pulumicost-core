# Feature Specification: Zerolog Distributed Tracing

**Feature Branch**: `004-zerolog-tracing`
**Created**: 2025-11-25
**Status**: Draft
**Input**: User description: "feat: Standardize on zerolog v1.34.0+ with distributed tracing"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Debugging Failed Cost Calculations (Priority: P1)

A DevOps engineer runs `pulumicost cost projected` and gets unexpected results or silent failures.
They need to understand what happened during the cost calculation process, including which plugins
were tried, which resources failed to price, and why certain fallbacks were used.

**Why this priority**: Without visibility into the cost calculation flow, users cannot effectively
troubleshoot issues. This is the core value proposition of implementing structured logging.

**Independent Test**: Can be fully tested by running any CLI command with `--debug` flag and
verifying that structured log output reveals the complete decision flow from command start to
finish.

**Acceptance Scenarios**:

1. **Given** a user runs `pulumicost cost projected --debug`, **When** the command executes,
   **Then** log output shows: command start, resource ingestion count, plugin lookup attempts for
   each resource, cost calculation results, and command completion with duration.

2. **Given** a plugin fails to return pricing for a resource, **When** the engine falls back to
   local specs, **Then** the log output clearly indicates the fallback occurred and explains why
   (e.g., "plugin kubecost returned no price for aws:ec2:Instance, falling back to local spec").

3. **Given** a user sets `PULUMICOST_LOG_FORMAT=json`, **When** they run any command, **Then** all
   log output is valid JSON that can be piped to tools like `jq` for filtering.

---

### User Story 2 - Correlating Logs Across Plugin Boundaries (Priority: P2)

A platform engineer needs to trace a request from the CLI through a plugin and back to understand
performance bottlenecks or errors that occur in the plugin communication layer.

**Why this priority**: Cross-boundary tracing is essential for diagnosing issues in the distributed
plugin architecture, but it depends on the basic logging infrastructure being in place first.

**Independent Test**: Can be verified by running a command with an installed plugin, capturing logs
from both core and plugin, and confirming the same trace ID appears in both log streams.

**Acceptance Scenarios**:

1. **Given** a trace ID is generated at CLI entry, **When** the core communicates with a plugin
   over gRPC, **Then** the trace ID is propagated in the gRPC metadata and appears in plugin-side
   logs.

2. **Given** a plugin returns an error, **When** the error is logged by the core, **Then** both
   the core error log and any plugin error logs share the same trace ID for correlation.

---

### User Story 3 - Configuring Log Verbosity for Production vs Development (Priority: P3)

An operator wants to run pulumicost in production with minimal log noise (INFO level, JSON format
for aggregation) while developers want verbose console output (DEBUG level, human-readable format).

**Why this priority**: Configuration flexibility is important but the logging must work first.
This story can be delivered after basic logging is functional.

**Independent Test**: Can be tested by setting different configuration values (file, environment,
CLI flag) and confirming the logger respects the settings hierarchy.

**Acceptance Scenarios**:

1. **Given** a configuration file sets `logging.level: warn`, **When** a user runs a command
   without flags, **Then** only WARN and ERROR level logs appear.

2. **Given** no configuration changes, **When** a user runs a command with `--debug` flag,
   **Then** DEBUG level logs appear regardless of configuration file settings.

3. **Given** `PULUMICOST_LOG_LEVEL=error` is set, **When** a configuration file sets
   `logging.level: debug`, **Then** the environment variable takes precedence and only ERROR
   logs appear.

---

### User Story 4 - Injecting External Trace IDs (Priority: P4)

An enterprise user running pulumicost as part of a larger orchestration pipeline wants to inject
their existing trace ID so logs from pulumicost correlate with their broader observability platform.

**Why this priority**: This is an advanced use case for enterprise users. It builds on the core
tracing functionality and can be delivered as an enhancement.

**Independent Test**: Can be tested by setting `PULUMICOST_TRACE_ID` environment variable and
verifying that value appears in all log entries instead of a generated ID.

**Acceptance Scenarios**:

1. **Given** `PULUMICOST_TRACE_ID=external-trace-12345` is set, **When** a user runs any command,
   **Then** all log entries use `external-trace-12345` as the trace_id value.

---

### Edge Cases

- What happens when the log file path is not writable? System logs to stderr and continues
  operation without failing the command.
- How does the system handle very long-running commands with thousands of resources? Trace ID
  remains consistent; logs do not cause performance degradation.
- What if a plugin does not support trace ID propagation? Core logs continue with trace ID;
  plugin logs may be uncorrelated (no failure).
- What happens when an invalid log level is configured? System falls back to INFO level and
  logs a warning about the invalid configuration.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST generate a unique trace identifier at the start of each CLI command
  execution.
- **FR-002**: System MUST include the trace identifier in all log entries produced during that
  command execution.
- **FR-003**: System MUST propagate the trace identifier through gRPC metadata when communicating
  with plugins.
- **FR-004**: System MUST support configurable log levels: TRACE, DEBUG, INFO, WARN, ERROR.
- **FR-005**: System MUST support both JSON and human-readable (console) output formats.
- **FR-006**: System MUST respect the configuration precedence: CLI flag > environment variable >
  configuration file > default.
- **FR-007**: System MUST include standard contextual fields in all log entries: timestamp, level,
  trace_id, component, message.
- **FR-008**: System MUST log the start and completion (with duration) of all major operations:
  command execution, plugin connection, cost calculation.
- **FR-009**: System MUST log all fallback decisions (plugin to spec, spec to none) with
  explanatory messages.
- **FR-010**: System MUST allow injection of external trace identifiers via environment variable.
- **FR-011**: System MUST NOT log sensitive information such as API keys, credentials, or
  authentication tokens.
- **FR-012**: System MUST provide a `--debug` CLI flag that enables DEBUG level logging.
- **FR-013**: System MUST write logs to stderr by default to keep stdout clean for command output.
- **FR-014**: System MUST support file output as an alternative destination.

### Key Entities

- **LogEntry**: A single structured log record containing timestamp, level, trace_id, component,
  operation, message, and optional contextual fields (duration_ms, resource_urn, plugin_name,
  cost_monthly, adapter).
- **TraceContext**: The correlation context containing trace_id and propagated through the entire
  request lifecycle.
- **LoggerConfiguration**: Settings controlling log level, output format, output destination, and
  optional features like caller info and stack traces.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of CLI commands produce at least one log entry showing command start and one
  showing command completion with duration.
- **SC-002**: 100% of log entries produced during a single command execution share the same trace
  identifier.
- **SC-003**: When a plugin is invoked, the trace identifier appears in gRPC metadata (verifiable
  via integration test).
- **SC-004**: JSON log output is valid JSON parseable by standard tools (100% of entries pass
  `jq .` validation).
- **SC-005**: Enabling DEBUG level logging adds no more than 5% overhead to command execution time
  for typical workloads.
- **SC-006**: All fallback scenarios produce a WARN level log entry explaining the fallback reason.
- **SC-007**: The logging package achieves at least 90% test coverage.
- **SC-008**: Zero instances of sensitive data (credentials, tokens) appear in any log output
  regardless of log level.

## Assumptions

- The existing slog-based implementation in `internal/logging/` can be completely replaced with
  zerolog without breaking downstream dependencies (the package is currently unused).
- Plugins that receive trace IDs via gRPC metadata will adopt their own logging standards to use
  them (out of scope for this feature).
- ULID format for trace IDs is acceptable; no specific ID format is mandated by external systems.
- The `github.com/rs/zerolog` package is stable and appropriate for production use.
- Log rotation is out of scope for the initial implementation; operators can use external tools
  like logrotate.
