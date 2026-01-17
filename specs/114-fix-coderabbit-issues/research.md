# Phase 0: Research & Validation

**Feature**: CodeRabbit Issue Resolution
**Date**: 2026-01-16

## Overview

This document confirms the codebase state for the reported issues and defines specific remediation strategies.

## Findings

### 1. Duplicate Constants (`EnvAnalyzerMode`)
- **Location 1**: `internal/cli/analyzer_serve.go`
- **Location 2**: `internal/pluginhost/process.go`
- **Current Value**: `"FINFOCUS_ANALYZER_MODE"`
- **Decision**: Create new package `github.com/rshade/finfocus/internal/constants` and move the definition there. This avoids circular imports between `cli` and `pluginhost`.

### 2. Error Swallowing in `plugin_inspect.go`
- **Issue**: `defer func() { _ = client.Close() }()` ignores errors.
- **Decision**: Replace with debug logging:
  ```go
  defer func() {
      if err := client.Close(); err != nil {
          logging.FromContext(ctx).Debugf("Failed to close client: %v", err)
      }
  }()
  ```
- **Issue**: `renderTable` returns no error.
- **Decision**: Change signature to `func renderTable(...) error` and propagate all write errors.

### 3. Hardcoded Paths
- **Issue**: `findPluginPath` uses `os.UserHomeDir`.
- **Decision**: Use `config.New().PluginDir` which already handles the `.finfocus/plugins` logic and platform differences.

### 4. Struct Tags
- **Issue**: `FieldMapping` emits empty fields.
- **Decision**: Add `omitempty` to `json` and `yaml` tags for `Condition` and `ExpectedType`.

### 5. Logging Standards
- **Issue**: Log level for launch failures not specified.
- **Decision**: Use **Debug Level** for plugin launch failures in `plugin_list` to avoid spamming users during normal operation.

## Alternatives Considered

- **Existing Config Package**: We could put `EnvAnalyzerMode` in `internal/config`, but `config` often imports other packages. A dedicated `internal/constants` package is safer for avoiding circular dependencies.
- **Ignoring Render Errors**: We considered keeping `renderTable` as-is, but this violates the "Reliable Error Handling" requirement (User Story 1).

## Conclusion

The reported issues are valid and verified. The plan to refactor into a `constants` package and enforce error handling is sound.