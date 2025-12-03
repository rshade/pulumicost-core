# Research: Complete Logging Integration

**Date**: 2025-12-01
**Feature**: 007-integrate-logging

## Research Tasks Completed

### 1. Existing Logging Package Analysis

**Question**: What capabilities does the existing `internal/logging` package provide?

**Findings**:

- **zerolog integration**: Already using `github.com/rs/zerolog v1.34.0`
- **Trace ID propagation**: `TracingHook` automatically injects `trace_id` from context
- **Sensitive data redaction**: `SafeStr()` function with `sensitivePatterns` list
- **Configuration support**: `logging.Config` struct with Level, Format, Output, File, Caller fields
- **Component loggers**: `ComponentLogger()` creates sub-loggers with component field
- **Context integration**: `FromContext()` extracts logger from context

**Decision**: Extend existing package rather than replace. Add audit logging as new feature.

**Rationale**: Package is well-designed with 76.9% coverage. Adding features is lower risk
than refactoring.

### 2. Config Package Integration Pattern

**Question**: How does config currently integrate with logging?

**Findings**:

- `config.LoggingConfig` struct exists with Level, Format, File, Outputs fields
- `config.GetLogLevel()` and `config.GetLogFile()` helper methods exist
- CLI currently creates logging config manually in `root.go` PersistentPreRunE
- Environment variable override already implemented (PULUMICOST_LOG_LEVEL, PULUMICOST_LOG_FORMAT)

**Decision**: Create bridge function in config package to convert `config.LoggingConfig` to
`logging.Config`.

**Rationale**: Keeps packages decoupled while providing clean integration path.

### 3. Audit Logging Best Practices

**Question**: What fields should audit entries contain for cost queries?

**Findings** (industry standards):

- **Timestamp**: ISO 8601 format (zerolog provides this)
- **Trace ID**: Request correlation (already have)
- **Command**: Which CLI command was run
- **Parameters**: Relevant input parameters (file paths, date ranges)
- **User context**: Not applicable for CLI tool (no auth)
- **Duration**: How long the operation took
- **Result summary**: Success/failure, resource count, total cost

**Decision**: Audit entries will be structured log entries at INFO level with `audit: true`
field for filtering.

**Rationale**: Using existing log infrastructure avoids complexity. Separate audit file
is configurable but optional.

**Alternatives Considered**:

1. Separate audit logging library - Rejected: Adds dependency, complexity
2. Database-backed audit log - Rejected: Overkill for CLI tool
3. Custom audit format - Rejected: Structured JSON logs are filterable

### 4. Log File Path Display Pattern

**Question**: Where should log file path be displayed in CLI output?

**Findings**:

- Current pattern: Logging starts in `PersistentPreRunE` before any command runs
- Message should go to stdout (not stderr) to keep separate from log output
- Should only display when logging to file (not stderr/stdout)

**Decision**: Print "Logging to: /path/to/file" to stdout in PersistentPreRunE when file
logging is configured.

**Rationale**: Early display ensures operators see the path before any operation.
Stdout keeps it separate from log stream.

### 5. Graceful Fallback Pattern

**Question**: How should file write failures be handled?

**Findings**:

- Current `createWriter()` in zerolog.go already falls back to stderr on file open error
- Logs warning but doesn't display prominently to user

**Decision**: Add explicit user-facing warning to stdout when fallback occurs.

**Rationale**: Operators need to know their logs aren't going where expected.

## Technology Decisions Summary

| Decision | Choice | Rationale |
| -------- | ------ | --------- |
| Audit log format | Structured JSON with `audit: true` | Filterable, consistent with existing logs |
| Audit log destination | Configurable (same file or separate) | Flexibility for compliance needs |
| Config bridge | Function in config package | Clean separation, no circular imports |
| Log path display | stdout in PersistentPreRunE | Early visibility, separate from log stream |
| Fallback notification | stdout warning message | Ensures operator awareness |

## No Clarifications Needed

All technical decisions resolved through analysis of existing code and industry best practices.
