# Data Model: Complete Logging Integration

**Date**: 2025-12-01
**Feature**: 007-integrate-logging

## Entities

### 1. LoggingConfig (Existing - config package)

Configuration for logging behavior read from `~/.pulumicost/config.yaml`.

```go
// LoggingConfig defines logging preferences (existing in config/config.go)
type LoggingConfig struct {
    Level   string      `yaml:"level"   json:"level"`   // trace, debug, info, warn, error
    Format  string      `yaml:"format"  json:"format"`  // json, console, text
    Outputs []LogOutput `yaml:"outputs" json:"outputs"` // Multiple destinations
    File    string      `yaml:"file"    json:"file"`    // Legacy single file

    // NEW: Audit configuration
    Audit   AuditConfig `yaml:"audit"   json:"audit"`
}

// AuditConfig defines audit logging settings (NEW)
type AuditConfig struct {
    Enabled bool   `yaml:"enabled" json:"enabled"` // Enable audit logging
    File    string `yaml:"file"    json:"file"`    // Separate audit file (optional)
}
```

**Validation Rules**:

- Level must be one of: trace, debug, info, warn, error (default: info)
- Format must be one of: json, console, text (default: json)
- File path must be writable or fallback to stderr
- Audit.File is optional; if empty, audit logs go to main log

### 2. Config (Existing - logging package)

Runtime logging configuration used by zerolog.

```go
// Config holds logging configuration settings (existing in logging/zerolog.go)
type Config struct {
    Level      string // Log level: trace, debug, info, warn, error
    Format     string // Output format: json, console, text
    Output     string // Output destination: stderr, stdout, file
    File       string // File path when Output is "file"
    Caller     bool   // Include file:line in output
    StackTrace bool   // Include stack trace on errors
}
```

**Relationship**: `config.LoggingConfig` → `logging.Config` via bridge function.

### 3. AuditEntry (NEW - logging package)

Represents a single audit log entry for cost queries.

```go
// AuditEntry represents an audit log record for cost operations (NEW)
type AuditEntry struct {
    Timestamp   time.Time         // When the operation occurred
    TraceID     string            // Request correlation ID
    Command     string            // CLI command name (e.g., "cost projected")
    Parameters  map[string]string // Relevant parameters (file path, dates, etc.)
    Duration    time.Duration     // How long the operation took
    Success     bool              // Whether operation succeeded
    ResultCount int               // Number of results returned
    TotalCost   float64           // Total cost calculated (if applicable)
    Error       string            // Error message if failed
}
```

**Fields**:

| Field | Type | Required | Description |
| ----- | ---- | -------- | ----------- |
| Timestamp | time.Time | Yes | UTC timestamp of operation |
| TraceID | string | Yes | ULID trace identifier |
| Command | string | Yes | CLI command path |
| Parameters | map[string]string | Yes | Input parameters (redacted if sensitive) |
| Duration | time.Duration | Yes | Operation duration |
| Success | bool | Yes | True if operation completed without error |
| ResultCount | int | No | Number of cost results returned |
| TotalCost | float64 | No | Aggregate cost if calculated |
| Error | string | No | Error message if Success=false |

**Lifecycle**:

1. Created at start of cost command
2. Parameters populated from command flags
3. Duration, Success, ResultCount, TotalCost populated after operation
4. Written to audit log before command returns

### 4. AuditLogger (NEW - logging package)

Interface for writing audit entries.

```go
// AuditLogger writes audit entries (NEW)
type AuditLogger interface {
    // Log writes an audit entry
    Log(ctx context.Context, entry AuditEntry)

    // Enabled returns whether audit logging is active
    Enabled() bool
}

// zerologAuditLogger implements AuditLogger using zerolog
type zerologAuditLogger struct {
    logger  zerolog.Logger
    enabled bool
}
```

## State Transitions

### Logging Initialization Flow

```text
CLI Start
    │
    ▼
Load config.yaml
    │
    ▼
Check --debug flag ──────────────────────┐
    │                                     │
    ▼                                     │
Check PULUMICOST_LOG_LEVEL env ──────────┤
    │                                     │
    ▼                                     │
Apply config.LoggingConfig ◄─────────────┘
    │                    (overrides in priority order)
    ▼
Convert to logging.Config
    │
    ▼
Create zerolog logger
    │
    ├─── Output=file ──► Open file ──► Success ──► Print path to stdout
    │                        │
    │                        └──► Failure ──► Print warning, use stderr
    │
    └─── Output=stderr/stdout ──► Use directly
    │
    ▼
Store logger in context
    │
    ▼
Initialize AuditLogger (if enabled)
```

### Audit Entry Lifecycle

```text
Command Start
    │
    ▼
Create AuditEntry {
    Timestamp: now,
    TraceID: from context,
    Command: cmd.Name(),
    Parameters: from flags
}
    │
    ▼
Execute command logic
    │
    ├─── Success ──► entry.Success = true
    │                entry.Duration = elapsed
    │                entry.ResultCount = len(results)
    │                entry.TotalCost = sum(costs)
    │
    └─── Error ──► entry.Success = false
                   entry.Error = err.Error()
                   entry.Duration = elapsed
    │
    ▼
AuditLogger.Log(ctx, entry)
    │
    ▼
Command returns
```

## Relationships

```text
┌─────────────────┐         ┌─────────────────┐
│ config.Config   │         │ logging.Config  │
│                 │ bridge  │                 │
│ LoggingConfig ──┼────────►│ Level           │
│   Level         │         │ Format          │
│   Format        │         │ Output          │
│   File          │         │ File            │
│   Audit         │         └─────────────────┘
└─────────────────┘                 │
        │                           │
        │                           ▼
        │                   ┌─────────────────┐
        │                   │ zerolog.Logger  │
        │                   └─────────────────┘
        │                           │
        ▼                           ▼
┌─────────────────┐         ┌─────────────────┐
│ AuditConfig     │────────►│ AuditLogger     │
│   Enabled       │         │                 │
│   File          │         │ Log(entry)      │
└─────────────────┘         │ Enabled()       │
                            └─────────────────┘
                                    │
                                    ▼
                            ┌─────────────────┐
                            │ AuditEntry      │
                            │                 │
                            │ Timestamp       │
                            │ TraceID         │
                            │ Command         │
                            │ Parameters      │
                            │ Duration        │
                            │ Success         │
                            │ ResultCount     │
                            │ TotalCost       │
                            │ Error           │
                            └─────────────────┘
```

## Configuration File Schema

```yaml
# ~/.pulumicost/config.yaml
logging:
  level: info          # trace, debug, info, warn, error
  format: json         # json, console, text
  file: ""             # path when writing to file (empty = stderr)

  audit:
    enabled: false     # enable audit logging
    file: ""           # separate audit log file (empty = main log)
```

## JSON Log Schema

### Standard Log Entry

```json
{
  "level": "info",
  "time": "2025-12-01T10:30:00Z",
  "trace_id": "01HQ7X2J3K4M5N6P7Q8R9S0T1U",
  "component": "cli",
  "message": "command started",
  "command": "cost projected"
}
```

### Audit Log Entry

```json
{
  "level": "info",
  "time": "2025-12-01T10:30:05Z",
  "trace_id": "01HQ7X2J3K4M5N6P7Q8R9S0T1U",
  "component": "audit",
  "audit": true,
  "command": "cost projected",
  "parameters": {
    "pulumi_json": "/path/to/plan.json",
    "output": "table"
  },
  "duration_ms": 1234,
  "success": true,
  "result_count": 5,
  "total_cost": 123.45,
  "message": "cost query completed"
}
```
