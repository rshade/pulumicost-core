# Integration Tests

This directory contains integration tests for PulumiCost Core, verifying cross-component communication and end-to-end workflows.

## Test Organization

```
test/integration/
├── helpers/           # Test helpers and utilities
│   └── cli_helper.go  # CLI command execution helper
├── cli/              # CLI → Engine workflow tests
│   └── cli_workflow_test.go
├── config/           # Configuration loading tests
│   └── config_loading_test.go
├── errors/           # Error propagation tests
│   └── error_propagation_test.go
├── output/           # Output format generation tests
│   └── output_format_test.go
├── plugin/           # Plugin communication and management tests
│   ├── plugin_communication_test.go  # gRPC integration
│   ├── setup_test.go                 # Mock registry infrastructure
│   ├── init_test.go                  # Plugin init command tests
│   ├── install_test.go               # Plugin install command tests
│   ├── update_test.go                # Plugin update command tests
│   ├── remove_test.go                # Plugin remove command tests
│   └── concurrency_test.go           # Concurrent operation tests
└── e2e/              # End-to-end tests (Phase 8)
    └── cli_workflow_test.go
```

## Test Categories

### 1. CLI Workflow Tests (`cli/`)
Tests complete CLI command flows from argument parsing through engine execution to output rendering.

**Coverage:**
- Projected cost calculation
- Table/JSON/NDJSON output formats
- Error handling (missing files, invalid JSON)
- Help and version commands
- Plugin list/validate commands

### 2. Configuration Loading Tests (`config/`)
Tests configuration loading and precedence across components.

**Coverage:**
- Default configuration values
- Environment variable loading
- Config file loading (YAML)
- Configuration precedence (flags > env > file > defaults)
- Invalid configuration handling

### 3. Error Propagation Tests (`errors/`)
Tests error handling and propagation across component boundaries.

**Coverage:**
- File I/O errors (missing files)
- JSON parsing errors
- Invalid plan structures
- Plugin communication errors (timeout, protocol, invalid data, unavailable)
- Graceful degradation (fallback to "none" adapter)

### 4. Output Format Tests (`output/`)
Tests output rendering in different formats (JSON, table, NDJSON).

**Coverage:**
- JSON structure and field validation
- Table formatting and headers
- NDJSON line-by-line parsing
- Empty result handling
- Currency formatting and cost precision
- Consistency across formats

### 5. Plugin Communication Tests (`plugin/`)
Tests integration between engine and plugin layer via gRPC.

**Coverage:**
- Basic gRPC connection
- Projected cost flow
- Actual cost flow
- Error injection and handling
- Timeout and latency simulation

### 6. Plugin Management Tests (`plugin/`)
Tests plugin lifecycle commands (init, install, update, remove).

**Coverage:**
- Plugin initialization with scaffolding (`init_test.go`)
- Plugin installation from mock registry (`install_test.go`)
- Plugin updates with version control (`update_test.go`)
- Plugin removal with config cleanup (`remove_test.go`)
- Concurrent operations (`concurrency_test.go`)
- Mock GitHub Release API (`setup_test.go`)

## Test Helpers

### CLIHelper (`helpers/cli_helper.go`)

Provides utilities for testing CLI commands programmatically.

**Key Methods:**
- `Execute(args...)` - Execute CLI command and capture output
- `ExecuteOrFail(args...)` - Execute and fail test on error
- `ExecuteExpectError(args...)` - Execute expecting failure
- `ExecuteJSON(v, args...)` - Execute and unmarshal JSON output
- `CreateTempFile(content)` - Create temporary test file
- `CreateTempDir()` - Create temporary test directory
- `WithEnv(env, fn)` - Execute with temporary environment variables
- `AssertContains(output, expected)` - Assert output contains string
- `AssertJSONField(output, field, expected)` - Assert JSON field value

