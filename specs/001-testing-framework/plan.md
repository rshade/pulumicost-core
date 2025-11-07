# Implementation Plan: Testing Framework and Strategy

**Branch**: `001-testing-framework` | **Date**: 2025-11-06 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-testing-framework/spec.md`
**GitHub Issue**: [#9](https://github.com/rshade/pulumicost-core/issues/9)

## Summary

Establish a comprehensive testing framework for PulumiCost Core that achieves 80% minimum test coverage (95% for critical paths) through unit, integration, and end-to-end tests. The framework includes configurable mock plugins for isolation testing, organized test fixtures covering AWS/Azure/GCP, and CI/CD automation that blocks pull requests failing tests or coverage thresholds. This directly supports the constitution's Test-Driven Development principle and provides the foundation for reliable cost calculations.

## Technical Context

**Language/Version**: Go 1.24.5 (already in use per go.mod)
**Primary Dependencies**:
- `testing` (Go standard library)
- `github.com/stretchr/testify` (assertions and mocks)
- `google.golang.org/grpc` (mock plugin gRPC server)
- `github.com/google/go-cmp` (deep equality comparisons)

**Storage**: File-based (test fixtures stored as JSON/YAML files, no database)
**Testing**: Go native testing framework with testify assertions
**Target Platform**: Linux (amd64, arm64), macOS (amd64, arm64), Windows (amd64) - per constitution
**Project Type**: Single Go project with CLI (existing structure)
**Performance Goals**:
- Unit tests: <5 minutes total execution time
- Integration tests: <3 minutes total execution time
- E2E tests: <2 minutes total execution time
- CI complete pipeline: <15 minutes total

**Constraints**:
- Must achieve 80% minimum coverage (95% for critical paths) per constitution
- Must run with `-race` flag to detect concurrency issues
- Must support parallel test execution (`go test -parallel`)
- Must integrate with existing CI/CD pipeline (`.github/workflows/ci.yml`)

**Scale/Scope**:
- ~10,000 lines of test code estimated across all categories
- ~100 test fixtures covering AWS, Azure, GCP scenarios
- 5 configurable mock plugin response types
- 3 error injection scenarios

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Verify compliance with PulumiCost Core Constitution (`.specify/memory/constitution.md`):

- [X] **Plugin-First Architecture**: Testing framework itself doesn't require plugins, but includes mock plugin infrastructure to test plugin communication. Compliant - this is orchestration logic.
- [X] **Test-Driven Development**: This feature IS the TDD infrastructure. Tests for the testing framework itself will follow TDD (meta-testing). Compliant.
- [X] **Cross-Platform Compatibility**: All tests will run on Linux, macOS, Windows. CI already verifies cross-platform builds. Compliant.
- [X] **Documentation as Code**: Will update docs/testing/ with testing guide for developers. Part of implementation. Compliant.
- [X] **Protocol Stability**: Mock plugin will implement current CostSource protocol. No protocol changes. Compliant.
- [X] **Quality Gates**: This feature IMPLEMENTS the quality gates. Will integrate with existing CI. Compliant.
- [X] **Multi-Repo Coordination**: Testing framework is core-only. No cross-repo dependencies. Compliant.

**Violations Requiring Justification**: None - all principles satisfied.

## Project Structure

### Documentation (this feature)

```text
specs/001-testing-framework/
├── plan.md              # This file (/speckit.plan command output)
├── spec.md              # Feature specification (completed)
├── research.md          # Phase 0 output (testing tools research)
├── data-model.md        # Phase 1 output (test entities)
├── quickstart.md        # Phase 1 output (how to write tests)
├── contracts/           # Phase 1 output (mock plugin API)
│   └── mock-plugin-api.md
├── checklists/          # Quality validation
│   └── requirements.md  # Specification checklist (completed)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created yet)
```

### Source Code (repository root)

Existing Go project structure - will add test directories:

```text
# Existing structure (unchanged)
cmd/pulumicost/          # CLI entry point
internal/
├── cli/                 # CLI commands
├── engine/              # Cost calculation engine
├── pluginhost/          # Plugin management
├── registry/            # Plugin discovery
├── ingest/              # Pulumi plan parsing
├── spec/                # Local pricing specs
├── proto/               # Protocol adapters
└── config/              # Configuration

