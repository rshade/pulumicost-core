# Data Model: Pulumi Analyzer Plugin

**Feature**: 008-analyzer-plugin
**Date**: 2025-12-05
**Status**: Complete

## Overview

This document defines the data structures and transformations required to implement the Pulumi Analyzer plugin. The core challenge is mapping between Pulumi's protocol buffer types and PulumiCost's internal types.

## Entities

### 1. Analyzer Server

The main gRPC service implementation that bridges Pulumi and PulumiCost.

```go
// internal/analyzer/server.go
package analyzer

import (
    "context"
    "sync"

    "github.com/pulumi/pulumi/sdk/v3/proto/go/pulumirpc"
    "github.com/rshade/pulumicost-core/internal/engine"
    "github.com/rs/zerolog"
    "google.golang.org/protobuf/types/known/emptypb"
)

// Server implements the pulumirpc.AnalyzerServer interface for cost estimation.
type Server struct {
    pulumirpc.UnimplementedAnalyzerServer

    engine  *engine.Engine
    logger  zerolog.Logger
    version string

    // Stack context from ConfigureStack
    stackName    string
    projectName  string
    organization string
    dryRun       bool

    // Cancellation support
    cancelMu sync.Mutex
    canceled bool
}

// NewServer creates a new Analyzer server with the given engine and logger.
func NewServer(e *engine.Engine, logger zerolog.Logger, version string) *Server {
    return &Server{
        engine:  e,
        logger:  logger,
        version: version,
    }
}
```

**Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `engine` | `*engine.Engine` | Cost calculation engine (existing) |
| `logger` | `zerolog.Logger` | Stderr-bound logger |
| `version` | `string` | Plugin version for diagnostics |
| `stackName` | `string` | From ConfigureStack RPC |
| `projectName` | `string` | From ConfigureStack RPC |
| `organization` | `string` | From ConfigureStack RPC |
| `dryRun` | `bool` | True if preview (always true for analyzer) |
| `canceled` | `bool` | Set by Cancel RPC |

### 2. Resource Mapping

Transformation from Pulumi's protocol buffer types to PulumiCost internal types.

**Source Type**: `pulumirpc.AnalyzerResource`

```protobuf
message AnalyzerResource {
    string type = 1;                       // e.g., "aws:ec2/instance:Instance"
    google.protobuf.Struct properties = 2; // Resource properties
    string urn = 3;                        // e.g., "urn:pulumi:dev::myproject::aws:ec2/instance:Instance::webserver"
    string name = 4;                       // Resource name
    AnalyzerResourceOptions options = 5;   // Resource options
    AnalyzerProviderResource provider = 6; // Provider info
    string parent = 7;                     // Parent URN
    repeated string dependencies = 8;      // Dependency URNs
    // ...
}
```

**Target Type**: `engine.ResourceDescriptor` (from `internal/engine/types.go`)

```go
type ResourceDescriptor struct {
    Type       string                 // Cloud resource type
    ID         string                 // Resource identifier (from URN)
    Provider   string                 // Cloud provider name
    Properties map[string]interface{} // Resource properties
}
```

**Mapper Function**:

```go
// internal/analyzer/mapper.go
package analyzer

import (
    "strings"

    "github.com/pulumi/pulumi/sdk/v3/proto/go/pulumirpc"
    "github.com/rshade/pulumicost-core/internal/engine"
    "google.golang.org/protobuf/types/known/structpb"
)

// MapResource converts a pulumirpc.AnalyzerResource to an engine.ResourceDescriptor.
func MapResource(r *pulumirpc.AnalyzerResource) engine.ResourceDescriptor {
    return engine.ResourceDescriptor{
        Type:       r.GetType(),
        ID:         extractResourceID(r.GetUrn()),
        Provider:   extractProvider(r),
        Properties: structToMap(r.GetProperties()),
    }
}

// MapResources converts a slice of AnalyzerResource to ResourceDescriptors.
func MapResources(resources []*pulumirpc.AnalyzerResource) []engine.ResourceDescriptor {
    result := make([]engine.ResourceDescriptor, len(resources))
    for i, r := range resources {
        result[i] = MapResource(r)
    }
    return result
}

// extractResourceID extracts the resource name from a Pulumi URN.
// URN format: urn:pulumi:stack::project::type::name
func extractResourceID(urn string) string {
    parts := strings.Split(urn, "::")
    if len(parts) >= 4 {
        return parts[len(parts)-1] // Last part is the name
    }
    return urn
}

// extractProvider extracts the provider name from the resource.
// Tries provider URN first, falls back to type prefix.
func extractProvider(r *pulumirpc.AnalyzerResource) string {
    // Try provider resource first
    if p := r.GetProvider(); p != nil {
        if providerType := p.GetType(); providerType != "" {
            // Format: pulumi:providers:aws
            parts := strings.Split(providerType, ":")
            if len(parts) >= 3 {
                return parts[2]
            }
        }
    }

    // Fall back to resource type prefix
    // Format: aws:ec2/instance:Instance
    parts := strings.Split(r.GetType(), ":")
    if len(parts) >= 1 {
        return parts[0]
    }

    return "unknown"
}

// structToMap converts a protobuf Struct to a Go map.
func structToMap(s *structpb.Struct) map[string]interface{} {
    if s == nil {
        return make(map[string]interface{})
    }
    return s.AsMap()
}
```

