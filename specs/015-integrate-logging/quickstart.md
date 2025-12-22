# Quickstart: Logging Configuration

This guide shows how to configure logging in PulumiCost.

## Default Behavior

By default, PulumiCost logs to stderr at INFO level in JSON format:

```bash
pulumicost cost projected --pulumi-json plan.json
# Logs appear on stderr, command output on stdout
```

## Enable Debug Logging

Use the `--debug` flag for verbose output:

```bash
pulumicost --debug cost projected --pulumi-json plan.json
# Debug messages now visible in console format
```

## Configure via Config File

Create `~/.pulumicost/config.yaml`:

```yaml
logging:
  level: debug         # trace, debug, info, warn, error
  format: json         # json or console
  file: /var/log/pulumicost/app.log

  audit:
    enabled: true
    file: /var/log/pulumicost/audit.log  # Optional separate file
```

When file logging is configured, the CLI shows where logs are written:

```bash
$ pulumicost cost projected --pulumi-json plan.json
Logging to: /var/log/pulumicost/app.log
...
```

## Configure via Environment Variables

Environment variables override config file settings:

```bash
export PULUMICOST_LOG_LEVEL=debug
export PULUMICOST_LOG_FORMAT=console
pulumicost cost projected --pulumi-json plan.json
```

## Priority Order

Configuration is applied in this order (later overrides earlier):

1. Config file (`~/.pulumicost/config.yaml`)
2. Environment variables (`PULUMICOST_LOG_LEVEL`, `PULUMICOST_LOG_FORMAT`)
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
cat /var/log/pulumicost/audit.log | jq 'select(.audit == true)'
```

## Log Rotation

PulumiCost does not manage log rotation. Use your system's log rotation tools:

**logrotate (Linux)**:

```text
/var/log/pulumicost/*.log {
    daily
    rotate 7
    compress
    missingok
    notifempty
}
```

**newsyslog (macOS)**:

```text
/var/log/pulumicost/app.log  644  7  *  @T00  J
```

## Troubleshooting

### Logs not appearing in file

Check that the directory exists and is writable:

```bash
mkdir -p /var/log/pulumicost
chmod 755 /var/log/pulumicost
```

If the file cannot be written, you'll see a warning:

```text
Warning: Cannot write to /var/log/pulumicost/app.log, logging to stderr
```

### Finding your trace ID

Every command generates a trace ID for correlation:

```bash
# Inject your own trace ID
export PULUMICOST_TRACE_ID=my-trace-123
pulumicost cost projected --pulumi-json plan.json
```

Search logs by trace ID:

```bash
grep "my-trace-123" /var/log/pulumicost/app.log
```

## Complete Configuration Examples

### Development Setup

Verbose logging to console for debugging:

```yaml
# ~/.pulumicost/config.yaml
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
# ~/.pulumicost/config.yaml
logging:
  level: info
  format: json
  file: /var/log/pulumicost/pulumicost.log
  audit:
    enabled: true
    file: /var/log/pulumicost/audit.log
```

### CI/CD Pipeline Setup

JSON logs for log aggregation, with external trace ID:

```bash
export PULUMICOST_LOG_LEVEL=info
export PULUMICOST_LOG_FORMAT=json
export PULUMICOST_TRACE_ID="${CI_PIPELINE_ID}-${CI_JOB_ID}"

pulumicost cost projected --pulumi-json plan.json
```

### Security-Sensitive Environment

Audit logging with sensitive data redaction:

```yaml
# ~/.pulumicost/config.yaml
logging:
  level: warn  # Minimize log verbosity
  format: json
  file: /var/log/pulumicost/pulumicost.log
  audit:
    enabled: true
    file: /var/log/pulumicost/audit.log
# Note: API keys, passwords, and tokens are automatically redacted
```

## Quick Reference

| Setting      | Config File              | Environment Variable     | CLI Flag  |
| ------------ | ------------------------ | ------------------------ | --------- |
| Log Level    | `logging.level`          | `PULUMICOST_LOG_LEVEL`   | `--debug` |
| Log Format   | `logging.format`         | `PULUMICOST_LOG_FORMAT`  | -         |
| Log File     | `logging.file`           | -                        | -         |
| Trace ID     | -                        | `PULUMICOST_TRACE_ID`    | -         |
| Audit Enable | `logging.audit.enabled`  | -                        | -         |
| Audit File   | `logging.audit.file`     | -                        | -         |