# NEW: Testing infrastructure
test/
├── unit/                # Unit tests by package
│   ├── engine/          # Engine unit tests
│   ├── cli/             # CLI unit tests
│   ├── pluginhost/      # Plugin host unit tests
│   ├── registry/        # Registry unit tests
│   ├── ingest/          # Ingest unit tests
│   └── config/          # Config unit tests
├── integration/         # Cross-component tests
│   ├── cli_workflow_test.go      # CLI command execution
│   ├── plugin_comm_test.go       # gRPC plugin communication
│   └── config_loading_test.go    # Configuration integration
├── e2e/                 # End-to-end scenarios
│   ├── projected_cost_test.go    # Full projected workflow
│   ├── actual_cost_test.go       # Full actual cost workflow
│   └── output_formats_test.go    # Golden file testing
├── fixtures/            # Test data files
│   ├── plans/           # Pulumi plan JSON files
│   │   ├── aws/         # AWS resource plans
│   │   ├── azure/       # Azure resource plans
│   │   └── gcp/         # GCP resource plans
│   ├── responses/       # Mock plugin responses
│   ├── configs/         # Test configuration files
│   └── specs/           # Test pricing specifications
├── mocks/               # Mock implementations
│   └── plugin/          # Mock plugin server
│       ├── server.go    # gRPC mock server
│       └── config.go    # Response configuration
└── benchmarks/          # Performance tests
    └── engine_bench_test.go
```

**Structure Decision**: Extend existing single Go project structure with comprehensive `/test` directory organization. Mirrors internal package structure for unit tests, adds integration/e2e/fixtures/mocks for advanced testing scenarios. Aligns with Go testing conventions and existing codebase patterns.

## Complexity Tracking

No constitution violations - complexity tracking not required.

---

## Phase 0: Research & Decisions

### Research Topics

1. **Testing Frameworks**: Evaluate Go testing approaches (standard library vs enhanced frameworks)
2. **Mock Strategies**: Research gRPC mocking patterns for plugin testing
3. **Coverage Tools**: Identify best coverage reporting tools for Go
4. **CI Integration**: Review existing GitHub Actions workflow for test integration points
5. **Golden File Testing**: Research golden file testing libraries for output validation
6. **Benchmark Patterns**: Investigate Go benchmark best practices for performance regression detection

### Initial Decisions (to be validated in research.md)

- **Testing Framework**: Go standard `testing` package + `testify` for assertions (already in use)
- **gRPC Mocking**: Custom mock plugin server (more control than generic mocks)
- **Coverage**: Go native `go test -cover` + CI coverage reports (already configured)
- **Golden Files**: `github.com/sebdah/goldie` or custom implementation
- **CI Integration**: Extend existing `.github/workflows/ci.yml` (already has test job)

---

## Phase 1: Design Artifacts

### Data Model

Key entities to be defined in `data-model.md`:

1. **TestSuite**: Metadata, categories, execution results
2. **MockPlugin**: Configuration, response types, error injection
3. **TestFixture**: File organization, loading mechanisms, naming conventions
4. **CoverageReport**: Package-level metrics, critical path identification
5. **BenchmarkResult**: Performance baselines, regression thresholds

### Contracts

Mock Plugin API to be defined in `contracts/mock-plugin-api.md`:

```
MockPlugin gRPC Service:
- Configure(ResponseConfig) → Status
- SetError(MethodName, ErrorType) → Status
- Reset() → Status

Implements CostSourceService protocol:
- Name() → PluginInfo
- GetProjectedCost(ResourceDescriptor[]) → CostResult[]
- GetActualCost(ActualCostRequest) → ActualCostResult[]
```

### Quickstart

Developer guide to be created in `quickstart.md`:

- How to write a unit test for a new function
- How to add integration test for cross-component flow
- How to create test fixtures
- How to run tests locally vs CI
- How to debug failing tests
- How to add benchmarks

---

## Next Steps

1. **Complete Phase 0**: Generate `research.md` with detailed research findings
2. **Complete Phase 1**: Generate `data-model.md`, `contracts/`, and `quickstart.md`
3. **Run Constitution Check**: Verify all principles still satisfied after design
4. **Generate Tasks**: Run `/speckit.tasks` to create implementation checklist

---

## Notes

- This feature directly implements the constitution's TDD principle
- Existing CI workflow (`.github/workflows/ci.yml`) already has test job - will enhance it
- Current coverage is 24.2% overall - this feature will bring it to 80%+ target
- Critical paths already identified in constitution: CLI, engine, pluginhost
- Mock plugin reusable for future plugin development testing