**Example Usage:**
```go
func TestMyCommand(t *testing.T) {
    h := helpers.NewCLIHelper(t)

    // Create test plan file
    planFile := h.CreateTempFile(`{"resources": []}`)

    // Execute command with JSON output
    var result map[string]interface{}
    h.ExecuteJSONOrFail(&result, "cost", "projected", "--pulumi-json", planFile)

    // Verify results
    assert.Contains(t, result, "summary")
}
```

## Running Integration Tests

### Run All Integration Tests
```bash
go test -v ./test/integration/...
```

### Run Specific Test Category
```bash
go test -v ./test/integration/cli/...       # CLI workflow tests
go test -v ./test/integration/config/...    # Config loading tests
go test -v ./test/integration/errors/...    # Error propagation tests
go test -v ./test/integration/output/...    # Output format tests
go test -v ./test/integration/plugin/...    # Plugin communication tests
```

### Run Specific Test
```bash
go test -v ./test/integration/cli -run TestCLIWorkflow_ProjectedCost
```

### Run with Coverage
```bash
go test -coverprofile=coverage.out ./test/integration/...
go tool cover -html=coverage.out
```

## Test Patterns

### Pattern 1: CLI Command Testing
```go
func TestCLI_MyCommand(t *testing.T) {
    h := helpers.NewCLIHelper(t)

    // Create test input
    input := h.CreateTempFile(`{"test": "data"}`)

    // Execute command
    output, err := h.Execute("my-command", "--input", input, "--output", "json")
    require.NoError(t, err)

    // Verify output
    var result map[string]interface{}
    err = json.Unmarshal([]byte(output), &result)
    require.NoError(t, err)
    assert.Equal(t, "expected", result["field"])
}
```

### Pattern 2: Error Testing
```go
func TestCLI_ErrorHandling(t *testing.T) {
    h := helpers.NewCLIHelper(t)

    // Execute command expecting error
    errMsg := h.ExecuteExpectError("my-command", "--invalid-flag")

    // Verify error message
    assert.Contains(t, errMsg, "unknown flag")
}
```

### Pattern 3: Environment Variable Testing
```go
func TestCLI_WithEnvironment(t *testing.T) {
    h := helpers.NewCLIHelper(t)

    // Set environment variables
    env := map[string]string{
        "MY_VAR": "test-value",
    }

    h.WithEnv(env, func() {
        // Execute command with env vars
        output := h.ExecuteOrFail("my-command")
        h.AssertContains(output, "test-value")
    })
}
```

### Pattern 4: Mock Plugin Integration
```go
func TestPlugin_Integration(t *testing.T) {
    // Start mock plugin server
    server, err := plugin.StartMockServerTCP()
    require.NoError(t, err)
    defer server.Stop()

    // Configure mock response
    server.Plugin.SetProjectedCostResponse("aws_instance",
        plugin.QuickResponse("USD", 100.0, 0.137))

    // Connect and test
    conn, err := grpc.NewClient(server.Address(),
        grpc.WithTransportCredentials(insecure.NewCredentials()))
    require.NoError(t, err)
    defer conn.Close()

    client := pb.NewCostSourceServiceClient(conn)
    // ... make gRPC calls
}
```

### Pattern 5: Output Format Validation
```go
func TestOutput_JSON(t *testing.T) {
    h := helpers.NewCLIHelper(t)

    output, err := h.Execute("my-command", "--output", "json")
    require.NoError(t, err)

    // Validate JSON structure
    var result map[string]interface{}
    err = json.Unmarshal([]byte(output), &result)
    require.NoError(t, err)

    // Verify required fields
    assert.Contains(t, result, "summary")
    assert.Contains(t, result, "resources")
}
```

## Best Practices

### 1. Use Test Helpers
Always use `CLIHelper` for CLI tests to ensure proper cleanup and error handling.

### 2. Test Isolation
Each test should be independent and not rely on state from other tests.

```go
// Good: Isolated test
func TestMyFeature(t *testing.T) {
    h := helpers.NewCLIHelper(t)
    tempFile := h.CreateTempFile("data")
    // Test uses its own resources
}

// Bad: Shared state
var sharedFile string // Don't share across tests
```

