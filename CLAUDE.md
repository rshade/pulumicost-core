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

# Plugin management
./bin/pulumicost plugin list
./bin/pulumicost plugin validate
```

### Test Requirements

- **Unit tests**: Must achieve 80% coverage minimum
- **Critical paths**: Must achieve 95% coverage
- **All error paths**: Must be tested
- **Performance regressions**: Must be detected via benchmarks
- **Integration scenarios**: Must include plugin communication flows
- **End-to-end workflows**: Must test complete CLI usage

### CI/CD Integration

The existing CI/CD pipeline automatically runs all tests including the new framework:

- Unit tests with coverage reporting
- Integration tests with timeout handling
- Linting and security scanning
- Cross-platform build verification

**Never complete a project without running:**

```bash
make test    # Run all tests
make lint    # Run linting
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

PulumiCost uses zerolog for structured logging with distributed tracing support.

#### Enabling Debug Output

```bash
# CLI flag
pulumicost cost projected --debug --pulumi-json plan.json

# Environment variable
export PULUMICOST_LOG_LEVEL=debug
export PULUMICOST_LOG_FORMAT=json    # json or console
export PULUMICOST_TRACE_ID=external-trace-123  # inject external trace ID
```

#### Configuration Precedence

1. CLI flags (`--debug`)
2. Environment variables (`PULUMICOST_LOG_LEVEL`)
3. Config file (`~/.pulumicost/config.yaml`)
4. Default (info level, console format)

#### Logging Patterns for Developers

```go
// Get logger from context (preferred - includes trace_id)
log := logging.FromContext(ctx)
log.Debug().
    Ctx(ctx).
    Str("component", "engine").
    Str("operation", "get_projected_cost").
    Int("resource_count", len(resources)).
    Msg("starting projected cost calculation")

// Create component sub-logger
logger = logging.ComponentLogger(logger, "registry")

// Log with duration
start := time.Now()
// ... operation ...
log.Info().
    Ctx(ctx).
    Dur("duration_ms", time.Since(start)).
    Msg("operation complete")
```

#### Standard Log Fields

| Field         | Purpose             | Example                      |
| ------------- | ------------------- | ---------------------------- |
| `trace_id`    | Request correlation | "01HQ7X2J3K4M5N6P7Q8R9S0T1U" |
| `component`   | Package identifier  | "cli", "engine", "registry"  |
| `operation`   | Current operation   | "get_projected_cost"         |
| `duration_ms` | Operation timing    | 245                          |

#### Log Levels

- **TRACE**: Property extraction, detailed calculations
- **DEBUG**: Function entry/exit, retries, intermediate values
- **INFO**: High-level operations (command start/end)
- **WARN**: Recoverable issues (fallbacks, deprecations)
- **ERROR**: Failures needing attention

#### Trace ID Management

```go
// Generate trace ID at entry point (usually in CLI PersistentPreRunE)
traceID := logging.GetOrGenerateTraceID(ctx)
ctx = logging.ContextWithTraceID(ctx, traceID)

// TracingHook automatically injects trace_id when using .Ctx(ctx)
```

## CI/CD Pipeline

Complete CI/CD pipeline setup with GitHub Actions for automated testing, building, and release management.

### CI Pipeline (.github/workflows/ci.yml)

Triggered on pull requests and pushes to main branch:

**Test Job:**

- Go 1.25.5 setup with caching
- Unit tests with race detection and coverage reporting
- Coverage threshold check (minimum 61%)
- Artifacts uploaded for coverage reports

**Lint Job:**

- golangci-lint with project-specific configuration
- Security scanning with gosec included
- Timeout set to 5 minutes

**Security Job:**

- govulncheck for dependency vulnerability scanning
- Checks for known vulnerabilities in Go dependencies

**Validation Job:**

- gofmt formatting checks
- go mod tidy verification
- go vet static analysis

**Build Job:**

- Cross-platform builds (Linux, macOS, Windows)
- Support for amd64 and arm64 architectures
- Build artifacts uploaded with proper naming

### Release Pipeline (.github/workflows/release.yml)

Triggered on version tags (v\*):

**Multi-Platform Binaries:**

- Linux: amd64, arm64
- macOS: amd64, arm64
- Windows: amd64
- Naming convention: `pulumicost-v{version}-{os}-{arch}`

**Release Features:**

