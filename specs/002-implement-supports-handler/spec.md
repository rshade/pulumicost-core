# Feature Specification: Implement Supports() gRPC Handler

**Feature Branch**: `002-implement-supports-handler`  
**Created**: 2025-11-22  
**Status**: Draft  
**Input**: User description: "Implement Supports() gRPC handler in pluginsdk..."

## Clarifications

### Session 2025-11-22
- Q: How should the gRPC handler report an internal plugin error to the client? → A: Propagate as a gRPC Internal status code, with a generic message like "plugin failed to execute."
- Q: How should the system report non-support for a valid request (e.g., wrong region)? → A: The system should return a 'supported: false' response with a clear, descriptive 'Reason' field explaining why the resource is not supported.
- Q: What is the maximum acceptable latency for a single Supports gRPC query to complete? → A: 50ms
- Q: What validation should be performed on the Resource entity fields? → A: Strict validation - validate against a known, pre-defined list of providers, resource types, and regions.
- Q: How should the system determine what providers, resource types, and regions are valid for strict validation? → A: Query registered plugins at runtime. A registry.json will contain plugin name, regions, and providers. Resource types remain queryable from plugins to allow independent plugin releases without core PRs.
- Q: What should happen when no plugins are registered for validation? → A: Two-step validation: first find plugin by provider/region from registry, then query that plugin's Supports method for resource_type validation.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Query Plugin Capabilities (Priority: P1)

As a client application developer, I want to query a plugin to determine if it supports a specific cloud resource, so that my application can dynamically handle different plugin capabilities.

**Why this priority**: This is the core functionality required to allow clients to introspect plugin capabilities, preventing errors and enabling more intelligent behavior.

**Independent Test**: Can be tested by sending a gRPC request to the `Supports` endpoint of a plugin that implements the capability. The test passes if a valid, accurate response is returned.

**Acceptance Scenarios**:

1. **Given** a plugin that implements the `Supports` capability for "aws_s3_bucket", **When** the client queries the `Supports` endpoint for "aws_s3_bucket", **Then** the system returns a response indicating that the resource is supported.
2. **Given** a plugin that implements the `Supports` capability and does *not* support "aws_gamelift_fleet", **When** the client queries the `Supports` endpoint for "aws_gamelift_fleet", **Then** the system returns a response indicating that the resource is not supported.

---

### User Story 2 - Handle Legacy Plugins Gracefully (Priority: P1)

As a client application developer, when I query a plugin that does *not* implement the `Supports` capability, I want to receive a clear, default response indicating that the capability is not supported, so that my application does not crash or hang.

**Why this priority**: Ensures backward compatibility and prevents the system from failing when interacting with older plugins that lack this new capability.

**Independent Test**: Can be tested by sending a gRPC request to the `Supports` endpoint of a plugin that does *not* implement the capability. The test passes if a default "not supported" response is returned.

**Acceptance Scenarios**:

1. **Given** a plugin that does not implement the `Supports` capability, **When** the client queries the `Supports` endpoint for any resource, **Then** the system returns a response indicating the resource is not supported and provides a reason that the capability is unimplemented.

### Edge Cases

- If a `Supports` request is received with fields that are empty, malformed, or do not match registry values (provider/region from registry.json), the system MUST reject it with a gRPC `InvalidArgument` status code.
- If no plugin is registered for the requested provider/region combination, the system MUST reject with gRPC `InvalidArgument` status code indicating no plugin available.
- Resource type validation is delegated to the matched plugin via its Supports method, allowing plugins to expand supported types independently.
- If a plugin's internal `Supports` method encounters an unexpected error (e.g., database connection fails), the gRPC handler MUST return a gRPC `Internal` status code. The detailed error MUST be logged on the server-side for debugging but MUST NOT be sent to the client.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST expose a `Supports` capability via its gRPC interface that is accessible to clients.
- **FR-002**: The `Supports` capability MUST accept a request containing resource details (such as provider and type).
- **FR-003**: The `Supports` capability MUST return a response indicating whether the specified resource is supported.
- **FR-004**: The response MUST include a descriptive reason if the resource is not supported. This applies whether non-support is determined by the core engine (e.g., no plugin available) or by the plugin itself (e.g., wrong region, unsupported resource type).
- **FR-005**: The system MUST delegate the `Supports` check to the underlying plugin if the plugin has implemented this specific capability.
- **FR-006**: The system MUST provide a default response indicating "not supported" if the underlying plugin has *not* implemented the `Supports` capability.
- **FR-007**: The system MUST validate incoming `Supports` requests. A request is considered invalid and MUST be rejected if the `provider` or `region` fields do not match values from the plugin registry (registry.json). Resource types are validated by querying the plugin at runtime, allowing plugins to expand support without core updates.

### Key Entities

- **ResourceDescriptor**: Represents a cloud resource that can be queried for support. It has the following key attributes:
    - `provider`: The cloud provider (e.g., "aws").
    - `resource_type`: The type of the resource (e.g., "ec2").
    - `region`: The geographical region of the resource.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of gRPC client calls to the `Supports` endpoint of any plugin receive a valid response without errors, regardless of whether the plugin implements the capability.
- **SC-002**: For plugins that implement the `Supports` capability, the response from the gRPC endpoint must exactly match the response from the plugin's internal implementation.
- **SC-003**: For plugins that do not implement the `Supports` capability, the gRPC endpoint must consistently return a standardized response indicating that the capability is not implemented.
- **SC-004**: The introduction of this feature results in zero regressions to the existing `GetProjectedCost` and `GetActualCost` functionalities.
- **SC-005**: 99% of all `Supports` gRPC queries complete within 50ms.
