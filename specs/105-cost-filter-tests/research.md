# Research Findings: Integration Tests for --filter Flag

**Date**: Tue Dec 16 2025
**Researcher**: AI Assistant
**Feature**: Integration Tests for --filter Flag

## Decision: Filter Implementation Status Clarification

**What was researched**: Whether `cost actual` supports a `--filter` flag as stated in the specification.

**Findings**:

- `cost projected` DOES support `--filter` flag (implemented in `internal/cli/cost_projected.go`)
- `cost actual` does NOT support `--filter` flag (confirmed by examining `internal/cli/cost_actual.go`)
- `cost actual` uses `--group-by "tag:key=value"` for tag-based filtering instead
- The specification incorrectly states that `--filter` is implemented for `cost actual`

**Decision**: Update specification to reflect reality - test the existing `--group-by` tag filtering mechanism for `cost actual`, not a non-existent `--filter` flag.

**Rationale**: Aligns with actual codebase implementation. Testing non-existent functionality would fail. The existing `--group-by` tag filtering provides equivalent functionality.

**Alternatives considered**:

- Implement `--filter` flag for `cost actual` to match spec - rejected as this would change existing API without justification
- Test only `cost projected` filtering - rejected as spec specifically requires both commands

## Decision: Test Fixture Scale

**What was researched**: Appropriate scale for test fixtures (number of resources in test plans).

**Findings**:

- Existing fixture `aws-multi-resource-plan.json` contains 4 resources (EC2 instance, S3 bucket, RDS instance, Lambda function)
- All resources are AWS-based, no cross-provider testing possible
- Resources have basic tagging (Environment: dev, Purpose: Static Assets, etc.)

**Decision**: Create new test fixture with 10-20 resources across AWS, Azure, and GCP providers to support comprehensive filtering tests.

**Rationale**: Provides sufficient variety for testing type/provider filters while staying within reasonable test execution time limits. Cross-provider coverage ensures filter logic works across different cloud platforms.

**Alternatives considered**:

- Use existing 4-resource fixture - insufficient for comprehensive testing
- Create 50+ resource fixture - would increase test execution time beyond acceptable limits

## Decision: Filter Syntax Error Messages

**What was researched**: Specific error messages for invalid filter syntax in `cost projected`.

**Findings**:

- Filter parsing happens in `engine.matchesFilter()` function
- Invalid filter syntax (not "key=value" format) causes all resources to be included (no error)
- No explicit error messages for malformed filters
- Current behavior: invalid filters are silently ignored

**Decision**: Tests should expect no error for invalid syntax (current behavior), but verify that invalid filters include all resources rather than filtering.

**Rationale**: Matches current implementation behavior. While not ideal UX, changing error handling would be outside scope of adding tests.

**Alternatives considered**:

- Add error handling for invalid filters - would require code changes beyond testing scope
- Test for error messages that don't exist - would cause test failures

## Decision: Case Sensitivity Implementation

**What was researched**: How filter case sensitivity is implemented.

**Findings**:

- `matchesFilter()` converts both filter key/value and resource properties to lowercase for comparison
- Provider/service extraction uses `strings.ToLower()`
- Resource type comparison uses `strings.Contains(strings.ToLower(resource.Type), value)`
- Filtering is effectively case-insensitive for all supported filter types

**Decision**: Tests should expect case-insensitive behavior - filters like `TYPE=AWS:EC2/INSTANCE` should work.

**Rationale**: Matches current implementation. This provides better UX than strict case sensitivity.

**Alternatives considered**:

- Test for case-sensitive behavior - would not match implementation
- Implement case-sensitive filtering - would require code changes

## Decision: Test Infrastructure Availability

**What was researched**: Available test infrastructure and helper functions.

**Findings**:

- `test/integration/helpers/cli_helper.go` provides comprehensive CLI testing utilities
- `CLIHelper` supports executing commands, capturing output, creating temp files
- Existing fixtures in `test/fixtures/plans/` include various resource types
- Integration test pattern established in `internal/cli/integration_test.go`

**Decision**: Use existing `CLIHelper` infrastructure - no additional setup required.

**Rationale**: Mature, well-tested infrastructure already handles the testing needs.

**Alternatives considered**:

- Build custom test helpers - unnecessary duplication of effort
- Use unit test patterns - insufficient for CLI integration testing

## Decision: Output Format Testing Approach

**What was researched**: How to test filtering across different output formats (table, json, ndjson).

**Findings**:

- CLI helper supports executing with different `--output` flags
- JSON/NDJSON outputs are machine-parseable
- Table output is human-readable text

**Decision**: Test JSON output for structural validation, test table output for content inclusion/exclusion, test NDJSON for line-based parsing.

**Rationale**: Covers all output formats with appropriate validation methods for each format type.

**Alternatives considered**:

- Test only JSON output - misses table/NDJSON validation
- Parse table output as structured data - brittle and unnecessary

## Decision: Actual Cost Testing Complexity

**What was researched**: Complexity of testing `cost actual` with tag filtering.

**Findings**:

- Requires mock plugin server for cost data
- Tag filtering uses `--group-by "tag:key=value"` syntax
- Existing integration tests may have mock server setup

**Decision**: Include `cost actual` tag filtering tests using existing mock infrastructure if available, or note as requiring additional setup.

**Rationale**: Maintains test coverage completeness. If mock setup is complex, document as future enhancement.

**Alternatives considered**:

- Skip `cost actual` testing - reduces coverage as specified in requirements
- Implement full mock server - may be beyond initial test implementation scope
