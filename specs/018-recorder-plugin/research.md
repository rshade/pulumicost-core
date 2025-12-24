# Research: Reference Recorder Plugin for DevTools

**Date**: 2025-12-11
**Feature**: 018-recorder-plugin

## Research Topics

### 1. pluginsdk v0.4.6 Patterns

**Research Task**: Find best practices for using pluginsdk v0.4.6 request validation helpers.

**Decision**: Use pluginsdk.BasePlugin with embedded struct pattern, register all providers with wildcard matcher, and use request validation helpers before processing.

**Rationale**: The aws-example plugin demonstrates this pattern effectively. The v0.4.6 release adds validation helpers that should be demonstrated in the recorder as a reference implementation.

**Pattern**:

```go
type RecorderPlugin struct {
    *pluginsdk.BasePlugin
    config *Config
    recorder *Recorder
}

func NewRecorderPlugin(cfg *Config) *RecorderPlugin {
    base := pluginsdk.NewBasePlugin("recorder")
    // Register as universal handler (accepts all resources)
    base.Matcher().AddProvider("*")

    return &RecorderPlugin{
        BasePlugin: base,
        config: cfg,
        recorder: NewRecorder(cfg.OutputDir),
    }
}
```

**Alternatives Considered**:

- Direct protobuf implementation without SDK: Rejected - violates reference implementation goal
- Selective resource registration: Rejected - recorder should capture ALL requests

### 2. JSON Serialization for Protobuf Messages

**Research Task**: Best practices for serializing gRPC protobuf messages to human-readable JSON.

**Decision**: Use `protojson.Marshal` from `google.golang.org/protobuf/encoding/protojson` with indent formatting.

**Rationale**: Standard protobuf JSON encoding ensures field names match proto definitions and handles all proto types correctly (timestamps, enums, oneofs).

**Pattern**:

```go
import "google.golang.org/protobuf/encoding/protojson"

func (r *Recorder) SerializeRequest(req proto.Message) ([]byte, error) {
    opts := protojson.MarshalOptions{
        Multiline:       true,
        Indent:          "  ",
        EmitUnpopulated: true,  // Include zero values for completeness
    }
    return opts.Marshal(req)
}
```

**Alternatives Considered**:

- Standard `encoding/json`: Rejected - doesn't handle proto-specific types properly
- Binary protobuf format: Rejected - not human-readable, defeats inspection purpose

### 3. Thread-Safe File Writing

**Research Task**: Best practices for concurrent file writes with unique filenames.

**Decision**: Use ULID for unique identifiers (time-ordered, no collisions), sync.Mutex for recorder state, individual file writes are atomic via os.WriteFile.

**Rationale**: ULIDs are monotonically increasing and contain timestamp, making recorded files naturally sorted by time. Mutex protects shared state; file writes are atomic.

**Pattern**:

```go
import "github.com/oklog/ulid/v2"

func generateFilename(method string) string {
    timestamp := time.Now().UTC().Format("20060102T150405Z")
    id := ulid.Make()
    return fmt.Sprintf("%s_%s_%s.json", timestamp, method, id.String())
}
```

**Alternatives Considered**:

- UUID v4: Rejected - not time-ordered, harder to correlate with events
- Timestamp only: Rejected - potential collisions under concurrent load
- Sequence numbers: Rejected - requires persistent state

### 4. Mock Response Generation

**Research Task**: Best practices for generating randomized but valid cost responses.

**Decision**: Use deterministic ranges for cost values ($0.01 - $1000/month for projected, $0.001 - $100/day for actual). Use SDK Calculator for response building.

**Rationale**: Ranges should be realistic enough to test aggregation logic but clearly fake. SDK Calculator ensures protocol compliance.

**Pattern**:

```go
import "math/rand"

func (m *Mocker) GenerateProjectedCost() float64 {
    // Range: $0.01 to $1000 per month (log scale for realistic distribution)
    min := 0.01
    max := 1000.0
    return min * math.Pow(max/min, rand.Float64())
}

func (m *Mocker) CreateProjectedCostResponse() *pb.GetProjectedCostResponse {
    cost := m.GenerateProjectedCost()
    return m.plugin.Calculator().CreateProjectedCostResponse(
        "USD",
        cost / 730, // Convert monthly to hourly
        fmt.Sprintf("Mock cost: $%.2f/month", cost),
    )
}
```

**Alternatives Considered**:

