# Feature Specification: E2E Cost Testing

**Feature Branch**: `008-e2e-cost-testing`
**Created**: 2025-12-03
**Status**: Draft
**Input**: User description: "Implement E2E testing for PulumiCost using Pulumi Automation API with projected and actual cost validation"

## Clarifications

### Session 2025-12-03
- Q: What is the testing strategy for the E2E framework (white-box vs black-box)? → A: Both white-box (importing packages) and black-box (CLI binary execution) tests will be implemented in separate files (`e2e_white_box_test.go`, `e2e_black_box_test.go`).
- Q: How should expected costs be defined for validation? → A: Use a static, hardcoded map of expected resource costs (e.g., `t3.micro: 7.59`) within the test code for deterministic validation.
- Q: What naming convention should be used for unique stack names? → A: Use ULID (Universally Unique Lexicographically Sortable Identifier) appended to a prefix (e.g., `e2e-test-`).
- Q: How should AWS resources for E2E tests be isolated to prevent interference? → A: A dedicated, isolated AWS account will be used exclusively for running E2E tests.
- Q: What is the maximum cleanup timeout for E2E tests? → A: The maximum cleanup timeout for E2E tests will be 60 minutes by default, but it will be configurable to allow for flexibility.
- Q: How should the AWS region be configured for E2E tests? → A: The AWS region will be configurable via Pulumi stack configuration, with a fallback to the `AWS_REGION` environment variable.

### Session 2025-12-04
- Q: How should E2E tests generate preview JSON for pulumicost validation? → A: **E2E tests MUST follow the exact user workflow**: create a real Pulumi project directory (Pulumi.yaml), run `pulumi preview --json > preview.json` via CLI, then pass that file to `pulumicost cost projected --pulumi-json preview.json`. This matches how users and GitHub Actions will use the tool. Do NOT use Automation API inline programs (no JSON output available) or state hacking approaches that users cannot replicate.
- Q: What Pulumi runtime should E2E test projects use? → A: Use **Pulumi YAML** for E2E test projects. YAML projects are much faster (~2.5 min vs 10+ min) because they don't require Go compilation or dependency downloads. The `pulumicost` tool only needs the preview JSON output - it doesn't care what language generated it.
- Q: Why not use Go projects for E2E tests? → A: Go projects require downloading ~100MB+ of SDK dependencies and compiling for each test. YAML projects are interpreted directly by Pulumi with no compilation step, making tests significantly faster.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Validate Projected Cost Accuracy (Priority: P1)

A developer runs the E2E test suite to verify that PulumiCost accurately calculates projected costs for AWS resources before deployment. The test creates a Pulumi program programmatically, generates a preview, and validates that the cost calculations match expected AWS pricing.

**Why this priority**: Projected cost calculation is the core value proposition of PulumiCost. Without accurate projected costs, users cannot make informed decisions about infrastructure spending before deployment.

**Independent Test**: Can be fully tested by running a single projected cost test that deploys a t3.micro EC2 instance and verifies the monthly cost estimate is within ±5% of known AWS pricing ($0.0104/hour = ~$7.59/month for us-east-1).

**Acceptance Scenarios**:

1. **Given** a Pulumi program with an EC2 t3.micro instance, **When** the E2E test runs projected cost calculation, **Then** the calculated monthly cost is non-zero and within ±5% of expected AWS list price
2. **Given** a Pulumi program with an 8GB gp3 EBS volume, **When** the E2E test runs projected cost calculation, **Then** the calculated monthly cost matches expected AWS pricing (±5%)
3. **Given** a Pulumi program with unsupported resource types, **When** the E2E test runs projected cost calculation, **Then** the system returns appropriate fallback or error indication without crashing

---

### User Story 2 - Validate Actual Cost Calculation (Priority: P2)

A developer runs the E2E test suite to verify that PulumiCost can calculate actual costs based on runtime duration. After deploying resources for a known period, the test validates that actual costs are proportional to the deployment duration using the fallback formula (projected_cost × runtime_hours / 730).

**Why this priority**: Actual cost validation demonstrates the full cost lifecycle - from projection to reality. This validates the fallback calculation mechanism when real billing APIs are not available.

**Independent Test**: Can be fully tested by deploying resources, waiting a defined period (5-10 minutes), then calculating actual costs and verifying proportionality to runtime.

**Acceptance Scenarios**:

1. **Given** a deployed EC2 instance running for 10 minutes, **When** the E2E test calculates actual cost, **Then** the result equals approximately (projected_monthly_cost / 730), assuming AWS bills a minimum of one full hour
2. **Given** a Pulumi stack with multiple resources, **When** the E2E test calculates actual cost with a time range, **Then** all resources have proportional actual costs
3. **Given** a deployment with start and end timestamps, **When** the E2E test queries actual cost, **Then** the runtime is correctly parsed and applied to the calculation

