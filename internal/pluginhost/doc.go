// Package pluginhost manages plugin lifecycle and gRPC communication.
//
// Plugins are external processes that implement the FinFocus plugin protocol.
// This package handles launching, connecting to, and communicating with plugins.
//
// # Plugin Launchers
//
// Two launcher types are available:
//   - ProcessLauncher: Launches plugins as TCP processes
//   - StdioLauncher: Uses stdin/stdout for plugin communication
//
// # Connection Management
//
// The Client type wraps plugin gRPC connections with:
//   - 10-second connection timeout
//   - 100ms retry delays
//   - Automatic cleanup on failure
//
// # Platform Support
//
// Binary detection is platform-aware:
//   - Unix: Checks executable permissions
//   - Windows: Looks for .exe extension
//
// # Cleanup
//
// Always call cmd.Wait() after Kill() to prevent zombie processes.
package pluginhost
