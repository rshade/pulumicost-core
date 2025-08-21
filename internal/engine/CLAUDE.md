# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Engine Package Overview

The `internal/engine` package is the core orchestration layer for PulumiCost, responsible for coordinating cost calculations between plugins and local pricing specifications, then rendering results in multiple output formats.

## Architecture

### Core Components

1. **Engine** (`engine.go`)
   - Central orchestrator managing plugin clients and spec loaders
   - Handles both projected and actual cost calculations
   - Implements graceful fallback from plugins to local specs

2. **Types System** (`types.go`)  
   - `ResourceDescriptor`: Represents cloud resources with type, provider, and properties
   - `CostResult`: Standardized cost output with breakdown, currency, and metadata
   - `PricingSpec`: YAML-based local pricing fallback structure

3. **Output Rendering** (`project.go`)
   - Multi-format output: table, JSON, NDJSON
   - Tabwriter-based table formatting with column truncation
   - Streaming NDJSON for large result sets

### Data Flow Architecture

```text
Resources → Engine.GetProjectedCost/GetActualCost → Plugin Clients (gRPC) → Results
                                                 ↓ (fallback)
                                              Spec Loader (YAML) → Results
                                                 ↓
                                              RenderResults → Output
```

## Testing Commands

```bash
# Run all engine tests (currently no test files exist)
go test ./internal/engine/...

# Test engine integration via CLI tests
go test ./internal/cli/... -run TestCLIIntegration

# Run engine functionality through example commands
./bin/pulumicost cost projected --pulumi-json examples/plans/aws-simple-plan.json
./bin/pulumicost cost actual --pulumi-json examples/plans/aws-simple-plan.json --from 2025-01-01
```

## Cost Calculation Strategy

### Projected Cost Flow

1. **Plugin First**: Query each plugin client via `GetProjectedCost` gRPC call
2. **Fallback to Spec**: If no plugin responds, use local YAML pricing spec
3. **Default Values**: If no spec available, return "none" adapter with zero cost
4. **Multi-Plugin Support**: Collect results from all responding plugins

### Actual Cost Flow  

1. **Plugin Only**: Query plugins via `GetActualCost` gRPC call with time range
2. **First Success**: Use first plugin that successfully returns data
3. **Time Calculations**: Convert historical data to monthly/hourly rates
4. **No Fallback**: Actual costs require live plugin data

### Cost Calculation Constants

```go
hoursPerDay = 24
hoursPerMonth = 730          // Standard business month
defaultEstimate = 100.0      // USD fallback estimate
monthlyConversion = 30.44    // Average days per month
```

## Output Format Handling

### Table Format

- Uses `text/tabwriter` for aligned columns
- Truncates long resource names (>50 chars) with "..."
- Headers: Resource, Adapter, Projected Monthly, Currency, Notes
- Column padding: 2 spaces between columns

### JSON Format

- Pretty-printed JSON with 2-space indentation
- Full array of CostResult objects
- Includes all fields: breakdown, metadata, etc.

### NDJSON Format

- Newline-delimited JSON for streaming
- One CostResult per line
- No array wrapper - suitable for large datasets

## Error Handling Patterns

### Plugin Communication

- **Continue on Error**: Plugin failures don't stop processing
- **Best Effort**: Collect all available results
- **Graceful Degradation**: Fall back to specs, then to "none" adapter

### Resource Processing

- **Per-Resource Isolation**: One resource failure doesn't affect others
- **Multi-Plugin Tolerance**: Continue if some plugins fail
- **Default Results**: Always return some result, even if placeholder

## Integration Points

### Plugin System

- Consumes `pluginhost.Client` instances with gRPC API
- Converts resource properties to proto format (`map[string]string`)
- Handles plugin failures transparently

### Spec System  

- `SpecLoader` interface for YAML pricing specs
- Type assertion for `*PricingSpec` objects
- Provider/service/SKU based lookups

### Proto Integration

- Uses real protobuf definitions from pulumicost-spec
- `GetProjectedCostRequest`/`GetActualCostRequest` messages
- Automatic property conversion from `interface{}` to `string`

## Key Implementation Details

### Property Conversion

```go
// Converts arbitrary properties to string map for gRPC
func convertToProto(properties map[string]interface{}) map[string]string
```

### Time Range Processing

```go
// Converts actual costs to monthly rate using average month length
Monthly: result.TotalCost * 30.44 / float64(to.Sub(from).Hours()/hoursPerDay)
```

### Service Extraction

```go
// Currently returns "default" - placeholder for future resource type parsing
func extractService(_ string) string { return "default" }
```

## Cost Result Structure

```go
type CostResult struct {
    ResourceType string             // Cloud resource type (e.g., "aws:ec2/instance:Instance")
    ResourceID   string             // Unique resource identifier/URN
    Adapter      string             // Plugin name or "local-spec"/"none"
    Currency     string             // ISO currency code (typically "USD")
    Monthly      float64            // Projected/actual monthly cost
    Hourly       float64            // Projected/actual hourly cost  
    Notes        string             // Human-readable cost details
    Breakdown    map[string]float64 // Detailed cost breakdown by component
}
```

## Common Usage Patterns

### Engine Initialization

```go
clients := []*pluginhost.Client{...}  // From registry
loader := spec.NewLoader(specDir)     // From config
engine := engine.New(clients, loader)
```

### Multi-Format Output

```go
results, err := engine.GetProjectedCost(ctx, resources)
engine.RenderResults(engine.OutputJSON, results)    // JSON
engine.RenderResults(engine.OutputTable, results)   // Table  
engine.RenderResults(engine.OutputNDJSON, results)  // NDJSON
```

This package acts as the central coordinating layer, making it critical for understanding the overall cost calculation flow in PulumiCost.

