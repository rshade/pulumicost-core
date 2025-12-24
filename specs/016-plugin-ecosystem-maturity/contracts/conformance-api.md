# Conformance API Contract

**Feature**: 016-plugin-ecosystem-maturity
**Date**: 2025-12-02

## Overview

This document defines the API contract for the conformance testing framework.
The conformance suite is invoked via CLI and programmatic Go API.

---

## CLI Interface

### Command: `pulumicost plugin conformance`

Run conformance tests against a plugin binary.

```bash
pulumicost plugin conformance [flags] <plugin-path>
```

**Arguments**:

| Argument      | Required | Description                    |
|---------------|----------|--------------------------------|
| `plugin-path` | Yes      | Path to plugin binary to test  |

**Flags**:

| Flag            | Type   | Default  | Description                       |
|-----------------|--------|----------|-----------------------------------|
| `--mode`        | string | tcp      | Communication mode: tcp, stdio    |
| `--verbosity`   | string | normal   | Output detail: quiet, normal, verbose, debug |
| `--output`      | string | table    | Output format: table, json, junit |
| `--output-file` | string | (stdout) | Write output to file              |
| `--timeout`     | string | 5m       | Global suite timeout              |
| `--category`    | string | (all)    | Filter by category (repeatable)   |
| `--filter`      | string | (all)    | Regex filter for test names       |

**Examples**:

```bash
# Basic conformance check
pulumicost plugin conformance ./plugins/aws-cost

# Verbose output with JSON
pulumicost plugin conformance --verbosity verbose --output json ./plugins/aws-cost

# Filter to protocol tests only
pulumicost plugin conformance --category protocol ./plugins/aws-cost

# JUnit XML for CI
pulumicost plugin conformance --output junit --output-file report.xml ./plugins/aws-cost

# Use stdio mode
pulumicost plugin conformance --mode stdio ./plugins/aws-cost
```

**Exit Codes**:

| Code | Meaning                                    |
|------|--------------------------------------------|
| 0    | All tests passed                           |
| 1    | One or more tests failed                   |
| 2    | Plugin crashed or connection failed        |
| 3    | Protocol version mismatch                  |
| 4    | Invalid arguments or configuration         |

---

## Go API

### Package: `internal/conformance`

#### NewSuite

```go
// NewSuite creates a conformance test suite with the given configuration.
func NewSuite(cfg SuiteConfig) (*Suite, error)
```

**SuiteConfig**:

```go
type SuiteConfig struct {
    PluginPath  string        // Required: path to plugin binary
    CommMode    CommMode      // Optional: "tcp" (default) or "stdio"
    Verbosity   Verbosity     // Optional: logging level (default: normal)
    Timeout     time.Duration // Optional: suite timeout (default: 5m)
    Categories  []Category    // Optional: filter categories
    TestFilter  string        // Optional: regex for test names
    Logger      zerolog.Logger // Optional: custom logger
}
```

**Returns**:

- `*Suite`: Configured test suite
- `error`: If plugin path invalid or configuration error

---

#### Suite.Run

```go
// Run executes all conformance tests and returns the report.
func (s *Suite) Run(ctx context.Context) (*Report, error)
```

**Parameters**:

- `ctx`: Context for cancellation and timeout

**Returns**:

- `*Report`: Test results and summary
- `error`: Only for infrastructure failures (not test failures)

---

#### Report.WriteJSON

```go
// WriteJSON writes the report as JSON to the given writer.
func (r *Report) WriteJSON(w io.Writer) error
```

---

#### Report.WriteJUnit

```go
// WriteJUnit writes the report as JUnit XML to the given writer.
func (r *Report) WriteJUnit(w io.Writer) error
```

---

#### Report.WriteTable

```go
// WriteTable writes the report as human-readable table to the given writer.
func (r *Report) WriteTable(w io.Writer) error
```

---

### Usage Example

```go
package main

import (
    "context"
    "os"
    "time"

    "github.com/rshade/pulumicost-core/internal/conformance"
)

func main() {
    suite, err := conformance.NewSuite(conformance.SuiteConfig{
        PluginPath: "./plugins/aws-cost",
        CommMode:   conformance.CommModeTCP,
        Verbosity:  conformance.VerbosityVerbose,
        Timeout:    5 * time.Minute,
        Categories: []conformance.Category{conformance.CategoryProtocol},
    })
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    report, err := suite.Run(ctx)
    if err != nil {
        log.Fatal(err)
    }

    // Output as JSON
    report.WriteJSON(os.Stdout)

    // Check results
    if report.Summary.Failed > 0 {
        os.Exit(1)
    }
}
```

---

## Output Formats

### Table (Human-Readable)

```text
CONFORMANCE TEST RESULTS
========================
Plugin: aws-cost v1.2.0 (protocol v1.0)
Mode:   TCP

TESTS
-----
✓ Name_ReturnsPluginIdentifier              [  50ms]
✓ Name_ReturnsProtocolVersion               [  45ms]
✓ GetProjectedCost_ValidResource            [ 120ms]
✗ GetProjectedCost_InvalidResource          [ 110ms]
  Error: expected NotFound, got InvalidArgument
✓ GetProjectedCost_BatchRequest             [ 890ms]
⊘ GetActualCost_RequiresCredentials         [  --  ] (skipped)

SUMMARY
-------
Total: 20 | Passed: 18 | Failed: 1 | Skipped: 1 | Duration: 4.5s
```

### JSON

```json
{
  "suite": "conformance",
  "plugin": {
    "path": "./plugins/aws-cost",
    "name": "aws-cost",
    "version": "1.2.0",
    "protocol_version": "1.0",
    "comm_mode": "tcp"
  },
  "results": [
    {
      "name": "Name_ReturnsPluginIdentifier",
      "category": "protocol",
      "status": "pass",
      "duration_ms": 50
    },
    {
      "name": "GetProjectedCost_InvalidResource",
      "category": "error",
      "status": "fail",
      "duration_ms": 110,
      "error": "expected NotFound, got InvalidArgument"
    }
  ],
  "summary": {
    "total": 20,
    "passed": 18,
    "failed": 1,
    "skipped": 1,
    "errors": 0
  },
  "duration_ms": 4500,
  "timestamp": "2025-12-02T10:30:00Z"
}
```

### JUnit XML

```xml
<?xml version="1.0" encoding="UTF-8"?>
<testsuites name="pulumicost-conformance" tests="20" failures="1"
            skipped="1" time="4.5">
  <testsuite name="conformance" tests="20" failures="1" skipped="1"
             time="4.5" timestamp="2025-12-02T10:30:00Z">
    <properties>
      <property name="plugin.name" value="aws-cost"/>
      <property name="plugin.version" value="1.2.0"/>
      <property name="protocol.version" value="1.0"/>
    </properties>
    <testcase name="Name_ReturnsPluginIdentifier"
              classname="protocol" time="0.05"/>
    <testcase name="GetProjectedCost_InvalidResource"
              classname="error" time="0.11">
      <failure message="expected NotFound, got InvalidArgument"
               type="AssertionError">
        Expected gRPC status NotFound but got InvalidArgument
      </failure>
    </testcase>
    <testcase name="GetActualCost_RequiresCredentials"
              classname="protocol" time="0">
      <skipped message="credentials not configured"/>
    </testcase>
  </testsuite>
</testsuites>
```
