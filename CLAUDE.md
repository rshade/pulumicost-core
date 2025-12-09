# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## CRITICAL INSTRUCTIONS

**DO NOT RUN `git commit`** - This is explicitly forbidden. Use `git add`, `git status`, `git diff`, and `git log` only. The user will commit manually.

**ALWAYS run `make lint` and `make test`** before claiming success.

**DO NOT modify `.golangci.yml`** without explicit approval.

## Project Overview

PulumiCost Core is a CLI tool and plugin host system for calculating cloud infrastructure costs from Pulumi infrastructure definitions. It provides both projected cost estimates and actual historical cost analysis through a plugin-based architecture.

## Build Commands

```bash
make build         # Build binary to bin/pulumicost
make test          # Run unit tests (default, fast)
make test-race     # Run with race detector
make test-integration  # Integration tests (slower)
make test-e2e      # E2E tests (requires AWS credentials)
make lint          # Run golangci-lint + markdownlint
make validate      # go mod tidy, go vet
make clean         # Remove build artifacts
make run           # Build and run with --help
make dev           # Build and run without args
```

### Single Package/Test Commands

```bash
go test -v ./internal/cli/...           # Test specific package
go test -v ./internal/engine/...        # Test engine package
go test -run TestSpecificFunction ./... # Run specific test

# Coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out        # View in browser
```

## Go Version

**Project Go Version**: 1.25.5 (see `go.mod`)

**CRITICAL**: Before claiming any Go version "doesn't exist" or suggesting version
changes, verify on <https://go.dev/dl/> first.

## SpecKit Feature Development Workflow

This project uses SpecKit for structured feature development. Features are developed in numbered branches with specifications, plans, and task lists.

### Slash Commands

- `/speckit.specify [description]` - Create feature specification from natural language
- `/speckit.clarify` - Identify underspecified areas and ask clarification questions
- `/speckit.plan` - Generate technical implementation plan
- `/speckit.tasks` - Generate actionable task list from plan
- `/speckit.implement` - Execute tasks from tasks.md
- `/speckit.analyze` - Cross-artifact consistency analysis
- `/speckit.checklist` - Generate custom validation checklist
- `/speckit.taskstoissues` - Convert tasks to GitHub issues

### Directory Structure

```text
specs/
└── NNN-feature-name/
    ├── spec.md          # Feature specification (what/why)
    ├── plan.md          # Technical plan (how)
    ├── tasks.md         # Actionable task list
    └── checklists/      # Validation checklists
```

### Constitution

See `.specify/memory/constitution.md` for non-negotiable principles:

- Plugin-First Architecture
- Test-Driven Development (80% coverage minimum, 95% for critical paths)
- Cross-Platform Compatibility
- Documentation as Code
- Protocol Stability

## Architecture

### Core Components

1. **CLI Layer** (`internal/cli/`) - Cobra-based commands:
   - `cost projected` - Estimate costs from Pulumi preview JSON
   - `cost actual` - Fetch historical costs with time ranges/grouping
   - `plugin list/validate/install/remove/update` - Plugin management
   - `analyzer serve` - Pulumi Analyzer gRPC server for zero-click cost estimation

2. **Engine** (`internal/engine/`) - Core cost calculation:
   - Orchestrates between plugins and local specs
   - Output formats: table, JSON, NDJSON
   - Cross-provider aggregation with time-based grouping
   - `hoursPerMonth = 730` for monthly calculations

3. **Plugin Host** (`internal/pluginhost/`) - gRPC plugin management:
   - `ProcessLauncher` (TCP) and `StdioLauncher` (stdin/stdout)
   - 10-second timeout with 100ms retry delays
   - Always call `cmd.Wait()` after `Kill()` to prevent zombies

4. **Registry** (`internal/registry/`) - Plugin discovery:
   - Scans `~/.pulumicost/plugins/<name>/<version>/`
   - Optional `plugin.manifest.json` validation

5. **Ingestion** (`internal/ingest/`) - Pulumi plan parsing:
   - Converts `pulumi preview --json` to resource descriptors
   - **CRITICAL**: Must inspect `newState` to extract `Inputs` correctly

6. **Analyzer** (`internal/analyzer/`) - Pulumi Analyzer protocol:
   - Implements `pulumirpc.AnalyzerServer` for zero-click cost estimation
   - ADVISORY enforcement (never blocks deployments)
   - Prints ONLY port number to stdout (Pulumi handshake protocol)

7. **TUI** (`internal/tui/`) - Shared Terminal UI components:
   - Built on Bubble Tea and Lip Gloss
   - Adaptive color schemes (light/dark terminal detection)
   - Reusable progress indicators, styled text, and tables

### Data Flow

```text
Pulumi JSON → Ingestion → Resource Descriptors → Engine
                                                    ↓
                                        Plugins (gRPC) or Local Specs
                                                    ↓
                                              Cost Results → Output Rendering
```

### Plugin Communication

Plugins communicate via gRPC using protocol buffers from `pulumicost-spec`:

- `Name()` - Plugin identification
- `GetProjectedCost()` - Estimated costs for resources
- `GetActualCost()` - Historical costs from cloud APIs

## Documentation

All documentation lives in `docs/` with GitHub Pages deployment.

