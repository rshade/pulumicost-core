# Feature Specification: Comprehensive Documentation Sync

**Feature Branch**: `010-sync-docs-codebase`
**Created**: 2025-12-08
**Status**: Draft
**Input**: Update documentation to sync with codebase (Analyzer, Plugins, CLI, Broken Links, AI Context)

## User Scenarios & Testing

### User Story 1 - Zero-Click Cost Estimation (Priority: P1)

As a Pulumi user, I want to easily find documentation on how to set up and use the Analyzer integration so that I can see cost estimates directly in my `pulumi preview` output.

**Why this priority**: The Analyzer is a key differentiator ("Zero-click") but is currently completely undocumented.

**Independent Test**: Can a user follow `docs/getting-started/analyzer-setup.md` (or the section in User Guide) and successfully configure the analyzer without external help?

**Acceptance Scenarios**:
1. **Given** a user reading the README, **When** they look for "Analyzer" or "Cost Estimation", **Then** they find a link to the Analyzer documentation.
2. **Given** the Analyzer documentation, **When** the user follows the configuration steps, **Then** `pulumi preview` works with the plugin.

### User Story 2 - Accurate CLI Reference (Priority: P1)

As a DevOps engineer, I want a complete reference of all `finfocus` CLI commands, including plugin management, so that I can automate installation and updates.

**Why this priority**: Users cannot effectively manage plugins or script the tool if commands are undocumented.

**Independent Test**: Verify every command in `finfocus --help` has a corresponding entry in `docs/reference/cli-commands.md`.

**Acceptance Scenarios**:
1. **Given** a user wants to install a plugin, **When** they check the CLI reference, **Then** they find `finfocus plugin install` documented with examples.

### User Story 3 - AI Agent Context Awareness (Priority: P1)

As a developer using AI assistants (Claude/Gemini) in this repo, I want the AI to be aware of the latest project structure and features so that it generates correct code and explanations.

**Why this priority**: AI assistants relying on outdated `CLAUDE.md` files will hallucinate or suggest obsolete patterns, wasting developer time.

**Independent Test**: Ask an AI with the updated context "How do I add a new analyzer feature?" and verify it references the correct architecture.

**Acceptance Scenarios**:
1. **Given** an AI agent session, **When** the user asks about the Analyzer architecture, **Then** the AI references `docs/architecture/analyzer.md` and `internal/analyzer/CLAUDE.md`.

### User Story 4 - Configuration & Troubleshooting (Priority: P2)

As a user debugging an issue, I want to verify my configuration and understand error codes so that I can resolve problems independently.

**Why this priority**: Reduces support burden and improves user experience during errors.

**Independent Test**: Can a user match an error code (e.g., `ErrNoCostData`) to a solution in the docs?

**Acceptance Scenarios**:
1. **Given** a `ErrNoCostData` error, **When** the user searches the docs, **Then** they find the `ErrNoCostData` entry in `docs/reference/error-codes.md`.
2. **Given** a need to change log levels, **When** the user checks `docs/reference/config-reference.md` or `environment-variables.md`, **Then** they find the correct setting.

### Edge Cases

- **Version Mismatch**: Users may read the new documentation but have an older CLI version installed.
  - *Handling*: Documentation should clearly state version requirements for new features (e.g., "Available in v0.5.0+").
- **Missing Code Implementation**: Documentation might describe features not yet merged to `main`.
  - *Handling*: Verify feature existence in `internal/` before documenting. Mark as "Preview" if experimental.

## Requirements

### Assumptions

- **Code Implementation**: It is assumed that the Analyzer, Plugin Kit, and Cross-Provider Aggregation features are already implemented in the codebase.
- **Tooling**: A linting tool (e.g., `markdown-link-check` or `make docs-lint`) is available or can be configured to verify links.

### Functional Requirements

#### New Documentation Files
- **FR-001**: System MUST include `docs/architecture/analyzer.md` detailing the Analyzer architecture, configuration, and protocol.
- **FR-002**: System MUST include `docs/reference/config-reference.md` defining the schema for `config.yaml` (output, logging, plugins).
- **FR-003**: System MUST include `docs/reference/error-codes.md` listing Engine and CLI errors with solutions.
- **FR-004**: System MUST include `docs/reference/environment-variables.md` listing all `FINFOCUS_*` variables.
- **FR-005**: System MUST include `docs/getting-started/analyzer-setup.md` (or equivalent section) for setting up the analyzer.
- **FR-006**: System MUST include `docs/deployment/cicd-integration.md` (basic version).

#### Updates to Existing Documentation
- **FR-007**: `README.md` MUST be updated to list new features (Analyzer, Cross-Provider, Plugin Kit), fix broken links, and update Quick Start.
- **FR-008**: `docs/reference/cli-commands.md` MUST be updated to include `plugin` (init, install, update, remove) and `analyzer serve` commands.
- **FR-009**: `docs/guides/user-guide.md` MUST include Analyzer usage, Cross-Provider aggregation, and Plugin management.
- **FR-010**: `docs/guides/developer-guide.md` MUST include Analyzer development and Cross-Provider engine features.
- **FR-011**: Empty stub files (`docs/plugins/README.md`, etc.) MUST be either completed or removed.

#### AI Agent Documentation Updates
- **FR-012**: `CLAUDE.md` (root) MUST be updated with Analyzer package docs, new CLI surface area, and Cross-Provider aggregation features.
- **FR-013**: `internal/cli/CLAUDE.md` MUST document `analyzer serve` and `plugin` commands.
- **FR-014**: `internal/engine/CLAUDE.md` MUST verify documentation of GroupBy and Error types.
- **FR-015**: `AGENTS.md` and `GEMINI.md` MUST be synchronized with `CLAUDE.md` updates.

### Key Entities

- **Documentation Source**: The `internal/` code (CLI, Engine, Analyzer) serves as the source of truth.
- **AI Context Files**: `CLAUDE.md`, `GEMINI.md`, `AGENTS.md` which direct AI behavior.

## Success Criteria

### Measurable Outcomes

- **SC-001**: `make docs-lint` (or equivalent link checker) passes with 0 broken links in `docs/README.md`.
- **SC-002**: All 5 new documentation files defined in FR-001 to FR-005 exist and are non-empty (>500 bytes).
- **SC-003**: `docs/reference/cli-commands.md` contains entries for "plugin" and "analyzer".
- **SC-004**: `CLAUDE.md` contains the string "Analyzer" and "Plugin Management".
- **SC-005**: `README.md` contains a working link to "Analyzer Integration".