---

### User Story 3 - Automated Resource Cleanup (Priority: P1)

A developer runs the E2E test suite and expects all AWS resources to be automatically cleaned up after test completion, regardless of test success or failure. This prevents orphaned resources and unexpected AWS charges.

**Why this priority**: Resource cleanup is critical for safe E2E testing. Orphaned resources lead to unexpected costs and could cause AWS account quota issues. This is a P1 alongside projected cost validation because tests without cleanup are unsafe to run.

**Independent Test**: Can be fully tested by running any E2E test and verifying via AWS console or API that no test resources remain after completion.

**Acceptance Scenarios**:

1. **Given** a successful E2E test run, **When** the test completes, **Then** all created AWS resources are destroyed
2. **Given** a failing E2E test run, **When** the test encounters an error, **Then** cleanup still executes via defer mechanism
3. **Given** an AWS API timeout during cleanup, **When** the cleanup retries, **Then** resources are eventually destroyed with exponential backoff

---

### User Story 4 - Pulumi Automation API Integration (Priority: P2)

A developer uses the E2E test framework which leverages the Pulumi Automation API for programmatic infrastructure deployment. This replaces shell script-based testing with type-safe Go code, providing better error handling and test integration.

**Why this priority**: The Automation API provides the foundation for reliable, maintainable E2E tests. It enables proper Go test integration, structured error handling, and lifecycle management.

**Independent Test**: Can be tested by verifying that a test successfully creates a Pulumi stack programmatically, deploys resources, and destroys them without manual intervention.

**Acceptance Scenarios**:

1. **Given** a Go test file using Pulumi Automation API, **When** the test runs, **Then** it programmatically creates a stack without Pulumi CLI interaction
2. **Given** a Pulumi Automation API stack, **When** deployment fails, **Then** structured error information is available for assertion
3. **Given** a test using inline Pulumi programs, **When** the program is defined in Go, **Then** resources are created with proper type checking

---

### User Story 5 - Cost Comparison and Validation (Priority: P3)

A developer receives detailed cost comparison reports from E2E tests, showing projected vs actual costs and identifying discrepancies. This helps catch regressions in cost calculation logic.

**Why this priority**: While not required for basic E2E testing, cost comparison reporting enables regression detection and builds confidence in cost accuracy over time.

**Independent Test**: Can be tested by running an E2E test and examining the validation output for projected/actual cost comparison data.

**Acceptance Scenarios**:

1. **Given** completed projected and actual cost calculations, **When** the E2E test validates results, **Then** a comparison report is generated
2. **Given** a cost discrepancy exceeding 5% tolerance, **When** validation runs, **Then** the test fails with descriptive error message
3. **Given** multiple resources with costs, **When** comparison runs, **Then** per-resource and aggregate comparisons are available

---

### User Story 6 - Plugin Integration E2E Testing (Priority: P1)

A developer runs E2E tests that validate the complete cost calculation chain including plugin installation, configuration, and cost retrieval. This ensures the entire system works as users would experience it.

**Why this priority**: Testing without the AWS pricing plugin validates CLI parsing but not real cost accuracy. The full chain test with `pulumicost-plugin-aws-public` is required to validate actual cost values against expected AWS pricing.

**Independent Test**: Can be tested by:
1. Installing `aws-public` plugin via `pulumicost plugin install aws-public`
2. Running cost calculation with the plugin
3. Validating cost output matches expected AWS pricing (~$7.59/month for t3.micro)

**Acceptance Scenarios**:

1. **Given** no plugins installed, **When** the E2E test runs cost calculation, **Then** the system returns $0.00 (validates CLI parsing works correctly)
2. **Given** the `aws-public` plugin is installed, **When** the E2E test runs cost calculation on a t3.micro EC2 instance, **Then** the calculated monthly cost is within ±5% of $7.59
3. **Given** a plugin installation fails, **When** the E2E test detects the failure, **Then** it provides a clear error message and skips cost validation tests
4. **Given** the E2E test completes (pass or fail), **When** cleanup runs, **Then** the plugin is optionally uninstalled to avoid test pollution

---

### Edge Cases

- What happens when AWS API calls fail during resource provisioning?
  - Retry logic with exponential backoff (max 3 attempts)
- What happens when cleanup takes longer than the test timeout?
  - Separate cleanup timeout (60 minutes default, configurable) with forced destruction fallback
- How does the system handle zero-cost resources?
  - Zero costs are valid and should not fail validation
- What happens when the plugin returns nil for actual costs?
  - Fallback calculation is triggered using projected costs and runtime