**Mapping Rules**:

| Source Field | Target Field | Transformation |
|--------------|--------------|----------------|
| `type` | `Type` | Direct copy |
| `urn` | `ID` | Extract last `::` segment |
| `provider.type` | `Provider` | Extract from `pulumi:providers:X` |
| `type` (fallback) | `Provider` | Extract first `:` segment |
| `properties` | `Properties` | `structpb.Struct.AsMap()` |

### 3. Diagnostic Conversion

Transformation from `engine.CostResult` to `pulumirpc.AnalyzeDiagnostic`.

**Source Type**: `engine.CostResult` (from `internal/engine/types.go`)

```go
type CostResult struct {
    ResourceType string             `json:"resourceType"`
    ResourceID   string             `json:"resourceId"`
    Adapter      string             `json:"adapter"`
    Currency     string             `json:"currency"`
    Monthly      float64            `json:"monthly"`
    Hourly       float64            `json:"hourly"`
    Notes        string             `json:"notes"`
    Breakdown    map[string]float64 `json:"breakdown"`
    // Actual cost fields (not used for projected)
    TotalCost  float64   `json:"totalCost,omitempty"`
    DailyCosts []float64 `json:"dailyCosts,omitempty"`
    // ...
}
```

**Target Type**: `pulumirpc.AnalyzeDiagnostic`

```protobuf
message AnalyzeDiagnostic {
    string policyName = 1;
    string policyPackName = 2;
    string policyPackVersion = 3;
    string description = 4;
    string message = 5;
    EnforcementLevel enforcementLevel = 7;
    string urn = 8;
    PolicySeverity severity = 9;
}
```

**Converter Function**:

```go
// internal/analyzer/diagnostics.go
package analyzer

import (
    "fmt"

    "github.com/pulumi/pulumi/sdk/v3/proto/go/pulumirpc"
    "github.com/rshade/pulumicost-core/internal/engine"
)

const (
    policyPackName = "pulumicost"
    policyNameCost = "cost-estimate"
    policyNameSum  = "stack-cost-summary"
)

// CostToDiagnostic converts a CostResult to an AnalyzeDiagnostic.
func CostToDiagnostic(cost engine.CostResult, urn string, version string) *pulumirpc.AnalyzeDiagnostic {
    message := formatCostMessage(cost)
    severity := pulumirpc.PolicySeverity_POLICY_SEVERITY_LOW

    // Elevate severity if no cost data available
    if cost.Monthly == 0 && cost.Notes != "" {
        severity = pulumirpc.PolicySeverity_POLICY_SEVERITY_MEDIUM
    }

    return &pulumirpc.AnalyzeDiagnostic{
        PolicyName:        policyNameCost,
        PolicyPackName:    policyPackName,
        PolicyPackVersion: version,
        Description:       "Estimated resource cost",
        Message:           message,
        EnforcementLevel:  pulumirpc.EnforcementLevel_ADVISORY,
        Urn:               urn,
        Severity:          severity,
    }
}

// StackSummaryDiagnostic creates a stack-level cost summary diagnostic.
func StackSummaryDiagnostic(costs []engine.CostResult, version string) *pulumirpc.AnalyzeDiagnostic {
    var totalMonthly float64
    var currency string = "USD"
    analyzed := 0

    for _, c := range costs {
        totalMonthly += c.Monthly
        if c.Currency != "" {
            currency = c.Currency
        }
        if c.Monthly > 0 {
            analyzed++
        }
    }

    message := fmt.Sprintf("Total Estimated Monthly Cost: $%.2f %s (%d resources analyzed)",
        totalMonthly, currency, analyzed)

    return &pulumirpc.AnalyzeDiagnostic{
        PolicyName:        policyNameSum,
        PolicyPackName:    policyPackName,
        PolicyPackVersion: version,
        Description:       "Stack cost summary",
        Message:           message,
        EnforcementLevel:  pulumirpc.EnforcementLevel_ADVISORY,
        Severity:          pulumirpc.PolicySeverity_POLICY_SEVERITY_LOW,
        // No URN - stack-level
    }
}

func formatCostMessage(cost engine.CostResult) string {
    if cost.Monthly > 0 {
        return fmt.Sprintf("Estimated Monthly Cost: $%.2f %s (source: %s)",
            cost.Monthly, cost.Currency, cost.Adapter)
    }
    if cost.Notes != "" {
        return cost.Notes
    }
    return "Unable to estimate cost"
}
```

### 4. Configuration Types

New configuration section for analyzer settings.

