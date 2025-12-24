# Tasks: Automated Nightly Failure Analysis

**Feature**: Automated Nightly Failure Analysis
**Branch**: `001-nightly-failure-analysis`
**Status**: Complete
**Spec**: [specs/001-nightly-failure-analysis/spec.md](specs/001-nightly-failure-analysis/spec.md)

## Implementation Strategy

- **Approach**: Inline Bash script within GitHub Actions for maximum portability and zero-dependency execution.
- **Verification**: Verified via manual label trigger and automated `workflow_run` completion.

## Phase 1: Workflow Setup

- [x] T001 Create `.github/workflows/nightly-analysis.yml` with `workflow_run` trigger
- [x] T002 Implement logic to find or create a failure issue using `gh` CLI
- [x] T003 Implement logic to fetch failed job metadata and logs using `gh api`

## Phase 2: Analysis Implementation

- [x] T004 Implement log truncation and context preparation in `context.txt`
- [x] T005 Integrate OpenCode CLI for LLM-based failure analysis
- [x] T006 Implement posting analysis report as a GitHub issue comment

## Phase 3: Polish

- [x] T007 Ensure `context.txt` and `analysis.md` are added to `.gitignore` (if applicable)
- [x] T008 Verify workflow permissions for issues and actions