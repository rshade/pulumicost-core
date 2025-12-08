# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## CRITICAL INSTRUCTIONS

**DO NOT RUN `git commit`** - This is explicitly forbidden. You may use `git add`, `git status`, `git diff`, and `git log`, but you are NOT allowed to run commit commands. The user will commit manually.

## Project Overview

PulumiCost Core is a CLI tool and plugin host system for calculating cloud infrastructure costs from Pulumi infrastructure definitions. It provides both projected cost estimates and actual historical cost analysis through a plugin-based architecture.

## Build Commands

- `make build` - Build the pulumicost binary to bin/pulumicost
- `make test` - Run all tests
- `make lint` - Run golangci-lint (requires installation)
- `make run` - Build and run with --help
- `make dev` - Build and run without arguments
- `make clean` - Remove build artifacts

**IMPORTANT** Always run `make lint` and `make test` before claiming success. See `.specify/memory/constitution.md` for quality gate requirements.

## Go Version Information

**Project Go Version**: 1.24.10

### Dependencies

- `github.com/spf13/cobra` - CLI framework
- `google.golang.org/grpc` - Plugin communication
- `gopkg.in/yaml.v3` - YAML spec parsing
- `github.com/rshade/pulumicost-spec` - Protocol definitions (via replace directive to ../pulumicost-spec)
- archive/tar
- archive/zip
- compress/gzip
- net/http
- github.com/spf13/cobra
- github.com/rs/zerolog
- github.com/oklog/ulid/v2
- gopkg.in/yaml.v3
- google.golang.org/grpc v1.77.0
- File system
  - ~/.pulumicost/plugins/
  - ~/.pulumicost/config.yaml

### Version Verification Protocol

**CRITICAL**: Before claiming any Go version "doesn't exist" or suggesting version changes:

1. **ALWAYS verify on <https://go.dev/dl/>** using WebFetch
2. Check `go.mod` for the project's actual Go version
3. Trust the versions specified in the repository
4. Never assume based on training data - Go releases frequently after knowledge cutoffs

Do NOT suggest version downgrades without explicit verification from go.dev.

## Documentation Commands

