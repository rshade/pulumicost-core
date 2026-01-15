# Internal API Contracts: Plugin Info and DryRun Discovery

## internal/proto/adapter.go

### Updated CostSourceClient Interface

```go
type CostSourceClient interface {
    Name(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*NameResponse, error)
    GetProjectedCost(ctx context.Context, in *GetProjectedCostRequest, opts ...grpc.CallOption) (*GetProjectedCostResponse, error)
    GetActualCost(ctx context.Context, in *GetActualCostRequest, opts ...grpc.CallOption) (*GetActualCostResponse, error)
    GetRecommendations(ctx context.Context, in *GetRecommendationsRequest, opts ...grpc.CallOption) (*GetRecommendationsResponse, error)
    
    // NEW METHODS
    GetPluginInfo(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*PluginMetadata, error)
    DryRun(ctx context.Context, in *DryRunRequest, opts ...grpc.CallOption) (*DryRunResponse, error)
}
```

### New Internal Structs

```go
type PluginMetadata struct {
    Name               string
    Version            string
    SpecVersion        string
    SupportedProviders []string
    Metadata           map[string]string
}

type DryRunRequest struct {
    Resource             *ResourceDescriptor
    SimulationParameters map[string]string
}

type DryRunResponse struct {
    FieldMappings           []*FieldMapping
    ConfigurationValid      bool
    ConfigurationErrors     []string
    ResourceTypeSupported   bool
}

type FieldMapping struct {
    FieldName    string
    Status       string // SUPPORTED, UNSUPPORTED, CONDITIONAL, DYNAMIC
    Condition    string
    ExpectedType string
}
```

## CLI Commands

### finfocus plugin inspect

**Usage**: `finfocus plugin inspect <plugin-name> <resource-type> [flags]`

**Arguments**:
- `plugin-name`: Name of the installed plugin (e.g., `aws-public`).
- `resource-type`: Pulumi resource type (e.g., `aws:ec2/instance:Instance`).

**Flags**:
- `--json`: Output in JSON format.
- `--version <v>`: Specific version of the plugin if multiple are installed.

### finfocus plugin list

**Updated Output**:
Will include `VERSION` and `SPEC` columns.

```text
NAME         VERSION   SPEC      STATUS    PROVIDERS
aws-public   v0.1.0    v0.4.11   Active    aws
vantage      v1.2.3    v0.4.11   Active    aws, azure, gcp
```