### 3. Cleanup with t.Cleanup
All test helpers automatically register cleanup functions:

```go
func (h *CLIHelper) CreateTempFile(content string) string {
    tmpFile, _ := os.CreateTemp("", "test-*")
    // ...
    h.t.Cleanup(func() {
        os.Remove(tmpFile.Name())  // Auto-cleanup
    })
    return tmpFile.Name()
}
```

### 4. Clear Error Messages
Use descriptive assertion messages:

```go
// Good
require.NoError(t, err, "Failed to parse JSON output: %s", output)

// Less helpful
require.NoError(t, err)
```

### 5. Test Both Success and Failure Paths
For each feature, test:
- ✅ Normal success path
- ❌ Error conditions
- ⚠️ Edge cases (empty input, large input, etc.)

### 6. Use Fixtures
Store complex test data in `test/fixtures/`:

```go
planFile := filepath.Join("..", "..", "fixtures", "plans", "aws-simple-plan.json")
```

## Mock Plugin Testing

Integration tests use the Phase 4 mock plugin infrastructure:

**Available Features:**
- **Scenarios**: Success, PartialData, HighCost, ZeroCost, MultiCurrency
- **Error Injection**: Timeout, Protocol, InvalidData, Unavailable
- **Latency Simulation**: Add delay in milliseconds
- **Custom Responses**: Configure projected/actual cost responses
- **TCP Server**: `StartMockServerTCP()` for real network testing
- **Bufconn Server**: `StartMockServer()` for in-memory testing

**Example:**
```go
// Start TCP server for integration testing
server, err := plugin.StartMockServerTCP()
require.NoError(t, err)
defer server.Stop()

// Configure scenario
server.Plugin.ConfigureScenario(plugin.ScenarioSuccess)

// Or configure custom response
server.Plugin.SetProjectedCostResponse("aws_instance",
    plugin.QuickResponse("USD", 73.0, 0.10))

// Get server address for connections
address := server.Address()  // e.g., "127.0.0.1:12345"
```

## Debugging Failed Tests

### 1. Run with Verbose Output
```bash
go test -v ./test/integration/cli -run TestCLIWorkflow_ProjectedCost
```

### 2. Check stderr Output
```go
output, err := h.Execute("cost", "projected", ...)
if err != nil {
    t.Logf("Stderr: %s", h.Stderr())  // Capture error output
}
```

### 3. Print JSON for Inspection
```go
var result map[string]interface{}
err := json.Unmarshal([]byte(output), &result)
if err != nil {
    t.Logf("Raw output: %s", output)  // See what was actually returned
}
```

### 4. Use require vs assert
- `require.*` - Stops test immediately on failure
- `assert.*` - Continues test, reports failure at end

## CI/CD Integration

Integration tests run automatically in CI:
- ✅ All integration tests must pass
- ✅ Coverage thresholds enforced (61% overall, 70% critical paths)
- ✅ Test failures block PR merging
- ✅ Coverage reports added as PR comments

## Future Enhancements

### Planned Improvements
- [ ] Add integration tests for actual cost commands
- [x] Add integration tests for plugin init commands
- [ ] Add performance benchmarks for integration flows
- [ ] Add integration tests with real plugin binaries
- [ ] Add integration tests for multi-plugin scenarios
- [ ] Add file locking support for concurrent plugin operations (documented in `concurrency_test.go`)

## Contributing

When adding new integration tests:
1. Choose the appropriate test category directory
2. Use the `CLIHelper` for CLI tests
3. Follow existing test patterns
4. Add test to this README under appropriate section
5. Ensure tests pass locally before committing
6. Verify tests pass in CI

## Related Documentation

- [Unit Tests](../unit/README.md) - Component-level testing
- [Mock Plugin](../mocks/plugin/README.md) - Mock plugin documentation
- [E2E Tests](e2e/README.md) - End-to-end workflow testing
- [Test Fixtures](../fixtures/README.md) - Test data documentation