```go
// internal/config/config.go (extension)

// AnalyzerConfig defines analyzer-specific configuration.
type AnalyzerConfig struct {
    Timeout AnalyzerTimeout           `yaml:"timeout" json:"timeout"`
    Plugins map[string]AnalyzerPlugin `yaml:"plugins" json:"plugins"`
}

// AnalyzerTimeout defines timeout settings for analysis.
type AnalyzerTimeout struct {
    PerResource   Duration `yaml:"per_resource"    json:"per_resource"`   // Per-resource timeout
    Total         Duration `yaml:"total"           json:"total"`          // Overall timeout
    WarnThreshold Duration `yaml:"warn_threshold"  json:"warn_threshold"` // Warning threshold
}

// AnalyzerPlugin defines a cost plugin configuration.
type AnalyzerPlugin struct {
    Path    string            `yaml:"path"    json:"path"`
    Enabled bool              `yaml:"enabled" json:"enabled"`
    Env     map[string]string `yaml:"env"     json:"env"` // Environment variables
}
```

**Example Configuration**:

```yaml
# ~/.pulumicost/config.yaml
analyzer:
  timeout:
    per_resource: 5s
    total: 60s
    warn_threshold: 30s
  plugins:
    vantage:
      path: ~/.pulumicost/plugins/vantage/v1.0.0/pulumicost-plugin-vantage
      enabled: true
      env:
        VANTAGE_API_KEY: "${VANTAGE_API_KEY}"
```

## State Transitions

### Analyzer Server Lifecycle

```text
                          ┌──────────────────────┐
                          │  Plugin Started      │
                          │  (by Pulumi engine)  │
                          └──────────┬───────────┘
                                     │
                                     ▼
                          ┌──────────────────────┐
                          │  Handshake RPC       │
                          │  (receive engine addr)│
                          └──────────┬───────────┘
                                     │
                                     ▼
                          ┌──────────────────────┐
                          │  ConfigureStack RPC  │
                          │  (receive stack info) │
                          └──────────┬───────────┘
                                     │
                                     ▼
         ┌───────────────────────────┴───────────────────────────┐
         │                                                        │
         ▼                                                        ▼
┌─────────────────────┐                              ┌─────────────────────┐
│  AnalyzeStack RPC   │                              │  GetAnalyzerInfo    │
│  (full stack)       │                              │  (metadata query)   │
└─────────┬───────────┘                              └─────────────────────┘
          │
          ▼
┌─────────────────────┐
│  Map Resources      │
│  (proto → internal) │
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│  Calculate Costs    │
│  (via engine)       │
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│  Create Diagnostics │
│  (costs → proto)    │
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│  Return Response    │
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐      ┌─────────────────────┐
│  Wait for Cancel    │─────▶│  Graceful Shutdown  │
│  or process exit    │      │  (cleanup)          │
└─────────────────────┘      └─────────────────────┘
```

### Request Flow

```text
Pulumi Engine                    Analyzer Plugin                    Cost Plugins
     │                                 │                                  │
     │──── AnalyzeStack ──────────────▶│                                  │
     │     ([]AnalyzerResource)        │                                  │
     │                                 │──── MapResources ────────────────▶
     │                                 │     ([]ResourceDescriptor)       │
     │                                 │                                  │
     │                                 │──── Engine.GetProjectedCost ────▶│
     │                                 │                                  │
     │                                 │◀──── []CostResult ───────────────│
     │                                 │                                  │
     │                                 │──── CostsToDiagnostics ─────────▶
     │                                 │     (per-resource + summary)     │
     │                                 │                                  │
     │◀─── AnalyzeResponse ────────────│                                  │
     │     ([]AnalyzeDiagnostic)       │                                  │
```

## Validation Rules

### ResourceDescriptor Validation

From `internal/engine/types.go`:

| Field | Constraint | Value |
|-------|------------|-------|
| `Type` | Required, max length | 256 bytes |
| `ID` | Max length | 1024 bytes |
| `Properties` | Max count | 100 properties |
| Property key | Valid chars, max length | alphanumeric/`_`/`-`/`.`, 128 bytes |
| Property value | Max length | 10KB |

### Diagnostic Validation

| Field | Constraint | Enforcement |
|-------|------------|-------------|
| `EnforcementLevel` | Must be ADVISORY | MVP requirement (FR-005) |
| `Message` | Non-empty | Always includes cost or error message |
| `PolicyPackName` | Constant | "pulumicost" |

## Error Handling

### Error Categories

| Category | Example | Diagnostic Behavior |
|----------|---------|---------------------|
| Mapping Error | Invalid URN format | Log warning, skip resource |
| Plugin Timeout | Cost plugin unresponsive | Use fallback, MEDIUM severity |
| Network Error | API unreachable | Use local specs, MEDIUM severity |
| Unsupported Type | Unknown resource | Debug log, skip silently |
| All Failures | No cost data at all | WARNING diagnostic |

### Error Aggregation

The analyzer collects errors without failing the entire request:

```go
type AnalysisResult struct {
    Costs      []engine.CostResult
    Errors     []AnalysisError
    TotalTime  time.Duration
    Analyzed   int
    Skipped    int
}

type AnalysisError struct {
    URN     string
    Type    string
    Error   error
    Skipped bool
}
```
