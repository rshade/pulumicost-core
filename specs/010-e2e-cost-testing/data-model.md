# Data Model: E2E Cost Testing

**Feature**: E2E Cost Testing
**Status**: Draft

## Entities

### E2E Test Context
Manages the lifecycle of a single E2E test run.

| Field | Type | Description |
|-------|------|-------------|
| `ID` | `string` | Unique identifier (ULID) for the test run. Used in stack name. |
| `StackName` | `string` | Full Pulumi stack name: `e2e-<test-name>-<id>`. |
| `T` | `*testing.T` | Go testing reference. |
| `StartTime` | `time.Time` | When the test started. |
| `Resources` | `[]Resource` | List of resources tracked for cleanup. |
| `Region` | `string` | AWS Region for this test context. |

### Cost Expectation
Defines the expected cost parameters for a test case.

| Field | Type | Description |
|-------|------|-------------|
| `ResourceName` | `string` | Pulumi resource name (URN suffix). |
| `ProjectedMonthly` | `float64` | Expected monthly cost in USD. |
| `ActualHourly` | `float64` | Expected hourly cost (for actual cost validation). |
| `TolerancePercent` | `float64` | Allowed deviation (default 5.0). |

### Validation Result
Output of a comparison between calculated and expected costs.

| Field | Type | Description |
|-------|------|-------------|
| `Matches` | `bool` | Whether the cost is within tolerance. |
| `Difference` | `float64` | Absolute difference in USD. |
| `PercentDiff` | `float64` | Percentage difference. |
| `Message` | `string` | Descriptive result message. |

## Interfaces

### CostValidator
Responsible for comparing results.

```go
type CostValidator interface {
    ValidateProjected(actual float64, expected float64) error
    ValidateActual(calculated float64, runtime time.Duration, expectedHourly float64) error
}
```

## Storage
- **State Backend**: Local filesystem (`~/.pulumi`) or ephemeral temp dir.
- **Logs**: Output to `stdout`/`stderr` via `testing.T.Log`.