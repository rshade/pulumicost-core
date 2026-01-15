# Research: Zerolog Distributed Tracing

**Feature Branch**: `004-zerolog-tracing`
**Date**: 2025-11-25

## Research Summary

This document captures research findings for implementing structured logging with zerolog and
distributed tracing throughout finfocus-core.

---

## 1. Zerolog Logger Configuration

### Decision

Use zerolog v1.34.0 with TracingHook pattern for automatic trace ID injection.

### Rationale

- zerolog is already a dependency in go.mod (v1.34.0)
- Zero-allocation JSON logging provides best performance
- TracingHook pattern allows automatic trace ID injection without manual passing
- Context integration via `zerolog.Ctx(ctx)` and `logger.WithContext(ctx)` enables clean API

### Alternatives Considered

| Alternative | Rejected Because |
|-------------|------------------|
| Continue with slog | Less mature ecosystem, no zero-allocation, already unused |
| Zap | More complex API, zerolog already in go.mod |
| Custom logging | Unnecessary reinvention, no ecosystem benefits |

### Implementation Pattern

```go
// TracingHook extracts trace_id from context and adds to all log events
type TracingHook struct{}

func (h TracingHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
    ctx := e.GetCtx()
    if ctx == nil {
        return
    }
    if traceID, ok := ctx.Value(traceIDKey).(string); ok {
        e.Str("trace_id", traceID)
    }
}

// Logger factory
func NewLogger(cfg LoggingConfig) zerolog.Logger {
    var output io.Writer
    if cfg.Format == "console" {
        output = zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
    } else {
        output = os.Stderr
    }

    return zerolog.New(output).
        With().
        Timestamp().
        Logger().
        Hook(TracingHook{}).
        Level(parseLevel(cfg.Level))
}
```

---

## 2. Trace ID Generation

### Decision

Use ULID format via `github.com/oklog/ulid/v2` for trace ID generation.

### Rationale

- ULIDs are lexicographically sortable (useful for log analysis)
- 26-character canonical string representation (compact)
- Monotonic within same millisecond (ensures ordering)
- No external dependencies beyond the ulid package
- Compatible with external trace ID injection (any string accepted)

### Alternatives Considered

| Alternative | Rejected Because |
|-------------|------------------|
| UUID v4 | Not sortable, longer string representation |
| UUID v7 | Less mature ecosystem, ULID equally good |
| Snowflake ID | Over-engineered for this use case |
| Random string | No ordering guarantees, harder to debug |

### Implementation Pattern

```go
import (
    "crypto/rand"
    "github.com/oklog/ulid/v2"
)

func GenerateTraceID() string {
    return ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String()
}

func GetOrGenerateTraceID(ctx context.Context) string {
    // Check for external trace ID first
    if envTraceID := os.Getenv("FINFOCUS_TRACE_ID"); envTraceID != "" {
        return envTraceID
    }
    // Check context
    if traceID, ok := ctx.Value(traceIDKey).(string); ok && traceID != "" {
        return traceID
    }
    // Generate new
    return GenerateTraceID()
}
```

---

## 3. gRPC Metadata Propagation

### Decision

Use gRPC client unary interceptor to inject trace ID into outgoing metadata with key
`x-finfocus-trace-id`.

### Rationale

- Standard gRPC pattern for distributed tracing
- Interceptor pattern ensures all RPCs automatically include trace ID
- Metadata key follows HTTP header naming conventions (lowercase, hyphenated)
- Compatible with OpenTelemetry-style propagation if needed later

### Alternatives Considered

| Alternative | Rejected Because |
|-------------|------------------|
| Pass trace ID in each RPC request | Invasive, requires proto changes |
| OpenTelemetry SDK | Over-engineered for current needs, can add later |
| Custom wire format | Non-standard, harder to debug |

### Implementation Pattern

```go
import (
    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
)

const TraceIDMetadataKey = "x-finfocus-trace-id"

// UnaryClientInterceptor injects trace_id into outgoing gRPC metadata
func TraceInterceptor() grpc.UnaryClientInterceptor {
    return func(
        ctx context.Context,
        method string,
        req, reply interface{},
        cc *grpc.ClientConn,
        invoker grpc.UnaryInvoker,
        opts ...grpc.CallOption,
    ) error {
        if traceID, ok := ctx.Value(traceIDKey).(string); ok {
            ctx = metadata.AppendToOutgoingContext(ctx, TraceIDMetadataKey, traceID)
        }
        return invoker(ctx, method, req, reply, cc, opts...)
    }
}

// Apply when creating gRPC connection
conn, err := grpc.Dial(address,
    grpc.WithUnaryInterceptor(TraceInterceptor()),
)
```

---

## 4. Log Level Mapping

### Decision

Map five log levels to zerolog levels: TRACE→Trace, DEBUG→Debug, INFO→Info, WARN→Warn,
ERROR→Error.

### Rationale

- Direct mapping to zerolog's native levels
- zerolog supports TraceLevel (lower than Debug)
- Consistent with industry standards
- Configuration validation ensures only valid levels accepted

### Implementation Pattern

