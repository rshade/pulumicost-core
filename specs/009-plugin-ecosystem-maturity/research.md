# Research: Plugin Ecosystem Maturity

**Feature**: 102-plugin-ecosystem-maturity
**Date**: 2025-12-02

## JUnit XML Report Generation

### Decision: go-junit-report v2 for pipeline integration

**Rationale**: Decouples test execution from report generation, allowing
standard `go test` to run normally while separately generating JUnit XML.
Works well with existing CI/CD pipeline.

**Alternatives Considered**:

- **gotestsum**: All-in-one wrapper but adds external dependency
  as test runner replacement
- **Custom implementation**: More control but significant effort
  for a standard format

**Implementation Pattern**:

```bash
go test -json ./internal/conformance/... 2>&1 | \
  go-junit-report -parser gojson > conformance-report.xml
```

For JSON output, simply write to file directly from Go.

---

## Test Case Organization

### Decision: Table-driven tests with t.Run() subtests

**Rationale**: Idiomatic Go pattern that enables selective test execution
via `-run` flags, clear test naming, and amortized test infrastructure
cost across all cases.

**Structure**:

```go
var protocolTests = []struct {
    name     string
    category string  // "protocol", "performance", "error"
    test     func(t *testing.T, client proto.CostSourceClient)
}{
    {"Name_ReturnsPluginIdentifier", "protocol", testNameRPC},
    {"GetProjectedCost_ValidResource", "protocol", testProjectedCost},
    // ...
}

for _, tc := range protocolTests {
    t.Run(tc.name, func(t *testing.T) {
        tc.test(t, client)
    })
}
```

---

## Plugin Process Lifecycle

### Decision: exec.CommandContext with WaitDelay (Go 1.20+)

**Rationale**: Built-in Go stdlib approach. WaitDelay ensures cleanup
even if plugin pipes stay open. Process groups enable reliable child
process cleanup on Unix.

**Pattern**:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

cmd := exec.CommandContext(ctx, pluginPath, "--port", strconv.Itoa(port))
cmd.WaitDelay = 5 * time.Second  // Grace period before SIGKILL

// Unix: kill entire process group
if runtime.GOOS != "windows" {
    cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
```

**Crash Handling** (per FR-020):

- Catch process exit via cmd.Wait()
- Log exit code, signal, and last request
- Restart plugin for remaining tests
- Mark current test as failed with crash details

**Alternatives Considered**:

- **HashiCorp go-plugin**: Full-featured but heavyweight for
  conformance testing use case
- **Manual process management**: More work, less reliable cleanup

---

## Protocol Version Checking

### Decision: gRPC metadata handshake on first RPC call

**Rationale**: Simple, no additional RPC methods needed. Plugin returns
version info in Name() response or via metadata. Conformance suite
validates before running tests.

**Implementation**:

```go
// Add version to outgoing context
ctx = metadata.AppendToOutgoingContext(ctx,
    "x-pulumicost-protocol-version", "1.0")

// Name() response includes version info
nameResp, err := client.Name(ctx, &proto.Empty{})
if nameResp.ProtocolVersion != expectedVersion {
    return fmt.Errorf("protocol version mismatch: got %s, want %s",
        nameResp.ProtocolVersion, expectedVersion)
}
```

**Alternatives Considered**:

- **Dedicated Version RPC**: Cleaner separation but requires
  protocol change in pulumicost-spec
- **HTTP/2 ALPN**: Works for TLS connections but we use plaintext
  for local plugins

---

## Verbosity Levels

### Decision: Standard Go log levels via zerolog

**Rationale**: Project already uses zerolog (v1.34.0). Conformance suite
reuses existing logging infrastructure.

**Mapping**:

| Level   | Output                                    |
|---------|-------------------------------------------|
| quiet   | Pass/fail summary only                    |
| normal  | Test names + results (default)            |
| verbose | + request/response payloads               |
| debug   | + timing, connection state, retry details |

**Implementation**: Use `--verbosity` flag on CLI command, maps to
zerolog level in conformance runner.

---

## E2E Testing Credential Handling

### Decision: Environment variables only

**Rationale**: Simple, widely supported by CI systems (GitHub Actions
secrets, GitLab CI variables). No additional dependencies.

**Variables**:

```bash
# AWS
AWS_ACCESS_KEY_ID
AWS_SECRET_ACCESS_KEY
AWS_REGION

# Azure
AZURE_CLIENT_ID
AZURE_CLIENT_SECRET
AZURE_TENANT_ID
AZURE_SUBSCRIPTION_ID

# GCP
GOOGLE_APPLICATION_CREDENTIALS  # Path to service account JSON
```

**Skip Logic**: Tests skip gracefully when credentials absent:

```go
func TestAWSCostAPI(t *testing.T) {
    if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
        t.Skip("AWS credentials not configured")
    }
    // ... test logic
}
```

---

## Output Format Implementation

### Decision: Both JUnit XML and JSON

**Rationale**: JUnit XML for CI integration (GitHub Actions, Jenkins),
JSON for programmatic access and custom tooling.

**JUnit XML Structure**:

```xml
<testsuite name="conformance" tests="20" failures="1" time="45.2">
  <testcase name="Name_ReturnsPluginIdentifier" time="0.05"/>
  <testcase name="GetProjectedCost_InvalidResource" time="0.12">
    <failure message="expected NotFound, got OK"/>
  </testcase>
</testsuite>
```

**JSON Structure**:

```json
{
  "suite": "conformance",
  "plugin": "/path/to/plugin",
  "version": "1.0.0",
  "results": [
    {"name": "Name_ReturnsPluginIdentifier", "status": "pass", "duration_ms": 50},
    {"name": "GetProjectedCost_InvalidResource", "status": "fail",
     "error": "expected NotFound, got OK", "duration_ms": 120}
  ],
  "summary": {"total": 20, "passed": 19, "failed": 1}
}
```

---

## References

- [Go Table-Driven Tests](https://go.dev/wiki/TableDrivenTests)
- [go-junit-report](https://github.com/jstemmer/go-junit-report)
- [gRPC-Go Metadata](https://github.com/grpc/grpc-go/blob/master/Documentation/grpc-metadata.md)
- [exec.CommandContext Best Practices](https://pkg.go.dev/os/exec#CommandContext)
- [zerolog Documentation](https://github.com/rs/zerolog)
