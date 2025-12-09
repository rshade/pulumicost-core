# Data Model: Documentation Information Architecture

**Feature**: `011-sync-docs-codebase`

## Information Architecture

The documentation is structured to separate concerns by audience and depth.

### 1. Analyzer Architecture (`docs/architecture/analyzer.md`)

| Section | Content Description | Source |
|---------|---------------------|--------|
| **Overview** | High-level explanation of "Zero-Click" cost estimation via `pulumi preview` | `internal/cli/analyzer_serve.go` |
| **Protocol** | Details of the Pulumi Analyzer gRPC protocol (handshake, resource analysis) | `internal/analyzer/` |
| **Configuration** | How to configure `Pulumi.yaml` to use the analyzer | `internal/config/config.go` |
| **Diagnostics** | Explanation of how costs are reported (ADVISORY diagnostics) | `internal/analyzer/` |

### 2. Configuration Reference (`docs/reference/config-reference.md`)

| Key | Type | Description | Default | Source |
|-----|------|-------------|---------|--------|
| `output.default_format` | string | `table`, `json`, `ndjson` | `table` | `internal/config/config.go` |
| `output.precision` | int | Decimal places for currency | `2` | `internal/config/config.go` |
| `logging.level` | string | `trace`, `debug`, `info`, `warn`, `error` | `info` | `internal/config/config.go` |
| `logging.format` | string | `console`, `json` | `console` | `internal/config/config.go` |
| `logging.file` | string | Path to log file | `~/.pulumicost/logs/...` | `internal/config/config.go` |
| `logging.outputs` | list | List of output sinks | `[{type: console}]` | `internal/config/config.go` |
| `analyzer.timeout` | object | Timeout settings | `{total: 60s}` | `internal/config/config.go` |
| `plugins` | map | Plugin-specific config | `{}` | `internal/config/config.go` |

### 3. Error Codes (`docs/reference/error-codes.md`)

| Code Constant | Message Pattern | Component | Solution |
|---------------|-----------------|-----------|----------|
| `ErrNoCostData` | "no cost data available" | Engine | Check plugin installation/coverage |
| `ErrMixedCurrencies` | "mixed currencies..." | Engine | Filter resources or exchange rates |
| `ErrInvalidGroupBy` | "invalid groupBy..." | Engine | Use `daily` or `monthly` |
| `ErrEmptyResults` | "empty results..." | Engine | Check date range or filters |
| `ErrInvalidDateRange` | "invalid date range..." | Engine | `to` must be after `from` |
| `ErrConfigCorrupted` | "configuration file..." | Config | Fix YAML syntax or delete file |

### 4. Environment Variables (`docs/reference/environment-variables.md`)

| Variable | Scope | Description | Source |
|----------|-------|-------------|--------|
| `PULUMICOST_LOG_LEVEL` | Global | Override log level | `internal/cli/root.go` |
| `PULUMICOST_LOG_FORMAT` | Global | Override log format | `internal/cli/root.go` |
| `PULUMICOST_LOG_FILE` | Global | Override log file | `internal/config/config.go` |
| `PULUMICOST_TRACE_ID` | Global | Inject trace ID | `internal/logging/zerolog.go` |
| `PULUMICOST_OUTPUT_FORMAT` | Global | Override output format | `internal/config/config.go` |
| `PULUMICOST_OUTPUT_PRECISION` | Global | Override precision | `internal/config/config.go` |
| `PULUMICOST_CONFIG_STRICT` | Global | Panic on config error | `internal/config/config.go` |
| `PULUMICOST_PLUGIN_*` | Plugin | Passed to plugins | `internal/pluginhost/` |

## Navigation Structure

- **Getting Started**
  - Installation
  - Quickstart
  - *Analyzer Setup (New)*
- **Guides**
  - User Guide (*Updated*)
  - Developer Guide (*Updated*)
- **Reference**
  - CLI Commands (*Updated*)
  - *Configuration (New)*
  - *Environment Variables (New)*
  - *Error Codes (New)*
- **Architecture**
  - System Overview
  - Plugin Protocol
  - *Analyzer Integration (New)*
