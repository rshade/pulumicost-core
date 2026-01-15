# Environment Variables Contract

**Version**: 1.0.0  
**Date**: 2025-12-10  
**Purpose**: Defines the standardized environment variables for FinFocus plugin communication

## Overview

This contract specifies the environment variables that must be set by the FinFocus core when launching plugins, and how plugins should read these variables.

## Required Environment Variables

### Plugin Communication

| Variable                 | Type    | Required                | Description                                     |
| ------------------------ | ------- | ----------------------- | ----------------------------------------------- |
| `FINFOCUS_PLUGIN_PORT` | integer | Yes                     | Primary port for plugin gRPC server binding     |
| `PORT`                   | integer | Yes (for compatibility) | Fallback port variable for legacy compatibility |

### Plugin Configuration

| Variable                | Type   | Required | Description                                 |
| ----------------------- | ------ | -------- | ------------------------------------------- |
| `FINFOCUS_LOG_LEVEL`  | string | No       | Logging verbosity: DEBUG, INFO, WARN, ERROR |
| `FINFOCUS_LOG_FORMAT` | string | No       | Log output format: json, console            |
| `FINFOCUS_TRACE_ID`   | string | No       | External trace ID for distributed tracing   |

## Contract Rules

### Setting Variables (Core Responsibility)

1. **Core MUST set both port variables** for backward compatibility
2. **Core MUST use pluginsdk constants** instead of hardcoded strings
3. **Port values MUST be valid** (1-65535)
4. **Core MUST NOT set sensitive data** in environment variables

### Reading Variables (Plugin Responsibility)

1. **Plugins MUST use pluginsdk functions** for environment variable access
2. **Plugins MUST call GetPort()** for port determination
3. **GetPort() prioritizes FINFOCUS_PLUGIN_PORT** over PORT
4. **Plugins MUST handle missing variables gracefully**

## Implementation Contract

### Core Implementation

```go
import "github.com/rshade/finfocus-spec/sdk/go/pluginsdk"

// Correct implementation
cmd.Env = append(os.Environ(),
    fmt.Sprintf("%s=%d", pluginsdk.EnvPort, port),
    fmt.Sprintf("%s=%d", pluginsdk.EnvPortFallback, port),
)

// INCORRECT - hardcoded strings not allowed
cmd.Env = append(os.Environ(),
    fmt.Sprintf("FINFOCUS_PLUGIN_PORT=%d", port),
    fmt.Sprintf("PORT=%d", port),
)
```

### Plugin Implementation

```go
import "github.com/rshade/finfocus-spec/sdk/go/pluginsdk"

// Correct implementation
func startServer() error {
    port := pluginsdk.GetPort()
    if port == 0 {
        return errors.New("no port specified")
    }
    // bind to port...
}

// INCORRECT - direct os.Getenv not allowed
func startServer() error {
    portStr := os.Getenv("PORT")
    // parse and use...
}
```

## Validation

### Automated Tests

- **Contract Test**: Verify core sets required environment variables
- **Integration Test**: Verify plugins can read variables correctly
- **Migration Test**: Verify both old and new variable names work

### Manual Verification

- Check that no hardcoded environment variable strings exist in core
- Verify plugins use pluginsdk.GetPort() instead of os.Getenv()
- Confirm backward compatibility during transition period

## Breaking Changes

None. This contract maintains backward compatibility by setting both PORT and FINFOCUS_PLUGIN_PORT.

## Future Evolution

- After all plugins migrate, PORT can be deprecated
- Additional environment variables can be added following this pattern
- pluginsdk/env.go will be the single source of truth for variable names
