# CLAUDE.md - internal/pluginhost

This file provides guidance to Claude Code (claude.ai/code) when working with the `internal/pluginhost` package.

## Package Overview

The `internal/pluginhost` package manages the lifecycle and communication with external plugin processes. It is responsible for launching plugins, establishing gRPC connections, and managing their termination.

**Key Responsibilities:**
- **Launch**: Spawning plugin processes (TCP or stdio).
- **Connect**: Establishing gRPC connections with retry logic.
- **Health**: Monitoring plugin health (via connection state).
- **Cleanup**: Ensuring plugins are terminated when no longer needed.

## Architecture

### Launchers

The package defines a `Launcher` interface with two implementations:

1.  **ProcessLauncher** (`process.go`): Launches plugins as independent processes listening on a TCP port.
    - Used for production plugins.
    - Supports concurrent plugin execution.
    - Handles port allocation and collision avoidance.
2.  **StdioLauncher** (`stdio.go`): Launches plugins communicating via stdin/stdout.
    - Used for testing or special environments.
    - Proxies stdio to a local TCP listener for gRPC compatibility.

### Client

The `Client` struct wraps the gRPC connection and the generated protobuf client. It provides a high-level API for interacting with plugins.

## Port Management (ProcessLauncher)

### Allocation Strategy

1.  **Allocate**: `allocatePortWithListener` opens a TCP listener on port 0 (OS-assigned random port).
2.  **Hold**: The listener is held open to reserve the port and prevent other processes from taking it.
3.  **Release**: Just before starting the plugin, `releasePortListener` closes the listener.
4.  **Bind**: The plugin process starts and binds to the now-available port.

### Race Condition Mitigation

There is a small window between releasing the port and the plugin binding to it. To mitigate this:
- **Retries**: `StartWithRetry` attempts to launch the plugin multiple times if binding fails.
- **Error Detection**: `isPortCollisionError` detects bind errors to trigger retries.

## Environment Variables

The `pluginhost` package manages environment variables passed to plugins:

-   `PULUMICOST_PLUGIN_PORT`: (Required) The port the plugin should listen on.
-   `PORT`: **REMOVED** (Deprecated). Core no longer sets this variable to avoid conflicts with cloud environment variables (e.g., Cloud Run).
-   `PULUMICOST_LOG_LEVEL`: Propagated from core configuration.
-   `PULUMICOST_TRACE_ID`: Propagated for distributed tracing.

**Note**: Plugins should prefer the `--port` flag over environment variables, but `PULUMICOST_PLUGIN_PORT` is provided for backward compatibility and ease of development.

## Testing

### Unit Tests

-   **Mock Commands**: Tests use shell scripts (`createMockServerCommand`) to simulate plugin behavior.
-   **Timeout Handling**: Tests verify that timeouts are respected during launch and connection.
-   **Race Conditions**: Specific tests target port allocation and concurrency.

### Integration Tests

-   **Real Binary**: Integration tests use a compiled `recorder` plugin or test binary.
-   **Full Lifecycle**: Tests verify the complete launch -> connect -> request -> close cycle.

## Critical Patterns

-   **Zombie Prevention**: Always call `cmd.Wait()` after `cmd.Process.Kill()`.
-   **Context Propagation**: Pass contexts through to all blocking operations (dialing, command execution).
-   **Error Wrapping**: Wrap errors with context (e.g., `fmt.Errorf("starting plugin: %w", err)`).
-   **Logging**: Use `logging.FromContext(ctx)` for structured logging with trace IDs.

## Recent Changes

-   **Remove PORT Env Var**: The `PORT` environment variable is no longer set by `ProcessLauncher`. Plugins must use `--port` flag or `PULUMICOST_PLUGIN_PORT` env var.
-   **Guidance Logging**: Added helpful log messages when plugins fail to bind, suggesting `--port` flag support.
-   **Debug Logging**: Added debug logs if `PORT` is detected in the user's environment (to indicate it's being ignored/shadowed).