```go
func parseLevel(level string) zerolog.Level {
    switch strings.ToLower(level) {
    case "trace":
        return zerolog.TraceLevel
    case "debug":
        return zerolog.DebugLevel
    case "info":
        return zerolog.InfoLevel
    case "warn", "warning":
        return zerolog.WarnLevel
    case "error":
        return zerolog.ErrorLevel
    default:
        return zerolog.InfoLevel // Default fallback
    }
}
```

---

## 5. Output Format Strategy

### Decision

Support two output formats: JSON (default for production) and Console (for development).

### Rationale

- JSON format enables log aggregation tools (Loki, CloudWatch, Datadog)
- Console format with colors improves developer experience
- Format selection via config file, environment variable, or --debug flag
- --debug flag implies console format and DEBUG level for convenience

### Implementation Pattern

```go
func createWriter(cfg LoggingConfig) io.Writer {
    var output io.Writer = os.Stderr

    if cfg.Output == "file" && cfg.File != "" {
        f, err := os.OpenFile(cfg.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
        if err != nil {
            // Fall back to stderr, log warning
            fmt.Fprintf(os.Stderr, "Warning: cannot open log file %s: %v\n", cfg.File, err)
        } else {
            output = f
        }
    }

    if cfg.Format == "console" || cfg.Format == "text" {
        return zerolog.ConsoleWriter{
            Out:        output,
            TimeFormat: time.RFC3339,
            NoColor:    cfg.NoColor,
        }
    }

    return output // JSON format (zerolog default)
}
```

---

## 6. Configuration Precedence

### Decision

Configuration precedence: CLI flag > Environment variable > Config file > Default.

### Rationale

- Standard Unix convention for configuration layering
- CLI flags allow per-invocation overrides
- Environment variables enable container/CI configuration
- Config file provides persistent defaults
- Sensible defaults ensure zero-config works

### Implementation Pattern

```go
// In CLI command setup
func resolveLogLevel(flagValue string) string {
    // 1. CLI flag takes precedence
    if flagValue != "" {
        return flagValue
    }
    // 2. Environment variable
    if envLevel := os.Getenv("FINFOCUS_LOG_LEVEL"); envLevel != "" {
        return envLevel
    }
    // 3. Config file (already loaded into cfg)
    if cfg.Logging.Level != "" {
        return cfg.Logging.Level
    }
    // 4. Default
    return "info"
}
```

---

## 7. Sensitive Data Protection

### Decision

Never log values for keys matching sensitive patterns. Implement allowlist/blocklist approach.

### Rationale

- Prevents accidental credential exposure
- Standard security practice (OWASP)
- Pattern-based detection catches common variations
- Explicit blocklist is more maintainable than trying to sanitize

### Sensitive Patterns to Block

- `api_key`, `apikey`, `api-key`
- `password`, `passwd`, `pwd`
- `secret`, `token`
- `credential`, `cred`
- `private_key`, `privatekey`
- `auth`, `authorization`
- `bearer`

### Implementation Pattern

```go
var sensitivePatterns = []string{
    "api_key", "apikey", "api-key",
    "password", "passwd", "pwd",
    "secret", "token",
    "credential", "cred",
    "private_key", "privatekey",
    "auth", "authorization", "bearer",
}

func isSensitiveKey(key string) bool {
    lower := strings.ToLower(key)
    for _, pattern := range sensitivePatterns {
        if strings.Contains(lower, pattern) {
            return true
        }
    }
    return false
}

// Use in logging calls - redact sensitive values
func SafeStr(e *zerolog.Event, key, value string) *zerolog.Event {
    if isSensitiveKey(key) {
        return e.Str(key, "[REDACTED]")
    }
    return e.Str(key, value)
}
```

---

## 8. Component Logger Pattern

### Decision

Create sub-loggers with `component` field for each package (cli, engine, registry, pluginhost,
ingest, spec).

### Rationale

- Enables filtering logs by component
- Clear origin of each log message
- Consistent with microservice logging patterns
- Sub-loggers inherit parent configuration

### Implementation Pattern

```go
// In each package, create a component logger
var logger zerolog.Logger

func init() {
    // Will be replaced when main initializes logging
    logger = zerolog.Nop()
}

func SetLogger(l zerolog.Logger) {
    logger = l.With().Str("component", "engine").Logger()
}

// Usage in package
func CalculateCosts(ctx context.Context, resources []Resource) {
    logger.Info().
        Ctx(ctx).
        Int("resource_count", len(resources)).
        Msg("starting cost calculation")

    // ... calculation logic ...

    logger.Info().
        Ctx(ctx).
        Float64("total_monthly", total).
        Dur("duration_ms", elapsed).
        Msg("cost calculation complete")
}
```

---

## References

- [zerolog GitHub](https://github.com/rs/zerolog)
- [gRPC-Go Metadata Interceptor Example](https://github.com/grpc/grpc-go/tree/master/examples/features/metadata_interceptor)
- [gRPC-Go Metadata Documentation](https://github.com/grpc/grpc-go/blob/master/Documentation/grpc-metadata.md)
- [ULID Specification](https://github.com/ulid/spec)
- [oklog/ulid Go Implementation](https://github.com/oklog/ulid)
