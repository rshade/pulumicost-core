# Analyzer Service Contract

**Feature**: 009-analyzer-plugin
**Date**: 2025-12-05
**Proto Source**: `github.com/pulumi/pulumi/sdk/v3/proto/go/pulumirpc`

## Overview

This document defines the contract for implementing the Pulumi Analyzer gRPC service for cost estimation. The implementation uses the official Pulumi SDK proto definitions.

## Service Definition

```protobuf
service Analyzer {
    // Handshake exchanges addresses between engine and analyzer
    rpc Handshake(AnalyzerHandshakeRequest) returns (AnalyzerHandshakeResponse);

    // ConfigureStack receives stack context before analysis
    rpc ConfigureStack(AnalyzerStackConfigureRequest) returns (AnalyzerStackConfigureResponse);

    // AnalyzeStack analyzes all resources at end of preview
    rpc AnalyzeStack(AnalyzeStackRequest) returns (AnalyzeResponse);

    // Analyze analyzes a single resource (optional for MVP)
    rpc Analyze(AnalyzeRequest) returns (AnalyzeResponse);

    // GetAnalyzerInfo returns metadata about the analyzer
    rpc GetAnalyzerInfo(google.protobuf.Empty) returns (AnalyzerInfo);

    // GetPluginInfo returns plugin version information
    rpc GetPluginInfo(google.protobuf.Empty) returns (PluginInfo);

    // Configure configures analyzer policies (optional for MVP)
    rpc Configure(ConfigureAnalyzerRequest) returns (google.protobuf.Empty);

    // Cancel signals graceful shutdown
    rpc Cancel(google.protobuf.Empty) returns (google.protobuf.Empty);
}
```

## Required RPCs (MVP)

### 1. Handshake

**Purpose**: Establish connection with Pulumi engine.

**Request**:

```protobuf
message AnalyzerHandshakeRequest {
    string engine_address = 1;      // gRPC address of Pulumi engine
    optional string root_directory = 2;    // Plugin root (if booted by engine)
    optional string program_directory = 3; // Working directory
}
```

**Response**:

```protobuf
message AnalyzerHandshakeResponse {
    // Empty for now
}
```

**Behavior**:

- Store `engine_address` for potential callbacks (not used in MVP)
- Log handshake success at DEBUG level
- Return empty response

### 2. ConfigureStack

**Purpose**: Receive stack context before analysis begins.

**Request**:

```protobuf
message AnalyzerStackConfigureRequest {
    string stack = 1;       // Stack name
    string project = 2;     // Project name
    string organization = 3; // Organization name
    bool dry_run = 4;       // True if preview (always true for analyzer)
    map<string, string> config = 7;       // Stack configuration
    map<string, string> tags = 8;         // Stack tags
}
```

**Response**:

```protobuf
message AnalyzerStackConfigureResponse {
    // Empty
}
```

**Behavior**:

- Store stack/project/organization for logging context
- Log stack configuration at DEBUG level
- Return empty response

### 3. AnalyzeStack

**Purpose**: Calculate costs for all resources in the stack.

**Request**:

```protobuf
message AnalyzeStackRequest {
    repeated AnalyzerResource resources = 1;
}

message AnalyzerResource {
    string type = 1;                       // e.g., "aws:ec2/instance:Instance"
    google.protobuf.Struct properties = 2; // Resource properties
    string urn = 3;                        // Full URN
    string name = 4;                       // Resource name
    AnalyzerResourceOptions options = 5;   // Resource options
    AnalyzerProviderResource provider = 6; // Provider info
    string parent = 7;                     // Parent URN
    repeated string dependencies = 8;      // Dependency URNs
}
```

**Response**:

```protobuf
message AnalyzeResponse {
    repeated AnalyzeDiagnostic diagnostics = 2;
}

message AnalyzeDiagnostic {
    string policyName = 1;         // "cost-estimate" or "stack-cost-summary"
    string policyPackName = 2;     // "pulumicost"
    string policyPackVersion = 3;  // Plugin version
    string description = 4;        // Human description
    string message = 5;            // Cost message
    EnforcementLevel enforcementLevel = 7; // ADVISORY only
    string urn = 8;                // Resource URN (empty for summary)
    PolicySeverity severity = 9;   // LOW or MEDIUM
}
```

**Behavior**:

1. Map `[]AnalyzerResource` → `[]engine.ResourceDescriptor`
2. Call `engine.GetProjectedCost(ctx, resources)`
3. Convert `[]engine.CostResult` → `[]AnalyzeDiagnostic`
4. Add stack summary diagnostic
5. Return response

### 4. GetAnalyzerInfo

**Purpose**: Return analyzer metadata and policy list.

**Response**:

```protobuf
message AnalyzerInfo {
    string name = 1;           // "pulumicost"
    string displayName = 2;    // "PulumiCost Analyzer"
    repeated PolicyInfo policies = 3;
    string version = 4;        // Plugin version
    bool supportsConfig = 5;   // false for MVP
    string description = 7;    // "Cost estimation for cloud resources"
}

message PolicyInfo {
    string name = 1;                    // "cost-estimate"
    string displayName = 2;             // "Cost Estimation"
    string description = 3;             // "Estimates monthly cost..."
    EnforcementLevel enforcementLevel = 5; // ADVISORY
}
```

**Behavior**:

- Return static metadata about pulumicost
- List the "cost-estimate" policy
- Version from build constants

### 5. GetPluginInfo

**Purpose**: Return plugin version.

**Response**:

```protobuf
message PluginInfo {
    string version = 1; // e.g., "0.1.0"
}
```

### 6. Cancel

**Purpose**: Signal graceful shutdown.

**Behavior**:

- Set `canceled` flag
- Interrupt any ongoing analysis
- Close plugin connections gracefully
- Return empty response

## Optional RPCs

### Analyze

Single-resource analysis. Can return `Unimplemented` in MVP.

### Configure

Policy configuration. Can return `Unimplemented` in MVP.

### Remediate

Property transformation. Not applicable for cost estimation.

## Error Handling

| Scenario | gRPC Status | Diagnostic Behavior |
|----------|-------------|---------------------|
| Normal operation | OK | Return diagnostics |
| Plugin timeout | OK | WARNING diagnostic |
| Network error | OK | WARNING diagnostic |
| Invalid request | INVALID_ARGUMENT | Error with details |
| Plugin shutdown | CANCELLED | Empty response |
| Unimplemented RPC | UNIMPLEMENTED | Skip gracefully |

## Enforcement Levels

| Level | Usage |
|-------|-------|
| `ADVISORY` | All cost diagnostics (MVP) |
| `MANDATORY` | Future: budget violations |
| `DISABLED` | Policy disabled by config |
| `REMEDIATE` | Not applicable |

## Severity Levels

| Level | Usage |
|-------|-------|
| `LOW` | Successful cost calculation |
| `MEDIUM` | Fallback or partial data |
| `HIGH` | Future: significant cost increase |
| `CRITICAL` | Future: budget threshold exceeded |
