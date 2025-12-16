# Feature Specification: Reference Recorder Plugin for DevTools

**Feature Branch**: `104-recorder-plugin`
**Created**: 2025-12-11
**Status**: Draft
**Input**: User description: "Add Reference Recorder Plugin for DevTools - canonical reference implementation and developer tool for inspecting data shapes and testing Core interactions without external dependencies"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Request Data Inspection (Priority: P1)

As a plugin developer, I want to capture and inspect the exact gRPC requests that PulumiCost Core sends to plugins, so that I can understand data shapes, field values, and request structures without building a full plugin first.

**Why this priority**: This directly addresses the "Blind Integration" problem - plugin developers cannot currently see what data the Core sends. This is the primary value proposition of the recorder plugin.

**Independent Test**: Can be tested by running the recorder against a sample Pulumi plan and verifying that JSON files are created containing full request payloads.

**Acceptance Scenarios**:

1. **Given** the recorder plugin is configured with an output directory, **When** `pulumicost cost projected` is executed with a Pulumi plan, **Then** each request is saved as a separate JSON file with timestamp, method name, and unique identifier in the filename.
2. **Given** the recorder is running with default configuration, **When** a request is processed, **Then** the JSON output includes all fields: resource types, providers, tags, regions, SKUs, and any other request metadata.
3. **Given** the output directory does not exist, **When** the recorder starts, **Then** the directory is created automatically with appropriate permissions.

---

### User Story 2 - Mock Response Generation (Priority: P2)

As a Core developer, I want the recorder to generate randomized but valid cost responses, so that I can test aggregation logic, output formatting, and reporting features without requiring real cloud credentials or external services.

**Why this priority**: This enables testing Core's aggregation and reporting logic in isolation. Without mock responses, developers need real cloud integrations to test end-to-end flows.

**Independent Test**: Can be tested by running the recorder with mock mode enabled and verifying that responses contain valid cost data structures with randomized values.

**Acceptance Scenarios**:

1. **Given** mock response mode is enabled, **When** `GetProjectedCost` is called, **Then** the recorder returns a valid response with randomized cost values in USD currency.
2. **Given** mock response mode is enabled, **When** `GetActualCost` is called, **Then** the recorder returns a valid response with randomized historical cost records following FOCUS 1.2 format.
3. **Given** mock response mode is disabled (default), **When** any cost method is called, **Then** the recorder returns empty/default responses indicating no cost data available.

---

### User Story 3 - Reference Implementation Study (Priority: P3)

As a new plugin developer, I want to study a complete, working plugin implementation that demonstrates best practices, so that I can use it as a template for building my own plugin.

**Why this priority**: While documentation helps, seeing working code is invaluable. The recorder serves as executable documentation of SDK patterns.

**Independent Test**: Can be tested by verifying the plugin builds, runs, and demonstrates all SDK features mentioned in documentation.

**Acceptance Scenarios**:

1. **Given** a developer reads the recorder source code, **When** they examine the implementation, **Then** they see examples of all gRPC interfaces (CostSource methods), error handling, and SDK builder usage.
2. **Given** a developer wants to create a new plugin, **When** they copy the recorder structure, **Then** they have a working foundation that includes configuration handling, graceful shutdown, and logging patterns.
3. **Given** the recorder is built and installed, **When** `pulumicost plugin validate` is run, **Then** the recorder passes all validation checks as a conformant plugin.

---

### User Story 4 - Contract Testing Support (Priority: P4)

As a Core maintainer, I want to use the recorder as a target for contract tests, so that I can detect regressions in plugin orchestration without depending on external plugins.

**Why this priority**: Internal testing reduces dependency on external repositories and speeds up CI feedback loops.

**Independent Test**: Can be tested by running Core's plugin integration tests against the recorder and verifying test pass/fail behavior.

**Acceptance Scenarios**:

1. **Given** the recorder is installed in the plugins directory, **When** Core's plugin integration tests run, **Then** the recorder is discovered and responds correctly to all protocol methods.
2. **Given** Core makes a breaking change to the plugin protocol, **When** CI runs against the recorder, **Then** tests fail clearly indicating the contract violation.

---

### Edge Cases

