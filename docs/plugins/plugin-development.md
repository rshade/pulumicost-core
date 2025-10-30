---
layout: default
title: Plugin Development Guide
description: Complete guide to building PulumiCost cost source plugins
---

This guide provides a complete walkthrough for building PulumiCost cost source
plugins. You'll learn how to implement a plugin from scratch, test it, and
deploy it for production use.

## Table of Contents

1. [Overview](#overview)
2. [Getting Started](#getting-started)
3. [Creating Your First Plugin](#creating-your-first-plugin)
4. [Implementing Core Functionality](#implementing-core-functionality)
5. [Testing Your Plugin](#testing-your-plugin)
6. [Deployment](#deployment)
7. [Best Practices](#best-practices)
8. [Advanced Topics](#advanced-topics)

---

## Overview

### What is a Plugin?

A PulumiCost plugin is a standalone gRPC server that provides cloud cost
information. Plugins can:

- Calculate **projected costs** for infrastructure resources
- Retrieve **actual historical costs** from cloud APIs
- Support multiple cloud providers (AWS, Azure, GCP, Kubernetes, etc)
- Provide custom pricing logic and business rules

### Plugin Architecture

```text
┌─────────────────────┐
│  PulumiCost Core    │
│  (Plugin Host)      │
└──────────┬──────────┘
           │ gRPC
    ┌──────┴──────┐
    ▼             ▼
┌────────┐    ┌────────┐
│Plugin A│    │Plugin B│
│(Kubecost)   │(Vantage)
└────────┘    └────────┘
```

### When to Build a Plugin

Build a plugin when you need to:

- Integrate with a cost management platform
  (Kubecost, Vantage, CloudHealth, etc)
- Implement custom pricing logic for your organization
- Support cloud providers not yet covered by existing plugins
- Connect to proprietary cost databases

---

## Getting Started

### Prerequisites

- **Go 1.24+** installed
- **Protocol Buffers** compiler (`protoc`)
- **PulumiCost Core** repository cloned
- Basic understanding of gRPC and protocol buffers

### Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install protocol buffer compiler (if not already installed)
# macOS:
brew install protobuf

# Linux:
sudo apt-get install protobuf-compiler

# Windows:
# Download from https://github.com/protocolbuffers/protobuf/releases
```

### Project Structure

A typical plugin project structure:

```text
my-plugin/
├── go.mod
├── go.sum
├── main.go                    # Plugin entry point
├── plugin.manifest.yaml       # Plugin metadata (optional)
├── internal/
│   ├── client/               # API client for cost source
│   │   └── client.go
│   ├── pricing/              # Pricing logic
│   │   └── calculator.go
│   └── config/               # Configuration handling
│       └── config.go
└── README.md
```

---

## Creating Your First Plugin

### Step 1: Initialize Go Module

```bash
mkdir my-plugin
cd my-plugin
go mod init github.com/yourusername/my-plugin

# Add required dependencies
go get github.com/rshade/pulumicost-core/pkg/pluginsdk
go get github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1
go get google.golang.org/grpc
```

### Step 2: Create Basic Plugin Structure

Create `main.go`:

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/rshade/pulumicost-core/pkg/pluginsdk"
    pbc "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
)

// MyPlugin implements the Plugin interface.
type MyPlugin struct {
    *pluginsdk.BasePlugin
}

// NewMyPlugin creates a new plugin instance.
func NewMyPlugin() *MyPlugin {
    base := pluginsdk.NewBasePlugin("my-plugin")

    // Configure supported providers
    base.Matcher().AddProvider("aws")

    // Add supported resource types
    base.Matcher().AddResourceType("aws:ec2:Instance")

    return &MyPlugin{
        BasePlugin: base,
    }
}

// GetProjectedCost calculates projected costs.
func (p *MyPlugin) GetProjectedCost(
    ctx context.Context,
    req *pbc.GetProjectedCostRequest,
) (*pbc.GetProjectedCostResponse, error) {
    resource := req.GetResource()

    if !p.Matcher().Supports(resource) {
        return nil, pluginsdk.NotSupportedError(resource)
    }

    // Simple example: flat rate for t3.micro instances
    unitPrice := 0.0104 // Hourly rate

    return p.Calculator().CreateProjectedCostResponse(
        "USD",
        unitPrice,
        "on-demand hourly rate",
    ), nil
}

// GetActualCost retrieves historical costs.
func (p *MyPlugin) GetActualCost(
    ctx context.Context,
    req *pbc.GetActualCostRequest,
) (*pbc.GetActualCostResponse, error) {
    // Implement actual cost retrieval
    return nil, pluginsdk.NoDataError(req.GetResourceId())
}

func main() {
    plugin := NewMyPlugin()

    // Set up graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigChan
        log.Println("Shutting down...")
        cancel()
    }()

    // Start serving
    config := pluginsdk.ServeConfig{
        Plugin: plugin,
        Port:   0, // Auto-select port
    }

    log.Printf("Starting %s plugin...", plugin.Name())
    if err := pluginsdk.Serve(ctx, config); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}
```

### Step 3: Build and Test

```bash
# Build the plugin
go build -o my-plugin .

# Run the plugin manually
./my-plugin
# Output: PORT=50051 (or similar)

# In another terminal, test with grpcurl
grpcurl -plaintext localhost:50051 list
```

---

## Implementing Core Functionality

### Resource Matching

Use the `ResourceMatcher` to determine which resources your plugin supports:

```go
func NewMyPlugin() *MyPlugin {
    base := pluginsdk.NewBasePlugin("my-plugin")

    // Support specific providers
    base.Matcher().AddProvider("aws")
    base.Matcher().AddProvider("azure")

    // Support specific resource types
    base.Matcher().AddResourceType("aws:ec2:Instance")
    base.Matcher().AddResourceType("aws:rds:Instance")
    base.Matcher().AddResourceType("azure:compute:VirtualMachine")

    return &MyPlugin{BasePlugin: base}
}
```

### Cost Calculation

Implement projected cost calculation with the `CostCalculator`:

```go
func (p *MyPlugin) GetProjectedCost(
    ctx context.Context,
    req *pbc.GetProjectedCostRequest,
) (*pbc.GetProjectedCostResponse, error) {
    resource := req.GetResource()

    if !p.Matcher().Supports(resource) {
        return nil, pluginsdk.NotSupportedError(resource)
    }

    // Extract resource properties
    instanceType := resource.GetTags()["instanceType"]
    region := resource.GetTags()["region"]

    // Calculate hourly rate
    unitPrice := p.lookupPrice(instanceType, region)

    // Use helper to create response
    return p.Calculator().CreateProjectedCostResponse(
        "USD",
        unitPrice,
        "on-demand pricing",
    ), nil
}
```

### Actual Cost Retrieval

Implement historical cost retrieval:

```go
func (p *MyPlugin) GetActualCost(
    ctx context.Context,
    req *pbc.GetActualCostRequest,
) (*pbc.GetActualCostResponse, error) {
    resourceID := req.GetResourceId()
    startTime := req.GetStart().AsTime()
    endTime := req.GetEnd().AsTime()

    // Query your cost API
    costs, err := p.apiClient.GetCosts(ctx, resourceID, startTime, endTime)
    if err != nil {
        return nil, fmt.Errorf("fetching costs: %w", err)
    }

    // Convert to protobuf results
    var results []*pbc.ActualCostResult
    for _, cost := range costs {
        results = append(results, &pbc.ActualCostResult{
            Timestamp:   timestamppb.New(cost.Timestamp),
            Cost:        cost.Amount,
            UsageAmount: cost.Hours,
            UsageUnit:   "hour",
            Source:      "my-api",
        })
    }

    return p.Calculator().CreateActualCostResponse(results), nil
}
```

### Configuration Management

Load configuration from `~/.pulumicost/config.yaml`:

```go
package config

import (
    "fmt"
    "os"
    "path/filepath"

    "gopkg.in/yaml.v3"
)

type Config struct {
    APIKey   string `yaml:"api_key"`
    Endpoint string `yaml:"endpoint"`
}

func Load(pluginName string) (*Config, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return nil, fmt.Errorf("getting home dir: %w", err)
    }

    configPath := filepath.Join(homeDir, ".pulumicost", "config.yaml")
    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil, fmt.Errorf("reading config: %w", err)
    }

    var fullConfig map[string]interface{}
    if err := yaml.Unmarshal(data, &fullConfig); err != nil {
        return nil, fmt.Errorf("parsing config: %w", err)
    }

    integrations, ok := fullConfig["integrations"].(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("missing integrations section")
    }

    pluginConfig, ok := integrations[pluginName].(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("missing config for plugin %s", pluginName)
    }

    return &Config{
        APIKey:   pluginConfig["api_key"].(string),
        Endpoint: pluginConfig["endpoint"].(string),
    }, nil
}
```

---

## Testing Your Plugin

### Unit Tests

Create `main_test.go`:

```go
package main

import (
    "context"
    "testing"

    pbc "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestGetProjectedCost(t *testing.T) {
    plugin := NewMyPlugin()

    tests := []struct {
        name          string
        resource      *pbc.ResourceDescriptor
        expectError   bool
        expectedPrice float64
    }{
        {
            name: "supported resource",
            resource: &pbc.ResourceDescriptor{
                Provider:     "aws",
                ResourceType: "aws:ec2:Instance",
                Tags: map[string]string{
                    "instanceType": "t3.micro",
                    "region":       "us-east-1",
                },
            },
            expectError:   false,
            expectedPrice: 0.0104,
        },
        {
            name: "unsupported provider",
            resource: &pbc.ResourceDescriptor{
                Provider:     "gcp",
                ResourceType: "gcp:compute:Instance",
            },
            expectError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := &pbc.GetProjectedCostRequest{
                Resource: tt.resource,
            }

            resp, err := plugin.GetProjectedCost(context.Background(), req)

            if tt.expectError {
                assert.Error(t, err)
                return
            }

            require.NoError(t, err)
            assert.Equal(t, tt.expectedPrice, resp.GetUnitPrice())
            assert.Equal(t, "USD", resp.GetCurrency())
        })
    }
}
```

### Integration Tests

Test with PulumiCost CLI:

```bash
# Build plugin
go build -o my-plugin .

# Install plugin
mkdir -p ~/.pulumicost/plugins/my-plugin/1.0.0
cp my-plugin ~/.pulumicost/plugins/my-plugin/1.0.0/

# Create manifest
cat > ~/.pulumicost/plugins/my-plugin/1.0.0/plugin.manifest.yaml <<EOF
name: my-plugin
version: 1.0.0
description: My custom cost plugin
author: Your Name
supported_providers:
  - aws
protocols:
  - grpc
binary: my-plugin
EOF

# Test with PulumiCost
pulumicost plugin list
pulumicost cost projected --pulumi-json test-plan.json
```

### Manual gRPC Testing

Test individual RPC methods:

```bash
# Start plugin
./my-plugin &
PLUGIN_PID=$!

# Test Name RPC
grpcurl -plaintext localhost:50051 \
  pulumicost.v1.CostSourceService/Name

# Test GetProjectedCost
grpcurl -plaintext -d '{
  "resource": {
    "provider": "aws",
    "resource_type": "aws:ec2:Instance",
    "tags": {"instanceType": "t3.micro"}
  }
}' localhost:50051 \
  pulumicost.v1.CostSourceService/GetProjectedCost

# Cleanup
kill $PLUGIN_PID
```

---

## Deployment

### Plugin Installation Structure

Plugins are installed in this directory structure:

```text
~/.pulumicost/
└── plugins/
    └── <plugin-name>/
        └── <version>/
            ├── <plugin-binary>
            └── plugin.manifest.yaml (optional)
```

### Creating a Manifest

Create `plugin.manifest.yaml`:

```yaml
name: my-plugin
version: 1.0.0
description: Custom cost source plugin
author: Your Name
supported_providers:
  - aws
  - azure
protocols:
  - grpc
binary: my-plugin
metadata:
  homepage: https://github.com/yourusername/my-plugin
  license: MIT
```

### Building Release Binaries

Use cross-compilation for multiple platforms:

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o my-plugin-linux-amd64 .

# macOS ARM64
GOOS=darwin GOARCH=arm64 go build -o my-plugin-darwin-arm64 .

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o my-plugin-windows-amd64.exe .
```

### Distribution

Create installation script:

```bash
#!/bin/bash
# install.sh

PLUGIN_NAME="my-plugin"
VERSION="1.0.0"
INSTALL_DIR="$HOME/.pulumicost/plugins/$PLUGIN_NAME/$VERSION"

# Detect platform
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS-$ARCH" in
    linux-x86_64)  BINARY="my-plugin-linux-amd64" ;;
    darwin-arm64)  BINARY="my-plugin-darwin-arm64" ;;
    darwin-x86_64) BINARY="my-plugin-darwin-amd64" ;;
    *)
        echo "Unsupported platform: $OS-$ARCH"
        exit 1
        ;;
esac

# Create install directory
mkdir -p "$INSTALL_DIR"

# Download and install
RELEASE_URL="https://github.com/yourusername/my-plugin/releases"
curl -L "$RELEASE_URL/download/v$VERSION/$BINARY" \
    -o "$INSTALL_DIR/my-plugin"
chmod +x "$INSTALL_DIR/my-plugin"

# Install manifest
curl -L "$RELEASE_URL/download/v$VERSION/plugin.manifest.yaml" \
    -o "$INSTALL_DIR/plugin.manifest.yaml"

echo "Installed $PLUGIN_NAME version $VERSION"
```

---

## Best Practices

### Error Handling

Always provide detailed error messages:

```go
func (p *MyPlugin) GetProjectedCost(
    ctx context.Context,
    req *pbc.GetProjectedCostRequest,
) (*pbc.GetProjectedCostResponse, error) {
    if req == nil {
        return nil, errors.New("GetProjectedCostRequest cannot be nil")
    }

    resource := req.GetResource()
    if resource == nil {
        return nil, errors.New("resource cannot be nil")
    }

    if !p.Matcher().Supports(resource) {
        return nil, pluginsdk.NotSupportedError(resource)
    }

    // Implementation...
}
```

### Context Handling

Respect context cancellation and timeouts:

```go
func (p *MyPlugin) GetActualCost(
    ctx context.Context,
    req *pbc.GetActualCostRequest,
) (*pbc.GetActualCostResponse, error) {
    // Check context before expensive operations
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    // Use context in API calls
    costs, err := p.apiClient.GetCostsWithContext(ctx, req.GetResourceId())
    if err != nil {
        return nil, fmt.Errorf("fetching costs: %w", err)
    }

    return p.Calculator().CreateActualCostResponse(costs), nil
}
```

### Logging

Use structured logging:

```go
import "log/slog"

func (p *MyPlugin) GetProjectedCost(
    ctx context.Context,
    req *pbc.GetProjectedCostRequest,
) (*pbc.GetProjectedCostResponse, error) {
    resource := req.GetResource()

    slog.Info("calculating projected cost",
        "provider", resource.GetProvider(),
        "resource_type", resource.GetResourceType(),
        "sku", resource.GetSku(),
    )

    // Implementation...
}
```

### Caching

Implement caching for frequently accessed data:

```go
import (
    "sync"
    "time"
)

type PriceCache struct {
    mu     sync.RWMutex
    prices map[string]*CachedPrice
}

type CachedPrice struct {
    Value     float64
    ExpiresAt time.Time
}

func (c *PriceCache) Get(key string) (float64, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    price, exists := c.prices[key]
    if !exists || time.Now().After(price.ExpiresAt) {
        return 0, false
    }

    return price.Value, true
}

func (c *PriceCache) Set(key string, value float64, ttl time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.prices[key] = &CachedPrice{
        Value:     value,
        ExpiresAt: time.Now().Add(ttl),
    }
}
```

### Rate Limiting

Implement rate limiting for API calls:

```go
import "golang.org/x/time/rate"

type RateLimitedClient struct {
    limiter *rate.Limiter
    client  *http.Client
}

func NewRateLimitedClient(requestsPerSecond int) *RateLimitedClient {
    return &RateLimitedClient{
        limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), 1),
        client:  &http.Client{},
    }
}

func (c *RateLimitedClient) Get(
    ctx context.Context, url string,
) (*http.Response, error) {
    if err := c.limiter.Wait(ctx); err != nil {
        return nil, fmt.Errorf("rate limit wait: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    return c.client.Do(req)
}
```

---

## Advanced Topics

### Health Checks

Implement the optional `ObservabilityService`:

```go
func (p *MyPlugin) HealthCheck(
    ctx context.Context,
    req *pbc.HealthCheckRequest,
) (*pbc.HealthCheckResponse, error) {
    // Check API connectivity
    if err := p.apiClient.Ping(ctx); err != nil {
        return &pbc.HealthCheckResponse{
            Status: pbc.HealthCheckResponse_STATUS_NOT_SERVING,
            Message: fmt.Sprintf("API unavailable: %v", err),
        }, nil
    }

    return &pbc.HealthCheckResponse{
        Status: pbc.HealthCheckResponse_STATUS_SERVING,
        Message: "healthy",
    }, nil
}
```

### Metrics and Monitoring

Expose Prometheus metrics:

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "plugin_request_duration_seconds",
            Help: "Request duration in seconds",
        },
        []string{"method"},
    )
)

func init() {
    prometheus.MustRegister(requestDuration)
}

func (p *MyPlugin) GetProjectedCost(
    ctx context.Context,
    req *pbc.GetProjectedCostRequest,
) (*pbc.GetProjectedCostResponse, error) {
    timer := prometheus.NewTimer(
        requestDuration.WithLabelValues("GetProjectedCost"))
    defer timer.ObserveDuration()

    // Implementation...
}

// Serve metrics on separate port
go func() {
    http.Handle("/metrics", promhttp.Handler())
    log.Fatal(http.ListenAndServe(":9090", nil))
}()
```

### Multi-Provider Plugins

Support multiple cloud providers in one plugin:

```go
type MultiCloudPlugin struct {
    *pluginsdk.BasePlugin

    awsClient   *AWSClient
    azureClient *AzureClient
    gcpClient   *GCPClient
}

func (p *MultiCloudPlugin) GetProjectedCost(
    ctx context.Context,
    req *pbc.GetProjectedCostRequest,
) (*pbc.GetProjectedCostResponse, error) {
    resource := req.GetResource()

    switch resource.GetProvider() {
    case "aws":
        return p.awsClient.GetCost(ctx, resource)
    case "azure":
        return p.azureClient.GetCost(ctx, resource)
    case "gcp":
        return p.gcpClient.GetCost(ctx, resource)
    default:
        return nil, pluginsdk.NotSupportedError(resource)
    }
}
```

---

## Related Documentation

- [Plugin SDK Reference](plugin-sdk.md) - SDK API documentation
- [Plugin Examples](plugin-examples.md) - Common patterns and examples
- [Plugin Protocol](../architecture/plugin-protocol.md) - Protocol spec
- [Developer Guide](../guides/developer-guide.md) - Contributing to core
