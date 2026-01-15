# Analysis Script Interface

**Script**: `scripts/analysis/analyze_failure.go`

## CLI Usage

```bash
go run scripts/analysis/analyze_failure.go --issue-body "<body_content>" --output context.txt
```

## Inputs

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--issue-body` | `string` | Yes | The raw body of the GitHub issue, used to extract the Run URL. |
| `--run-id` | `string` | No | Explicit Run ID (overrides extraction from body). |
| `--output` | `string` | Yes | Path to write the generated context file. |
| `--repo` | `string` | No | "owner/repo" (default: current repo from git config or env). |

## Output Format (Context File)

The output file is a plain text file optimized for LLM consumption:

```text
# NIGHTLY FAILURE CONTEXT
Run ID: 123456789
Repo: rshade/finfocus

## FAILED JOB: Integration Tests (Linux)
Step: Run Tests
Error: 
... (Last 50 lines of logs) ...

## FAILED JOB: E2E Tests
Step: Verify Cost
Error:
... (Last 50 lines of logs) ...
```

## Environment Variables

- `GH_TOKEN`: Required for GitHub API authentication.
