---
title: Environment Variables Reference
description: PulumiCost environment variable configuration options
layout: default
---

PulumiCost's behavior can be influenced by environment variables, which provide a
flexible way to override configuration settings. This is especially useful in CI/CD
environments or for sensitive information. Environment variables always take precedence
over settings in `~/.pulumicost/config.yaml`.

## Global Configuration Overrides

These variables affect the overall behavior of PulumiCost.

| Variable | Description | Values / Examples |
| --- | --- | --- |
| `PULUMICOST_LOG_LEVEL` | Overrides logging level | `trace`, `debug`, `info`, `warn`, `error` |
| `PULUMICOST_LOG_FORMAT` | Overrides log output format | `text`, `json` |
| `PULUMICOST_LOG_FILE` | Absolute path for log output | `/var/log/pulumicost.log` |
| `PULUMICOST_TRACE_ID` | External trace ID for distributed tracing | Any string |
| `PULUMICOST_OUTPUT_FORMAT` | Default output format for commands | `table`, `json`, `ndjson` |
| `PULUMICOST_OUTPUT_PRECISION` | Decimal places for cost values | `0` to `10` |
| `PULUMICOST_CONFIG_STRICT` | Enables strict config parsing | `true`, `1` |

## Plugin-Specific Configuration

Environment variables can also be used to configure individual plugins. This is
particularly useful for injecting sensitive credentials (e.g., API keys, secrets)
without hardcoding them in the `config.yaml` file.

The naming convention for plugin environment variables is
`PULUMICOST_PLUGIN_<PLUGIN_NAME>_<KEY_NAME>`.

- `<PLUGIN_NAME>`: The name of the plugin (e.g., `AWS`, `KUBECOST`, `VANTAGE`).
- `<KEY_NAME>`: The configuration key for that plugin (e.g., `ACCESS_KEY_ID`).

**Examples:**

```bash
# Configure AWS Plugin credentials
export PULUMICOST_PLUGIN_AWS_ACCESS_KEY_ID="AKIA..."
export PULUMICOST_PLUGIN_AWS_SECRET_ACCESS_KEY="your-secret-key"
export PULUMICOST_PLUGIN_AWS_REGION="us-east-1"

# Configure Kubecost Plugin endpoint
export PULUMICOST_PLUGIN_KUBECOST_ENDPOINT="http://kubecost.monitoring:9090"

# Configure Vantage Plugin API Key
export PULUMICOST_PLUGIN_VANTAGE_API_KEY="vantage-api-key-123"
```

Refer to the specific plugin's documentation for a list of supported `KEY_NAME`s.
