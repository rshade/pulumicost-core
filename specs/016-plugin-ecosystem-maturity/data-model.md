# Data Model: Plugin Ecosystem Maturity

**Feature**: 102-plugin-ecosystem-maturity
**Date**: 2025-12-02

## Entities

### ConformanceTestCase

Represents a single protocol compliance check.

| Field             | Type     | Description                                    |
|-------------------|----------|------------------------------------------------|
| Name              | string   | Unique test identifier (e.g., "Name_Returns")  |
| Category          | string   | Test category: protocol, performance, error    |
| Description       | string   | Human-readable test description                |
| Timeout           | Duration | Max time for this test (default: 10s)          |
| RequiredMethods   | []string | gRPC methods this test validates               |

**Validation Rules**:

- Name must be non-empty and unique within suite
- Category must be one of: "protocol", "performance", "error", "context"
- Timeout must be positive and <= 60 seconds

---

### TestResult

Represents the outcome of running a single test.

| Field       | Type     | Description                                |
|-------------|----------|--------------------------------------------|
| TestName    | string   | Reference to ConformanceTestCase.Name      |
| Status      | Status   | pass, fail, skip, error                    |
| Duration    | Duration | Actual execution time                      |
| Error       | string   | Error message if Status != pass            |
| Details     | string   | Additional context (request/response logs) |
| Timestamp   | Time     | When test completed                        |

**State Transitions**:

```text
pending -> running -> pass
                   -> fail (assertion failed)
                   -> error (plugin crash, timeout)
                   -> skip (precondition not met)
```

**Status Values**:

- `pass`: Test assertions passed
- `fail`: Test assertions failed (expected behavior not observed)
- `skip`: Test skipped (missing credentials, version mismatch)
- `error`: Infrastructure error (plugin crash, timeout, connection lost)

---

### PluginUnderTest

Represents the plugin binary being validated.

| Field           | Type   | Description                                 |
|-----------------|--------|---------------------------------------------|
| Path            | string | Absolute path to plugin binary              |
| Name            | string | Plugin name (from Name() RPC)               |
| Version         | string | Plugin version (from Name() RPC)            |
| ProtocolVersion | string | Protocol version implemented                |
| CommMode        | string | Communication mode: "tcp" or "stdio"        |

**Validation Rules**:

- Path must exist and be executable
- CommMode must be "tcp" or "stdio"
- ProtocolVersion must match conformance suite version

---

### ConformanceSuiteConfig

Configuration for running the conformance suite.

| Field         | Type     | Description                              |
|---------------|----------|------------------------------------------|
| PluginPath    | string   | Path to plugin binary                    |
| CommMode      | string   | "tcp" or "stdio"                         |
| Verbosity     | string   | quiet, normal, verbose, debug            |
| OutputFormat  | string   | table, json, junit                       |
| OutputPath    | string   | Path for output file (optional)          |
| Timeout       | Duration | Global timeout for entire suite          |
| Categories    | []string | Filter to specific test categories       |
| TestFilter    | string   | Regex filter for test names              |

---

### E2ETestConfig

Configuration for E2E tests with real cloud providers.

| Field           | Type     | Description                              |
|-----------------|----------|------------------------------------------|
| Provider        | string   | aws, azure, gcp                          |
| AccountID       | string   | Test account identifier                  |
| Region          | string   | Cloud region for testing                 |
| CostTolerance   | float64  | Acceptable cost variance (0.05 = 5%)     |
| ResourceFilters | []string | Filter to specific resource types        |
| StartDate       | Time     | Cost query start date                    |
| EndDate         | Time     | Cost query end date                      |

**Credential Sources** (environment variables):

| Provider | Variables                                              |
|----------|--------------------------------------------------------|
| AWS      | AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION   |
| Azure    | AZURE_CLIENT_ID, AZURE_CLIENT_SECRET, AZURE_TENANT_ID  |
| GCP      | GOOGLE_APPLICATION_CREDENTIALS                         |

---

### SuiteReport

Aggregate report for conformance suite run.

| Field        | Type         | Description                           |
|--------------|--------------|---------------------------------------|
| SuiteName    | string       | "conformance" or "e2e"                |
| Plugin       | PluginInfo   | Plugin metadata                       |
| Results      | []TestResult | Individual test results               |
| Summary      | Summary      | Aggregate counts                      |
| StartTime    | Time         | Suite start timestamp                 |
| EndTime      | Time         | Suite end timestamp                   |
| TotalTime    | Duration     | Total execution time                  |

**Summary**:

| Field   | Type | Description           |
|---------|------|-----------------------|
| Total   | int  | Total tests executed  |
| Passed  | int  | Tests with pass status|
| Failed  | int  | Tests with fail status|
| Skipped | int  | Tests with skip status|
| Errors  | int  | Tests with error status|

---

## Relationships

```text
ConformanceSuiteConfig
    └── PluginUnderTest (1:1)

SuiteReport
    ├── PluginUnderTest (1:1)
    └── TestResult (1:N)
        └── ConformanceTestCase (N:1, by TestName)

E2ETestConfig
    └── (standalone, no foreign keys)
```

---

## Go Type Definitions

```go
// Status represents test outcome.
type Status string

const (
    StatusPass  Status = "pass"
    StatusFail  Status = "fail"
    StatusSkip  Status = "skip"
    StatusError Status = "error"
)

// Category represents test grouping.
type Category string

const (
    CategoryProtocol    Category = "protocol"
    CategoryPerformance Category = "performance"
    CategoryError       Category = "error"
    CategoryContext     Category = "context"
)

// Verbosity represents logging detail level.
type Verbosity string

const (
    VerbosityQuiet   Verbosity = "quiet"
    VerbosityNormal  Verbosity = "normal"
    VerbosityVerbose Verbosity = "verbose"
    VerbosityDebug   Verbosity = "debug"
)

// CommMode represents plugin communication mode.
type CommMode string

const (
    CommModeTCP   CommMode = "tcp"
    CommModeStdio CommMode = "stdio"
)
```
