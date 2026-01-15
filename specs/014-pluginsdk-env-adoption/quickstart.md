# Quickstart: Adopt pluginsdk/env.go

**Audience**: Plugin developers, core maintainers  
**Time**: 15 minutes  
**Prerequisites**: finfocus-spec v0.4.5+ with pluginsdk/env.go

## Overview

This guide shows how to migrate from hardcoded environment variable strings to the standardized pluginsdk/env.go constants and functions.

## Step 1: Update Dependencies

Ensure your go.mod includes the latest finfocus-spec:

```bash
go get github.com/rshade/finfocus-spec/sdk/go/pluginsdk@latest
```

## Step 2: Import pluginsdk

Add the import to your Go files:

```go
import (
    "github.com/rshade/finfocus-spec/sdk/go/pluginsdk"
)
```

## Step 3: Replace Hardcoded Strings (Core)

### Before

```go
cmd.Env = append(os.Environ(),
    fmt.Sprintf("PORT=%d", port),
    fmt.Sprintf("FINFOCUS_PLUGIN_PORT=%d", port),
)
```

### After

```go
cmd.Env = append(os.Environ(),
    fmt.Sprintf("%s=%d", pluginsdk.EnvPort, port),
    fmt.Sprintf("%s=%d", pluginsdk.EnvPortFallback, port),
)
```

## Step 4: Use Helper Functions (Plugins)

### Before

```go
func getPort() int {
    portStr := os.Getenv("PORT")
    if port, err := strconv.Atoi(portStr); err == nil {
        return port
    }
    return 0
}
```

### After

```go
func getPort() int {
    return pluginsdk.GetPort() // Handles FINFOCUS_PLUGIN_PORT first, then PORT
}
```

## Step 5: Update Tests

If your tests mock environment variables, update them to use constants:

### Before

```go
os.Setenv("PORT", "8080")
defer os.Unsetenv("PORT")
```

### After

```go
os.Setenv(pluginsdk.EnvPort, "8080")
defer os.Unsetenv(pluginsdk.EnvPort)
```

## Step 6: Verify Migration

Run tests to ensure everything works:

```bash
make test
make lint
```

## Common Patterns

### Setting Multiple Environment Variables

```go
env := []string{
    fmt.Sprintf("%s=%d", pluginsdk.EnvPort, port),
    fmt.Sprintf("%s=%d", pluginsdk.EnvPortFallback, port),
    fmt.Sprintf("%s=%s", pluginsdk.EnvLogLevel, "INFO"),
    fmt.Sprintf("%s=%s", pluginsdk.EnvLogFormat, "json"),
}
cmd.Env = append(os.Environ(), env...)
```

### Reading Configuration in Plugins

```go
func configurePlugin() {
    port := pluginsdk.GetPort()
    logLevel := pluginsdk.GetLogLevel()
    logFormat := pluginsdk.GetLogFormat()
    traceID := pluginsdk.GetTraceID()

    // Use values...
}
```

## Troubleshooting

### Plugin Won't Start

- Check that core is setting both `FINFOCUS_PLUGIN_PORT` and `PORT`
- Verify plugin is using `pluginsdk.GetPort()` not `os.Getenv("PORT")`

### Tests Failing

- Update test environment variable names to use constants
- Ensure test cleanup unsets the correct variable names

### Import Errors

- Verify finfocus-spec version is v0.4.5 or later
- Run `go mod tidy` to update dependencies

## Next Steps

- Update CLAUDE.md with new patterns
- Add integration tests for environment variable propagation
- Consider migrating logging and tracing variables in follow-up work
