# Research: `Supports` Handler Implementation Pattern

## 1. Unknowns Investigated

The primary unknown was the exact location and implementation pattern for the gRPC server and its handlers within the `pluginsdk`. The initial assumption of `pkg/pluginsdk/server.go` was incorrect.

## 2. Findings

### Source Code Analysis

- **File**: `pkg/pluginsdk/sdk.go`
- **Dependencies**: `go.mod` confirms the project uses `github.com/rshade/finfocus-spec/sdk/go` for its protobuf definitions (`pbc`).
- **Key Function**: `func Serve(ctx context.Context, config ServeConfig) error` is the entry point for starting a plugin server.
- **Implementation Pattern**:
    1.  The `Serve` function instantiates a standard `grpc.NewServer()`.
    2.  It creates a local `*Server` struct which wraps the actual plugin logic (via a `Plugin` interface).
    3.  The local `*Server` is registered with the gRPC server via `pbc.RegisterCostSourceServiceServer(grpcServer, server)`.
    4.  The local `*Server` struct has methods for each RPC call (e.g., `GetProjectedCost`), which simply delegate the call to the wrapped `plugin` instance.
    5.  The `Plugin` interface, also defined in this file, does **not** currently include a `Supports` method.

## 3. Decisions

- **Implementation File**: All changes will be made in `pkg/pluginsdk/sdk.go` and its corresponding test file.
- **Implementation Strategy**:
    1.  A new `SupportsProvider` interface will be defined in `pkg/pluginsdk/sdk.go`.
    2.  A `Supports` method will be added to the `*Server` struct.
    3.  This method will use a type assertion to check if the contained `plugin` implements `SupportsProvider`.
    4.  If the type assertion succeeds, the call will be delegated to the plugin's `Supports` method.
    5.  If the type assertion fails, a default response (`supported: false`) will be returned, as required by the specification.

This approach aligns perfectly with the existing architecture, ensuring consistency and adhering to the optional, capability-based design.
