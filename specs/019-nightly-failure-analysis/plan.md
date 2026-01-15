# Implementation Plan: Automated Nightly Failure Analysis

**Branch**: `019-nightly-failure-analysis` | **Date**: 2025-12-16 | **Spec**: [specs/019-nightly-failure-analysis/spec.md](specs/019-nightly-failure-analysis/spec.md)
**Input**: Feature specification from `specs/019-nightly-failure-analysis/spec.md`

## Summary

Implement an automated workflow that triggers on `nightly-failure` issues, retrieves the corresponding build logs via GitHub API, analyzes the failure using OpenCode/LLM, and posts a triage report as a comment. This reduces manual investigation time for nightly regressions.

## Technical Context

**Language/Version**: GitHub Actions YAML, Scripting (Go vs JS vs Python) [NEEDS CLARIFICATION: Best choice for analysis script?]
**Primary Dependencies**: GitHub Actions, `gh` CLI, OpenCode CLI/API
**Storage**: N/A (Stateless workflow)
**Testing**: Workflow unit tests? [NEEDS CLARIFICATION: How to test GH Actions logic locally?]
**Target Platform**: GitHub Actions Runner (Ubuntu-latest)
**Project Type**: CI/CD Automation
**Performance Goals**: Analysis posted < 15 mins
**Constraints**: GitHub API Rate Limits, Log size limits
**Scale/Scope**: Nightly builds (1/day)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Verify compliance with FinFocus Core Constitution (`.specify/memory/constitution.md`):

- [x] **Plugin-First Architecture**: N/A (CI/CD Infrastructure)
- [x] **Test-Driven Development**: Logic scripts will be TDD'd; Workflow behavior verified via integration tests.
- [x] **Cross-Platform Compatibility**: Workflow runs on Linux; scripts must be cross-platform compatible.
- [x] **Documentation as Code**: Internal docs for the workflow.
- [x] **Protocol Stability**: N/A
- [x] **Quality Gates**: Will adhere to linting/formatting.
- [x] **Multi-Repo Coordination**: N/A

**Violations Requiring Justification**: None.

## Project Structure

### Documentation (this feature)

```text
specs/019-nightly-failure-analysis/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output
```

### Source Code (repository root)

```text
.github/
└── workflows/
    └── nightly-analysis.yml  # New workflow

scripts/
└── analysis/                 # [NEEDS CLARIFICATION: Script location?]
    └── analyze_failure.go    # Analysis logic
```

**Structure Decision**: A new GitHub Actions workflow file coupled with a helper script (likely Go to match project language) for complex logic like log parsing and API interaction.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | | |