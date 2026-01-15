# Implementation Tasks: Comprehensive Documentation Sync

**Feature**: `010-sync-docs-codebase`

## Implementation Strategy

- **Approach**: Documentation-driven development (DDD).
- **Execution**:
  - Phase 1 & 2: Build the foundation (New architecture & reference files).
  - Phase 3: Update visible guides (User Guide, README) referencing new files.
  - Phase 4: Complete CLI reference & errors.
  - Phase 5: Update AI context to reflect new docs.
  - Phase 6: Final polish & linting.

## Dependencies

- **US1 (Zero-Click)**: Depends on `analyzer.md` (T003), `config-reference.md` (T004).
- **US2 (CLI)**: Depends on nothing external, but logically follows analyzer setup.
- **US3 (AI Context)**: Should run LAST to capture all doc changes.
- **US4 (Troubleshooting)**: Depends on `error-codes.md` (T010), `config-reference.md` (T004).

## Parallel Execution Examples

- T003, T004, T005 can run in parallel (Foundational content).
- T012, T013, T014 can run in parallel (AI context updates).
- T009 and T010 can run in parallel (CLI and Errors).

## Phase 1: Setup & Cleanup

**Goal**: Clean up existing documentation structure to prepare for new content.

- [x] T001 Remove empty stub files (`docs/plugins/README.md`, `docs/reference/README.md`, `docs/deployment/README.md`, `docs/support/README.md`, `docs/getting-started/examples/README.md`)
- [x] T002 Ensure directory structure (`docs/architecture`, `docs/reference`, `docs/deployment`) exists

## Phase 2: Foundational Content

**Goal**: Create definitive reference documentation for Architecture, Configuration, and Environment Variables.

- [x] T003 [P] Create `docs/architecture/analyzer.md` (Overview, Protocol, Configuration, Diagnostics)
- [x] T004 [P] Create `docs/reference/config-reference.md` (Output, Logging, Plugins schema)
- [x] T005 [P] Create `docs/reference/environment-variables.md` (Global & Plugin env vars)

## Phase 3: User Story 1 - Zero-Click Cost Estimation

**Goal**: Enable users to set up the Analyzer integration ("Zero-Click").
**Independent Test**: User can successfully configure analyzer using `docs/getting-started/analyzer-setup.md`.

- [x] T006 [US1] Create `docs/getting-started/analyzer-setup.md` (Prerequisites, Config, Usage)
- [x] T007 [US1] Update `docs/guides/user-guide.md` (Add Analyzer section, Cross-Provider Aggregation)
- [x] T008 [US1] Update `docs/README.md` (Add Key Features, Quick Start update, Fix broken links)

## Phase 4: User Story 2 & 4 - CLI Reference & Troubleshooting

**Goal**: Provide complete command reference and troubleshooting aids.
**Independent Test**: `finfocus --help` commands match docs; Error codes in docs match codebase.

- [x] T009 [US2] Update `docs/reference/cli-commands.md` (Add plugin init/install/update/remove, analyzer serve)
- [x] T010 [P] [US4] Create `docs/reference/error-codes.md` (Engine & Config errors)
- [x] T011 [US4] Create `docs/deployment/cicd-integration.md` (Basic CI integration guide)

## Phase 5: User Story 3 - AI Agent Context Awareness

**Goal**: Update AI context files to prevent hallucinations about the codebase.
**Independent Test**: Ask AI about "Analyzer architecture" and receive correct reference.

- [x] T012 [P] [US3] Update root `CLAUDE.md` (Analyzer package, CLI surface, Cross-provider)
- [x] T013 [P] [US3] Update `internal/cli/CLAUDE.md` (Analyzer serve, Plugin commands)
- [x] T014 [P] [US3] Update `internal/engine/CLAUDE.md` (Verify GroupBy/Error types doc)
- [x] T015 [US3] Sync `AGENTS.md` and `GEMINI.md` with new features

## Phase 6: Final Polish

**Goal**: Ensure high-quality, lint-free documentation.

- [x] T016 Update `docs/guides/developer-guide.md` (Analyzer development, testing patterns)
- [x] T017 Verify all links with `make docs-lint` (or equivalent link checker)
