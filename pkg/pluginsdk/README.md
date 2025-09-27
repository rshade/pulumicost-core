# PulumiCost Plugin SDK

The PulumiCost Plugin SDK provides a comprehensive development framework for creating cost calculation plugins for cloud providers. This SDK simplifies the process of building plugins that integrate with the PulumiCost ecosystem.

## Overview

The Plugin SDK consists of:

- **Core SDK** (`sdk.go`) - Main plugin interface and server implementation
- **Helper Utilities** (`helpers.go`) - Common patterns and calculations
- **Testing Framework** (`testing.go`) - Testing utilities for plugin development
- **Manifest System** (`manifest.go`) - Plugin metadata and validation

## Quick Start

### 1. Initialize a New Plugin

```bash
pulumicost plugin init my-provider --author "Your Name" --providers aws,azure
```

This creates a complete plugin project structure with:
- Go module configuration
- Plugin manifest
- Boilerplate implementation
- Build scripts
- Documentation templates
- Example tests

### 2. Implement the Plugin Interface

```go
package main

import (
    "context"
    pbc "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
    "github.com/rshade/pulumicost-core/pkg/pluginsdk"
)

type MyPlugin struct {
    *pluginsdk.BasePlugin
}

func NewMyPlugin() *MyPlugin {
    base := pluginsdk.NewBasePlugin("my-provider")
    
    // Configure supported providers and resource types
    base.Matcher().AddProvider("aws")
    base.Matcher().AddResourceType("aws:ec2:Instance")
    
    return &MyPlugin{BasePlugin: base}
}

func (p *MyPlugin) GetProjectedCost(ctx context.Context, req *pbc.GetProjectedCostRequest) (*pbc.GetProjectedCostResponse, error) {
    if !p.Matcher().Supports(req.Resource) {
        return nil, pluginsdk.NotSupportedError(req.Resource)
    }
    
    unitPrice := calculatePrice(req.Resource) // Your pricing logic
    return p.Calculator().CreateProjectedCostResponse("USD", unitPrice, "description"), nil
}

func (p *MyPlugin) GetActualCost(ctx context.Context, req *pbc.GetActualCostRequest) (*pbc.GetActualCostResponse, error) {
    // Implement actual cost retrieval from cloud provider APIs
    return nil, pluginsdk.NoDataError(req.ResourceId)
}

func main() {
    plugin := NewMyPlugin()
    config := pluginsdk.ServeConfig{Plugin: plugin}
    pluginsdk.Serve(context.Background(), config)
}
```

### 3. Build and Test

```bash
make build
make test
make install  # Install to local plugin registry
```

## Core Interfaces

### Plugin Interface

Every plugin must implement the `Plugin` interface:

```go
type Plugin interface {
    Name() string
    GetProjectedCost(ctx context.Context, req *pbc.GetProjectedCostRequest) (*pbc.GetProjectedCostResponse, error)
    GetActualCost(ctx context.Context, req *pbc.GetActualCostRequest) (*pbc.GetActualCostResponse, error)
}
```

### BasePlugin

The `BasePlugin` provides common functionality:

```go
base := pluginsdk.NewBasePlugin("my-plugin-name")

// Configure supported providers
base.Matcher().AddProvider("aws")
base.Matcher().AddProvider("azure")

// Configure supported resource types  
base.Matcher().AddResourceType("aws:ec2:Instance")
base.Matcher().AddResourceType("azure:compute:VirtualMachine")
```

## Helper Utilities

### ResourceMatcher

Helps determine if a plugin supports specific resources:

```go
matcher := pluginsdk.NewResourceMatcher()
matcher.AddProvider("aws")
matcher.AddResourceType("aws:ec2:Instance")

// Check if resource is supported
if matcher.Supports(resource) {
    // Handle the resource
}
```

### CostCalculator

Provides utilities for cost calculations:

```go
calc := pluginsdk.NewCostCalculator()

// Convert between hourly and monthly costs
monthlyCost := calc.HourlyToMonthly(0.10) // $0.10/hour → $73/month
hourlyCost := calc.MonthlyToHourly(73.0)  // $73/month → $0.10/hour

// Create standardized responses
response := calc.CreateProjectedCostResponse("USD", 0.10, "EC2 instance hourly cost")
```

### Error Handling

Standard error functions for common scenarios:

```go
// Resource not supported
return nil, pluginsdk.NotSupportedError(resource)

// No cost data available
return nil, pluginsdk.NoDataError(resourceID)
```

## Testing Framework

The SDK includes comprehensive testing utilities:

### Test Server Setup

```go
func TestMyPlugin(t *testing.T) {
    plugin := NewMyPlugin()
    testPlugin := pluginsdk.NewTestPlugin(t, plugin)
    
    // Test plugin name
    testPlugin.TestName("my-plugin")
    
    // Test projected cost calculation
    resource := pluginsdk.CreateTestResource("aws", "aws:ec2:Instance", map[string]string{
        "instanceType": "t3.micro",
    })
    response := testPlugin.TestProjectedCost(resource, false) // false = expect success
    
    // Test unsupported resource
    unsupported := pluginsdk.CreateTestResource("gcp", "gcp:compute:Instance", nil)
    testPlugin.TestProjectedCost(unsupported, true) // true = expect error
}
```

