# Tasks: Automated Nightly Failure Analysis

**Feature**: Automated Nightly Failure Analysis
**Branch**: `019-nightly-failure-analysis`
**Spec**: [specs/019-nightly-failure-analysis/spec.md](specs/019-nightly-failure-analysis/spec.md)
**Plan**: [specs/019-nightly-failure-analysis/plan.md](specs/019-nightly-failure-analysis/plan.md)

## Implementation Strategy

- **Approach**: Build the "Brain" (Analysis Script) first, then the "Body" (GitHub Actions Workflow).
- **Testing**:
    - **Unit**: Go script logic (log truncation, ID extraction) will be tested via standard Go tests.
    - **Integration**: Manual trigger of the workflow using the "Quickstart" dummy issue method.
- **Dependencies**: The workflow depends on the script being functional. US2 depends on the workflow (US1) being able to run the script.

## Phase 1: Setup

- [x] T001 Create `scripts/analysis` directory
- [x] T002 Initialize `scripts/analysis/go.mod` if standalone dependencies are required (or verify root `go.mod` can be used)

## Phase 2: Foundational (Analysis Script)

*Goal: specific logic to extract GitHub Run IDs and fetch/truncate logs for LLM consumption.*

- [x] T003 [P] [US1] Create unit tests for ID extraction and log parsing in `scripts/analysis/analyze_failure_test.go` (TDD First)
- [x] T004 [P] [US1] Create `scripts/analysis/analyze_failure.go` with CLI flag parsing (using `flag` package)
- [x] T005 [P] [US1] Implement `FailureContext` and `FailedJob` structs in `scripts/analysis/analyze_failure.go`
- [x] T006 [P] [US1] Implement logic to extract Run ID from GitHub Issue Body text in `scripts/analysis/analyze_failure.go`
- [x] T007 [US1] Implement GitHub API client logic to fetch workflow run logs in `scripts/analysis/analyze_failure.go`
- [x] T008 [US1] Implement log parsing logic to populate `FailedJob` structs (parsing specific jobs/steps) and truncation logic, writing to `context.txt` in `scripts/analysis/analyze_failure.go`

## Phase 3: User Story 1 - Automated Failure Triage

*Goal: The workflow triggers automatically and prepares the context.*

- [x] T009 [US1] Create `.github/workflows/nightly-analysis.yml` with `on: issues` (types: labeled) trigger
- [x] T010 [US1] Add job-level conditional `if: github.event.label.name == 'nightly-failure'` to `nightly-analysis.yml`
- [x] T011 [US1] Add steps to checkout repo, setup Go environment, and install OpenCode CLI in `nightly-analysis.yml`
- [x] T012 [US1] Add step to run `go run scripts/analysis/analyze_failure.go` passing the issue body and outputting to `context.txt`

## Phase 4: User Story 2 - Comprehensive Analysis Report

*Goal: The LLM analyzes the context and posts a structured report.*

- [x] T013 [US2] Define the OpenCode prompt with strict formatting requirements (Summary, Root Cause, Fixes) in `nightly-analysis.yml`
- [x] T014 [US2] Add workflow step to execute `opencode run` using `context.txt` and capture output to `analysis.md`
- [x] T015 [US2] Add workflow step to post `analysis.md` content as a GitHub Issue comment using `gh issue comment`

## Phase 5: Polish & Cross-Cutting Concerns

- [x] T016 Update `README.md` or internal docs to describe the Nightly Failure Analysis workflow
- [x] T017 Ensure `context.txt` and `analysis.md` are added to `.gitignore` to prevent accidental commits during local testing

## Dependencies

1. **Analysis Script (Phase 2)** must be complete before **Workflow Execution (Phase 3)** can be fully tested.
2. **Workflow Trigger (Phase 3)** is the prerequisite for **Analysis Reporting (Phase 4)**.

## Parallel Execution Opportunities

- **T003-T005**: Struct definitions and pure logic (ID extraction) can be written in parallel with Workflow YAML creation (T009-T010).
- **T008 (Tests)**: Can be written alongside T005/T007.
