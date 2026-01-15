# Feature Specification: Complete Logging Integration

**Feature Branch**: `007-integrate-logging`
**Created**: 2025-12-01
**Status**: Draft
**Input**: GitHub Issue #124 - Logging package exists but isn't used

## Context

The logging package (`internal/logging`) has been partially integrated into CLI, engine, and pluginhost components. However, the integration is incomplete:

1. **Current State**: Basic structured logging with trace IDs works in CLI, engine, and pluginhost
2. **Missing**: Audit logging for cost queries, configuration persistence (log rotation is intentionally not in scope - operators use external tools)
3. **Coverage**: Logging package has 76.9% test coverage

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Configure Logging Behavior (Priority: P1)

As an operator deploying FinFocus in a production environment, I want to configure logging behavior through configuration files so that logs integrate with my existing log management infrastructure.

**Why this priority**: Production deployments require configurable logging for operational visibility and compliance.
Without configuration, operators cannot integrate FinFocus with existing monitoring systems.

**Independent Test**: Can be fully tested by setting logging configuration and verifying log output format, level, and destination match the configuration.

**Acceptance Scenarios**:

1. **Given** a configuration file with `logging.level: debug`, **When** I run any finfocus command, **Then** debug-level messages appear in the output
2. **Given** a configuration file with `logging.format: json`, **When** I run any finfocus command, **Then** all log output is valid JSON with consistent schema
3. **Given** a configuration file with `logging.output: file` and `logging.file: /var/log/finfocus.log`, **When** I run any finfocus command, **Then** logs are written to the specified file

---

### User Story 2 - Clear Log Location Communication (Priority: P2)

As an operator, I want to see where logs are being written when I run commands so that I know where to look for troubleshooting and can configure external log rotation tools.

**Why this priority**: Without clear communication about log destinations, operators waste time searching for logs or miss important diagnostic information. This is essential for production operations.

**Independent Test**: Can be fully tested by running any command with file logging enabled and verifying the log file path is displayed in CLI output.

**Acceptance Scenarios**:

1. **Given** logging configured to write to a file, **When** I run any finfocus command, **Then** the CLI displays the log file path at startup (e.g., "Logging to: /var/log/finfocus.log")
2. **Given** logging configured to write to stderr (default), **When** I run any finfocus command with --debug, **Then** no log path message is shown (logs appear inline)
3. **Given** logging configured to a file that cannot be written, **When** I run any finfocus command, **Then** the CLI displays a warning about the fallback to stderr

---

### User Story 3 - Audit Logging for Cost Queries (Priority: P3)

As a security auditor, I want to track all cost queries with timestamps, users, and query parameters so that I can review who accessed what cost data and when.

**Why this priority**: Audit logging is important for compliance and security but is not essential for core cost calculation functionality.

**Independent Test**: Can be fully tested by running cost queries and verifying audit log entries contain required information.

**Acceptance Scenarios**:

1. **Given** audit logging is enabled, **When** I run `finfocus cost projected --pulumi-json plan.json`, **Then** an audit entry is logged with timestamp, trace_id, command, and input file path
2. **Given** audit logging is enabled, **When** I run `finfocus cost actual` with date range flags,
   **Then** an audit entry is logged with timestamp, trace_id, command, and date range
3. **Given** audit logging writes to a separate audit log file, **When** I review the audit log, **Then** I can filter by date, command type, or trace_id

---

### Edge Cases

- What happens when the log file destination is not writable? System falls back to stderr with warning displayed to user
- What happens when disk is full? Logs to stderr with error, continues operation
- What happens when configuration has invalid log level? Uses default (info) level with warning
- How does system behave when audit log file is locked by another process? Retries briefly, then logs warning and continues without audit

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST read logging configuration from `~/.finfocus/config.yaml` under the `logging` section
- **FR-002**: System MUST support configuration of log level (trace, debug, info, warn, error)
- **FR-003**: System MUST support configuration of log format (json, console/text)
- **FR-004**: System MUST support configuration of log output destination (stderr, stdout, file)
- **FR-005**: System MUST allow log file path configuration when output is set to file
- **FR-006**: System MUST display log file path in CLI output when logging to a file
- **FR-007**: System MUST display a warning when falling back to stderr due to file write failure
- **FR-008**: System MUST write audit entries for all cost query operations
- **FR-009**: Audit entries MUST include timestamp, trace_id, command name, and relevant parameters
- **FR-010**: System MUST gracefully fall back to stderr if configured file destination is unavailable
- **FR-011**: CLI flags (--debug) MUST override configuration file settings
- **FR-012**: Environment variables (FINFOCUS_LOG_LEVEL, FINFOCUS_LOG_FORMAT) MUST override configuration file settings
- **FR-013**: System MUST automatically redact sensitive data patterns (api_key, password, token, secret, credential, auth) in all log output

### Configuration Schema

```yaml
logging:
  level: info          # trace, debug, info, warn, error
  format: json         # json, console, text
  output: stderr       # stderr, stdout, file
  file: ""             # path when output is "file"
  caller: false        # include file:line in logs

  audit:
    enabled: false
    file: ""           # separate audit log file (optional)
```

**Note**: Log rotation is intentionally not managed by FinFocus. Operators should use external tools (logrotate, systemd journal, etc.) to manage log file rotation based on their infrastructure requirements.

### Key Entities

- **Log Entry**: Structured log record with timestamp, level, message, trace_id, component, and contextual fields
- **Audit Entry**: Specialized log entry for cost queries containing command, parameters, duration, and result summary

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can configure logging via config file and see changes take effect immediately on next command
- **SC-002**: When file logging is enabled, operators can immediately identify where logs are written from CLI output before any command processing begins
- **SC-003**: 100% of cost commands generate audit log entries when audit logging is enabled
- **SC-004**: System continues operating normally if log file cannot be written (graceful degradation)
- **SC-005**: All log entries include trace_id for request correlation across components
- **SC-006**: Configuration errors are reported clearly with actionable messages

## Clarifications

### Session 2025-12-01

- Q: Should the spec explicitly require automatic redaction of sensitive data in all log output? → A: Yes, require automatic redaction of sensitive patterns (api_key, password, token, etc.)

## Assumptions

- Log rotation is the responsibility of the operator using external tools (logrotate, systemd, etc.)
- Audit logs follow the same structured format as regular logs but with additional fields
- Log file path messages are written to stdout (not stderr) to keep them separate from log output
