# Data Model: Integration Tests for --filter Flag

**Date**: Tue Dec 16 2025
**Feature**: Integration Tests for --filter Flag

## Overview

The data model centers around testing the `--filter` flag functionality across cost commands. The primary entities are test fixtures and test execution infrastructure.

## Core Entities

### TestPlan

**Purpose**: Represents a Pulumi plan JSON file used as input for cost calculation testing.

**Attributes**:

- `resources`: Array of resource objects (10-20 resources for comprehensive testing)
- `metadata`: Plan metadata (Pulumi version, project info)
- `configuration`: Provider configurations

**Relationships**:

- Used by `CLIHelper` for test execution
- Contains multiple `Resource` instances

**Validation Rules**:

- Must contain resources from multiple providers (AWS, Azure, GCP)
- Must include varied resource types (compute, storage, database)
- Must have consistent tagging scheme for tag-based filtering tests
- Must be actual exported Pulumi plan (not synthetic)

**State Transitions**:

- Created → Used in test → Cleaned up

### TestHelper (CLIHelper)

**Purpose**: Provides utilities for executing CLI commands in integration tests.

**Attributes**:

- `command`: Cobra command instance
- `stdout/stderr`: Captured output buffers
- `tempFiles`: List of temporary files to clean up

**Relationships**:

- Executes commands against `TestPlan` instances
- Produces `TestResult` instances

**Validation Rules**:

- Must suppress logging to avoid output pollution
- Must handle JSON parsing of command output
- Must clean up temporary files after test completion

**State Transitions**:

- Initialized → Command executed → Output captured → Resources cleaned up

### Resource

**Purpose**: Individual cloud resource within a TestPlan.

**Attributes**:

- `type`: Resource type (e.g., "aws:ec2/instance:Instance")
- `provider`: Cloud provider ("aws", "azure", "gcp")
- `properties`: Resource configuration and tags
- `urn`: Unique resource identifier

**Relationships**:

- Belongs to exactly one `TestPlan`
- Filtered by `FilterCriteria`

**Validation Rules**:

- Type must follow Pulumi URN format
- Properties must include relevant tags for filtering tests
- Must be valid for cost calculation

### FilterCriteria

**Purpose**: Represents filtering parameters applied to resource collections.

**Attributes**:

- `filterType`: Type of filter ("type", "provider", "tag")
- `filterValue`: Filter value to match against
- `isCaseSensitive`: Whether filtering is case-sensitive (false)

**Relationships**:

- Applied to `Resource` collections
- Produces filtered `Resource` subsets

**Validation Rules**:

- Must support "type=", "provider=" syntax for projected costs
- Must support "tag:key=value" syntax for actual costs
- Invalid syntax should not produce errors (include all resources)

### TestResult

**Purpose**: Outcome of a filter test execution.

**Attributes**:

- `command`: Executed CLI command
- `output`: Captured command output
- `exitCode`: Command exit status
- `filteredResources`: Count of resources in result
- `expectedResources`: Expected count after filtering

**Relationships**:

- Produced by `TestHelper` execution
- Validates against `AcceptanceCriteria`

**Validation Rules**:

- Output must match expected format (table/json/ndjson)
- Resource counts must match filtering expectations
- No unexpected errors for valid operations

## Data Flow

1. **Test Setup**: `CLIHelper` creates `TestPlan` with temporary file
2. **Command Execution**: `CLIHelper` executes cost command with `FilterCriteria`
3. **Result Processing**: `TestResult` captures and validates output
4. **Cleanup**: Temporary files removed, buffers cleared

## Cross-References

- **TestPlan**: Must align with existing fixture structure in `test/fixtures/plans/`
- **CLIHelper**: Must use existing implementation in `test/integration/helpers/`
- **FilterCriteria**: Must match implementation in `internal/engine/matchesFilter()`
