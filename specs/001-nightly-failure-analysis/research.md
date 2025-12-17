# Phase 0: Research Findings

**Feature**: Automated Nightly Failure Analysis
**Date**: 2025-12-16

## Decisions

### 1. Workflow Trigger Strategy
**Decision**: Use `on: issues` with `types: [labeled]` and a job-level conditional check.
**Rationale**: GitHub Actions does not support filtering by label name at the trigger level. We must trigger on all labeled events and filter within the job using `if: github.event.label.name == 'nightly-failure'`.

### 2. Log Retrieval & Processing
**Decision**: Use a custom Go script (`scripts/analysis/analyze_failure.go`) to fetch and process logs.
**Rationale**:
- **Consistency**: The project is written in Go.
- **Granularity**: We need to fetch specific job logs, not the entire run zip, to save bandwidth and context tokens.
- **Processing**: We need to parse/truncate logs to fit within LLM context limits (approx. 100k tokens for XAI, but smaller is better for latency/cost).
- **Testability**: Go logic can be unit tested with mocked HTTP clients.

### 3. OpenCode Integration
**Decision**: Install the `opencode` CLI in the workflow and invoke it with a constructed prompt.
**Rationale**:
- **Pattern Matching**: Matches the existing `opencode-review-fix.yml` workflow.
- **Flexibility**: Allows passing a constructed context file (logs + metadata) to the agent.
- **Command**: `opencode run --model xai/grok-code-free --file context.txt "Analyze this failure..."`

### 4. Analysis Output
**Decision**: The `opencode` agent will be instructed to post the comment directly or output markdown that the workflow posts via `gh` CLI.
**Refinement**: To ensure formatting control, the Go script or a `gh` command should post the result. However, `opencode run` is an agent loop. It might be safer to ask it to "Write the analysis to a file named report.md" and then have a subsequent step post `report.md` as a comment. This avoids the agent "hallucinating" a comment action if tools aren't perfectly aligned.

## Technical approach

1.  **Trigger**: Nightly workflow fails -> Creates issue -> Labels `nightly-failure`.
2.  **Analysis Workflow**:
    -   Triggered by label.
    -   **Step 1**: Checkout & Setup Go.
    -   **Step 2**: Install OpenCode CLI.
    -   **Step 3**: Run `go run scripts/analysis/analyze_failure.go`.
        -   Input: Issue Body (to extract Run URL/ID).
        -   Action: Fetch failed job logs.
        -   Output: `failure_context.txt` (Summary + Logs).
    -   **Step 4**: Run OpenCode.
        -   Input: `failure_context.txt`.
        -   Prompt: "Analyze these logs. Identify the root cause. Suggest fixes. Write output to analysis.md."
    -   **Step 5**: Post Comment.
        -   Command: `gh issue comment ${{ github.event.issue.number }} -F analysis.md`

## Unknowns Resolved

-   **Language**: Go is confirmed.
-   **Testing**: Unit tests for Go script; manual integration test for workflow.
-   **Dependencies**: `actions/setup-go`, `gh`, `opencode` CLI.
