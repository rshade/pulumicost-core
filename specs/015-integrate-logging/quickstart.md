# Quickstart: Logging Configuration

This guide shows how to configure logging in FinFocus.

## Default Behavior

By default, FinFocus logs to stderr at INFO level in JSON format:

```bash
finfocus cost projected --pulumi-json plan.json
# Logs appear on stderr, command output on stdout
```

## Enable Debug Logging

Use the `--debug` flag for verbose output:

```bash
finfocus --debug cost projected --pulumi-json plan.json
# Debug messages now visible in console format
```

## Configure via Config File

Create `~/.finfocus/config.yaml`:

```yaml
logging:
  level: debug         # trace, debug, info, warn, error
  format: json         # json or console
  file: /var/log/finfocus/app.log

  audit:
    enabled: true
    file: /var/log/finfocus/audit.log  # Optional separate file
```

When file logging is configured, the CLI shows where logs are written:

```bash
$ finfocus cost projected --pulumi-json plan.json
Logging to: /var/log/finfocus/app.log
...
```

## Configure via Environment Variables

Environment variables override config file settings:

```bash
export FINFOCUS_LOG_LEVEL=debug
export FINFOCUS_LOG_FORMAT=console
finfocus cost projected --pulumi-json plan.json
```

## Priority Order

Configuration is applied in this order (later overrides earlier):

1. Config file (`~/.finfocus/config.yaml`)
2. Environment variables (`FINFOCUS_LOG_LEVEL`, `FINFOCUS_LOG_FORMAT`)
3. CLI flags (`--debug`)

## Audit Logging

When audit logging is enabled, all cost queries are logged with:

- Timestamp and trace ID
- Command and parameters
- Duration and results

Example audit entry (JSON format):

```json
{
  "time": "2025-12-01T10:30:05Z",
  "trace_id": "01HQ7X2J3K4M5N6P7Q8R9S0T1U",
  "audit": true,
  "command": "cost projected",
  "parameters": {"pulumi_json": "/path/to/plan.json"},
  "duration_ms": 1234,
  "success": true,
  "result_count": 5,
  "total_cost": 123.45
}
```

Filter audit entries with jq:

```bash
cat /var/log/finfocus/audit.log | jq 'select(.audit == true)'
```

## Log Rotation

FinFocus does not manage log rotation. Use your system's log rotation tools:

**logrotate (Linux)**:

```text
/var/log/finfocus/*.log {
    daily
    rotate 7
    compress
    missingok
    notifempty
}
```

**newsyslog (macOS)**:

```text
/var/log/finfocus/app.log  644  7  *  @T00  J
```

## Troubleshooting

### Logs not appearing in file

Check that the directory exists and is writable:

```bash
mkdir -p /var/log/finfocus
chmod 755 /var/log/finfocus
```

If the file cannot be written, you'll see a warning:

```text
Warning: Cannot write to /var/log/finfocus/app.log, logging to stderr
```

### Finding your trace ID

Every command generates a trace ID for correlation:

```bash
# Inject your own trace ID
export FINFOCUS_TRACE_ID=my-trace-123
finfocus cost projected --pulumi-json plan.json
```

Search logs by trace ID:

```bash
grep "my-trace-123" /var/log/finfocus/app.log
```

## Complete Configuration Examples

### Development Setup

Verbose logging to console for debugging:

```yaml
# ~/.finfocus/config.yaml
logging:
  level: debug
  format: console
  # No file = outputs to stderr
  audit:
    enabled: false
```

### Production Setup

Structured logs to file with audit trail:

```yaml
# ~/.finfocus/config.yaml
logging:
  level: info
  format: json
  file: /var/log/finfocus/finfocus.log
  audit:
    enabled: true
    file: /var/log/finfocus/audit.log
```

### CI/CD Pipeline Setup

JSON logs for log aggregation, with external trace ID:

```bash
export FINFOCUS_LOG_LEVEL=info
export FINFOCUS_LOG_FORMAT=json
export FINFOCUS_TRACE_ID="${CI_PIPELINE_ID}-${CI_JOB_ID}"

finfocus cost projected --pulumi-json plan.json
```

### Security-Sensitive Environment

Audit logging with sensitive data redaction:

```yaml
# ~/.finfocus/config.yaml
logging:
  level: warn  # Minimize log verbosity
  format: json
  file: /var/log/finfocus/finfocus.log
  audit:
    enabled: true
    file: /var/log/finfocus/audit.log
# Note: API keys, passwords, and tokens are automatically redacted
```

## Quick Reference

| Setting      | Config File              | Environment Variable     | CLI Flag  |
| ------------ | ------------------------ | ------------------------ | --------- |
| Log Level    | `logging.level`          | `FINFOCUS_LOG_LEVEL`   | `--debug` |
| Log Format   | `logging.format`         | `FINFOCUS_LOG_FORMAT`  | -         |
| Log File     | `logging.file`           | -                        | -         |
| Trace ID     | -                        | `FINFOCUS_TRACE_ID`    | -         |
| Audit Enable | `logging.audit.enabled`  | -                        | -         |
| Audit File   | `logging.audit.file`     | -                        | -         |
