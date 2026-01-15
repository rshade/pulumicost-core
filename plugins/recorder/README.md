# Recorder Plugin

A reference implementation plugin that records all gRPC requests to JSON files and optionally returns mock cost responses. This plugin serves as both a developer tool for inspecting Core-to-plugin data shapes and a canonical reference implementation demonstrating pluginsdk v0.4.6 patterns.

## Features

- **Request Recording**: Captures all gRPC requests to JSON files for inspection
- **Mock Responses**: Optionally returns randomized but valid cost responses
- **Reference Implementation**: Demonstrates pluginsdk v0.4.6 best practices
- **Thread-Safe**: Safe for concurrent use with sync.Mutex protection
- **Graceful Shutdown**: Proper signal handling and cleanup

## Installation

```bash
# From finfocus-core repository root
make install-recorder

# Verify installation
./bin/finfocus plugin list
```

This builds the plugin and installs it to `~/.finfocus/plugins/recorder/0.1.0/`.

## Configuration

Configuration is via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `FINFOCUS_RECORDER_OUTPUT_DIR` | `./recorded_data` | Directory for recorded JSON files |
| `FINFOCUS_RECORDER_MOCK_RESPONSE` | `false` | Enable randomized mock responses |

## Usage

### Basic Recording (Default Mode)

```bash
# Set output directory (optional)
export FINFOCUS_RECORDER_OUTPUT_DIR=./my-recordings

# Run cost calculation - requests will be recorded
./bin/finfocus cost projected --pulumi-json plan.json

# View recorded files
ls -la ./my-recordings/
cat ./my-recordings/*.json | jq .
```

### Mock Response Mode

```bash
# Enable mock responses
export FINFOCUS_RECORDER_MOCK_RESPONSE=true

# Run cost calculation - will return randomized costs
./bin/finfocus cost projected --pulumi-json plan.json --output json
```

## Recorded Request Format

Each request creates a JSON file with the following structure:

```json
{
  "timestamp": "2025-12-11T14:30:52Z",
  "method": "GetProjectedCost",
  "requestId": "01JEK7X2J3K4M5N6P7Q8R9S0T1",
  "request": {
    "resource": {
      "resourceType": "aws:ec2:Instance",
      "provider": "aws",
      "sku": "t3.medium",
      "region": "us-east-1"
    }
  },
  "metadata": {
    "receivedAt": "2025-12-11T14:30:52.123456Z",
    "processingTimeMs": 2
  }
}
```

### File Naming

Files are named with the format: `<timestamp>_<method>_<ulid>.json`

Example: `20251211T143052Z_GetProjectedCost_01JEK7X2J3K4M5N6P7Q8R9S0T1.json`

## Use Cases

### 1. Debug Plugin Integration

See exactly what data Core sends to plugins:

```bash
FINFOCUS_RECORDER_OUTPUT_DIR=./debug ./bin/finfocus cost projected --pulumi-json my-plan.json
cat ./debug/*.json | jq '.request.resource'
```

### 2. Test Core Aggregation

Generate mock data to test output formatting:

```bash
FINFOCUS_RECORDER_MOCK_RESPONSE=true ./bin/finfocus cost projected \
  --pulumi-json plan.json \
  --output table
```

### 3. Contract Testing

Use as a reference plugin in integration tests:

```bash
# In test setup
export FINFOCUS_RECORDER_OUTPUT_DIR=/tmp/test-recordings
export FINFOCUS_RECORDER_MOCK_RESPONSE=true

# Run tests
go test ./test/integration/... -v
```

## SDK Patterns Demonstrated

This plugin demonstrates the following pluginsdk v0.4.6 patterns:

1. **BasePlugin Embedding**: Using `*pluginsdk.BasePlugin` for common functionality
2. **Request Validation**: Using `pluginsdk.ValidateProjectedCostRequest()` and similar helpers
3. **Graceful Shutdown**: Context-based cancellation with signal handling
4. **Structured Logging**: Using zerolog for consistent log output
5. **Thread Safety**: Using sync.Mutex for concurrent request handling

## Development

### Running Tests

```bash
# Run all tests
go test -v ./plugins/recorder/...

# Run with coverage
go test -cover ./plugins/recorder/...

# Run benchmarks
go test -bench=. ./plugins/recorder/...
```

### Code Structure

```text
plugins/recorder/
├── cmd/
│   └── main.go           # Plugin entry point
├── config.go             # Configuration loading
├── config_test.go        # Config tests
├── mocker.go             # Mock response generation
├── mocker_test.go        # Mocker tests
├── plugin.go             # Main plugin implementation
├── plugin_test.go        # Plugin tests
├── recorder.go           # Request recording
├── recorder_test.go      # Recorder tests
├── plugin.manifest.json  # Plugin metadata
└── README.md             # This file
```

## Troubleshooting

### Plugin Not Found

```bash
# Check plugin directory structure
ls -la ~/.finfocus/plugins/recorder/

# Verify binary is executable
chmod +x ~/.finfocus/plugins/recorder/0.1.0/finfocus-plugin-recorder
```

### No Files Being Recorded

```bash
# Check output directory exists and is writable
mkdir -p "$FINFOCUS_RECORDER_OUTPUT_DIR"
touch "$FINFOCUS_RECORDER_OUTPUT_DIR/test.txt" && rm "$FINFOCUS_RECORDER_OUTPUT_DIR/test.txt"

# Check plugin is being used
./bin/finfocus cost projected --debug --pulumi-json plan.json 2>&1 | grep recorder
```

### Mock Responses Not Working

```bash
# Verify environment variable is set correctly
echo $FINFOCUS_RECORDER_MOCK_RESPONSE  # Should be "true"

# Case-insensitive values work: true, TRUE, 1, yes, on
export FINFOCUS_RECORDER_MOCK_RESPONSE=TRUE
```

## License

Part of FinFocus Core. See repository license for details.
