# Data Model: Adopt pluginsdk/env.go for Environment Variable Handling

**Date**: 2025-12-10  
**Status**: Complete  
**Related**: [research.md](research.md), [plan.md](plan.md)

## Overview

This refactor standardizes environment variable handling for plugin communication. The data model focuses on the environment variables and plugin process configuration that need to be managed consistently.

## Key Entities

### Environment Variables

Standardized environment variables for plugin communication and configuration.

| Field                  | Type         | Description                                | Validation                          | Example        |
| ---------------------- | ------------ | ------------------------------------------ | ----------------------------------- | -------------- |
| PULUMICOST_PLUGIN_PORT | string (int) | Primary port for plugin gRPC communication | Must be valid port number (1-65535) | "8080"         |
| PORT                   | string (int) | Fallback port variable for compatibility   | Must be valid port number (1-65535) | "8080"         |
| PULUMICOST_LOG_LEVEL   | string       | Logging verbosity level                    | One of: DEBUG, INFO, WARN, ERROR    | "INFO"         |
| PULUMICOST_LOG_FORMAT  | string       | Log output format                          | One of: json, console               | "json"         |
| PULUMICOST_TRACE_ID    | string       | External trace ID for distributed tracing  | Valid trace ID format               | "abc123def456" |

### Plugin Process

Child process launched by core with standardized environment configuration.

| Field            | Type              | Description                | Validation                  |
| ---------------- | ----------------- | -------------------------- | --------------------------- |
| Command          | string            | Executable path for plugin | Must be valid file path     |
| Arguments        | []string          | Command line arguments     | Optional                    |
| Environment      | map[string]string | Environment variables      | Must include port variables |
| WorkingDirectory | string            | Process working directory  | Must be valid directory     |

## Relationships

```
Core Application
    ├── Launches Plugin Process
    │       ├── Sets Environment Variables (from pluginsdk constants)
    │       └── Plugin reads via pluginsdk.GetPort()
    └── Plugin Process
        ├── Binds to port from PULUMICOST_PLUGIN_PORT or PORT
        ├── Configures logging from PULUMICOST_LOG_*
        └── Injects trace ID from PULUMICOST_TRACE_ID
```

## State Transitions

### Plugin Launch Sequence

1. **Pre-launch**: Core validates port availability
2. **Environment Setup**: Core sets PULUMICOST_PLUGIN_PORT and PORT
3. **Process Start**: Plugin process begins execution
4. **Port Binding**: Plugin reads port via pluginsdk.GetPort() and binds
5. **Ready**: Plugin signals readiness via gRPC health check

### Error States

- **Invalid Port**: pluginsdk.GetPort() returns 0, plugin fails to start
- **Port Conflict**: Port already in use, plugin bind fails
- **Missing Environment**: No port variables set, plugin cannot determine bind address

## Validation Rules

### Environment Variable Validation

- Port values must be integers between 1-65535
- Log level must be one of: DEBUG, INFO, WARN, ERROR
- Log format must be one of: json, console
- Trace ID should be non-empty when provided

### Process Validation

- Plugin executable must exist and be executable
- Working directory must exist
- At least one port variable must be set
- Environment variables must not contain sensitive data in logs

## Migration Path

### Before (Current)

```go
cmd.Env = append(os.Environ(),
    fmt.Sprintf("PORT=%d", port),
    fmt.Sprintf("PULUMICOST_PLUGIN_PORT=%d", port),
)
```

### After (Target)

```go
cmd.Env = append(os.Environ(),
    fmt.Sprintf("%s=%d", pluginsdk.EnvPort, port),
    fmt.Sprintf("%s=%d", pluginsdk.EnvPortFallback, port),
)
```

## Implementation Notes

- All environment variable access in plugins should use pluginsdk functions
- Core sets both port variables for backward compatibility
- Future cleanup can remove PORT once all plugins migrate to pluginsdk.GetPort()
- Constants ensure consistency across core and all plugins
