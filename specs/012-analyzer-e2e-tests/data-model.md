# Data Model: Analyzer E2E Tests

**Feature**: Analyzer E2E Tests
**Date**: 2025-12-08

## Entities

### AnalyzerTestContext

Extends the existing `TestContext` from `test/e2e/main_test.go` with analyzer-specific functionality.

| Field         | Type        | Description                          |
| ------------- | ----------- | ------------------------------------ |
| T             | *testing.T  | Test instance                        |
| StackName     | string      | Unique stack name with ULID suffix   |
| WorkDir       | string      | Temp directory for test project      |
| BinaryPath    | string      | Path to finfocus binary            |
| PreviewOutput | string      | Captured stdout from pulumi preview  |
| PreviewStderr | string      | Captured stderr from pulumi preview  |

### AnalyzerDiagnostic

Represents an expected diagnostic pattern in preview output.

| Field    | Type   | Description                             |
| -------- | ------ | --------------------------------------- |
| Pattern  | string | Regex or substring to match             |
| Required | bool   | Whether diagnostic must be present      |
| Count    | int    | Expected number of occurrences (0=any)  |

### TestFixtureProject

Represents a Pulumi project fixture for testing.

| Field        | Type     | Description                           |
| ------------ | -------- | ------------------------------------- |
| Name         | string   | Project name                          |
| Path         | string   | Path relative to test/e2e/projects    |
| AnalyzerPath | string   | Path to analyzer binary in Pulumi.yaml|
| Resources    | []string | List of resource URNs expected        |

## Relationships

```text
AnalyzerTestContext
    |
    +-- TestFixtureProject (1:1 - each test uses one project)
    |
    +-- AnalyzerDiagnostic (1:N - validates multiple diagnostics)
```

## State Transitions

### Test Lifecycle

```text
[Initial] --> [Setup] --> [Preview] --> [Validate] --> [Cleanup] --> [Complete]

Initial:   Binary built, Pulumi CLI available
Setup:     Create temp dir, copy project, init stack
Preview:   Run pulumi preview, capture output
Validate:  Check diagnostics against expectations
Cleanup:   Destroy stack, remove temp dir
Complete:  Test passed or failed
```

## Validation Rules

### Diagnostic Validation

1. Stack summary diagnostic MUST appear exactly once
2. Per-resource diagnostics MUST appear for each resource in preview
3. Cost values MUST be parseable as float64
4. Policy pack name MUST be "finfocus"

### Output Parsing

1. Preview stdout MUST contain diagnostic messages
2. Preview stderr MAY contain debug logging (ignored)
3. Exit code MUST be 0 for successful preview
