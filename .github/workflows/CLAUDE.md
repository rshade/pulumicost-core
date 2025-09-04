# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## GitHub Actions Workflows Overview

This directory contains GitHub Actions workflows for CI/CD, automated code review, and Claude Code integration.

## Workflow Files

### CI/CD Workflows

**ci.yml** - Complete CI/CD pipeline triggered on PRs and main branch pushes:
- **Test Job**: Go 1.24.5 setup, unit tests with race detection, coverage reporting (20% minimum threshold)
- **Lint Job**: golangci-lint with project-specific configuration, security scanning with gosec
- **Security Job**: govulncheck for dependency vulnerability scanning
- **Validation Job**: gofmt formatting checks, go mod tidy verification, go vet static analysis
- **Build Job**: Cross-platform builds (Linux/macOS/Windows, amd64/arm64), artifact upload

**release.yml** - Multi-platform binary releases triggered on version tags (v*):
- Builds for Linux (amd64, arm64), macOS (amd64, arm64), Windows (amd64)
- Automatic changelog generation, SHA256 checksums, GitHub Release creation
- Binary naming: `pulumicost-v{version}-{os}-{arch}`

### Claude Code Integration Workflows

**claude-review-fix.yml** - Systematic review issue fixing:
- **Trigger**: Comments containing `/claude-review-fix` on PRs
- **Purpose**: Fixes ALL review issues from CodeRabbit, Claude Review, and Copilot in one attempt
- **Features**: 
  - Atomic single-change workflow with rollback capability
  - Concurrency control to prevent overlapping runs per PR
  - Comprehensive validation after each fix
  - Zero new issues guarantee
- **Requirements**: `ANTHROPIC_API_KEY` secret must be configured
- **Usage**: Comment `/claude-review-fix` on any PR to trigger

**claude-code-review.yml** - Automated code review:
- **Trigger**: PR creation or synchronization
- **Purpose**: Provides automated code review using Claude
- **Scope**: Configurable to filter by file paths or contributor types
- **Permissions**: Read-only access to repository and PRs
- **Status**: Optional filters available for external contributors or first-time contributors

**claude.yml** - General Claude assistance:
- **Trigger**: Comments mentioning `@claude` in issues, PRs, or reviews
- **Purpose**: General development assistance and question answering
- **Flexibility**: Responds to various GitHub events (issues, PR comments, reviews)
- **Usage**: Tag `@claude` in any comment to get development assistance
- **Permissions**: Read access to contents, PRs, issues, and CI results

## Workflow Architecture Patterns

### Security and Permissions
- All workflows use least-privilege permission models
- OIDC token authentication for secure access
- Separate permissions for read vs write operations
- Concurrency controls to prevent resource conflicts

### Error Handling and Reliability
- Timeout configurations for long-running operations
- Artifact upload with proper naming conventions
- Context cancellation handling for interrupted workflows
- Rollback mechanisms in review-fix workflow

### Integration Points
- Workflows can read CI results from other workflows
- Cross-workflow artifact sharing capabilities
- GitHub API integration for issue/PR management
- External tool integrations (golangci-lint, govulncheck)

## Development Commands for Workflows

```bash
# Validate workflow syntax
gh workflow validate .github/workflows/ci.yml
gh workflow validate .github/workflows/release.yml
gh workflow validate .github/workflows/claude-review-fix.yml

# Trigger workflows manually (if configured)
gh workflow run ci.yml
gh workflow run claude-code-review.yml

# View workflow runs and status
gh run list
gh run view <run-id>
gh run watch <run-id>

# Check workflow permissions and secrets
gh secret list
gh auth status
```

## Claude Code Usage Patterns

### Review Fixing Workflow
1. Create/update PR with code changes
2. Wait for automated reviews (CodeRabbit, etc.)
3. Comment `/claude-review-fix` on PR
4. Claude systematically fixes all issues in atomic changes
5. Each fix is validated before proceeding to next
6. Process completes with zero new issues

### General Assistance
- Use `@claude` in any GitHub comment for development help
- Ask about architecture decisions, debugging, or implementation guidance
- Request code explanations or suggestions for improvements
- Get help with testing strategies or CI/CD configuration

### Code Review Integration
- Automated reviews trigger on PR events
- Reviews focus on code quality, security, and best practices
- Integrates with existing CI pipeline results
- Provides actionable feedback for developers

## Workflow Development Guidelines

### Always Review Existing Workflows First
**CRITICAL**: Before creating new workflows, always examine existing workflows to understand established patterns, tool versions, and configurations. Inconsistency leads to maintenance issues and different behavior across workflows.

**Required Review Process**:
1. Examine `ci.yml` for current tool versions and patterns
2. Check action versions used (e.g., `actions/checkout@v5`, `actions/setup-go@v5`)
3. Identify official actions vs manual installations
4. Match timeout, caching, and configuration patterns
5. Ensure consistent permission models

### Established Tool Patterns

**golangci-lint**:
- ✅ **Use**: `golangci/golangci-lint-action@v8` with `version: v2.3.1` and `args: --timeout=5m`
- ❌ **Avoid**: Manual installation via curl scripts
- **Rationale**: Official action provides better caching, error handling, and maintenance

**Go Setup**:
- ✅ **Standard**: `actions/setup-go@v5` with `go-version: '1.24.5'` and `cache: true`
- **Consistency**: All workflows should use identical Go version and caching

**Checkout**:
- ✅ **Standard**: `actions/checkout@v5`
- **Special cases**: Use `fetch-depth: 0` only when git history is needed

### Concurrency Control Patterns
Always add concurrency control to prevent resource conflicts:
```yaml
concurrency:
  group: ${{ github.workflow }}-${{ github.event.issue.number }}  # For PR-based workflows
  cancel-in-progress: true
```

## Common Workflow Patterns

### Multi-Platform Builds
- Use matrix strategies for OS/architecture combinations
- Proper binary naming with platform suffixes
- LDFLAGS for version information embedding
- Artifact organization by platform

### Coverage and Quality Gates
- Minimum coverage thresholds with flexibility for project maturity
- Security scanning integrated into CI pipeline
- Multiple validation layers (formatting, linting, testing)
- Fail-fast approach with clear error reporting

### Dependency Management Integration
- Works with Renovate and Dependabot configurations
- Automated security vulnerability detection
- Semantic commit formatting for changelog generation
- Rate limiting to prevent notification spam