### Commands

```bash
make docs-lint     # Lint markdown
make docs-build    # Build Jekyll site
make docs-serve    # Serve at http://localhost:4000/pulumicost-core/
make docs-validate # Validate structure
```

### Structure

```text
docs/
├── guides/           # Audience-specific (User, Developer, Architect, Business)
├── getting-started/  # Quickstart, installation, examples
├── architecture/     # System design, core concepts, roadmap
├── plugins/          # Plugin development, SDK, per-plugin docs
├── reference/        # CLI, API, configuration, error codes
├── deployment/       # Installation, Docker, CI/CD, security
└── support/          # FAQ, troubleshooting, contributing
```

### Key Files

- `docs/README.md` - Documentation hub with navigation
- `docs/llms.txt` - Machine-readable index for AI tools
- `docs/_config.yml` - Jekyll configuration

## Key Patterns

### CLI Package (`internal/cli/`)

- Use `RunE` not `Run` for error handling
- Use `cmd.Printf()` for output (not `fmt.Printf()`)
- Defer cleanup functions immediately after obtaining resources
- Support multiple date formats: "2006-01-02", RFC3339

### Logging (Zerolog)

```go
// Get logger from context (includes trace_id)
log := logging.FromContext(ctx)
log.Debug().Ctx(ctx).Str("component", "engine").Msg("starting")
```

Enable debug: `--debug` flag or `PULUMICOST_LOG_LEVEL=debug`

**Log Levels**: TRACE (property extraction) → DEBUG (function entry/exit) →
INFO (high-level ops) → WARN (recoverable issues) → ERROR (failures)

**Standard Fields**: `trace_id`, `component`, `operation`, `duration_ms`

## Testing

### Test Directory Structure

```text
test/
├── unit/           # Unit tests by package (engine, config, spec)
├── integration/    # Cross-component tests (plugin communication)
├── e2e/            # End-to-end tests (separate Go module)
├── fixtures/       # Test data (plans, specs, configs, responses)
├── mocks/          # Mock implementations (plugin server)
└── benchmarks/     # Performance tests
```

### E2E Testing

**Location**: `test/e2e/` (separate Go module)

**Prerequisites**: AWS credentials, Pulumi CLI, `make build`

```bash
eval "$(aws configure export-credentials --format env)"
export PATH="$HOME/.pulumi/bin:$PATH"
export PULUMI_CONFIG_PASSPHRASE="e2e-test-passphrase"
make test-e2e
```

**CRITICAL**: E2E tests MUST call actual pulumicost CLI binary.
Never simulate cost values or stub CLI execution.

### Local Plugin Development

To debug plugin issues during Core development:

1. Clone the plugin repository (e.g., `pulumicost-plugin-aws-public`)
2. Modify the plugin code (add logging, fix type mapping)
3. Build: `make build-region REGION=us-east-1`
4. Install: Copy binary to `~/.pulumicost/plugins/<plugin>/<version>/`
5. Run Core E2E tests to verify

## Important Files

- `cmd/pulumicost/main.go` - CLI entry point
- `internal/engine/engine.go` - Core orchestration
- `internal/pluginhost/host.go` - Plugin client management
- `internal/ingest/pulumi_plan.go` - Pulumi plan parsing
- `.specify/memory/constitution.md` - Project principles and quality gates
- `examples/plans/aws-simple-plan.json` - Sample plan for testing

## Pulumi Integration Notes

### Plan JSON Parsing

The `pulumi preview --json` output nests resource details under `newState`.
Ingestion MUST inspect `newState` to extract `inputs` and `type`. Without this,
property extraction fails and plugins return `InvalidArgument` errors.

### Property Extraction

The adapter (`internal/proto/adapter.go`) relies on the `Inputs` map to extract:

- **SKU**: from `instanceType`, `type`, etc.
- **Region**: from `availabilityZone`, `region`

If ingestion fails to populate `Inputs`, these fields are empty.

### Resource Type Compatibility

Pulumi provides types like `aws:ec2/instance:Instance` (Type Token). Plugins may
expect `aws:ec2:Instance` or just `ec2`. Plugins should handle the standard
Pulumi format or normalize internally.

### Pulumi SDK Import Path

For Analyzer development, use the correct import:

```go
pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
// NOT: github.com/pulumi/pulumi/sdk/v3/proto/go/pulumirpc
```

## Multi-Repository Ecosystem

PulumiCost operates across three repositories:

- **pulumicost-core** (this repo) - CLI tool, plugin host, orchestration
- **pulumicost-spec** - Protocol buffer definitions, SDK generation
- **pulumicost-plugin** - Plugin implementations (Kubecost, Vantage, etc.)

Cross-repo changes follow the protocol in `.specify/memory/constitution.md`.

## Dependencies

Key dependencies (see `go.mod` for versions):

- `github.com/spf13/cobra` - CLI framework
- `google.golang.org/grpc` - Plugin communication
- `github.com/rs/zerolog` - Structured logging
- `github.com/rshade/pulumicost-spec` - Protocol definitions
- `github.com/pulumi/pulumi/sdk/v3` - Pulumi SDK for Analyzer (v3.210.0+)
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - TUI styling
