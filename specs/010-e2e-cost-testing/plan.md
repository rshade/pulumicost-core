# Implementation Plan: E2E Cost Testing

**Branch**: `010-e2e-cost-testing` | **Date**: 2025-12-03 | **Spec**: [Feature Spec](./spec.md)
**Input**: Feature specification from `/specs/010-e2e-cost-testing/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Implement an End-to-End (E2E) testing framework for PulumiCost using the Pulumi Automation API. This framework will programmatically deploy AWS resources (e.g., T3 micro instances), run cost calculations, and validate that:
1. Projected costs match AWS list prices within ±5%.
2. Actual costs are proportional to runtime duration.
3. All resources are automatically cleaned up after tests.

## Technical Context

**Language/Version**: Go 1.25
**Primary Dependencies**: 
- `github.com/pulumi/pulumi/sdk/v3/go/auto` (Automation API)
- `github.com/rshade/pulumicost/plugin-aws-public` (Cost data source)
- `github.com/stretchr/testify` (Assertions)
**Storage**: Local Pulumi state (ephemeral), no persistent DB.
**Testing**: `go test` with `//go:build e2e` tag.
**Target Platform**: Linux, macOS, Windows.
**Project Type**: CLI Tool / Testing Framework.
**Performance Goals**: Tests must complete within 60 minutes (default, configurable).
**Constraints**: Must not leave orphaned resources. Must run in CI environment.
**Scale/Scope**: Covers EC2 and EBS initially.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Verify compliance with PulumiCost Core Constitution (`.specify/memory/constitution.md`):

- [x] **Plugin-First Architecture**: Is this feature implemented as a plugin or orchestration logic? (Orchestration/Testing)
- [x] **Test-Driven Development**: Are tests planned before implementation? (80% minimum coverage)
- [x] **Cross-Platform Compatibility**: Will this work on Linux, macOS, Windows?
- [x] **Documentation as Code**: Are audience-specific docs planned? (Quickstart for devs)
- [x] **Protocol Stability**: Do protocol changes follow semantic versioning? (N/A)
- [x] **Quality Gates**: Are all CI checks (tests, lint, security) passing?
- [x] **Multi-Repo Coordination**: Are cross-repo dependencies documented? (Depends on AWS plugin)

**Violations Requiring Justification**: None.

## Project Structure

### Documentation (this feature)

```text
specs/010-e2e-cost-testing/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (N/A for this feature)
└── tasks.md             # Phase 2 output
```

### Source Code (repository root)

```text
test/e2e/
├── infra/               # Inline Pulumi programs
│   └── aws/
│       ├── ec2.go
│       └── ebs.go
├── main_test.go         # Test entry point and setup
├── e2e_white_box_test.go # White-box tests (importing packages)
├── e2e_black_box_test.go # Black-box tests (CLI binary)
├── cleanup.go           # Resource cleanup helpers
└── utils.go             # Helper functions (ULID, etc.)
```

**Structure Decision**: Option 1 (Single project) adapted for test directory structure.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | | |