# Quickstart: Automated Nightly Analysis

## Prerequisites

- GitHub CLI (`gh`) installed and authenticated.
- Access to the repository Actions.

## Testing the Workflow

You can manually trigger the analysis logic by simulating a failure issue.

1.  **Create a Dummy Issue**:
    ```bash
    gh issue create --title "Test Nightly Failure" --body "## Nightly Test Failure\n\n**Run URL:** https://github.com/rshade/finfocus/actions/runs/YOUR_FAILED_RUN_ID" --label "nightly-failure"
    ```
    *Replace `YOUR_FAILED_RUN_ID` with a real ID of a failed run (or a recent run).*

2.  **Verify Workflow Trigger**:
    Go to `Actions` tab -> `Analyze Nightly Failures` workflow.
    Ensure it picked up the new issue.

3.  **Check Output**:
    Wait for the workflow to complete.
    Check the issue for a new comment from the `github-actions` bot (or whichever identity runs the workflow).

## Running the Script Locally

To test the log fetching logic without a full workflow run:

1.  **Set Token**:
    ```bash
    export GH_TOKEN=$(gh auth token)
    ```

2.  **Run Script**:
    ```bash
    go run scripts/analysis/analyze_failure.go --run-id <REAL_RUN_ID> --output context.txt
    ```

3.  **Inspect Context**:
    ```bash
    cat context.txt
    ```

