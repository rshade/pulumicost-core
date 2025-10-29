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

## Documentation Commands

- `make docs-lint` - Lint documentation markdown files
- `make docs-build` - Build documentation site with Jekyll
- `make docs-serve` - Serve documentation locally (http://localhost:4000/pulumicost-core/)
- `make docs-validate` - Validate documentation structure and completeness

## Playwright MCP Integration

### Overview
The project is configured with Playwright MCP for automated browser testing and documentation validation. The configuration is in `.mcp.json` and uses chromium in headless, isolated mode.

### Configuration
Located in `.mcp.json`:
```json
{
  "playwright": {
    "command": "npx",
    "args": ["-y", "@playwright/mcp@latest", "--browser", "chromium", "--headless", "--isolated"],
    "env": {
      "PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD": "0"
    }
  }
}
```

### Key Features
- **Browser**: Chromium (automatically installed via npx)
- **Mode**: Headless and isolated (no persistent profile)
- **Auto-installation**: PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=0 ensures chromium installs on first use

### Initial Setup
If you encounter chromium installation issues, manually install:
```bash
npx playwright install chromium
```

### Common Use Cases

**1. Documentation Site Validation**
```bash
# Navigate to local docs and take screenshot
mcp__playwright__browser_navigate(url: "http://localhost:4000/pulumicost-core/")
mcp__playwright__browser_snapshot()
mcp__playwright__browser_take_screenshot(filename: "docs-homepage.png")
```

**2. GitHub Pages Validation**
```bash
# Check deployed documentation
mcp__playwright__browser_navigate(url: "https://rshade.github.io/pulumicost-core/")
mcp__playwright__browser_snapshot()
```

**3. Link Checking and Navigation Testing**
```bash
# Test documentation navigation
mcp__playwright__browser_navigate(url: "https://rshade.github.io/pulumicost-core/")
mcp__playwright__browser_click(element: "User Guide link", ref: "a[href='/guides/user-guide']")
mcp__playwright__browser_snapshot()
```

**4. Form Testing (Future Plugin Integration)**
```bash
# Test interactive documentation features
mcp__playwright__browser_fill_form(fields: [...])
mcp__playwright__browser_click(element: "Submit button", ref: "button[type='submit']")
```

**5. Network Request Monitoring**
```bash
# Monitor API calls in documentation examples
mcp__playwright__browser_navigate(url: "https://rshade.github.io/pulumicost-core/examples/")
mcp__playwright__browser_network_requests()
```

### Troubleshooting

**Issue: "Chromium distribution 'chrome' is not found"**
- Solution: Run `npx playwright install chromium`
- Root cause: Chromium not installed or wrong browser channel specified

**Issue: Hanging on launch**
- Solution: Ensure `--headless` and `--isolated` flags are set in `.mcp.json`
- Check: `PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=0` is set to allow installation

**Issue: Permission denied in WSL**
- Solution: Add `--no-sandbox --disable-setuid-sandbox` to launchOptions if needed
- Note: Already configured in current setup

### Best Practices
1. **Always use snapshots first**: `browser_snapshot()` is faster than screenshots and provides better context
2. **Screenshots for visual verification**: Use `browser_take_screenshot()` for visual regression testing
3. **Network monitoring**: Use `browser_network_requests()` to validate API calls in documentation examples
4. **Cleanup**: Browser instances are automatically cleaned up due to `--isolated` flag
5. **Documentation validation workflow**:
   - Start local docs: `make docs-serve`
   - Navigate: `browser_navigate(url: "http://localhost:4000/pulumicost-core/")`
   - Validate: `browser_snapshot()` and verify content
   - Test links: Click through navigation and verify no 404s

### Integration with CI/CD
For future automated testing, Playwright can be integrated into GitHub Actions:
```yaml
- name: Install Playwright
  run: npx playwright install chromium
- name: Test Documentation
  run: npx playwright test
```

### Actual Usage Examples
Real-world Playwright MCP usage for testing GitHub Pages:
```bash
# Navigate and verify page loads
mcp__playwright__browser_navigate(url: "https://rshade.github.io/pulumicost-core/")

# Take full page screenshot for visual verification
mcp__playwright__browser_take_screenshot(filename: "site-screenshot.png", fullPage: true)

# Check network requests to verify CSS and assets loaded
mcp__playwright__browser_network_requests()
# Returns: All HTTP requests with status codes (useful for debugging 404s)
```

## GitHub Pages and Jekyll Documentation Setup

### Critical Setup Requirements

**1. Entry Point File:**
- GitHub Pages requires `index.md` or `index.html` as the landing page
- Jekyll does NOT automatically convert `README.md` to `index.html`
- Always create an explicit `index.md` file in the docs directory

**2. Jekyll Plugin Dependencies:**
- Plugins must be installed BEFORE Jekyll can use their template tags
- Common error: `Liquid syntax error: Unknown tag 'seo'` means plugin not loaded
- Solution: Either install the plugin or remove the template tag

**3. Custom CSS Integration:**
- Custom stylesheets must be explicitly linked in `_layouts/default.html`
- Path format: `{{ "/assets/css/style.css?v=" | append: site.github.build_revision | relative_url }}`
- SCSS files in `docs/assets/css/style.scss` are automatically processed by Jekyll

**4. Layout and Content Separation:**
- Avoid duplicate H1 headings between layout header and page content
- Layout typically provides site title/header
- Page content should start with introductory text, not repeat the title

### Common Jekyll Build Errors

**Error: `Unknown tag 'seo'`**
- Cause: `jekyll-seo-tag` plugin not installed or not in `_config.yml` plugins list
- Fix: Either add plugin to Gemfile and _config.yml, or replace `{% seo %}` with manual tags
- Manual alternative:
  ```html
  <title>{{ page.title | default: site.title }}</title>
  <meta name="description" content="{{ page.description | default: site.description }}" />
  ```

**Error: 404 on GitHub Pages**
- Cause: Missing `index.md` or `index.html` in docs directory
- Fix: Create `index.md` with proper frontmatter:
  ```yaml
  ---
  layout: default
  title: Your Title
  description: Your description
  ---
  ```

**Error: No CSS styling on deployed site**
- Cause: Missing stylesheet link in `_layouts/default.html`
- Fix: Add link tag in `<head>` section:
  ```html
  <link rel="stylesheet" href="{{ "/assets/css/style.css?v=" | append: site.github.build_revision | relative_url }}">
  ```

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
  validation-summary:  # Good: specific and unique
    runs-on: ubuntu-latest
    if: always()       # Proper indentation under job
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
@import "{{ site.theme }}";  // Import base theme first

/* Then add custom overrides */
table { /* Enhanced table styling */ }
.wrapper { /* Layout adjustments */ }
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
- **docs/_config.yml** - Jekyll configuration

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

### ✅ CORE-6 Completed - Cross-Provider Aggregation System

- **Status**: Advanced cross-provider cost aggregation with comprehensive validation
- **Key Features**:
  - **Currency Validation**: Ensures consistent currency across all cost results (ErrMixedCurrencies)
  - **Input Validation**: Comprehensive checks for empty results, invalid date ranges, and grouping types
  - **Time-Based Aggregation**: Daily and monthly cost aggregation with intelligent cost conversion
  - **Provider Extraction**: Automatic provider identification from resource types ("aws:ec2:Instance" → "aws")
  - **Type Safety**: GroupBy validation methods (IsValid(), IsTimeBasedGrouping(), String())
  - **Error Handling**: Specific error types for different validation failures
  - **Cost Intelligence**: Prefers actual costs (TotalCost) over projections with automatic daily/monthly conversion

### Architecture Changes

- **New Engine Method**: `GetActualCostWithOptions()` with flexible querying
- **Enhanced Data Structures**: `ActualCostRequest` with advanced filtering options
- **Tag Matching**: `matchesTags()` helper for resource filtering
- **Cost Aggregation**: Daily/monthly cost breakdown logic
- **Output Enhancement**: Rich table formatting for actual cost results
- **Cross-Provider Functions**: `CreateCrossProviderAggregation()` with comprehensive validation pipeline
- **Currency System**: Centralized currency validation with defaulting to USD
- **Time Processing**: Intelligent cost calculation for different time periods
- **Error Types**: New error constants for specific validation scenarios

**New Error Types for Cross-Provider Aggregation**:

```go
var (
    ErrNoCostData       = errors.New("no cost data available")
    ErrMixedCurrencies  = errors.New("mixed currencies not supported in cross-provider aggregation")
    ErrInvalidGroupBy   = errors.New("invalid groupBy type for cross-provider aggregation")
    ErrEmptyResults     = errors.New("empty results provided for aggregation")
    ErrInvalidDateRange = errors.New("invalid date range: end date must be after start date")
)
```

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
- `golangci-lint: enabled: true` - Integrates with project's existing linting
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

### Commands

```bash
@coderabbitai resolve          # Mark all previous comments as resolved
@coderabbitai configuration    # Show current configuration
@coderabbitai plan            # Plan code edits for comments
```