### Manual Testing

For manual testing during development:

```go
func TestManual(t *testing.T) {
    plugin := NewMyPlugin()
    server := pluginsdk.NewTestServer(t, plugin)
    defer server.Close()
    
    client := server.Client()
    
    // Make direct gRPC calls
    resp, err := client.Name(context.Background(), &pbc.NameRequest{})
    // ... test manually
}
```

## Manifest System

### Creating Manifests

```go
manifest := pluginsdk.CreateDefaultManifest("my-plugin", "Your Name", []string{"aws"})
manifest.Description = "Custom description"
manifest.Version = "2.0.0"

// Save as YAML or JSON
manifest.SaveManifest("manifest.yaml")
manifest.SaveManifest("manifest.json")
```

### Loading and Validation

```go
// Load from file
manifest, err := pluginsdk.LoadManifest("manifest.yaml")
if err != nil {
    log.Fatal(err)
}

// Validate manually
if err := manifest.Validate(); err != nil {
    log.Fatal(err)
}
```

### Manifest Schema

```yaml
name: my-plugin
version: 1.0.0
description: Plugin description
author: Plugin Author
supported_providers: [aws, azure]
protocols: [grpc]
binary: ./bin/pulumicost-plugin-my-plugin
metadata:
  repository: https://github.com/example/my-plugin
  docs: https://example.com/docs
```

## Advanced Patterns

### Multi-Provider Support

```go
func (p *MyPlugin) GetProjectedCost(ctx context.Context, req *pbc.GetProjectedCostRequest) (*pbc.GetProjectedCostResponse, error) {
    switch req.Resource.Provider {
    case "aws":
        return p.calculateAWSCost(req.Resource)
    case "azure":
        return p.calculateAzureCost(req.Resource)
    default:
        return nil, pluginsdk.NotSupportedError(req.Resource)
    }
}
```

### Configuration and Credentials

```go
type MyPlugin struct {
    *pluginsdk.BasePlugin
    awsClient *aws.Client
    config    Config
}

func NewMyPlugin(config Config) (*MyPlugin, error) {
    awsClient, err := aws.NewClient(config.AWSCredentials)
    if err != nil {
        return nil, err
    }
    
    return &MyPlugin{
        BasePlugin: pluginsdk.NewBasePlugin("my-plugin"),
        awsClient:  awsClient,
        config:     config,
    }, nil
}
```

### Caching and Performance

```go
type MyPlugin struct {
    *pluginsdk.BasePlugin
    priceCache map[string]float64
    cacheMutex sync.RWMutex
}

func (p *MyPlugin) getCachedPrice(key string) (float64, bool) {
    p.cacheMutex.RLock()
    defer p.cacheMutex.RUnlock()
    price, exists := p.priceCache[key]
    return price, exists
}

func (p *MyPlugin) setCachedPrice(key string, price float64) {
    p.cacheMutex.Lock()
    defer p.cacheMutex.Unlock()
    p.priceCache[key] = price
}
```

## Best Practices

### Error Handling

1. **Use SDK Error Functions**: Use `NotSupportedError` and `NoDataError` for consistency
2. **Graceful Degradation**: Continue processing when some resources fail
3. **Contextual Errors**: Include resource information in error messages

### Performance

1. **Resource Matching**: Use the ResourceMatcher for efficient filtering
2. **Caching**: Cache pricing data to avoid repeated API calls
3. **Concurrent Processing**: Process multiple resources concurrently when possible

### Testing

1. **Unit Tests**: Test individual calculation functions
2. **Integration Tests**: Test gRPC interface using SDK test utilities
3. **Error Cases**: Test unsupported resources and error conditions
4. **Edge Cases**: Test with missing properties and invalid data

### Documentation

1. **Resource Support**: Document supported resource types and properties
2. **Configuration**: Document required credentials and configuration
3. **Examples**: Provide usage examples and sample Pulumi plans

## Examples

- [AWS Example Plugin](../examples/plugins/aws-example/) - Complete AWS plugin implementation
- [Plugin Template](../examples/plugins/template/) - Generated by `pulumicost plugin init`

## API Reference

### Types

- `Plugin` - Main plugin interface
- `BasePlugin` - Base implementation with common utilities
- `ResourceMatcher` - Resource filtering utilities
- `CostCalculator` - Cost calculation helpers
- `Manifest` - Plugin metadata structure
- `TestPlugin` - Testing framework utilities

### Functions

- `NewBasePlugin(name)` - Create a new base plugin
- `NewResourceMatcher()` - Create a resource matcher
- `NewCostCalculator()` - Create a cost calculator
- `Serve(ctx, config)` - Start plugin gRPC server
- `LoadManifest(path)` - Load manifest from file
- `CreateDefaultManifest(...)` - Create default manifest

For complete API documentation, see the Go package documentation or use `go doc` commands:

```bash
go doc github.com/rshade/pulumicost-core/pkg/pluginsdk
go doc github.com/rshade/pulumicost-core/pkg/pluginsdk.Plugin
go doc github.com/rshade/pulumicost-core/pkg/pluginsdk.BasePlugin
```