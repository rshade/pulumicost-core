# Feature Specification: Pulumi Analyzer Plugin Integration

**Feature Branch**: `008-analyzer-plugin`  
**Created**: 2025-12-05  
**Status**: Draft  
**Input**: User description: "Pulumi Analyzer Plugin Integration"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Automatic Cost Estimation in Preview (Priority: P1)

As a Cloud Infrastructure Engineer, I want to see the estimated cost impact of my changes directly within the `pulumi preview` output so that I can optimize spend *before* resources are provisioned, without context switching to a separate tool or exporting files manually.

**Why this priority**: This is the core value proposition—replacing the manual "export plan -> run tool" workflow with a seamless "zero-click" experience.

**Independent Test**: Can be fully tested by running `pulumi preview` on a project with the analyzer configured and verifying that cost estimates appear in the standard output.

**Acceptance Scenarios**:

1. **Given** a Pulumi project with `pulumicost` analyzer configured and installed, **When** I run `pulumi preview` adding a costly resource (e.g., AWS RDS), **Then** I see an INFO diagnostic message in the CLI output stating the "Estimated Monthly Cost".
2. **Given** a valid project, **When** I run `pulumi preview` with no resource changes, **Then** the analyzer runs successfully and reports the current stack cost (or delta $0).
3. **Given** the analyzer is running, **When** it calculates costs, **Then** it does not block or crash the main Pulumi engine even if calculation takes a moment.

---

### User Story 2 - Plugin Installation & Handshake (Priority: P1)

As a Developer, I want the plugin to reliably start and connect to the Pulumi engine so that my workflow is not interrupted by tool failures.

**Why this priority**: The plugin architecture relies on a specific handshake (random port, stdout). If this fails, the feature doesn't work.

**Independent Test**: Can be tested by manually running `pulumi-analyzer-cost --serve` and verifying it prints a port and listens on it, and by verifying Pulumi can load it via `Pulumi.yaml`.

**Acceptance Scenarios**:

1. **Given** the `pulumi-analyzer-cost` binary in `~/.pulumi/plugins/analyzer-cost-vX.Y.Z/`, **When** Pulumi starts the plugin, **Then** the binary starts, listens on a random TCP port, and prints only that port number to stdout.
2. **Given** a `Pulumi.yaml` with `analyzers: - cost`, **When** `pulumi preview` starts, **Then** it successfully discovers and connects to the plugin without "plugin not found" or "connection refused" errors.

---

### User Story 3 - Robust Error Handling (Priority: P2)

As a User, I want the preview to complete successfully even if cost estimation fails (e.g., network issue), receiving a warning instead of a crash.

**Why this priority**: Cost estimation is auxiliary to deployment. It should never prevent a critical deployment (unless explicitly configured as a gate, which is out of scope for this MVP).

**Independent Test**: Simulate a pricing API failure or invalid resource graph and ensure `pulumi preview` still finishes.

**Acceptance Scenarios**:

1. **Given** the pricing engine cannot fetch rates (e.g., no internet), **When** I run `pulumi preview`, **Then** the analyzer returns a WARNING diagnostic ("Unable to estimate costs: ...") but allows the preview to complete.
2. **Given** a resource type that is not supported, **When** analyzed, **Then** it is silently ignored or logged to debug, without crashing the plugin.

### Edge Cases

- **Large Stacks**: For stacks with 1000+ resources, the plugin uses configurable timeout limits with optimized heuristics. No hard-coded maximum; timeouts are user-configurable via `~/.pulumicost/config.yaml`. A warning is logged if analysis exceeds the configured threshold.
- **Concurrency**: What happens if multiple previews run simultaneously? The random port selection must avoid collisions.
- **Logging**: Since `stdout` is used for the handshake, any other logs (debug/info) written to `stdout` during startup will break the integration. They must go to `stderr` or a log file.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users see a "Estimated Monthly Cost" line in their `pulumi preview` output for supported resources, both as a stack summary and as individual per-resource diagnostics.
- **SC-002**: The plugin handshake (startup to connection) succeeds in 100% of valid installation scenarios.
- **SC-003**: Adding the analyzer adds less than 2 seconds of latency to the `pulumi preview` of a small stack (under 50 resources).
- **SC-004**: The plugin handles pricing API failures by downgrading to a warning, ensuring 0% blocking of deployments due to pricing service unavailability (in default mode).

## Clarifications

### Session 2025-12-05

- Q: How does `pulumicost-core` discover, load, and manage these individual cost calculation plugins (e.g., for AWS, Azure, GCP)? → A: Config-Registered
- Q: For this MVP, should the `pulumicost-core` analyzer provide a configuration option to fail the `pulumi preview` or `pulumi up` if cost increases exceed a defined threshold, or should all cost diagnostics strictly remain informational (INFO/WARN) without blocking deployment? → A: Strictly Informational (MVP)
- Q: What is the preferred mechanism for `pulumicost-core` to pass the `PluginConfig` for a specific child cost plugin to that child plugin? → A: Environment Variables
- Q: How should resource mapping be implemented? → A: New Separate Mapper (mapping down to the existing interface)
- Q: Which approach is preferred for invoking the analyzer mode? → A: Distinct Subcommand (non-hidden)
- Q: How should a child cost plugin communicate pricing data retrieval errors or other failures back to `pulumicost-core` to enable appropriate `AnalyzeDiagnostic` messages? → A: Via `ErrorDetail` in gRPC response
- Q: Where should the analyzer's cost plugin configuration be stored? → A: Existing `~/.pulumicost/config.yaml` with new `analyzer` section
- Q: What is the acceptable maximum latency for analyzing a large stack (1000+ resources)? → A: No hard limit; configurable timeouts with optimized heuristics, log warning if exceeded
- Q: How should the analyzer plugin configuration in `~/.pulumicost/config.yaml` be structured? → A: Map of objects (key=name, value={path, env, enabled})
- Q: Should cost diagnostics be reported per-resource or as a stack-wide summary? → A: Both (per-resource details + stack summary)
- Q: What is the default logging destination? → A: Use existing Zerolog configuration

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST implement the `pulumirpc.Analyzer` gRPC server interface.
- **FR-002**: System MUST implement the `AnalyzeStack` method to receive and process the full list of resources from the Pulumi engine.
- **FR-003**: System MUST create a new mapper for `pulumirpc.Resource` protobuf messages to the internal `pulumicost.ResourceDescriptor` format.
- **FR-004**: System MUST utilize the existing `internal/engine` package to calculate projected costs, reusing the existing `Engine` and `SpecLoader` implementations.
- **FR-005**: System MUST return cost estimates as `AnalyzeDiagnostic` messages with `INFO` or `WARNING` severity; it MUST NOT return `ERROR` severity for cost overruns in this MVP to prevent blocking deployments. Cost estimates MUST be reported as both a stack-level summary and as individual per-resource diagnostics.
- **FR-006**: System MUST expose the analyzer mode via a distinct subcommand (e.g., `pulumicost analyzer serve`).
- **FR-007**: When starting in serve mode, System MUST listen on a random available TCP port and print *only* the port number to `stdout` as the first line.
- **FR-008**: System MUST utilize the existing `internal/logging` subsystem for all application logs. It MUST ensure that application logs do not corrupt the `stdout` handshake.
- **FR-009**: System MUST utilize the existing `internal/registry` and `internal/config` packages to discover, load, and configure cost plugins.
- **FR-010**: System MUST support configurable timeout limits by utilizing the existing timeout mechanisms in `internal/engine`.