- Fixed responses: Rejected - doesn't test aggregation edge cases
- Truly random (negative values possible): Rejected - would break Core assertions
- Seeded random: Considered for reproducibility, but not required for dev tool

### 5. Environment Variable Configuration

**Research Task**: Best practices for plugin configuration via environment variables.

**Decision**: Use standard `os.Getenv` with defaults, parse booleans case-insensitively, log configuration at startup.

**Rationale**: Consistent with Core patterns (PULUMICOST_* prefix). Simple, no external config library needed.

**Pattern**:

```go
type Config struct {
    OutputDir    string
    MockResponse bool
}

func LoadConfig() *Config {
    cfg := &Config{
        OutputDir:    "./recorded_data",
        MockResponse: false,
    }

    if dir := os.Getenv("PULUMICOST_RECORDER_OUTPUT_DIR"); dir != "" {
        cfg.OutputDir = dir
    }

    if mock := os.Getenv("PULUMICOST_RECORDER_MOCK_RESPONSE"); mock != "" {
        cfg.MockResponse = strings.EqualFold(mock, "true") || mock == "1"
    }

    return cfg
}
```

**Alternatives Considered**:

- YAML config file: Rejected - adds complexity, env vars match plugin pattern
- Viper/Cobra flags: Rejected - plugin receives only --port/--stdio from launcher

### 6. Graceful Shutdown Pattern

**Research Task**: Best practices for graceful shutdown in gRPC plugins.

**Decision**: Use context cancellation with signal handling, defer cleanup in main, flush recorder on shutdown.

**Rationale**: Matches aws-example pattern. Ensures pending writes complete before exit.

**Pattern**:

```go
func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Signal handling
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigCh
        cancel()
    }()

    plugin := NewRecorderPlugin(LoadConfig())
    defer plugin.Shutdown() // Flush pending writes

    config := pluginsdk.ServeConfig{
        Plugin: plugin,
        Port:   0,
    }

    if err := pluginsdk.Serve(ctx, config); err != nil {
        os.Exit(1)
    }
}
```

**Alternatives Considered**:

- No explicit shutdown: Rejected - could lose buffered data
- Fixed timeout: Considered but context cancellation is cleaner

### 7. Plugin Manifest Structure

**Research Task**: Best practices for plugin.manifest.json format.

**Decision**: Follow manifest.yaml pattern from aws-example, use JSON for consistency with other JSON outputs.

**Rationale**: Registry expects optional manifest for validation; JSON is more portable.

**Pattern**:

```json
{
  "name": "recorder",
  "version": "0.1.0",
  "description": "Reference plugin that records all gRPC requests and optionally returns mock responses",
  "author": "PulumiCost Team",
  "supported_providers": ["*"],
  "protocols": ["grpc"],
  "binary": "pulumicost-plugin-recorder",
  "metadata": {
    "repository": "https://github.com/rshade/pulumicost-core",
    "docs": "https://github.com/rshade/pulumicost-core/tree/main/plugins/recorder",
    "reference_implementation": true
  }
}
```

**Alternatives Considered**:

- YAML format: Rejected - JSON matches other recorded outputs
- No manifest: Rejected - spec requires FR-014

### 8. Build System Integration

**Research Task**: Best practices for adding plugin build target to Makefile.

**Decision**: Add `build-recorder` target that builds to `bin/pulumicost-plugin-recorder`, include in CI workflow.

**Rationale**: Consistent with existing build patterns, cross-platform via GOOS/GOARCH.

**Pattern**:

```makefile
.PHONY: build-recorder
build-recorder:
	go build -o bin/pulumicost-plugin-recorder ./plugins/recorder

.PHONY: build-all
build-all: build build-recorder
```

**Alternatives Considered**:

- Separate Makefile in plugins/recorder: Rejected - harder to maintain
- Shell script: Rejected - Makefile is standard in this project

## Summary

All research topics resolved. No blocking unknowns remain. Key decisions:

1. **SDK Pattern**: Embedded BasePlugin with wildcard provider matcher
2. **Serialization**: protojson with indentation
3. **Unique IDs**: ULID for time-ordered, collision-free filenames
4. **Mock Values**: Log-scale random in realistic ranges
5. **Config**: Environment variables with PULUMICOST_RECORDER_* prefix
6. **Shutdown**: Context cancellation with signal handling
7. **Manifest**: JSON format with reference implementation metadata
8. **Build**: Makefile target `build-recorder`