- Automatic changelog generation from git history
- SHA256 checksums for all binaries
- GitHub Release creation with proper metadata
- Asset upload with verification instructions
- Pre-release detection for tags containing hyphens

### Quality Gates

**Code Quality:**

- golangci-lint with essential linters (errcheck, govet, staticcheck, gosec, etc.)
- Security scanning integrated into CI pipeline
- Formatting and import organization enforced

**Coverage Requirements:**

- Minimum 20% code coverage (adjustable as project matures)
- Coverage reports generated and uploaded as artifacts
- Automatic threshold validation in CI

**Build Verification:**

- Cross-platform compilation verification
- Binary naming consistency
- Version information embedded in binaries

### Troubleshooting Commands

```bash
# Fix parallel linting conflicts
pkill golangci-lint || true

# Check coverage details
go tool cover -html=coverage.out

# Test release build locally
GOOS=linux GOARCH=amd64 make build

# Validate workflow syntax
gh workflow validate .github/workflows/ci.yml
```

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
- `github.com/rshade/pulumicost-spec` - Protocol definitions and pluginsdk
- `github.com/pulumi/pulumi/sdk/v3` - Pulumi SDK for Analyzer (v3.210.0+)
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - TUI styling
- `golang.org/x/term` - Terminal detection utilities

## Environment Variable Constants (pluginsdk)

PulumiCost uses standardized environment variable constants from `pluginsdk` for
consistency between core and plugins.

### Available Constants

```go
import "github.com/rshade/pulumicost-spec/sdk/go/pluginsdk"

// Plugin communication
pluginsdk.EnvPort        // "PULUMICOST_PLUGIN_PORT" - Primary port for gRPC

// Logging configuration
pluginsdk.EnvLogLevel    // "PULUMICOST_LOG_LEVEL" - Log verbosity
pluginsdk.EnvLogFormat   // "PULUMICOST_LOG_FORMAT" - Log format (json/text)
pluginsdk.EnvLogFile     // "PULUMICOST_LOG_FILE" - Log file path

// Distributed tracing
pluginsdk.EnvTraceID     // "PULUMICOST_TRACE_ID" - Trace ID for request correlation
```

### Usage Patterns

**Setting environment variables (core → plugin):**

```go
cmd.Env = append(os.Environ(),
    fmt.Sprintf("%s=%d", pluginsdk.EnvPort, port),
    fmt.Sprintf("PORT=%d", port), // Legacy fallback
)
```

**Reading environment variables (in tests):**

```go
os.Setenv(pluginsdk.EnvLogLevel, "debug")
defer os.Unsetenv(pluginsdk.EnvLogLevel)
```

**Note**: Config-specific variables like `PULUMICOST_OUTPUT_FORMAT` and
`PULUMICOST_CONFIG_STRICT` are NOT in pluginsdk as they are core-specific.

## Package-Specific Documentation

### internal/cli

The CLI package implements the Cobra-based command-line interface. Key patterns:

- Use `RunE` not `Run` for error handling
- Always use `cmd.Printf()` for output (not `fmt.Printf()`)
- Defer cleanup functions immediately after obtaining resources
- Support multiple date formats: "2006-01-02", RFC3339
- See `internal/cli/CLAUDE.md` for detailed CLI architecture and patterns

### internal/engine

The engine package orchestrates cost calculations between plugins and specs:

- Tries plugins first, falls back to local YAML specs
- Supports three output formats: table, JSON, NDJSON
- Uses `hoursPerMonth = 730` for monthly calculations
- Always returns some result, even if placeholder
- **Actual Cost Pipeline Features**:
  - `GetActualCostWithOptions()` - Advanced querying with time ranges and filters
  - Resource filtering with `matchesTags()` helper for tag-based filtering
  - Cost aggregation logic for daily/monthly breakdowns
  - Grouping support (resource, type, provider, date)
  - Multiple date format parsing ("2006-01-02", RFC3339)
- **Cross-Provider Aggregation Features**:
  - `CreateCrossProviderAggregation()` - Time-based multi-provider cost analysis
  - Currency validation system with `ErrMixedCurrencies` protection
  - Advanced input validation (empty results, invalid date ranges, grouping types)
  - GroupBy type safety with `IsValid()`, `IsTimeBasedGrouping()`, `String()` methods
  - Intelligent cost calculation (actual vs projected with time period conversion)
  - Provider extraction from resource types ("aws:ec2:Instance" → "aws")
  - Sorted chronological output for trend analysis
