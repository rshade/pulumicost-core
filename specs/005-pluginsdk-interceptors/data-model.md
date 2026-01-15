# Data Model: Add UnaryInterceptors Support to ServeConfig

**Branch**: `005-pluginsdk-interceptors`
**Date**: 2025-11-28

## Entities

### ServeConfig (Extended)

Configuration structure for serving a FinFocus plugin via gRPC.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| Plugin | `Plugin` | Yes | The plugin implementation to serve |
| Port | `int` | No | Port number (0 = auto from PORT env or random) |
| Registry | `RegistryLookup` | No | Custom registry lookup (nil = default) |
| Logger | `*zerolog.Logger` | No | Custom logger (nil = default) |
| **UnaryInterceptors** | `[]grpc.UnaryServerInterceptor` | No | **NEW** Additional interceptors to chain |

**Zero-Value Behavior**:

- `UnaryInterceptors = nil`: Only built-in `TracingUnaryServerInterceptor()` runs
- `UnaryInterceptors = []`: Same as nil (empty slice)
- `UnaryInterceptors = [a, b]`: Chain is [Tracing, a, b]

### UnaryServerInterceptor (gRPC Type)

```go
type UnaryServerInterceptor func(
    ctx context.Context,
    req interface{},
    info *UnaryServerInfo,
    handler UnaryHandler,
) (resp interface{}, err error)
```

**Contract**:

- MUST call `handler(ctx, req)` to continue the chain
- MUST return handler's response or custom error
- MAY modify context before calling handler
- MAY inspect/modify request/response

## Relationships

```text
ServeConfig
    │
    ├── Plugin (required)
    │       └── Implements CostSourceService
    │
    ├── UnaryInterceptors[] (optional)
    │       └── Each implements grpc.UnaryServerInterceptor
    │
    └── Serve() creates:
            └── grpc.Server with ChainUnaryInterceptor(
                    TracingUnaryServerInterceptor(),  // always first
                    config.UnaryInterceptors...       // user interceptors
                )
```

## State Transitions

This feature has no state transitions. Configuration is immutable after
`Serve()` is called.

```text
[ServeConfig created] → [Serve() called] → [gRPC server running]
                                                    │
                                               (immutable)
```

## Validation Rules

| Rule | Enforced By | Error Handling |
|------|-------------|----------------|
| Plugin != nil | Existing code | Existing error |
| UnaryInterceptors elements != nil | gRPC library | Panic (standard gRPC behavior) |
| Port in valid range | Existing code | Existing error |

**Note**: We do NOT add validation for nil elements in UnaryInterceptors.
This matches standard gRPC behavior where passing nil interceptors to
`ChainUnaryInterceptor()` causes a panic. Plugin developers are expected
to know their interceptor values are non-nil.
