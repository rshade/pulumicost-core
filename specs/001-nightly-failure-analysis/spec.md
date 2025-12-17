# Feature Specification: Automated Nightly Failure Analysis

**Feature Branch**: `001-nightly-failure-analysis`
**Created**: 2025-12-16
**Status**: Draft
**Input**: User description: "Research: Automated nightly failure analysis with OpenCode..."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Automated Failure Triage (Priority: P1)

As a developer, I want the system to automatically analyze nightly build failures immediately after they are reported, so that I don't have to manually hunt for logs and root causes.

**Why this priority**: Reducing manual triage time for nightly failures is the core value proposition of this feature.

**Independent Test**: Can be tested by manually creating an issue with the `nightly-failure` label and verifying that the analysis workflow triggers and produces a report.

**Acceptance Scenarios**:

1.  **Given** a new GitHub issue is created with the label `nightly-failure`, **When** the issue is saved, **Then** the analysis workflow automatically starts.
2.  **Given** an existing issue is labeled with `nightly-failure`, **When** the label is applied, **Then** the analysis workflow automatically starts.
3.  **Given** the analysis workflow runs, **When** it completes, **Then** a comment is posted to the issue with the analysis results.

---

### User Story 2 - Comprehensive Analysis Report (Priority: P1)

As a developer, I want the analysis comment to include a summary of the failure, the likely root cause, and suggested fix steps, so that I can quickly assess if it's a real regression or a flaky test.

**Why this priority**: The analysis is only useful if it provides actionable intelligence beyond what the raw logs offer.

**Independent Test**: Can be tested by mocking a failure log and verifying the generated comment content structure.

**Acceptance Scenarios**:

1.  **Given** a failed nightly run with a specific error in the logs, **When** the analysis is posted, **Then** it explicitly quotes or references the error message.
2.  **Given** a failure, **When** the analysis is posted, **Then** it includes a "Root Cause Hypothesis" section.
3.  **Given** a failure, **When** the analysis is posted, **Then** it includes a "Suggested Investigation" section.

---

### Edge Cases

- What happens when the workflow logs are too large for the context window? The system should truncate or summarize logs intelligently.
- What happens if the run ID cannot be extracted from the issue body? The system should post a comment indicating it could not find the linked run.
- What happens if the analysis API fails or times out? The workflow should fail gracefully and potentially post a generic "analysis failed" message.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST trigger an analysis workflow upon the creation of a GitHub issue with the `nightly-failure` label.
- **FR-002**: System MUST trigger an analysis workflow when the `nightly-failure` label is added to an existing issue.
- **FR-003**: System MUST extract the GitHub Actions Run ID or URL from the body of the failure issue.
- **FR-004**: System MUST retrieve the build logs for the specified Run ID via the GitHub API.
- **FR-005**: System MUST filter or parse logs to identify the specific job(s) and step(s) that failed.
- **FR-006**: System MUST pass the relevant failure context (logs, workflow configuration) to the analysis engine (OpenCode/LLM).
- **FR-007**: System MUST post the resulting analysis as a new comment on the triggering issue.
- **FR-008**: The analysis output MUST be structured with clear headings (e.g., Summary, Root Cause, Recommendations).

### Key Entities

- **Failure Issue**: The GitHub issue representing the failed nightly run.
- **Workflow Run**: The specific execution of the `nightly.yml` workflow that failed.
- **Analysis Report**: The generated markdown content containing the root cause analysis.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Analysis comment is posted within 15 minutes of the issue creation/labeling.
- **SC-002**: The system successfully retrieves and processes logs for 100% of accessible failed runs.
- **SC-003**: The generated report identifies the correct failing test case or error message in >90% of failures.
- **SC-004**: The system handles log files up to standard limits (e.g., 10MB) without crashing.

## Research & Feasibility Goals

<!--
  This section tracks the specific research questions to be answered during implementation,
  as this is a research-focused feature.
-->

- **RQ-001**: Verify GitHub Actions trigger latency for `issues` events.
- **RQ-002**: Determine exact permissions required for fetching logs from within a workflow.
- **RQ-003**: Validate the format of retrieved logs (zip vs text) and extraction methods.
- **RQ-004**: Assess API rate limits for log retrieval and LLM analysis.