- See `internal/engine/CLAUDE.md` for detailed calculation flows

**Error Types for Cross-Provider Aggregation**:

- `ErrMixedCurrencies`: Different currencies detected (USD vs EUR)
- `ErrInvalidGroupBy`: Non-time-based grouping used for cross-provider aggregation
- `ErrEmptyResults`: Empty or nil results provided for aggregation
- `ErrInvalidDateRange`: EndDate before StartDate in cost results

### internal/pluginhost

The pluginhost package manages plugin communication via gRPC:

- Two launcher types: ProcessLauncher (TCP) and StdioLauncher (stdin/stdout)
- 10-second timeout with 100ms retry delays
- Platform-specific binary detection (Unix permissions vs Windows .exe)
- Always call `cmd.Wait()` after `Kill()` to prevent zombies
- See `internal/pluginhost/CLAUDE.md` for detailed plugin lifecycle

### internal/registry

The registry package handles plugin discovery and lifecycle:

- Scans `~/.pulumicost/plugins/<name>/<version>/` structure
- Optional `plugin.manifest.json` validation
- Graceful handling of missing directories and invalid binaries
- Platform-specific executable detection
- See `internal/registry/CLAUDE.md` for detailed discovery patterns

### internal/analyzer

The analyzer package implements the Pulumi Analyzer gRPC protocol for zero-click cost estimation during `pulumi preview`:

- **Server**: Implements `pulumirpc.AnalyzerServer` interface with `AnalyzeStack`, `Handshake`, `ConfigureStack`, `Cancel`, `GetAnalyzerInfo`, `GetPluginInfo` RPCs
- **Resource Mapping**: Converts `pulumirpc.AnalyzerResource` to `engine.ResourceDescriptor` for cost calculation
- **Diagnostics**: Generates `pulumirpc.AnalyzeDiagnostic` from cost results with ADVISORY enforcement (never blocks deployments)
- **Graceful Degradation**: Errors produce warning diagnostics, preview continues even if cost calculation fails

**Key Components**:

- `Server` - gRPC server implementing Pulumi Analyzer protocol
- `MapResource/MapResources` - Convert Pulumi resources to internal format
- `CostToDiagnostic` - Convert cost results to Pulumi diagnostics
- `StackSummaryDiagnostic` - Aggregate stack-level cost summary
- `WarningDiagnostic` - Generate warning for error conditions

**CLI Integration**:

- `pulumicost analyzer serve` - Starts gRPC server on random TCP port
- Prints ONLY port number to stdout (Pulumi handshake protocol)
- All logging goes to stderr exclusively
- Graceful shutdown on SIGINT/SIGTERM

**Protocol Requirements**:

- Port handshake: Server prints port to stdout, Pulumi engine connects
- Stack configuration: Engine sends stack/project context before analysis
- Resource analysis: Engine sends resources, analyzer returns diagnostics
- Cancellation: Engine can request analysis cancellation

**Testing**:

```bash
# Run analyzer unit tests
go test ./internal/analyzer/...

# Run integration tests
go test ./test/integration/...

# Check coverage (target: 80%, achieved: 92.7%)
go test -coverprofile=coverage.out ./internal/analyzer/...
go tool cover -func=coverage.out
```

## Common Error Types

- `ErrNoCostData`: No cost data available for a resource.
- `ErrMixedCurrencies`: Multiple currencies detected in cross-provider aggregation.
- `ErrInvalidGroupBy`: Invalid grouping type used for time-based aggregation.
- `ErrEmptyResults`: Attempted aggregation on empty results.
- `ErrInvalidDateRange`: Invalid date range (end date before start date).
- `ErrResourceValidation`: Internal resource validation failed.
- `ErrConfigCorrupted`: Configuration file is malformed.

## CodeRabbit Configuration

### Setup

The repository includes a comprehensive `.coderabbit.yaml` configuration optimized for Go development with the following key settings:

**PR Blocking Configuration:**

- `fail_commit_status: true` - Blocks PR merging on critical issues
- `request_changes_workflow: true` - Formally requests changes for issues
- `profile: assertive` - Uses stricter analysis profile

**Comment Management:**