- `make docs-lint` - Lint documentation markdown files
- `make docs-build` - Build documentation site with Jekyll
- `make docs-serve` - Serve documentation locally (http://localhost:4000/pulumicost-core/)
- `make docs-validate` - Validate documentation structure and completeness

### GitHub Actions Workflow Best Practices

**1. npm Cache Configuration:**

- Only use `cache: 'npm'` if `package-lock.json` exists
- For dynamic npm installs without lockfile, omit the cache parameter
- Example fix:
  ```yaml
  - name: Setup Node.js
    uses: actions/setup-node@v6
    with:
      node-version: '24'
      # cache: 'npm'  # Remove if no package-lock.json
  ```

**2. Job Naming Conflicts:**

- Avoid reserved keywords like `summary`, `status`, `output`
- Use descriptive prefixes: `validation-summary`, `build-status`, etc.
- Proper indentation is critical for YAML:
  ```yaml
  validation-summary: # Good: specific and unique
    runs-on: ubuntu-latest
    if: always() # Proper indentation under job
    needs: [build]
  ```

**3. Testing Jekyll Builds Locally:**

- Always test Jekyll builds before committing changes
- Use Playwright MCP to verify the deployed site visually
- Check browser console for 404 errors or missing resources
- Workflow: Local build → Deploy → Playwright test → Commit

### Jekyll + GitHub Pages Testing Workflow

**Complete testing workflow using Playwright MCP:**

```bash
# 1. Test local Jekyll build first (if possible)
# make docs-serve

# 2. After deployment, test live site
mcp__playwright__browser_navigate(url: "https://rshade.github.io/pulumicost-core/")

# 3. Take screenshot to verify visual appearance
mcp__playwright__browser_take_screenshot(filename: "docs-check.png", fullPage: true)

# 4. Check network requests for 404s or missing CSS
mcp__playwright__browser_network_requests()
# Look for: HTTP 200 on style.css, fonts, and JavaScript files

# 5. Verify no duplicate content or layout issues in snapshot
mcp__playwright__browser_snapshot()
```

### Documentation Styling Best Practices

**Custom SCSS Structure:**

```scss
---
---

@import '{{ site.theme }}'; // Import base theme first

/* Then add custom overrides */
table {
  /* Enhanced table styling */
}
.wrapper {
  /* Layout adjustments */
}
```

**Common Styling Improvements:**

- Table borders, padding, and alternating row colors
- Wider content area (1200px max-width vs default 860px)
- Better link colors and hover states
- GitHub-style code blocks with proper syntax highlighting
- Responsive breakpoints for mobile devices

### Key Learnings from GitHub Pages Issues

1. **Always create index.md explicitly** - Don't rely on README.md conversion
2. **Test plugins before using template tags** - Jekyll fails silently in builds but errors in Actions
3. **Use Playwright MCP for visual verification** - Screenshot + network requests catch most issues
4. **Avoid duplicate titles** - Check both layout and content files
5. **Test locally when possible** - Catches issues before CI/CD failures
6. **Monitor GitHub Actions logs** - Liquid syntax errors show exact file and line number

## Documentation Architecture

### Location

All documentation is in the `docs/` directory with GitHub Pages deployed from that folder.

### Key Files

- **docs/README.md** - Documentation home page with navigation
- **docs/plan.md** - Complete documentation architecture and strategy
- **docs/llms.txt** - Machine-readable index for LLM/AI tools
- **docs/\_config.yml** - Jekyll configuration

### Directory Structure

```
docs/
├── guides/                # Audience-specific guides (User, Engineer, Architect, CEO)
├── getting-started/       # Quick onboarding and examples
├── architecture/          # System design and diagrams
├── plugins/              # Plugin documentation and development
├── reference/            # CLI, API, and configuration reference
├── deployment/           # Installation, configuration, and operations
└── support/              # FAQ, troubleshooting, contributing, support
```

### Audience-Specific Guides

- **guides/user-guide.md** - For end users: "How do I use this?"
- **guides/developer-guide.md** - For engineers: "How do I extend this?"
- **guides/architect-guide.md** - For architects: "How is this designed?"
- **guides/business-value.md** - For CEO/product: "What problem does this solve?"

### Plugin Documentation

- **plugins/plugin-development.md** - How to build a PulumiCost plugin
- **plugins/plugin-sdk.md** - Plugin SDK reference
- **plugins/vantage/** - Vantage plugin example (IN PROGRESS)
- **plugins/kubecost/** - Kubecost plugin docs (PLANNED)
- **plugins/flexera/** - Flexera plugin docs (FUTURE)
- **plugins/cloudability/** - Cloudability plugin docs (FUTURE)

### Documentation Standards

- Follow Google style guide for markdown
- All code examples must be tested
- Keep llms.txt updated (updated automatically by GitHub Actions)
- Run `make docs-lint` before committing documentation changes
- Use frontmatter YAML with `title`, `description`, and `layout` fields

### GitHub Actions for Docs

- **docs-build-deploy.yml** - Builds and deploys docs to GitHub Pages on main branch
- **docs-validate.yml** - Validates markdown, links, and structure on every commit
- Automated linting prevents documentation drift
- Link checking catches broken documentation references

## Architecture

### Core Components

1. **CLI Layer** (`internal/cli/`) - Cobra-based command interface with subcommands:
   - `cost projected` - Calculate projected costs from Pulumi preview JSON
   - `cost actual` - Fetch actual historical costs with time ranges
   - `plugin list` - List installed plugins
   - `plugin validate` - Validate plugin installations

2. **Engine** (`internal/engine/`) - Core cost calculation logic:
   - Orchestrates between plugins and local pricing specs
   - Handles resource mapping and cost aggregation
   - Supports multiple output formats (table, JSON, NDJSON)
   - **Actual Cost Pipeline**: Advanced cost querying with time ranges, filtering, and grouping
     - `GetActualCostWithOptions()` - Flexible actual cost queries with filtering
     - Tag-based filtering using `tag:key=value` syntax
     - Grouping by resource, type, provider, or date dimensions
     - Daily and monthly cost aggregation

3. **Plugin Host System** (`internal/pluginhost/`) - gRPC plugin management:
   - `Client` - Wraps plugin gRPC connections
   - `ProcessLauncher` - Launches plugins as TCP processes
   - `StdioLauncher` - Alternative stdio-based plugin communication

4. **Registry** (`internal/registry/`) - Plugin discovery and lifecycle:
   - Scans `~/.pulumicost/plugins/<name>/<version>/` for binaries
   - Manages plugin manifests and metadata

5. **Ingestion** (`internal/ingest/`) - Pulumi plan parsing:
   - Converts `pulumi preview --json` output to resource descriptors
   - Extracts provider and resource type information

6. **Spec System** (`internal/spec/`) - Local pricing specification:
   - YAML-based pricing specs in `~/.pulumicost/specs/`
   - Fallback when plugins don't provide pricing

### Plugin Protocol

Plugins communicate via gRPC using protocol buffers defined in the `pulumicost-spec` repository. Current implementation uses mock protobuf definitions (`internal/proto/mock.go`) until the spec repository is fully implemented.

Key plugin methods:

- `Name()` - Plugin identification
- `GetProjectedCost()` - Calculate estimated costs for resources
- `GetActualCost()` - Retrieve historical costs from cloud APIs

## Development Workflow

1. **Resource Flow**: Pulumi JSON → Resource Descriptors → Plugin Queries → Cost Results → Output Rendering

2. **Plugin Discovery**: Registry scans plugin directories → Launches processes → Establishes gRPC connections → Makes API calls

3. **Cost Calculation**: Try plugins first → Fallback to local specs → Aggregate results → Render output

## Key Files

- `cmd/pulumicost/main.go` - CLI entry point
- `internal/engine/engine.go` - Core orchestration logic
- `internal/pluginhost/host.go` - Plugin client management
- `internal/ingest/pulumi_plan.go` - Pulumi plan parsing
- `examples/plans/aws-simple-plan.json` - Sample Pulumi plan for testing
- `examples/specs/aws-ec2-t3-micro.yaml` - Sample pricing specification

## E2E Testing Implementation Details

**Crucial Learnings for E2E Tests:**

1.  **Pulumi Plan JSON Structure**: The `pulumi preview --json` output nests resource details (including `inputs` and `type`) under a `newState` object for operations like `create`, `update`, and `same`. The ingestion logic (`internal/ingest/pulumi_plan.go`) **MUST** inspect `newState` to correctly extract these fields. Failing to do so results in empty `Inputs`, which causes property extraction to fail.

2.  **Property Extraction Dependencies**: The Core (`internal/proto/adapter.go`) relies on the `Inputs` map being populated to extract `SKU` (from `instanceType`, `type`, etc.) and `Region` (from `availabilityZone`, `region`). If ingestion fails to populate `Inputs`, these fields will be empty, leading to `InvalidArgument` errors from strict plugins.

3.  **Plugin Resource Type Compatibility**: Plugins may have strict validation on `ResourceType`. Pulumi typically provides types like `aws:ec2/instance:Instance` (Type Token), while some plugins (or pricing APIs) might expect `aws:ec2:Instance` or just `ec2`.
    - **Current Strategy**: Plugins should handle the standard Pulumi Type Token format (`pkg:module:type`).
    - **Patching**: If a plugin rejects a valid type, it likely needs a patch to normalize or map the type string to its internal service identifier (e.g., `detectService` helper).

**Local Plugin Development Workflow:**

To debug or fix plugin issues during Core development:

1.  Clone the plugin repository (e.g., `pulumicost-plugin-aws-public`).
2.  Modify the plugin code (e.g., add logging, fix type mapping).
3.  Build the plugin: `make build-region REGION=us-east-1`.
4.  Install locally: Overwrite the binary in `~/.pulumicost/plugins/<plugin>/<version>/...`.
5.  Run Core E2E tests to verify the fix.

## Project Management

### Cross-Repository Project

- **GitHub Project**: https://github.com/users/rshade/projects/3
- **Scope**: Manages issues across three repositories:
  - `pulumicost-core` (this repository) - CLI tool and plugin host
  - `pulumicost-spec` - Protocol buffer definitions and specifications
  - `pulumicost-plugin` - Plugin implementations and SDK

### Product Manager Responsibilities

- Keep issues synchronized across all three repositories
- Manage cross-repo dependencies and coordination
- Track feature development across the entire ecosystem
- Ensure consistent issue labeling and milestone alignment

### GitHub CLI Commands for Project Management

```bash
# View project overview
gh project view 3 --owner rshade

# Add issues to project (when creating cross-repo issues)
gh issue edit ISSUE --repo OWNER/REPO --add-project "PulumiCost Development"
```

```go
var (
    ErrNoCostData       = errors.New("no cost data available")
    ErrMixedCurrencies  = errors.New("mixed currencies not supported in cross-provider aggregation")
    ErrInvalidGroupBy   = errors.New("invalid groupBy type for cross-provider aggregation")
    ErrEmptyResults     = errors.New("empty results provided for aggregation")
    ErrInvalidDateRange = errors.New("invalid date range: end date must be after start date")
)
```

## CI/CD Pipeline

### Overview

Complete CI/CD pipeline setup with GitHub Actions for automated testing, building, and release management.

### CI Pipeline (.github/workflows/ci.yml)

Triggered on pull requests and pushes to main branch:

**Test Job:**

- Go 1.24.10 setup with caching
- Unit tests with race detection and coverage reporting
- Coverage threshold check (minimum 20%)
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

### Commands for Local Development

```bash
# Basic development workflow
make build       # Build binary
make test        # Run all unit tests
make lint        # Code linting
make validate    # Go vet and formatting checks
make dev         # Build and run binary without args
make run         # Build and run with --help
make clean       # Remove build artifacts

# Single package testing
go test -v ./internal/cli/...           # Test only CLI package
go test -v ./internal/engine/...        # Test only engine package
go test -run TestSpecificFunction ./... # Run specific test function

# Coverage analysis
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out  # View in browser
go tool cover -func=coverage.out | grep total  # Check total coverage

# Development testing with examples
./bin/pulumicost cost projected --pulumi-json examples/plans/aws-simple-plan.json
./bin/pulumicost cost actual --start-date 2024-01-01 --end-date 2024-01-31
./bin/pulumicost cost actual --group-by resource --filter "tag:env=prod"
./bin/pulumicost cost actual --output json --start-date 2024-01-01T00:00:00Z
./bin/pulumicost plugin list
./bin/pulumicost plugin validate
```

## CI/CD Implementation Learnings

### golangci-lint Configuration

- **Issue**: Original .golangci.yml was overly complex (449 lines) with deprecated/invalid linters
- **Solution**: Simplified to essential linters (errcheck, govet, staticcheck, gosec, revive, unused, ineffassign)
- **Key learnings**:
  - `typecheck` and `gofmt` are not valid linters in newer golangci-lint versions
  - `goimports` is a formatter, not a linter in v2+
  - Use `--allow-parallel-runners` flag in Makefile to prevent conflicts
  - Project-specific configuration should match codebase maturity

### Coverage Thresholds

- **Current State**: 24.2% overall coverage, 67.2% in CLI package
- **Threshold Set**: 20% (adjusted from initial 80% for realistic expectations)
- **Strategy**: Start conservative, increase as project matures and more tests added
- **Command**: `go tool cover -func=coverage.out | grep total` for threshold checking

### Security Scanning Integration

- **gosec**: Already included in golangci-lint configuration
- **govulncheck**: Separate step for dependency vulnerability scanning
- **Common Issues**: File permissions (G306), potential file inclusion (G304), subprocess usage (G204)
- **Test exclusions**: Security issues in test files are often acceptable and should be excluded

### Cross-Platform Build Patterns

- **Binary naming**: `pulumicost-v{version}-{os}-{arch}` with `.exe` for Windows
- **Architecture matrix**: Linux/macOS (amd64, arm64), Windows (amd64 only)
- **LDFLAGS**: Proper shell escaping needed for version embedding
- **Build verification**: All platforms should compile successfully in CI

### GitHub Actions Best Practices

- **Deprecated actions**: Avoid `actions/create-release@v1`, use `softprops/action-gh-release@v2`
- **Artifact management**: Use `actions/upload-artifact@v4` with proper naming
- **HEREDOC usage**: Essential for multiline strings in workflow files
- **Matrix excludes**: Use to skip unsupported combinations (e.g., Windows ARM64)

### Release Automation Patterns

- **Tag detection**: `${GITHUB_REF#refs/tags/}` pattern for version extraction
- **Changelog generation**: Git history works well with `git log ${PREV_TAG}..${CURRENT_TAG}`
- **Checksums**: SHA256 for all binaries with verification instructions
- **Pre-release detection**: Use `contains(steps.version.outputs.tag, '-')` for beta/alpha tags

### Dependency Management Strategy

- **Dual approach**: Renovate + Dependabot with different schedules (avoid conflicts)
- **Rate limiting**: Prevent PR spam with `prConcurrentLimit` and `prHourlyLimit`
- **Semantic commits**: Enable conventional commit format for changelog automation
- **Security alerts**: Immediate notification for vulnerability PRs

### Common Linting Issues Found

- **errcheck (23 issues)**: Unchecked error returns, especially in defer statements and fmt functions
- **gosec (8 issues)**: File permissions, subprocess usage, file inclusion patterns
- **revive (50 issues)**: Missing package comments, exported type documentation
- **staticcheck (4 issues)**: Deprecated gRPC functions (grpc.DialContext, grpc.WithBlock)

### Testing Strategy Insights

- **Race detection**: Use `-race` flag for concurrent code testing
- **Coverage modes**: `atomic` mode recommended for accurate concurrent coverage
- **Integration testing**: Include CLI workflow testing in CI pipeline
- **Test exclusions**: Some linting rules should be relaxed for test files

### Project-Specific Notes

- **Test distribution**: CLI package well-tested (67.2%), other packages need attention
- **Architecture**: Plugin system will need careful testing as it develops
- **Proto integration**: Real protobuf definitions working, mock phase complete
- **Build system**: Well-structured with proper version/commit embedding

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

### Comprehensive Testing Framework

The project includes a comprehensive testing framework organized in the `/test` directory:

```
/test
├── unit/              # Unit tests by package (engine, config, spec)
├── integration/       # Cross-component tests (plugin communication, e2e)
├── fixtures/          # Test data (plans, specs, configs, responses)
├── mocks/             # Mock implementations (plugin server)
└── benchmarks/        # Performance tests
```

**Test Categories:**

- **Unit Tests** (80% coverage target): Individual component logic
- **Integration Tests**: Plugin communication, CLI workflows
- **End-to-End Tests**: Complete CLI workflows with real binaries
- **Performance Tests**: Benchmarks for cost calculations
- **Mock Tests**: Configurable plugin server for testing

**Running Tests:**

```bash
# All tests (including existing + new framework)
make test

# New testing framework only
go test ./test/...

# Specific categories
go test ./test/unit/...           # Unit tests
go test ./test/integration/...     # Integration tests
go test ./test/benchmarks/...      # Performance benchmarks
go test ./test/mocks/plugin/...    # Mock plugin tests

# With coverage
go test -coverprofile=coverage.out ./test/...
go tool cover -html=coverage.out

# With race detection
go test -race ./test/...
```

### E2E Testing (Real AWS Infrastructure)

The E2E test framework validates cost calculations against real AWS infrastructure using the Pulumi Automation API.

**Location**: `test/e2e/` (separate Go module)

**Running E2E Tests:**

```bash
# Set required environment variables
eval "$(aws configure export-credentials --format env)"
export PATH="$HOME/.pulumi/bin:$PATH"
export PULUMI_CONFIG_PASSPHRASE="e2e-test-passphrase"
export AWS_REGION=us-east-1

# Run all E2E tests
make test-e2e

# Run specific E2E test
make test-e2e TEST_ARGS="-run TestProjectedCost_EC2"

# Direct go test invocation
go test -v -tags e2e -timeout 60m ./...
```

**Prerequisites:**

- AWS credentials configured (via AWS CLI or environment variables)
- Pulumi CLI installed (`~/.pulumi/bin/pulumi`)
- Go 1.24.10+
- pulumicost binary built (`make build`)

**Environment Variables:**

| Variable                   | Description                               | Default     |
| -------------------------- | ----------------------------------------- | ----------- |
| `AWS_ACCESS_KEY_ID`        | AWS access key                            | Required    |
| `AWS_SECRET_ACCESS_KEY`    | AWS secret key                            | Required    |
| `AWS_SESSION_TOKEN`        | AWS session token (if using SSO/MFA)      | Optional    |
| `AWS_REGION`               | AWS region for tests                      | `us-east-1` |
| `E2E_REGION`               | Override AWS region                       | `us-east-1` |
| `E2E_TIMEOUT_MINS`         | Maximum test duration                     | `60`        |
| `PULUMI_CONFIG_PASSPHRASE` | Passphrase for local state encryption     | Required    |
| `PATH`                     | Must include Pulumi CLI (`~/.pulumi/bin`) | Required    |

**Test Categories:**

- `TestProjectedCost_EC2` - Validates EC2 projected cost against AWS pricing
- `TestProjectedCost_EBS` - Validates EBS projected cost against AWS pricing
- `TestActualCost_Runtime` - Validates actual cost calculation over time
- `TestCleanupVerification` - Verifies resource cleanup
- `TestUnsupportedResourceTypes` - Validates graceful handling of unsupported resources
- `TestCLIExecution` - Black-box CLI binary testing

**Safety Features:**

- ULID-prefixed stack names for test isolation
- Automatic cleanup on test completion
- Signal handling for cleanup on interrupt (Ctrl+C)
- 60-minute timeout with retry logic for AWS operations

**CRITICAL: No Simulation/Stubbing in E2E Tests:**

E2E tests MUST call the actual pulumicost CLI binary and validate real cost calculations. Never:

- Simulate cost values (e.g., `calculatedCost := expectedCost * 1.01`)
- Stub CLI execution with hardcoded responses
- Skip real infrastructure deployment
- Use mock pricing data in E2E tests

The purpose of E2E tests is to validate the complete system works correctly. Tests that simulate behavior defeat this purpose. If you need faster feedback during development, use unit tests or integration tests instead.

**Test Fixtures Available:**

- AWS, Azure, GCP Pulumi plans (`test/fixtures/plans/`)
- Pricing specifications (`test/fixtures/specs/`)
- Mock API responses (`test/fixtures/responses/`)
- Configuration examples (`test/fixtures/configs/`)

**Mock Plugin Server:**
The testing framework includes a configurable gRPC plugin server for testing plugin communication:

```go
mockPlugin := plugin.NewMockPlugin("test-plugin")
mockPlugin.SetProjectedCostResponse("aws_instance", customResponse)
mockPlugin.SetError("GetActualCost", simulatedError)
```

### Manual Testing Commands

Use the provided example files for manual testing:

```bash
# Projected cost calculation
./bin/pulumicost cost projected --pulumi-json examples/plans/aws-simple-plan.json

# Actual cost queries with time ranges
./bin/pulumicost cost actual --start-date 2024-01-01 --end-date 2024-01-31

# Actual cost with filtering and grouping
./bin/pulumicost cost actual --group-by resource --filter "tag:env=prod" --output table
./bin/pulumicost cost actual --group-by daily --start-date 2024-01-01T00:00:00Z --end-date 2024-01-31T23:59:59Z

# Cross-provider aggregation (NEW)
./bin/pulumicost cost actual --group-by daily --start-date 2024-01-01 --end-date 2024-01-31 --output json
./bin/pulumicost cost actual --group-by monthly --start-date 2024-01-01 --end-date 2024-12-31 --filter "tag:env=prod"
./bin/pulumicost cost actual --group-by daily --output table  # Shows cross-provider daily breakdown

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

## Logging (Zerolog)

PulumiCost uses zerolog for structured logging with distributed tracing support.

### Enabling Debug Output

```bash
# CLI flag
pulumicost cost projected --debug --pulumi-json plan.json

# Environment variable
export PULUMICOST_LOG_LEVEL=debug
export PULUMICOST_LOG_FORMAT=json    # json or console
export PULUMICOST_TRACE_ID=external-trace-123  # inject external trace ID
```

### Configuration Precedence

1. CLI flags (`--debug`)
2. Environment variables (`PULUMICOST_LOG_LEVEL`)
3. Config file (`~/.pulumicost/config.yaml`)
4. Default (info level, console format)

### Logging Patterns for Developers

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

### Standard Log Fields

| Field         | Purpose             | Example                      |
| ------------- | ------------------- | ---------------------------- |
| `trace_id`    | Request correlation | "01HQ7X2J3K4M5N6P7Q8R9S0T1U" |
| `component`   | Package identifier  | "cli", "engine", "registry"  |
| `operation`   | Current operation   | "get_projected_cost"         |
| `duration_ms` | Operation timing    | 245                          |

### Log Levels

- **TRACE**: Property extraction, detailed calculations
- **DEBUG**: Function entry/exit, retries, intermediate values
- **INFO**: High-level operations (command start/end)
- **WARN**: Recoverable issues (fallbacks, deprecations)
- **ERROR**: Failures needing attention

### Trace ID Management

```go
// Generate trace ID at entry point (usually in CLI PersistentPreRunE)
traceID := logging.GetOrGenerateTraceID(ctx)
ctx = logging.ContextWithTraceID(ctx, traceID)

// TracingHook automatically injects trace_id when using .Ctx(ctx)
```

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
- **Cross-Provider Aggregation Features** (NEW):
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
- Go 1.25.4 (008-analyzer-plugin)
- `~/.pulumicost/config.yaml` for plugin configuration (existing infrastructure) (008-analyzer-plugin)

- Go 1.24.10 + testing (stdlib), github.com/stretchr/testify (001-engine-test-coverage)
- Go 1.24.10 + zerolog v1.34.0, cobra v1.10.1, yaml.v3 (007-integrate-logging)
- File system (`~/.pulumicost/config.yaml`, log files) (007-integrate-logging)

## Recent Changes

- 001-engine-test-coverage: Added Go 1.24.10 + testing (stdlib), github.com/stretchr/testify
- 007-integrate-logging: Added zerolog v1.34.0 logging integration across all components
