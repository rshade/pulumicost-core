# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## PluginHost Package Overview

The `internal/pluginhost` package implements the plugin communication layer for PulumiCost, providing a flexible architecture for launching and managing gRPC-based plugins through multiple transport mechanisms.

## Architecture

### Core Components

1. **Client** (`host.go`)
   - Wraps gRPC connection and proto API client
   - Manages plugin lifecycle with Close function
   - Automatically retrieves plugin name on initialization via `Name()` RPC call

2. **Launcher Interface**
   - Abstract interface for different plugin launching strategies
   - Returns: `*grpc.ClientConn`, cleanup function, and error
   - Two implementations: ProcessLauncher and StdioLauncher

3. **ProcessLauncher** (`process.go`)
   - Launches plugins as separate TCP server processes
   - Auto-allocates ports using ephemeral port binding
   - Passes port via `--port` flag and `PULUMICOST_PLUGIN_PORT` env var
   - Implements connection retry with exponential backoff

4. **StdioLauncher** (`stdio.go`)
   - Launches plugins using stdin/stdout communication
   - Creates local TCP proxy to bridge stdio to gRPC
   - Passes `--stdio` flag to plugin binary
   - Useful for debugging and constrained environments

### Plugin Lifecycle

```text
1. Launcher.Start() → Spawn process/setup communication
2. Connect to plugin via gRPC
3. Call plugin.Name() to verify connection
4. Return Client with API handle
5. Client.Close() → Cleanup connections and kill process
```

## Testing Commands

```bash
# Run all pluginhost tests
go test ./internal/pluginhost/...

# Run with race detection
go test -race ./internal/pluginhost/...

# Run integration tests (includes mock plugin creation)
go test -v ./internal/pluginhost/... -run TestIntegration

# Skip integration tests (short mode)
go test -short ./internal/pluginhost/...

# Test specific launcher
go test -v ./internal/pluginhost/... -run TestProcessLauncher
go test -v ./internal/pluginhost/... -run TestStdioLauncher
```

## Communication Patterns

### ProcessLauncher Flow

1. Allocate ephemeral port (bind to :0, get port, close)
2. Start plugin with `--port=XXXXX` argument
3. Set `PULUMICOST_PLUGIN_PORT` environment variable
4. Connect via gRPC to `127.0.0.1:port`
5. Retry connection with 100ms delays up to 10s timeout

### StdioLauncher Flow

1. Create stdin/stdout pipes to plugin process
2. Start plugin with `--stdio` argument
3. Create local TCP listener on ephemeral port
4. Spawn goroutine to proxy between TCP and stdio
5. Connect via gRPC to local proxy address

### Connection Management

- **Timeouts**: 10 seconds for both launchers
- **Retry Logic**: 100ms delay between connection attempts
- **State Checking**: Verify connectivity.Ready or connectivity.Idle states
- **Cleanup**: Always kill process and close connections on error or shutdown

## Critical Implementation Details

### Port Allocation Pattern

```go
// Allocate ephemeral port safely
lc := &net.ListenConfig{}
listener, err := lc.Listen(ctx, "tcp", "127.0.0.1:0")
port := listener.Addr().(*net.TCPAddr).Port
listener.Close() // Port remains reserved briefly
```

### Process Cleanup Pattern

```go
defer func() {
    conn.Close()      // Close gRPC connection first
    cmd.Process.Kill() // Then kill process
    cmd.Wait()        // Reap zombie process
}()
```

### Error Handling with Cleanup

```go
if err != nil {
    if closeErr := closeFn(); closeErr != nil {
        return fmt.Errorf("action: %w (close error: %w)", err, closeErr)
    }
    return fmt.Errorf("action: %w", err)
}
```

## Testing Patterns

### Mock Plugin Creation

- Tests create temporary executable scripts
- Unix: Shell scripts with proper shebang
- Windows: Batch files or PowerShell scripts
- Set executable permissions: `os.Chmod(path, 0755)`

### Integration Testing

- Use `testing.Short()` to skip integration tests
- Create failing mock plugins to test error paths
- Test launcher switching with same binary
- Verify no zombie processes after tests

### Concurrent Client Testing

- Test multiple clients with same launcher
- Verify port allocation doesn't conflict
- Ensure cleanup doesn't affect other clients

## Common Gotchas

1. **Port Race Condition**: Brief window between port allocation and plugin startup where port could be taken
2. **Process Zombies**: Always call `cmd.Wait()` after `Kill()` to reap zombie processes
3. **Connection State**: Check both `Ready` and `Idle` states as valid
4. **Stdout/Stderr Routing**: Plugin stderr goes to parent stderr for debugging
5. **Context Cancellation**: Properly handle context cancellation during connection attempts
6. **Windows Differences**: Different executable patterns and process management

## Plugin Protocol Requirements

Plugins must implement:

1. **TCP Mode**: Accept `--port=XXXXX` flag and listen on specified port
2. **Stdio Mode**: Accept `--stdio` flag and use stdin/stdout for gRPC
3. **Name RPC**: Respond to `Name()` call with plugin identifier
4. **CostSource Service**: Implement full proto.CostSourceService interface

## Environment Variables

- `PULUMICOST_PLUGIN_PORT`: Port number passed to plugin (ProcessLauncher)
- Plugin inherits full parent environment via `os.Environ()`

## Timeout Configuration

- **Default Timeout**: 10 seconds for both launchers
- **Connection Delay**: 100ms between retry attempts  
- **Connection Test Timeout**: 100ms for state change verification

These timeouts are currently hardcoded constants but could be made configurable if needed.

