---
title: Configuration Reference
description: Configuration options for FinFocus Core
layout: default
---

FinFocus is configured via a configuration file (default:
`~/.finfocus/config.yaml`) and environment variables.

## File Format

The configuration file is in YAML format.

```yaml
output:
  default_format: table # table, json, ndjson
  precision: 2

logging:
  level: info # debug, info, warn, error

plugins:
  dir: ~/.finfocus/plugins
```

## Sections

### Output

- `default_format`: The default output format for commands.
- `precision`: Number of decimal places for cost values.

### Logging

- `level`: The verbosity of logs.

### Plugins

- `dir`: The directory where plugins are installed.
