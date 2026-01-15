# Implementation Plan: Plugin Ecosystem Maturity

**Branch**: `016-plugin-ecosystem-maturity` | **Date**: 2025-12-02 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/016-plugin-ecosystem-maturity/spec.md`

## Summary

Implement a comprehensive plugin conformance testing framework and optional E2E
testing infrastructure to ensure reliable plugin ecosystem. The conformance
suite will validate that plugins correctly implement the FinFocus gRPC
protocol, while E2E tests will verify cost data accuracy against real cloud
provider APIs.

## Technical Context

**Language/Version**: Go 1.25
**Primary Dependencies**: google.golang.org/grpc v1.77.0, github.com/rshade/finfocus-spec v0.4.1, github.com/stretchr/testify v1.11.1
**Storage**: N/A (test framework, no persistent storage)
**Testing**: Go testing with testify assertions, table-driven tests
**Target Platform**: Linux (amd64, arm64), macOS (amd64, arm64), Windows (amd64)
**Project Type**: Single project (Go CLI tool)
**Performance Goals**: Full conformance suite completes in under 5 minutes
**Constraints**: 10-second timeout per plugin call, 100ms retry delay
**Scale/Scope**: ~20 conformance test cases, 1000 resource batch limit

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Verify compliance with FinFocus Core Constitution:

- [x] **Plugin-First Architecture**: This feature validates plugin compliance,
  not core functionality. Conformance suite tests plugins via gRPC protocol.
- [x] **Test-Driven Development**: Tests ARE the feature. Conformance tests
  will be written first, then test infrastructure to run them.
- [x] **Cross-Platform Compatibility**: Go code, no platform-specific
  dependencies. CI will verify builds on Linux, macOS, Windows.
- [x] **Documentation as Code**: Quickstart guide planned for plugin
  developers. E2E test setup documentation required.
- [x] **Protocol Stability**: Conformance tests ensure protocol compliance.
  Version checking (FR-008a) prevents mismatches.
- [x] **Quality Gates**: Standard CI checks apply. 80% coverage minimum.
- [x] **Multi-Repo Coordination**: Depends on finfocus-spec for protocol
  definitions. No cross-repo changes required for initial implementation.

**Violations Requiring Justification**: None

## Project Structure

### Documentation (this feature)

```text
specs/016-plugin-ecosystem-maturity/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output (via /speckit.tasks)
```

### Source Code (repository root)

```text
internal/conformance/
├── doc.go               # Package documentation (exists, empty)
├── suite.go             # ConformanceSuite orchestrator
├── runner.go            # TestRunner executes individual tests
├── reporter.go          # JUnit XML and JSON output
├── tests/               # Individual conformance test cases
│   ├── protocol.go      # Protocol compliance tests (FR-001 to FR-004)
│   ├── timeout.go       # Timeout behavior tests (FR-005)
│   ├── batch.go         # Batch handling tests (FR-006)
│   ├── error.go         # Error code tests (FR-003)
│   └── context.go       # Context cancellation tests (FR-002)
└── version.go           # Protocol version checking (FR-008a)

test/e2e/
├── aws/                 # AWS E2E tests (optional, manual)
│   ├── cost_test.go     # AWS cost API validation
│   └── fixtures/        # Expected cost ranges
├── azure/               # Azure E2E tests (future)
└── gcp/                 # GCP E2E tests (future)

test/mocks/plugin/       # Existing mock plugin (enhanced)
├── server.go            # Already exists
├── api.go               # Already exists
└── conformance.go       # New: reference implementation for testing

internal/cli/
└── plugin_conformance.go # New: CLI command for conformance testing
```

**Structure Decision**: Conformance suite goes in `internal/conformance/` as
core infrastructure. E2E tests go in `test/e2e/` as optional integration tests.
CLI integration via new `plugin conformance` subcommand.

## Complexity Tracking

> No constitution violations. Standard Go package structure.
