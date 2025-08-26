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
   - Enhanced table format with cost summaries and breakdowns
   - Aggregated JSON with totals by provider, service, and adapter
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
2. **Smart Spec Fallback**: Multi-level fallback pattern:
   - Try `provider-service-sku` pattern (e.g., `aws-ec2-t3.micro`)
   - Fallback to `provider-service-default` pattern (e.g., `aws-ec2-default`)
   - Try common SKUs: `standard`, `basic`, `default`
3. **Intelligent Cost Calculation**: Extract costs from spec pricing data:
   - Direct monthly estimates (`monthlyEstimate`)
   - Hourly rates (`onDemandHourly`, `hourlyRate`)
   - Storage calculations with size multipliers (`pricePerGBMonth`)
   - Fallback estimates based on resource type
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

- Enhanced layout with cost summary sections
- Summary: Total monthly/hourly costs and resource count
- Breakdowns: By provider, service, and adapter
- Resource details: Individual resource costs and notes
- Truncates long resource names (>40 chars) with "..."

### JSON Format

- Aggregated results with summary and resource details
- Includes totals by provider, service, and adapter
- Full cost breakdown and analysis
- Pretty-printed JSON with 2-space indentation

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

## New Features (Issue #5 Implementation)

### Enhanced Spec Fallback System

The engine now implements intelligent spec loading with multiple fallback patterns:

```go
// 1. Try exact SKU match: aws-ec2-t3.micro.yaml
// 2. Try service default: aws-ec2-default.yaml  
// 3. Try common patterns: aws-ec2-standard.yaml, aws-ec2-basic.yaml
func (e *Engine) getProjectedCostFromSpec(resource ResourceDescriptor) *CostResult
```

### SKU and Service Extraction

Smart extraction of resource characteristics for spec lookup:

```go
func extractService(resourceType string) string    // "aws:ec2:Instance" → "ec2"
func extractSKU(resource ResourceDescriptor) string // Properties or type → "t3.micro"
```

### Cost Calculation from Specs

Flexible cost calculation supporting multiple pricing models:

```go
func calculateCostsFromSpec(spec *PricingSpec, resource ResourceDescriptor) (monthly, hourly float64)
```

Supports:
- Direct monthly estimates (`monthlyEstimate: 7.59`)
- Hourly rates (`onDemandHourly: 0.0104`)
- Storage pricing with size multipliers (`pricePerGBMonth: 0.10`)
- Intelligent fallbacks by resource type (compute: $20, database: $50, storage: $5)

### Cost Aggregation and Analysis

New aggregation system for comprehensive cost analysis:

```go
type CostSummary struct {
    TotalMonthly float64                   // Total monthly cost across all resources
    TotalHourly  float64                   // Total hourly cost across all resources
    Currency     string                    // Currency for all costs
    ByProvider   map[string]float64        // Costs grouped by cloud provider
    ByService    map[string]float64        // Costs grouped by service (ec2, rds, etc.)
    ByAdapter    map[string]float64        // Costs grouped by data source
}

func AggregateResults(results []CostResult) *AggregatedResults
```

### Resource Filtering

Flexible filtering system supporting multiple criteria:

```go
func FilterResources(resources []ResourceDescriptor, filter string) []ResourceDescriptor
```

Filter patterns:
- `provider=aws` - Filter by cloud provider
- `type=ec2` - Filter by service type
- `service=rds` - Filter by extracted service
- `instanceType=t3.micro` - Filter by any resource property
- `id=i-123` - Filter by resource ID

### Enhanced Output Formatting

Improved table output with comprehensive summaries:

```text
COST SUMMARY
============
Total Monthly Cost:    57.59 USD
Total Hourly Cost:     0.0788 USD
Total Resources:       3

BY PROVIDER
-----------
aws:        57.59 USD

BY SERVICE  
----------
ec2:        7.59 USD
rds:        50.00 USD

BY ADAPTER
----------
local-spec: 57.59 USD

RESOURCE DETAILS
================
Resource                     Adapter     Monthly  Hourly   Currency  Notes
--------                     -------     -------  ------   --------  -----
aws:ec2:Instance/i-123       local-spec  7.59     0.0104   USD       Calculated from local spec
aws:rds:Instance/db-456      local-spec  50.00    0.0685   USD       Calculated from local spec
```

This package acts as the central coordinating layer, making it critical for understanding the overall cost calculation flow in PulumiCost.

