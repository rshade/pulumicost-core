# Contract: CostSourceService Implementation

**Date**: 2025-12-11
**Feature**: 018-recorder-plugin

## Overview

The Recorder plugin implements the `CostSourceService` gRPC interface from finfocus-spec. This document specifies the exact contract the recorder fulfills.

## Service Definition

```protobuf
// From finfocus-spec/proto/finfocus/v1/cost_source.proto
service CostSourceService {
  rpc Name(NameRequest) returns (NameResponse);
  rpc GetProjectedCost(GetProjectedCostRequest) returns (GetProjectedCostResponse);
  rpc GetActualCost(GetActualCostRequest) returns (GetActualCostResponse);
}
```

## RPC Contracts

### 1. Name

**Purpose**: Return plugin identification.

**Request**: Empty `NameRequest`

**Response**:

```json
{
  "name": "recorder"
}
```

**Behavior**:

- Always returns "recorder"
- Records request to JSON file
- No error conditions

### 2. GetProjectedCost

**Purpose**: Calculate projected cost for a resource (or return mock).

**Request**:

```json
{
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
}
```

**Response (Mock Mode = false)**:

```json
{
  "costPerMonth": 0.0,
  "unitPrice": 0.0,
  "currency": "USD",
  "billingDetail": "Recorder plugin - mock responses disabled"
}
```

**Response (Mock Mode = true)**:

```json
{
  "costPerMonth": 73.42,
  "unitPrice": 0.1006,
  "currency": "USD",
  "billingDetail": "Mock cost: $73.42/month (recorder plugin)"
}
```

**Behavior**:

1. Validate request using pluginsdk v0.4.6 validation helpers
2. Record full request to JSON file
3. If MockResponse enabled: Generate randomized cost ($0.01-$1000/month)
4. If MockResponse disabled: Return zero cost with explanatory note
5. Return response

**Error Conditions**:

| Condition | gRPC Status | Message |
|-----------|-------------|---------|
| Invalid request (validation fails) | `INVALID_ARGUMENT` | Validation error details |
| Recording fails (disk full, etc.) | Continue with warning | Logged, not returned to client |

### 3. GetActualCost

**Purpose**: Return historical cost data (or mock).

**Request**:

```json
{
  "resourceId": "arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0",
  "start": "2025-12-01T00:00:00Z",
  "end": "2025-12-11T00:00:00Z",
  "tags": {
    "environment": "production"
  }
}
```

**Response (Mock Mode = false)**:

```json
{
  "results": []
}
```

**Response (Mock Mode = true)**:

```json
{
  "results": [
    {
      "source": "recorder-mock",
      "cost": 24.56
    }
  ]
}
```

**Behavior**:

1. Validate request using pluginsdk v0.4.6 validation helpers
2. Record full request to JSON file
3. If MockResponse enabled: Generate randomized cost result
4. If MockResponse disabled: Return empty results array
5. Return response

**Error Conditions**:

| Condition | gRPC Status | Message |
|-----------|-------------|---------|
| Invalid request (validation fails) | `INVALID_ARGUMENT` | Validation error details |
| Recording fails (disk full, etc.) | Continue with warning | Logged, not returned to client |

## Recorded Request Format

All requests are recorded regardless of mock mode setting.

**File Location**: `<OutputDir>/<timestamp>_<method>_<ulid>.json`

**Schema**:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["timestamp", "method", "requestId", "request"],
  "properties": {
    "timestamp": {
      "type": "string",
      "format": "date-time",
      "description": "ISO8601 timestamp when request was received"
    },
    "method": {
      "type": "string",
      "enum": ["Name", "GetProjectedCost", "GetActualCost"],
      "description": "gRPC method name"
    },
    "requestId": {
      "type": "string",
      "pattern": "^[0-9A-Z]{26}$",
      "description": "ULID unique identifier"
    },
    "request": {
      "type": "object",
      "description": "Full protobuf request as JSON (protojson format)"
    },
    "metadata": {
      "type": "object",
      "properties": {
        "receivedAt": {
          "type": "string",
          "format": "date-time"
        },
        "processingTimeMs": {
          "type": "integer",
          "minimum": 0
        }
      }
    }
  }
}
```

## Environment Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `FINFOCUS_RECORDER_OUTPUT_DIR` | string | `./recorded_data` | Directory for recorded files |
| `FINFOCUS_RECORDER_MOCK_RESPONSE` | bool | `false` | Enable mock response generation |

## Plugin Discovery

**Binary Name**: `finfocus-plugin-recorder`

**Installation Path**: `~/.finfocus/plugins/recorder/<version>/finfocus-plugin-recorder`

**Manifest** (`plugin.manifest.json`):

```json
{
  "name": "recorder",
  "version": "0.1.0",
  "description": "Reference plugin that records all gRPC requests and optionally returns mock responses",
  "author": "FinFocus Team",
  "supported_providers": ["*"],
  "protocols": ["grpc"],
  "binary": "finfocus-plugin-recorder",
  "metadata": {
    "repository": "https://github.com/rshade/finfocus",
    "docs": "https://github.com/rshade/finfocus/tree/main/plugins/recorder",
    "reference_implementation": true
  }
}
```

## Conformance Requirements

The recorder plugin MUST:

1. Respond to all three CostSourceService RPCs
2. Record all requests to JSON files (unless I/O fails)
3. Return valid protobuf responses (never panic)
4. Handle graceful shutdown on SIGINT/SIGTERM
5. Support both TCP (--port) and stdio (--stdio) modes
6. Use pluginsdk v0.4.6+ validation helpers
7. Pass `finfocus plugin validate` checks
