# Research: Comprehensive Documentation Sync

**Feature Branch**: `011-sync-docs-codebase`
**Date**: 2025-12-08

## Technical Findings

### 1. Analyzer Architecture
- **Source**: `internal/cli/analyzer_serve.go`, `internal/analyzer/`
- **Command**: `pulumicost analyzer serve`
- **Protocol**: gRPC (Pulumi Analyzer Protocol)
- **Configuration**: `analyzer` section in `config.yaml`
  - `timeout`: `per_resource`, `total`, `warn_threshold`
  - `plugins`: map of plugin-specific settings

### 2. CLI Commands
- **Plugin Management**: `internal/cli/plugin_*.go`
  - `init`: Initialize new plugin project
  - `install`: Install plugin from registry/url
  - `list`: List installed plugins
  - `remove`: Remove installed plugin
  - `update`: Update plugin
  - `validate`: Validate plugin binary
- **Analyzer**: `internal/cli/analyzer_serve.go`
  - `serve`: Start analyzer server

### 3. Configuration Schema
- **Source**: `internal/config/config.go`
- **Location**: `~/.pulumicost/config.yaml`
- **Environment Variable Overrides**:
  - `PULUMICOST_OUTPUT_FORMAT` -> `output.default_format`
  - `PULUMICOST_OUTPUT_PRECISION` -> `output.precision`
  - `PULUMICOST_LOG_LEVEL` -> `logging.level`
  - `PULUMICOST_LOG_FORMAT` -> `logging.format`
  - `PULUMICOST_LOG_FILE` -> `logging.file`
  - `PULUMICOST_PLUGIN_<NAME>_<KEY>` -> `plugins.<name>.<key>`
  - `PULUMICOST_CONFIG_STRICT`: Enforce strict config parsing

### 4. Error Codes
- **Engine Errors** (`internal/engine/types.go`, `internal/engine/engine.go`):
  - `ErrNoCostData`: "no cost data available"
  - `ErrMixedCurrencies`: "mixed currencies not supported in cross-provider aggregation"
  - `ErrInvalidGroupBy`: "invalid groupBy type for cross-provider aggregation"
  - `ErrEmptyResults`: "empty results provided for aggregation"
  - `ErrInvalidDateRange`: "invalid date range: end date must be after start date"
  - `ErrResourceValidation`: "resource validation failed"
- **Config Errors** (`internal/config/config.go`):
  - `ErrConfigCorrupted`: "configuration file appears corrupted"

### 5. Environment Variables
- `PULUMICOST_LOG_LEVEL`
- `PULUMICOST_LOG_FORMAT`
- `PULUMICOST_LOG_FILE`
- `PULUMICOST_TRACE_ID`
- `PULUMICOST_OUTPUT_FORMAT`
- `PULUMICOST_OUTPUT_PRECISION`
- `PULUMICOST_CONFIG_STRICT`
- `PULUMICOST_PLUGIN_*`

## Decisions

- **Documentation Source of Truth**: The `internal/` code is the definitive source.
- **Error Code Grouping**: Group errors by component (Engine, Config, CLI) in `error-codes.md`.
- **Config Reference Structure**: Use nested tables for `logging.outputs` and `analyzer` settings to handle complexity.

## Rationale

- **Why exhaustive research?**: The documentation was flagged as significantly out of sync. verifying the exact exported symbols and config keys ensures the new docs are accurate 1:1 with the code.
- **Why strict mode?**: Discovered `PULUMICOST_CONFIG_STRICT` during research, which is a critical feature for CI/CD pipelines and must be documented.