- What happens when the output directory is not writable? Plugin should log an error and continue operating without recording.
- What happens when disk space runs out during recording? Plugin should handle gracefully, log warning, and stop recording while continuing to respond to requests.
- How does the recorder handle malformed requests? Plugin should record the raw request anyway (for debugging purposes) and return appropriate gRPC error codes.
- What happens with very large requests (e.g., plans with thousands of resources)? Each request is recorded individually; system limits (file size, disk space) are respected.
- What happens if multiple Core instances call the recorder simultaneously? Plugin must be thread-safe and use unique identifiers to prevent filename collisions.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST implement all CostSourceService gRPC methods: `Name()`, `GetProjectedCost()`, and `GetActualCost()`.
- **FR-002**: System MUST serialize every incoming gRPC request to a JSON file when request capture is enabled.
- **FR-003**: System MUST use the filename format `<ISO8601-timestamp>_<method-name>_<ULID>.json` for recorded files.
- **FR-004**: System MUST support configuring the output directory via `PULUMICOST_RECORDER_OUTPUT_DIR` environment variable (default: `./recorded_data`).
- **FR-005**: System MUST support enabling/disabling mock responses via `PULUMICOST_RECORDER_MOCK_RESPONSE` environment variable (default: `false`).
- **FR-006**: System MUST generate randomized but structurally valid cost responses when mock mode is enabled.
- **FR-007**: System MUST return empty/default responses when mock mode is disabled.
- **FR-008**: System MUST use the pluginsdk library from pulumicost-spec v0.4.6 or later for SDK patterns (BasePlugin, Matcher, Serve, request validation helpers).
- **FR-009**: System MUST create the output directory if it does not exist.
- **FR-010**: System MUST support both TCP (ProcessLauncher) and stdio (StdioLauncher) communication modes.
- **FR-011**: System MUST handle graceful shutdown on SIGINT/SIGTERM, flushing any pending writes.
- **FR-012**: System MUST be thread-safe for concurrent requests.
- **FR-013**: System MUST follow the standard plugin binary naming convention: `pulumicost-plugin-recorder`.
- **FR-014**: System MUST include a `plugin.manifest.json` file with plugin metadata.
- **FR-015**: System MUST use the pluginsdk request validation helpers (from v0.4.6) to validate incoming requests before processing, demonstrating best practices for input sanitation.

### Key Entities

- **RecordedRequest**: Represents a captured gRPC request containing timestamp, method name, unique ID, full request payload, and optional metadata (client info, duration).
- **MockCostResponse**: Represents a generated mock response including randomized cost values, currency (USD), billing details, and FOCUS 1.2 compliant structure for actual costs.
- **PluginConfiguration**: Represents runtime configuration including output directory path, mock mode enabled flag, and any future configuration options.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Plugin developers can inspect complete request data within 5 minutes of running the recorder (measured by time to first recorded file).
- **SC-002**: Core developers can run full end-to-end tests using mock responses without any external credentials or services.
- **SC-003**: The recorder plugin passes all Core plugin validation checks (`pulumicost plugin validate`).
- **SC-004**: Recorded JSON files are human-readable and contain all fields present in the original gRPC requests.
- **SC-005**: Mock responses are valid according to the pulumicost-spec protocol definitions (can be deserialized without errors).
- **SC-006**: The recorder integrates with the standard build system via a single command (`make build-recorder`).
- **SC-007**: CI/CD pipeline builds and tests the recorder on every commit to detect protocol regressions.
- **SC-008**: Documentation enables a new developer to use the recorder within 10 minutes (quickstart guide).

## Dependencies

- **pulumicost-spec v0.4.6+**: Required for pluginsdk with request validation helpers
  - Release: https://github.com/rshade/pulumicost-spec/releases/tag/v0.4.6
  - Key features used: BasePlugin, Matcher, Serve, request validation helpers, FOCUS 1.2 builders
  - The v0.4.6 release adds request validation helpers that enable proper input sanitation and error handling

## Assumptions

- The pulumicost-spec SDK (`pluginsdk`) v0.4.6+ provides all necessary building blocks (BasePlugin, Matcher, Serve, builders, validation helpers).
- The recorder will be shipped as part of pulumicost-core releases or as an easily buildable target.
- JSON is the appropriate format for recorded data (human-readable, widely supported).
- Environment variables are the preferred configuration mechanism (consistent with Core patterns).
- The recorder does not need to persist state across restarts (ephemeral recording).
- Mock responses use USD as the default currency.
- The recorder targets development/testing use cases, not production deployments.

## Out of Scope

- Replay functionality (playing back recorded requests to test plugins)
- Encryption or compression of recorded data
- Remote storage integration (S3, GCS, etc.)
- Web UI for browsing recorded data
- Real cloud cost lookups or integrations
- Budget and Recommendation gRPC interfaces (future enhancement)
