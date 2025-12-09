---
title: Configuration Reference
description: PulumiCost YAML configuration file schema and options
layout: default
---

PulumiCost uses a YAML-based configuration file located at `~/.pulumicost/config.yaml`.
This file allows you to customize various aspects of the tool's behavior, including
output formatting, logging, and plugin-specific settings.

## Configuration File Location

The default location for the configuration file is `~/.pulumicost/config.yaml`.

## Schema

### `output`

Defines preferences for how PulumiCost formats its output.

| Key | Type | Default | Description |
| --- | --- | --- | --- |
| `default_format` | string | `"table"` | Default output format: `"table"`, `"json"`, or `"ndjson"`. |
| `precision` | int | `2` | Decimal places for monetary values. |

### `logging`

Configures the logging behavior of PulumiCost.

| Key | Type | Default | Description |
| --- | --- | --- | --- |
| `level` | string | `"info"` | Minimum logging level: `"trace"`, `"debug"`, `"info"`, `"warn"`, `"error"`. |
| `format` | string | `"text"` | Log format: `"text"` (console) or `"json"`. |
| `file` | string | (see below) | Path to write logs. Default: `~/.pulumicost/logs/pulumicost.log` |
| `outputs` | array | (see below) | Structured logging output destinations. |
| `audit` | object | (see below) | Configuration for audit logging. |

#### `logging.outputs` Object Schema

| Key | Type | Default | Description |
| --- | --- | --- | --- |
| `type` | string | (required) | Output type: `"console"`, `"file"`, `"syslog"`. |
| `level` | string | (global) | Overrides global logging level for this output. |
| `path` | string | (required for file) | File path for `type: "file"`. Must be absolute. |
| `format` | string | (global) | Overrides global logging format for this output. |
| `max_size_mb` | int | `0` | Max size in MB before rotation (0 = unlimited). |
| `max_files` | int | `0` | Max old log files to retain (0 = unlimited). |

#### `logging.audit`

Specific configuration for audit logging of cost query operations.

| Key | Type | Default | Description |
| --- | --- | --- | --- |
| `enabled` | bool | `false` | Set to `true` to enable audit logging. |
| `file` | string | (empty) | Separate file path for audit logs. |

### `plugins`

Contains a map of plugin-specific configurations, where each key is the plugin name.

```yaml
plugins:
  aws:
    region: "us-east-1"
    access_key_id: "AKIA..."
  kubecost:
    endpoint: "http://localhost:9090"
```

These values can be overridden by environment variables following the pattern
`PULUMICOST_PLUGIN_<PLUGIN_NAME>_<KEY_NAME>`.

### `analyzer`

Configuration specific to the Pulumi Analyzer integration.

| Key | Type | Default | Description |
| --- | --- | --- | --- |
| `timeout` | object | (see below) | Timeout settings for cost analysis operations. |
| `plugins` | map | `{}` | Plugin configurations for the analyzer. |

#### `analyzer.timeout` Object Schema

| Key | Type | Default | Description |
| --- | --- | --- | --- |
| `per_resource` | duration | `"5s"` | Max time per resource calculation. |
| `total` | duration | `"60s"` | Overall max time for analysis. |
| `warn_threshold` | duration | `"30s"` | Duration after which a warning is emitted. |

#### `analyzer.plugins` Object Schema

This section specifies which plugins the Analyzer should use and any specific
configurations for them.

| Key | Type | Default | Description |
| --- | --- | --- | --- |
| `path` | string | (empty) | Absolute path to plugin binary. |
| `enabled` | bool | `true` | Whether plugin is enabled for analyzer. |
| `env` | map | `{}` | Environment variables to pass to plugin. |
