<!--
Sync Impact Report - Constitution v1.0.0 (Initial Ratification)
===============================================================

Version Change: [TEMPLATE] → 1.0.0
Change Type: Initial ratification (MAJOR)

Principles Established:
- I. Plugin-First Architecture
- II. Test-Driven Development (NON-NEGOTIABLE)
- III. Cross-Platform Compatibility
- IV. Documentation as Code
- V. Protocol Stability

Sections Added:
- Quality Gates (CI/CD requirements)
- Multi-Repo Governance (cross-repo coordination)
- Governance (amendment process)

Templates Requiring Updates:
- ✅ plan-template.md (Constitution Check section aligned)
- ✅ spec-template.md (Requirements align with quality standards)
- ✅ tasks-template.md (Test requirements match TDD principle)

Follow-up TODOs:
- None (all placeholders filled)

Date: 2025-11-06
-->

# PulumiCost Core Constitution

## Core Principles

### I. Plugin-First Architecture

All cost data sources MUST be implemented as gRPC plugins. The core system
MUST remain provider-agnostic and serve only as an orchestration layer.
Plugins MUST communicate via the protocol buffer definitions in
`pulumicost-spec` repository. Direct provider integrations within core are
forbidden.

**Rationale**: Enables extensibility without modifying core, supports
third-party plugin development, and maintains clear separation of concerns
between orchestration and data sourcing.

### II. Test-Driven Development (NON-NEGOTIABLE)

Tests MUST be written before implementation. All code changes MUST maintain
minimum 80% test coverage. Critical paths (CLI entry points, cost calculation
engine, plugin communication) MUST achieve 95% coverage. Tests MUST pass in
CI before any pull request merges.

**Rationale**: Ensures reliability in cost calculations where accuracy is
paramount. TDD prevents regressions and serves as living documentation of
system behavior.

### III. Cross-Platform Compatibility

All code MUST compile and run correctly on Linux (amd64, arm64), macOS
(amd64, arm64), and Windows (amd64). CI MUST verify cross-platform builds on
every commit. Platform-specific code MUST be isolated and documented.

**Rationale**: Infrastructure teams use diverse operating systems.
Cross-platform support is essential for adoption and prevents vendor lock-in.

### IV. Documentation as Code

Every user-facing feature MUST have corresponding documentation before
release. Documentation MUST include audience-specific guides (User, Developer,
Architect, Business). All code examples in documentation MUST be tested.
Documentation linting MUST pass in CI.

**Rationale**: Complex systems require comprehensive documentation.
Audience-specific guides ensure all stakeholders can adopt and extend
PulumiCost effectively.

### V. Protocol Stability

Breaking changes to protocol buffer definitions MUST follow semantic
versioning. Protocol changes MUST be coordinated across all three
repositories (core, spec, plugin). Backward compatibility MUST be maintained
for at least two minor versions.

**Rationale**: Plugins are developed by third parties. Protocol stability
prevents breaking existing integrations and ensures ecosystem reliability.

## Quality Gates

All pull requests MUST pass the following automated checks before merging:

- **Go Tests**: All unit, integration, and contract tests pass with `-race`
  flag
- **Code Coverage**: Minimum 80% overall, 95% for critical paths (enforced
  via CI)
- **Linting**: `golangci-lint` passes with project configuration
  (`.golangci.yml`)
- **Security**: `govulncheck` reports no high/critical vulnerabilities
- **Formatting**: All Go code formatted with `gofmt` and `goimports`
- **Documentation**: Markdown linting passes (`markdownlint-cli2`)
- **Cross-Platform Build**: Successful compilation on Linux, macOS, Windows

## Multi-Repo Governance

PulumiCost operates as a three-repository ecosystem:

- **pulumicost-core**: CLI tool, plugin host, orchestration engine (this
  repository)
- **pulumicost-spec**: Protocol buffer definitions, SDK generation
- **pulumicost-plugin**: Plugin implementations (Kubecost, Vantage, etc.)

**Cross-Repo Change Protocol**:

1. Protocol changes MUST originate in `pulumicost-spec` with version bump
2. Core MUST update dependencies and adapt to new protocol
3. Plugins MUST update to new SDK version with compatibility testing
4. Changes MUST be coordinated via GitHub Project board (project #3)
5. Documentation MUST be updated across all affected repositories

**Dependency Rules**:

- Core depends on spec (protocol definitions)
- Plugins depend on spec (SDK)
- Core MUST NOT depend on specific plugin implementations
- Spec MUST remain independent (no external dependencies)

## Governance

This constitution supersedes all other development practices and conventions.
All code reviews, pull requests, and architectural decisions MUST verify
compliance with these principles.

**Amendment Process**:

- Amendments require documentation of rationale and impact analysis
- Breaking changes to principles require MAJOR version bump
- New principles or expanded guidance require MINOR version bump
- Clarifications and non-semantic changes require PATCH version bump
- All amendments MUST update this document and propagate changes to dependent
  templates

**Complexity Justification**:

Any violation of principles MUST be explicitly justified with:

1. Clear explanation of why the principle cannot be followed
2. Description of simpler alternatives considered and rejected
3. Documented plan to remediate in future versions

**Runtime Development Guidance**:

For day-to-day development conventions, tooling preferences, and
project-specific patterns, refer to `CLAUDE.md` in the repository root. That
file provides operational context while this constitution defines
non-negotiable architectural principles.

**Version**: 1.0.0 | **Ratified**: 2025-11-06 | **Last Amended**: 2025-11-06
