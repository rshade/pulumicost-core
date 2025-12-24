# Feature Specification: Add UnaryInterceptors Support to ServeConfig

**Feature Branch**: `005-pluginsdk-interceptors`
**Created**: 2025-11-28
**Status**: Draft
**Input**: User description: "feat(pluginsdk): Add UnaryInterceptors support to ServeConfig"
**GitHub Issue**: [#188](https://github.com/rshade/pulumicost-core/issues/188)
**Target Repository**: pulumicost-spec (sdk/go/pluginsdk)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Plugin Developer Registers Tracing Interceptor (Priority: P1)

As a plugin developer, I want to register the `TracingUnaryServerInterceptor()` provided by the pluginsdk so that my plugin automatically propagates trace IDs from incoming requests without manually extracting them in each handler.

**Why this priority**: This is the primary use case driving this feature. Trace propagation is essential for distributed debugging across the pulumicost-core host and plugins. Without this, plugin developers must duplicate tracing logic in every gRPC handler.

**Independent Test**: Can be fully tested by creating a plugin that registers the tracing interceptor and verifying that trace IDs from incoming gRPC metadata are automatically injected into the context.

**Acceptance Scenarios**:

1. **Given** a plugin developer creating a new plugin, **When** they configure `ServeConfig` with `TracingUnaryServerInterceptor()` in the `UnaryInterceptors` field, **Then** the interceptor is applied to all incoming gRPC requests.

2. **Given** a request from pulumicost-core with a `trace_id` in gRPC metadata, **When** the plugin receives the request with the tracing interceptor registered, **Then** the trace ID is available in the request context without manual extraction.

3. **Given** a plugin with the tracing interceptor registered, **When** the plugin logs messages using the context, **Then** all log entries include the propagated trace ID.

---

### User Story 2 - Plugin Developer Chains Multiple Interceptors (Priority: P2)

As a plugin developer, I want to register multiple interceptors (e.g., tracing, logging, metrics) so that I can compose cross-cutting concerns without modifying handler logic.

**Why this priority**: Composability is a common requirement for production plugins. While tracing alone (P1) provides immediate value, the ability to chain interceptors enables authentication, rate limiting, and other middleware patterns.

**Independent Test**: Can be tested by registering two or more interceptors and verifying they execute in order for each request.

**Acceptance Scenarios**:

1. **Given** a plugin with multiple interceptors configured, **When** a gRPC request is received, **Then** all interceptors execute in the order they were registered.

2. **Given** interceptors that modify the context, **When** the handler receives the request, **Then** all context modifications from prior interceptors are visible.

---

### User Story 3 - Plugin Operates Without Interceptors (Priority: P3)

As a plugin developer, I want my existing plugins to continue working without changes so that I can adopt interceptors incrementally.

**Why this priority**: Backward compatibility ensures existing plugins continue to function. This is lower priority because it's a non-functional constraint rather than new capability.

**Independent Test**: Can be tested by running an existing plugin without any interceptors configured and verifying it operates identically to before.

**Acceptance Scenarios**:

1. **Given** a plugin with no interceptors configured, **When** the plugin starts, **Then** it operates exactly as before this change.

2. **Given** a plugin with an empty `UnaryInterceptors` slice, **When** the plugin starts, **Then** no interceptors are applied and no errors occur.

---

### Edge Cases

- What happens when an interceptor panics? (Assumed: standard gRPC panic recovery behavior applies)
- What happens when an interceptor returns an error? (Assumed: standard gRPC error propagation terminates the call)
- What happens when `nil` is passed as the slice? (Handled: nil slice treated as empty, no additional interceptors)
- What happens when a nil element exists in the slice? (Standard gRPC behavior: panic during server creation - not our responsibility to handle)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: ServeConfig MUST accept an optional list of interceptors that can be applied to incoming requests.
- **FR-002**: The plugin server MUST apply configured interceptors to all incoming unary RPC calls.
- **FR-003**: Interceptors MUST execute in the order they are registered.
- **FR-004**: The system MUST function normally when no interceptors are configured (backward compatibility).
- **FR-005**: The system MUST handle an empty interceptor list without errors.
- **FR-006**: The system MUST support the existing `TracingUnaryServerInterceptor()` from the pluginsdk without modifications.

### Key Entities

- **ServeConfig**: Configuration structure for plugin servers. Extended with interceptor registration capability.
- **UnaryInterceptor**: A middleware function that wraps incoming unary RPC calls. Provided by plugin developers.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Plugin developers can enable automatic trace propagation by adding one configuration field, eliminating per-handler trace extraction code.
- **SC-002**: Existing plugins without interceptor configuration continue to work without code changes.
- **SC-003**: 100% of interceptors registered in the configuration are invoked for each incoming request.
- **SC-004**: Plugin startup time is not measurably impacted by interceptor configuration (within 5% of baseline).

## Assumptions

- The `TracingUnaryServerInterceptor()` is already implemented and tested in `pulumicost-spec/sdk/go/pluginsdk/logging.go`.
- Standard gRPC `ChainUnaryInterceptor` semantics apply (order-preserved, left-to-right execution).
- No streaming interceptors are needed at this time (only unary RPC calls are used in the plugin protocol).
- Panic recovery and error handling follow standard gRPC interceptor conventions.

## Out of Scope

- Streaming interceptors (can be added in a future iteration if needed).
- Built-in interceptors beyond what the pluginsdk already provides.
- Dynamic interceptor registration after server startup.
- Interceptor metrics or observability (handled by individual interceptor implementations).
