# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

PulumiCost Core is a CLI tool and plugin host system for calculating cloud infrastructure costs from Pulumi infrastructure definitions. It provides both projected cost estimates and actual historical cost analysis through a plugin-based architecture.

## Build Commands

- `make build` - Build the pulumicost binary to bin/pulumicost
- `make test` - Run all tests
- `make lint` - Run golangci-lint (requires installation)
- `make run` - Build and run with --help
- `make dev` - Build and run without arguments
- `make clean` - Remove build artifacts

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

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `google.golang.org/grpc` - Plugin communication
- `gopkg.in/yaml.v3` - YAML spec parsing
- `github.com/rshade/pulumicost-spec` - Protocol definitions (via replace directive to ../pulumicost-spec)

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

### Dependency & Milestone Tracker

**Milestones Created:**
- `2025-Q1 - Spec v0.1.0 MVP` (Due: Aug 20, 2025) - Protocol definitions
- `2025-Q1 - Core v0.1.0 MVP` (Due: Sep 6, 2025) - CLI and plugin host
- `2025-Q1 - Kubecost Plugin v0.1.0 MVP` (Due: Sep 6, 2025) - Plugin implementation

**Critical Path Dependencies:**
- SPEC-1 → CORE-3 (Plugin Host Bootstrap)  
- SPEC-1 → PLUG-KC-1 → CORE-5 (Actual Cost Pipeline)
- SPEC-2 → PLUG-KC-3 → CORE-4 (Projected Cost Pipeline)

**Week 1 (Parallel Work):**
- Core: CLI Skeleton (#3), Pulumi JSON Ingest (#4)
- Spec: Freeze proto & schema
- Plugin: Stub API client, manifest

**Week 2 (Dependencies unlock):**
- Core: Plugin Host Bootstrap (#2)
- Plugin: Kubecost API Client + Supports()

**Week 3 (Feature completion):**
- Core: Projected Cost Pipeline (#5), Actual Cost Pipeline (#6)
- Plugin: Projected Cost Logic

**Week 4 (Integration):**
- End-to-end examples and MVP stabilization

## Protocol Integration Status

### ✅ SPEC-1 Completed - Proto Integration
- **Status**: costsource.proto v0.1.0 is frozen and integrated
- **Location**: `/mnt/c/GitHub/go/src/github.com/rshade/pulumicost-spec/proto/pulumicost/v1/costsource.proto`
- **Generated SDK**: Available at `github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1`
- **Integration**: Core now uses real proto definitions via `internal/proto/adapter.go`

### Proto Integration Details
- Removed mock proto implementation (`internal/proto/mock.go`) 
- Created adapter layer (`internal/proto/adapter.go`) to bridge engine expectations with real proto types
- Updated dependencies: gRPC v1.74.2, protobuf v1.36.7
- Core engine successfully uses `CostSourceServiceClient` from pulumicost-spec

### Verified Working Commands
```bash
# Basic CLI functionality verified
./bin/pulumicost --help
./bin/pulumicost cost projected --help

# Projected cost calculation (shows resources but "none" adapter since no plugins)
./bin/pulumicost cost projected --pulumi-json examples/plans/aws-simple-plan.json

# Plugin management (correctly reports no plugins installed)
./bin/pulumicost plugin list
./bin/pulumicost plugin validate
```

### ✅ CORE-5 Completed - Actual Cost Pipeline
- **Status**: Comprehensive actual cost pipeline implemented with advanced features
- **Implementation**: PR #36 - Added cost aggregation, filtering, and grouping capabilities
- **Key Features**:
  - Time range queries with flexible date parsing ("2006-01-02", RFC3339)
  - Resource filtering by tags/metadata with `tag:key=value` syntax
  - Cost aggregation with daily/monthly breakdowns
  - Grouping by resource, type, provider, or date dimensions
  - Multiple output formats (table, JSON, NDJSON)
  - Comprehensive cost reporting with actual vs projected comparisons

### Architecture Changes
- **New Engine Method**: `GetActualCostWithOptions()` with flexible querying
- **Enhanced Data Structures**: `ActualCostRequest` with advanced filtering options
- **Tag Matching**: `matchesTags()` helper for resource filtering
- **Cost Aggregation**: Daily/monthly cost breakdown logic
- **Output Enhancement**: Rich table formatting for actual cost results

### Next Steps Unlocked
With SPEC-1 and CORE-5 complete, the following work can now proceed:
- **CORE-3**: Plugin Host Bootstrap (depends on SPEC-1) 
- **PLUG-KC-1**: Kubecost API Client (depends on SPEC-1)
- Integration testing with actual plugins

## CI/CD Pipeline

### Overview
Complete CI/CD pipeline setup with GitHub Actions for automated testing, building, and release management.

### CI Pipeline (.github/workflows/ci.yml)
Triggered on pull requests and pushes to main branch:

**Test Job:**
- Go 1.24.5 setup with caching
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
Triggered on version tags (v*):

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

### Dependency Management

**Renovate Configuration (.github/renovate.json):**
- Weekly updates on Monday mornings (UTC)
- Grouped updates by dependency type
- Semantic commit messages with conventional format
- Security vulnerability alerts with priority labeling
- Rate limiting to prevent spam

**Dependabot Configuration (.github/dependabot.yml):**
- Go modules and GitHub Actions monitoring
- Weekly schedule with proper time zone handling
- Automatic assignee and reviewer assignment
- Conventional commit message formatting

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

### Release Process

1. Create and push version tag: `git tag v1.0.0 && git push origin v1.0.0`
2. GitHub Actions automatically builds multi-platform binaries
3. Release created with changelog and downloadable assets
4. Checksums provided for verification

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

Use the provided example files:
```bash
# Projected cost calculation
./bin/pulumicost cost projected --pulumi-json examples/plans/aws-simple-plan.json

# Actual cost queries with time ranges
./bin/pulumicost cost actual --start-date 2024-01-01 --end-date 2024-01-31

# Actual cost with filtering and grouping
./bin/pulumicost cost actual --group-by resource --filter "tag:env=prod" --output table
./bin/pulumicost cost actual --group-by date --start-date 2024-01-01T00:00:00Z --end-date 2024-01-31T23:59:59Z

# Plugin management
./bin/pulumicost plugin list
./bin/pulumicost plugin validate
```
- Never complete a project without running `make lint`.

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
- See `internal/engine/CLAUDE.md` for detailed calculation flows

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