- How does the system handle concurrent E2E test runs?
  - Each test uses unique stack names with ULID suffixes to prevent conflicts

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide an E2E test framework in `test/e2e/` directory that tests PulumiCost cost calculations
- **FR-002**: System MUST use Pulumi Automation API (not shell scripts) for programmatic infrastructure deployment
- **FR-003**: System MUST support projected cost calculation tests that validate against known AWS pricing (±5% tolerance)
- **FR-004**: System MUST support actual cost calculation tests using the fallback formula (projected × runtime / 730)
- **FR-005**: System MUST automatically clean up all AWS resources after each test using defer statements
- **FR-006**: System MUST implement retry logic for AWS API calls with exponential backoff
- **FR-007**: System MUST enforce a 60-minute default maximum timeout for E2E test cleanup, configurable via an environment variable.
- **FR-008**: System MUST generate unique stack names for concurrent test isolation
- **FR-009**: System MUST validate that projected costs are non-zero for supported resources
- **FR-010**: System MUST validate that actual costs are proportional to resource runtime
- **FR-011**: System MUST handle unsupported resource types gracefully without crashing
- **FR-012**: System MUST parse time ranges for runtime calculations from ISO 8601 and YYYY-MM-DD formats
- **FR-013**: System MUST include `e2e_white_box_test.go` (importing packages) and `e2e_black_box_test.go` (CLI execution) for comprehensive coverage
- **FR-014**: System MUST use a hardcoded map of expected resource prices (e.g., `t3.micro` = 7.59) for deterministic validation, avoiding external pricing API calls
- **FR-015**: System MUST generate unique stack names using a ULID (Universally Unique Lexicographically Sortable Identifier) appended to a prefix like `e2e-test-` for isolation and traceability.
- **FR-016**: System MUST execute E2E tests within a dedicated, isolated AWS account to prevent interference with other environments.
- **FR-017**: System MUST configure the AWS region via Pulumi stack configuration, falling back to the `AWS_REGION` environment variable if not set.
- **FR-018**: System MUST support testing without plugins installed to validate CLI parsing and JSON output generation.
- **FR-019**: System MUST support testing with `aws-public` plugin installed to validate full cost calculation chain.
- **FR-020**: System MUST install plugins programmatically via CLI command (`pulumicost plugin install aws-public`) during E2E test setup.
- **FR-021**: System MUST optionally cleanup installed plugins after E2E tests to prevent test pollution.

### Key Entities

- **E2E Test Harness**: Go test framework using `testing.T` with Pulumi Automation API integration
- **Test Program**: Inline Pulumi program defined in Go that creates AWS resources (EC2, EBS)
- **Cost Validator**: Component that compares calculated costs against expected values with tolerance
- **Cleanup Manager**: Deferred cleanup mechanism ensuring resource destruction on test completion
- **Stack Context**: Unique Pulumi stack per test run with configuration and state management
- **Pricing Reference**: Hardcoded map of resource types to expected monthly costs (USD)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: E2E tests complete successfully within 60 minutes (default) including resource provisioning, cost calculation, and cleanup, and respect configurable timeouts.
- **SC-002**: Projected cost calculations match expected AWS list pricing within ±5% tolerance
- **SC-003**: Actual cost calculations are proportionally accurate to runtime duration
- **SC-004**: 100% of created AWS resources are cleaned up after test completion
- **SC-005**: E2E test failures provide actionable error messages identifying the specific validation that failed
- **SC-006**: E2E tests can run concurrently without resource conflicts or name collisions
- **SC-007**: Failed tests do not leave orphaned AWS resources
- **SC-008**: Test framework integrates with `go test` and standard Go testing patterns

## Assumptions

- **A-001**: AWS credentials are available in the test environment (via environment variables or AWS config)
- **A-002**: The pulumicost-plugin-aws-public plugin v0.0.1+ is installed and accessible
- **A-003**: Pulumi CLI is installed for Automation API to leverage
- **A-004**: Test resources use smallest possible instances (t3.micro, 8GB EBS) to minimize costs
- **A-005**: The fallback actual cost formula (projected × runtime / 730) is acceptable for MVP validation
- **A-006**: 730 hours/month is the standard monthly hour calculation constant
- **A-007**: A dedicated AWS account has been provisioned and configured for E2E test execution.

## Dependencies

- **External**: pulumicost-plugin-aws-public#24 (Fallback GetActualCost implementation)
- **External**: pulumicost-plugin-aws-public#26 (E2E test support)
- **Blocked By**: None (plugin v0.0.1 already released)
- **Blocks**: pulumicost-core#180 (CI/CD pipeline for E2E tests)

## Out of Scope

- Multi-region testing (deferred to issue #185)
- Real AWS Cost Explorer integration (future CostExplorer plugin)
- Reserved Instance / Savings Plan discount calculations
- Integration with CI/CD pipeline (separate issue #180)
- AWS test account setup automation (separate issue #181)
- E2E documentation (separate issue #182)
