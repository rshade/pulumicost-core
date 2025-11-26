# Data Model: Zerolog Distributed Tracing

**Feature Branch**: `004-zerolog-tracing`
**Date**: 2025-11-25

## Entities

### LogEntry

A structured log record produced by the logging system.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `time` | RFC3339 timestamp | Yes | When the log entry was created |
| `level` | string | Yes | Log level: trace, debug, info, warn, error |
| `trace_id` | string | Yes* | Request correlation ID (ULID format, 26 chars) |
| `component` | string | Yes | Package name: cli, engine, registry, pluginhost, ingest, spec |
| `message` | string | Yes | Human-readable log message |
| `operation` | string | No | Function or operation name |
| `duration_ms` | float64 | No | Execution time in milliseconds |
| `resource_urn` | string | No | Pulumi resource URN being processed |
| `plugin_name` | string | No | Plugin identifier (e.g., "kubecost") |
| `cost_monthly` | float64 | No | Calculated monthly cost |
| `adapter` | string | No | Cost source: "plugin", "spec", "none" |
| `error` | string | No | Error message (for error-level logs) |
| `caller` | string | No | File:line of log call (optional, for debug) |

*trace_id is required when context is available; may be absent for startup logs before context
is created.

**JSON Example:**

```json
{
  "time": "2025-11-25T10:30:00.001Z",
  "level": "info",
  "trace_id": "01JDP8K2M3N4P5Q6R7S8T9V0W1",
  "component": "engine",
  "operation": "CalculateCosts",
  "message": "cost calculation complete",
  "duration_ms": 125.5,
  "resource_urn": "urn:pulumi:dev::myproject::aws:ec2/instance:Instance::webserver",
  "cost_monthly": 62.59,
  "adapter": "plugin"
}
```

**Console Example:**

```text
10:30:00 INF cost calculation complete component=engine duration_ms=125.5 trace_id=01JDP8K2M3...
```

---

### TraceContext

The correlation context propagated through the request lifecycle.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `trace_id` | string | Yes | ULID or externally-injected identifier |

**Context Key:** `pulumicost.trace_id` (typed context key, not string)

**Propagation:**

1. Generated at CLI entry point (`cmd/pulumicost/main.go`)
2. Stored in `context.Context` via `context.WithValue()`
3. Extracted by TracingHook for automatic injection into log events
4. Propagated to plugins via gRPC metadata key `x-pulumicost-trace-id`

**State Transitions:** None (immutable once created)

---

### LoggerConfiguration

Settings controlling logger behavior.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `level` | string | "info" | Minimum log level: trace, debug, info, warn, error |
| `format` | string | "json" | Output format: json, console |
| `output` | string | "stderr" | Output destination: stderr, stdout, file |
| `file` | string | "" | File path when output=file |
| `caller` | bool | false | Include file:line in output |
| `stack_trace` | bool | true | Include stack trace on errors |

**YAML Configuration Example:**

```yaml
logging:
  level: info
  format: json
  output: stderr
  file: ""
  caller: false
  stack_trace: true
```

**Environment Variable Overrides:**

| Variable | Overrides |
|----------|-----------|
| `PULUMICOST_LOG_LEVEL` | logging.level |
| `PULUMICOST_LOG_FORMAT` | logging.format |
| `PULUMICOST_TRACE_ID` | Injects external trace ID |

---

## Validation Rules

### LogLevel Validation

- Must be one of: trace, debug, info, warn, error
- Case-insensitive comparison
- Invalid values fall back to "info" with warning log

### LogFormat Validation

- Must be one of: json, console, text (text is alias for console)
- Invalid values fall back to "json"

### TraceID Validation

- No validation on format (accepts any non-empty string)
- External trace IDs accepted as-is for integration with other systems
- Generated trace IDs use ULID format (26 uppercase alphanumeric characters)

### File Path Validation

- Must be absolute path if specified
- Parent directory must exist or be creatable
- File must be writable
- Falls back to stderr on error (does not fail command)

---

## Relationships

```text
┌─────────────────┐
│  CLI Command    │
│  (generates     │
│   trace_id)     │
└────────┬────────┘
         │
         │ context.Context with trace_id
         ▼
┌─────────────────┐     ┌─────────────────┐
│     Engine      │────▶│    Registry     │
│  (logs with     │     │  (logs with     │
│   trace_id)     │     │   trace_id)     │
└────────┬────────┘     └─────────────────┘
         │
         │ gRPC call with x-pulumicost-trace-id
         ▼
┌─────────────────┐
│   PluginHost    │
│  (propagates    │
│   trace_id)     │
└────────┬────────┘
         │
         │ metadata: x-pulumicost-trace-id
         ▼
┌─────────────────┐
│     Plugin      │
│  (receives      │
│   trace_id)     │
└─────────────────┘
```

---

## Component Identifiers

Standard component names for the `component` field:

| Component | Package | Description |
|-----------|---------|-------------|
| `cli` | internal/cli | Command-line interface handlers |
| `engine` | internal/engine | Cost calculation orchestration |
| `registry` | internal/registry | Plugin discovery and lifecycle |
| `pluginhost` | internal/pluginhost | gRPC plugin communication |
| `ingest` | internal/ingest | Pulumi plan parsing |
| `spec` | internal/spec | Local pricing spec loading |
| `config` | internal/config | Configuration management |
| `main` | cmd/pulumicost | Entry point and initialization |