- `auto_reply: true` - Enables automatic comment responses
- `abort_on_close: true` - Stops processing when PR is closed
- `auto_incremental_review: true` - Reviews new commits automatically

**Go-Specific Settings:**

- Custom path instructions for `**/*.go` files focusing on Go best practices
- Enhanced test review instructions for `**/*_test.go` files
- Enabled golangci-lint, gitleaks, yamllint, and markdownlint
- Docstring and unit test generation enabled

**Tool Configuration:**

- `golangci-lint: enabled: false` - Integrates with project's existing linting
- `markdownlint: enabled: true` - Validates documentation
- `gitleaks: enabled: true` - Scans for secrets
- `actionlint: enabled: true` - Validates GitHub Actions
- `semgrep: enabled: true` - Advanced security analysis

### Usage

CodeRabbit now:

1. **Blocks PRs** with critical issues by setting commit status to failed
2. **Updates comments** automatically on new commits
3. **Resolves outdated comments** when issues are fixed
4. **Provides detailed Go-specific feedback** on code quality
5. **Integrates with existing CI/CD** tools and workflows

## Active Technologies

- Go 1.25.5 + testing (stdlib), github.com/stretchr/testify, github.com/oklog/ulid/v2 (103-analyzer-e2e-tests)
- Local Pulumi state (`file://` backend), temp directories for test fixtures (103-analyzer-e2e-tests)
- Go 1.25.5 (008-analyzer-plugin)
- `~/.pulumicost/config.yaml` for plugin configuration (existing infrastructure) (008-analyzer-plugin)
- Go 1.25.5 + testing (stdlib), github.com/stretchr/testify (001-engine-test-coverage)
- Go 1.25.5 + zerolog v1.34.0, cobra v1.10.1, yaml.v3 (007-integrate-logging)
- File system (`~/.pulumicost/config.yaml`, log files) (007-integrate-logging)

## Recent Changes

- 001-engine-test-coverage: Added Go 1.25.5 + testing (stdlib), github.com/stretchr/testify
- 007-integrate-logging: Added zerolog v1.34.0 logging integration across all components
- 008-test-infra-hardening: Added comprehensive test infrastructure hardening

## Session Analysis - Recommended Updates

Based on recent development sessions, consider adding:

### Go Version Management

- **Version Consistency**: When updating Go versions, update both `go.mod` and ALL markdown files simultaneously
- **Search Pattern**: Use `grep "Go.*1\." --include="*.md"` to find all version references in documentation
- **Files to Check**: go.mod, all .md files in docs/, specs/, examples/, and root-level documentation
- **Docker Images**: Update Docker base images (e.g., `golang:1.24` → `golang:1.25.5`) in documentation examples

### Systematic Version Updates

- **Process**: 1) Update go.mod first, 2) Find all references with grep, 3) Update each file systematically, 4) Verify with final grep search
- **Common Patterns**: Update both specific versions (1.24.10 → 1.25.5) and minimum requirements (Go 1.24+ → Go 1.25.5+)
- **CI Workflows**: Update GitHub Actions go-version parameters in documentation examples

This ensures complete version consistency across the entire codebase and documentation.

## AI Agent File Maintenance

This file (CLAUDE.md) provides guidance for Claude Code and other AI assistants. To maintain its effectiveness:

### Update Requirements

- **Review regularly** when significant codebase changes occur
- **Update version information** immediately when Go versions change
- **Document new patterns** and conventions as they emerge
- **Include new technologies** and dependencies as they are added
- **Update build/test commands** when processes change
- **Maintain architecture documentation** as the system evolves

### When to Update

- New major features are implemented
- Build or testing processes change
- New dependencies are added
- Coding standards evolve
- Project structure changes significantly
- New tools or workflows are introduced

### Integration with GitHub Copilot

- This file is automatically read by GitHub Copilot via `.github/instructions/ai-agent-files.instructions.md`
- Use it as the authoritative source for development practices
- Reference these instructions when working with AI assistants
- Keep instructions current to ensure consistent AI assistance

### Maintenance Checklist

- [ ] Go version information is current
- [ ] Build commands work as documented
- [ ] Test commands produce expected results
- [ ] Architecture documentation reflects current state
- [ ] Dependencies are accurately listed
- [ ] Security practices are up to date
- [ ] Performance guidelines remain relevant
