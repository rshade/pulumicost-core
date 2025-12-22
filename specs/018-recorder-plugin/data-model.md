# Data Model: Reference Recorder Plugin

**Date**: 2025-12-11
**Feature**: 018-recorder-plugin

## Entities

### 1. Config

Runtime configuration loaded from environment variables.

| Field | Type | Source | Default | Description |
|-------|------|--------|---------|-------------|
| OutputDir | string | `PULUMICOST_RECORDER_OUTPUT_DIR` | `./recorded_data` | Directory for recorded JSON files |
| MockResponse | bool | `PULUMICOST_RECORDER_MOCK_RESPONSE` | `false` | Enable randomized mock responses |

**Validation Rules**:

- OutputDir: Must be a valid path; created if not exists
- MockResponse: Parsed from "true"/"false"/"1"/"0" (case-insensitive)

### 2. RecordedRequest

Represents a serialized gRPC request captured to disk.

| Field | Type | Description |
|-------|------|-------------|
| Timestamp | string (ISO8601) | When request was received (UTC) |
| Method | string | gRPC method name (e.g., "GetProjectedCost") |
| RequestID | string (ULID) | Unique identifier for this request |
| Request | object | Full protobuf request serialized as JSON |
| Metadata | RequestMetadata | Optional metadata about the request |

**File Naming**:

```text
<timestamp>_<method>_<ulid>.json
Example: 20251211T143052Z_GetProjectedCost_01JEK7X2J3K4M5N6P7Q8R9S0T1.json
```

**JSON Structure**:

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
      "region": "us-east-1",
      "tags": {
        "instanceType": "t3.medium",
        "environment": "production"
      }
    }
  },
  "metadata": {
    "receivedAt": "2025-12-11T14:30:52.123456Z",
    "processingTimeMs": 2
  }
}
```

### 3. RequestMetadata

Optional metadata captured alongside the request.

| Field | Type | Description |
|-------|------|-------------|
| ReceivedAt | string (RFC3339Nano) | Precise timestamp when request arrived |
| ProcessingTimeMs | int64 | Time spent processing request (milliseconds) |

### 4. MockCostResponse

Internal representation for generating mock responses (not persisted).

| Field | Type | Range | Description |
|-------|------|-------|-------------|
| MonthlyCost | float64 | $0.01 - $1000 | Randomized monthly cost |
| HourlyCost | float64 | Derived | MonthlyCost / 730 |
| Currency | string | "USD" | Always USD for mock responses |
| BillingDetail | string | - | Human-readable mock indicator |

**Generation Algorithm**:

```go
// Log-scale distribution for realistic cost spread
cost := minCost * math.Pow(maxCost/minCost, rand.Float64())
```

### 5. RecorderPlugin

Main plugin struct implementing CostSourceService.

| Field | Type | Description |
|-------|------|-------------|
| BasePlugin | *pluginsdk.BasePlugin | Embedded SDK base (provides Matcher, Calculator) |
| config | *Config | Runtime configuration |
| recorder | *Recorder | Request serialization handler |
| mocker | *Mocker | Mock response generator (nil if disabled) |
| mu | sync.Mutex | Protects shared state |

**State Transitions**:

```text
[Created] --> [Initialized] --> [Serving] --> [Shutting Down] --> [Stopped]
     |              |               |                |
     +-- NewRecorderPlugin()       |                |
                    +-- Serve()    |                |
                                   +-- Signal/Cancel
                                                    +-- Shutdown()
```

## Relationships

```text
RecorderPlugin
├── Config (1:1) - Configuration loaded at startup
├── Recorder (1:1) - Handles request serialization
│   └── RecordedRequest (1:N) - Each request creates one file
└── Mocker (0:1) - Only present if MockResponse=true
    └── MockCostResponse (transient) - Generated per request
```

## File System Layout

```text
<OutputDir>/
├── 20251211T143052Z_Name_01JEK7X2J3K4M5N6P7Q8R9S0T1.json
├── 20251211T143052Z_GetProjectedCost_01JEK7X2J3K4M5N6P7Q8R9S1T2.json
├── 20251211T143053Z_GetProjectedCost_01JEK7X2J3K4M5N6P7Q8R9S2T3.json
├── 20251211T143054Z_GetActualCost_01JEK7X2J3K4M5N6P7Q8R9S3T4.json
└── ... (one file per request)
```

## Protocol Buffer Types (from pulumicost-spec)

The plugin uses these types from `github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1`:

### Input Types

- `NameRequest` - Empty request for Name RPC
- `GetProjectedCostRequest` - Contains ResourceDescriptor
- `GetActualCostRequest` - Contains ResourceId, Start, End timestamps, Tags

### Output Types

- `NameResponse` - Contains Name string
- `GetProjectedCostResponse` - Contains CostPerMonth, UnitPrice, Currency, BillingDetail
- `GetActualCostResponse` - Contains Results array of ActualCostResult

### Shared Types

- `ResourceDescriptor` - ResourceType, Provider, Sku, Region, Tags map
- `ActualCostResult` - Source, Cost fields for historical data

## Validation Rules

### FR-015 Request Validation (pluginsdk v0.4.6)

Before processing any request, validate using SDK helpers:

```go
// GetProjectedCost validation
if err := pluginsdk.ValidateProjectedCostRequest(req); err != nil {
    // Record invalid request for debugging
    r.recorder.RecordInvalidRequest("GetProjectedCost", req, err)
    return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
}
```

### Output Directory Validation

- Create directory if not exists: `os.MkdirAll(dir, 0755)`
- Check write permissions on startup
- Log warning and disable recording if not writable
