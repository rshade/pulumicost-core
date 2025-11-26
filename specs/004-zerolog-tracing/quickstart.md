# Quickstart: Zerolog Distributed Tracing

**Feature Branch**: `004-zerolog-tracing`
**Date**: 2025-11-25

## For Users

### Enable Debug Logging

Use the `--debug` flag on any command to enable verbose logging:

```bash
pulumicost cost projected --pulumi-json plan.json --debug
```

Output will show the complete decision flow:

```text
10:30:00 INF command started component=cli operation=cost_projected trace_id=01JDP8K2...
10:30:00 DBG loading pulumi plan component=ingest file=plan.json trace_id=01JDP8K2...
10:30:00 DBG parsed 3 resources component=ingest resource_count=3 trace_id=01JDP8K2...
10:30:00 DBG looking up plugin component=registry resource_type=aws:ec2:Instance trace_id=01JDP8K2...
10:30:00 WRN plugin returned no price, falling back to spec component=engine adapter=spec trace_id=01JDP8K2...
10:30:00 INF cost calculation complete component=engine total_monthly=62.59 trace_id=01JDP8K2...
10:30:00 INF command complete component=cli duration_ms=250 trace_id=01JDP8K2...
```

### Configure via Environment Variables

```bash
# Set log level
export PULUMICOST_LOG_LEVEL=debug

# Set output format (json for log aggregation, console for development)
export PULUMICOST_LOG_FORMAT=json

# Inject external trace ID (for pipeline integration)
export PULUMICOST_TRACE_ID=external-pipeline-trace-12345
```

### Configure via Config File

Edit `~/.pulumicost/config.yaml`:

```yaml
logging:
  level: info           # trace, debug, info, warn, error
  format: json          # json, console
  output: stderr        # stderr, stdout, file
  file: ""              # path when output=file
```

---

## For Developers

### Adding Logging to New Code

1. **Import the logging package:**

```go
import "github.com/rshade/pulumicost-core/internal/logging"
```

2. **Get logger from context:**

```go
func ProcessResource(ctx context.Context, resource Resource) error {
    log := logging.FromContext(ctx).With().
        Str("component", "engine").
        Str("operation", "ProcessResource").
        Logger()

    log.Info().
        Str("resource_urn", resource.URN).
        Msg("processing resource")

    // ... process logic ...

    if err != nil {
        log.Error().Err(err).Msg("resource processing failed")
        return err
    }

    log.Debug().
        Float64("cost_monthly", cost).
        Msg("resource cost calculated")

    return nil
}
```

3. **Log operation duration:**

```go
func CalculateCosts(ctx context.Context, resources []Resource) (float64, error) {
    log := logging.FromContext(ctx)
    start := time.Now()

    log.Info().
        Int("resource_count", len(resources)).
        Msg("starting cost calculation")

    // ... calculation logic ...

    log.Info().
        Float64("total_monthly", total).
        Dur("duration_ms", time.Since(start)).
        Msg("cost calculation complete")

    return total, nil
}
```

### Log Level Guidelines

| Level | Use For |
|-------|---------|
| TRACE | Property extraction, detailed calculations |
| DEBUG | Function entry/exit, retries, intermediate values |
| INFO | High-level operations (command start/end, major milestones) |
| WARN | Recoverable issues (fallbacks, deprecations) |
| ERROR | Failures needing attention |

### Required Context Fields

Always include these fields when available:

```go
log.Info().
    Str("trace_id", traceID).      // Automatic via hook
    Str("component", "engine").     // Package name
    Str("operation", "FuncName").   // Function being performed
    Msg("message")
```

---

## For Testing

### Capture Logs in Tests

```go
func TestCostCalculation(t *testing.T) {
    // Create buffer to capture logs
    var buf bytes.Buffer
    logger := zerolog.New(&buf).With().Timestamp().Logger()

    // Create context with logger
    ctx := logger.WithContext(context.Background())
    ctx = logging.ContextWithTraceID(ctx, "test-trace-123")

    // Run function under test
    result, err := CalculateCosts(ctx, resources)

    // Assert on logs
    logOutput := buf.String()
    assert.Contains(t, logOutput, "test-trace-123")
    assert.Contains(t, logOutput, "cost calculation complete")
}
```

### Disable Logging in Tests

```go
func TestQuietOperation(t *testing.T) {
    // Use no-op logger
    logger := zerolog.Nop()
    ctx := logger.WithContext(context.Background())

    // Function runs silently
    result, err := CalculateCosts(ctx, resources)
}
```

---

## Trace ID Correlation

### Finding Related Logs

All logs from a single command share the same trace_id. Use it to filter:

```bash
# With jq
pulumicost cost projected --pulumi-json plan.json 2>&1 | jq 'select(.trace_id == "01JDP8K2M3N4P5Q6R7S8T9V0W1")'

# In log aggregation tools (Loki example)
{app="pulumicost"} | json | trace_id="01JDP8K2M3N4P5Q6R7S8T9V0W1"
```

### Pipeline Integration

Inject your pipeline's trace ID to correlate with broader observability:

```bash
# GitHub Actions example
export PULUMICOST_TRACE_ID="gh-$GITHUB_RUN_ID-$GITHUB_RUN_ATTEMPT"
pulumicost cost projected --pulumi-json plan.json

# Jenkins example
export PULUMICOST_TRACE_ID="jenkins-$BUILD_ID"
pulumicost cost projected --pulumi-json plan.json
```

### Plugin Correlation

Plugins receive the trace ID via gRPC metadata. Plugin logs should include the same trace_id
for end-to-end tracing:

```go
// In plugin code
func (s *server) GetProjectedCost(ctx context.Context, req *pb.Request) (*pb.Response, error) {
    md, _ := metadata.FromIncomingContext(ctx)
    traceID := md.Get("x-pulumicost-trace-id")

    s.logger.Info().
        Str("trace_id", traceID[0]).
        Msg("processing cost request")

    // ...
}
```
