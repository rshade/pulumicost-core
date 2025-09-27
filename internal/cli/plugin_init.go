package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rshade/pulumicost-core/pkg/pluginsdk"
	"github.com/spf13/cobra"
)

type PluginInitOptions struct {
	Name      string
	Author    string
	Providers []string
	OutputDir string
	Force     bool
}

// the generated project into the specified output directory.
func NewPluginInitCmd() *cobra.Command {
	var opts PluginInitOptions

	cmd := &cobra.Command{
		Use:   "init <plugin-name>",
		Short: "Initialize a new plugin development project",
		Long: `Initialize a new plugin development project with boilerplate code and manifest.

This command creates a new directory structure for plugin development including:
- Go module initialization
- Plugin manifest (manifest.yaml)
- Boilerplate main.go and plugin implementation
- Makefile with build scripts
- README.md with development instructions
- Example tests`,
		Args: cobra.ExactArgs(1),
		Example: `  # Initialize an AWS plugin
  pulumicost plugin init aws-plugin --author "Your Name" --providers aws

  # Initialize a multi-provider plugin
  pulumicost plugin init cloud-plugin --author "Your Name" --providers aws,azure,gcp

  # Initialize in a specific directory
  pulumicost plugin init my-plugin --author "Your Name" --providers aws --output-dir /path/to/plugins`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name = args[0]
			return runPluginInit(cmd, &opts)
		},
	}

	cmd.Flags().StringVar(&opts.Author, "author", "", "Plugin author name (required)")
	cmd.Flags().
		StringSliceVar(&opts.Providers, "providers", []string{}, "Supported cloud providers (e.g., aws,azure,gcp)")
	cmd.Flags().StringVar(&opts.OutputDir, "output-dir", ".", "Output directory for plugin project")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Overwrite existing files if directory exists")

	_ = cmd.MarkFlagRequired("author")
	_ = cmd.MarkFlagRequired("providers")

	return cmd
}

// runPluginInit validates the provided PluginInitOptions, creates the target project directory (honoring the force flag),
// generates the boilerplate project files, and prints progress and next-step instructions to the command output.
// It returns an error if validation fails (invalid name or no providers), if the directory cannot be created, or if file generation fails.
func runPluginInit(cmd *cobra.Command, opts *PluginInitOptions) error {
	// Validate plugin name
	if !isValidPluginName(opts.Name) {
		return fmt.Errorf(
			"invalid plugin name: %s (must contain only lowercase letters, numbers, and hyphens)",
			opts.Name,
		)
	}

	// Validate providers
	if len(opts.Providers) == 0 {
		return fmt.Errorf("at least one provider must be specified")
	}

	// Create project directory
	projectDir := filepath.Join(opts.OutputDir, opts.Name)
	if err := createProjectDirectory(projectDir, opts.Force); err != nil {
		return fmt.Errorf("creating project directory: %w", err)
	}

	cmd.Printf("Initializing plugin project: %s\n", opts.Name)
	cmd.Printf("Project directory: %s\n", projectDir)
	cmd.Printf("Author: %s\n", opts.Author)
	cmd.Printf("Supported providers: %s\n", strings.Join(opts.Providers, ", "))

	// Generate project files
	generator := &projectGenerator{
		name:       opts.Name,
		author:     opts.Author,
		providers:  opts.Providers,
		projectDir: projectDir,
		cmd:        cmd,
	}

	if err := generator.generateAll(); err != nil {
		return fmt.Errorf("generating project files: %w", err)
	}

	cmd.Printf("\n✅ Plugin project initialized successfully!\n\n")
	cmd.Printf("Next steps:\n")
	cmd.Printf("1. cd %s\n", projectDir)
	cmd.Printf("2. go mod tidy\n")
	cmd.Printf("3. make build\n")
	cmd.Printf("4. Edit internal/pricing/calculator.go to implement your pricing logic\n")
	cmd.Printf("5. Edit internal/client/client.go to implement your cloud provider client\n\n")
	cmd.Printf("For more information, see the README.md file in your project.\n")

	return nil
}

type projectGenerator struct {
	name       string
	author     string
	providers  []string
	projectDir string
	cmd        *cobra.Command
}

func (g *projectGenerator) generateAll() error {
	steps := []struct {
		name string
		fn   func() error
	}{
		{"Creating directory structure", g.createDirectories},
		{"Generating go.mod", g.generateGoMod},
		{"Generating plugin manifest", g.generateManifest},
		{"Generating main.go", g.generateMainGo},
		{"Generating plugin implementation", g.generatePlugin},
		{"Generating pricing calculator", g.generatePricingCalculator},
		{"Generating cloud client", g.generateCloudClient},
		{"Generating Makefile", g.generateMakefile},
		{"Generating README.md", g.generateReadme},
		{"Generating example tests", g.generateTests},
	}

	for _, step := range steps {
		g.cmd.Printf("  %s...\n", step.name)
		if err := step.fn(); err != nil {
			return fmt.Errorf("%s: %w", step.name, err)
		}
	}

	return nil
}

func (g *projectGenerator) createDirectories() error {
	dirs := []string{
		"cmd/plugin",
		"internal/pricing",
		"internal/client",
		"examples",
		"bin",
	}

	for _, dir := range dirs {
		fullDir := filepath.Join(g.projectDir, dir)
		if err := os.MkdirAll(fullDir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}

	return nil
}

func (g *projectGenerator) generateGoMod() error {
	content := fmt.Sprintf(`module github.com/example/%s

go 1.21

require (
	github.com/rshade/pulumicost-core v0.1.0
	github.com/rshade/pulumicost-spec v0.1.0
	google.golang.org/grpc v1.74.2
)
`, g.name)

	return g.writeFile("go.mod", content)
}

func (g *projectGenerator) generateManifest() error {
	manifest := pluginsdk.CreateDefaultManifest(g.name, g.author, g.providers)
	return manifest.SaveManifest(filepath.Join(g.projectDir, "manifest.yaml"))
}

func (g *projectGenerator) generateMainGo() error {
	content := fmt.Sprintf(`package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/example/%s/internal/pricing"
	"github.com/rshade/pulumicost-core/pkg/pluginsdk"
)

func main() {
	// Create the plugin implementation
	plugin := pricing.NewCalculator()

	// Set up context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received interrupt signal, shutting down...")
		cancel()
	}()

	// Start serving the plugin
	config := pluginsdk.ServeConfig{
		Plugin: plugin,
		Port:   0, // Let the system choose a port
	}

	log.Printf("Starting %%s plugin...", plugin.Name())
	if err := pluginsdk.Serve(ctx, config); err != nil {
		log.Fatalf("Failed to serve plugin: %%v", err)
	}
}
`, g.name)

	return g.writeFile("cmd/plugin/main.go", content)
}

func (g *projectGenerator) generatePlugin() error {
	providersStr := strings.Join(g.providers, `", "`)
	content := fmt.Sprintf(`package pricing

import (
	"context"

	pbc "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
	"github.com/rshade/pulumicost-core/pkg/pluginsdk"
)

// Calculator implements the PulumiCost plugin interface for %s.
type Calculator struct {
	*pluginsdk.BasePlugin
}

// NewCalculator creates a new %s cost calculator plugin.
func NewCalculator() *Calculator {
	base := pluginsdk.NewBasePlugin("%s")
	
	// Configure supported providers
	providers := []string{"%s"}
	for _, provider := range providers {
		base.Matcher().AddProvider(provider)
	}

	return &Calculator{
		BasePlugin: base,
	}
}

// GetProjectedCost calculates projected costs for resources.
func (c *Calculator) GetProjectedCost(ctx context.Context, req *pbc.GetProjectedCostRequest) (*pbc.GetProjectedCostResponse, error) {
	// Check if we support this resource
	if !c.Matcher().Supports(req.Resource) {
		return nil, pluginsdk.NotSupportedError(req.Resource)
	}

	// TODO: Implement your pricing logic here
	// This is a placeholder implementation
	unitPrice := 0.0
	billingDetail := "Pricing not implemented"

	// Example: Basic EC2 instance pricing
	switch req.Resource.ResourceType {
	case "aws:ec2:Instance":
		unitPrice = c.calculateEC2InstanceCost(req.Resource)
		billingDetail = "EC2 instance hourly cost"
	default:
		return nil, pluginsdk.NotSupportedError(req.Resource)
	}

	return c.Calculator().CreateProjectedCostResponse("USD", unitPrice, billingDetail), nil
}

// GetActualCost retrieves actual historical costs.
func (c *Calculator) GetActualCost(ctx context.Context, req *pbc.GetActualCostRequest) (*pbc.GetActualCostResponse, error) {
	// TODO: Implement actual cost retrieval from your cloud provider's billing API
	// This is a placeholder implementation
	return nil, pluginsdk.NoDataError(req.ResourceId)
}

// calculateEC2InstanceCost is an example pricing calculation.
func (c *Calculator) calculateEC2InstanceCost(resource *pbc.ResourceDescriptor) float64 {
	// TODO: Implement actual EC2 pricing logic
	// This is a simplified example - real implementation should:
	// 1. Parse instance type from resource properties
	// 2. Look up pricing from AWS Pricing API or local pricing data
	// 3. Consider region, operating system, tenancy, etc.
	
	instanceType := resource.Tags["instanceType"]
	if instanceType == "" {
		instanceType = "t3.micro" // default
	}

	// Simplified pricing - replace with real pricing data
	switch instanceType {
	case "t3.micro":
		return 0.0104 // $0.0104/hour
	case "t3.small":
		return 0.0208 // $0.0208/hour
	case "t3.medium":
		return 0.0416 // $0.0416/hour
	default:
		return 0.0104 // fallback
	}
}
`, g.name, g.name, g.name, providersStr)

	return g.writeFile("internal/pricing/calculator.go", content)
}

func (g *projectGenerator) generatePricingCalculator() error {
	content := `package pricing

// This file contains pricing calculation utilities and data structures.
// Implement your cloud provider pricing logic here.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// PricingData represents pricing information for resources.
type PricingData struct {
	Provider     string             ` + "`json:\"provider\"`" + `
	Region       string             ` + "`json:\"region\"`" + `
	ResourceType string             ` + "`json:\"resource_type\"`" + `
	Pricing      map[string]float64 ` + "`json:\"pricing\"`" + `
}

// LoadPricingData loads pricing data from a JSON file.
func LoadPricingData(path string) ([]PricingData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading pricing data: %w", err)
	}

	var pricing []PricingData
	if err := json.Unmarshal(data, &pricing); err != nil {
		return nil, fmt.Errorf("parsing pricing data: %w", err)
	}

	return pricing, nil
}

// SavePricingData saves pricing data to a JSON file.
func SavePricingData(path string, data []PricingData) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling pricing data: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return fmt.Errorf("writing pricing data: %w", err)
	}

	return nil
}

// Example pricing data structure - customize for your provider
func CreateExamplePricingData() []PricingData {
	return []PricingData{
		{
			Provider:     "aws",
			Region:       "us-east-1",
			ResourceType: "aws:ec2:Instance",
			Pricing: map[string]float64{
				"t3.micro":  0.0104,
				"t3.small":  0.0208,
				"t3.medium": 0.0416,
				"t3.large":  0.0832,
				"t3.xlarge": 0.1664,
			},
		},
	}
}
`

	return g.writeFile("internal/pricing/data.go", content)
}

func (g *projectGenerator) generateCloudClient() error {
	content := `package client

// This file contains the cloud provider client implementation.
// Implement API clients for your specific cloud provider here.

import (
	"context"
	"fmt"
)

// Client represents a client for your cloud provider's APIs.
type Client struct {
	// Add your cloud provider client configuration here
	// Examples:
	// - AWS SDK client
	// - Azure SDK client  
	// - GCP SDK client
	// - Custom API client
}

// Config holds configuration for the cloud provider client.
type Config struct {
	// Add configuration fields specific to your provider
	// Examples:
	// APIKey      string
	// Region      string
	// Credentials *Credentials
}

// NewClient creates a new cloud provider client.
func NewClient(config Config) (*Client, error) {
	// TODO: Initialize your cloud provider client here
	// Example:
	// return &Client{
	//     awsClient: aws.New(config.AWSConfig),
	// }, nil
	
	return &Client{}, nil
}

// GetResourceCost retrieves actual cost data for a specific resource.
func (c *Client) GetResourceCost(ctx context.Context, resourceID string, startTime, endTime int64) (float64, error) {
	// TODO: Implement actual cost retrieval from your cloud provider
	// This should call the appropriate billing/cost management API
	// Examples:
	// - AWS Cost Explorer API
	// - Azure Cost Management API
	// - GCP Cloud Billing API
	
	return 0.0, fmt.Errorf("not implemented: GetResourceCost for resource %s", resourceID)
}

// ValidateCredentials checks if the client credentials are valid.
func (c *Client) ValidateCredentials(ctx context.Context) error {
	// TODO: Implement credential validation
	// Make a simple API call to verify credentials work
	
	return fmt.Errorf("not implemented: ValidateCredentials")
}

// GetSupportedRegions returns the list of supported regions.
func (c *Client) GetSupportedRegions(ctx context.Context) ([]string, error) {
	// TODO: Implement region discovery
	// Return the list of regions supported by your cloud provider
	
	return []string{"us-east-1", "us-west-2"}, nil
}
`

	return g.writeFile("internal/client/client.go", content)
}

func (g *projectGenerator) generateMakefile() error {
	content := fmt.Sprintf(`# Makefile for %s plugin

.PHONY: build test clean lint help install

# Variables
PLUGIN_NAME = %s
BINARY_NAME = pulumicost-plugin-$(PLUGIN_NAME)
BUILD_DIR = bin
CMD_DIR = cmd/plugin

# Default target
help:
	@echo "Available commands:"
	@echo "  build     - Build the plugin binary"
	@echo "  test      - Run tests"
	@echo "  clean     - Clean build artifacts"
	@echo "  lint      - Run linters"
	@echo "  install   - Install plugin to local registry"
	@echo "  help      - Show this help"

# Build the plugin binary
build:
	@echo "Building $(PLUGIN_NAME) plugin..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "✅ Plugin built: $(BUILD_DIR)/$(BINARY_NAME)"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "✅ Clean complete"

# Run linters
lint:
	@echo "Running linters..."
	@golangci-lint run --allow-parallel-runners
	@echo "✅ Linting complete"

# Install plugin to local registry
install: build
	@echo "Installing plugin to local registry..."
	@mkdir -p ~/.pulumicost/plugins/$(PLUGIN_NAME)/1.0.0
	@cp $(BUILD_DIR)/$(BINARY_NAME) ~/.pulumicost/plugins/$(PLUGIN_NAME)/1.0.0/
	@cp manifest.yaml ~/.pulumicost/plugins/$(PLUGIN_NAME)/1.0.0/plugin.manifest.json
	@echo "✅ Plugin installed to ~/.pulumicost/plugins/$(PLUGIN_NAME)/1.0.0/"

# Development build with debug info
build-debug:
	@echo "Building $(PLUGIN_NAME) plugin with debug info..."
	@mkdir -p $(BUILD_DIR)
	@go build -gcflags="all=-N -l" -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "✅ Debug build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatting complete"

# Update dependencies
deps:
	@echo "Updating dependencies..."
	@go mod tidy
	@go mod download
	@echo "✅ Dependencies updated"

# Check for security vulnerabilities
security:
	@echo "Checking for security vulnerabilities..."
	@govulncheck ./...
	@echo "✅ Security check complete"
`, g.name, g.name)

	return g.writeFile("Makefile", content)
}

func (g *projectGenerator) generateReadme() error {
	providersStr := strings.Join(g.providers, ", ")
	content := fmt.Sprintf(`# %s

PulumiCost plugin for %s cost calculation.

## Overview

This plugin provides cost calculation capabilities for %s resources in PulumiCost. It implements both projected cost estimation and actual cost retrieval functionality.

**Supported Providers:** %s

## Installation

### From Source

1. Clone the repository:
   `+"```bash"+`
   git clone <repository-url>
   cd %s
   `+"```"+`

2. Build the plugin:
   `+"```bash"+`
   make build
   `+"```"+`

3. Install to local plugin registry:
   `+"```bash"+`
   make install
   `+"```"+`

### Configuration

The plugin may require cloud provider credentials to function properly. See the configuration section for details.

## Usage

Once installed, the plugin will be automatically discovered by PulumiCost:

`+"```bash"+`
# List installed plugins
pulumicost plugin list

# Validate plugin installation
pulumicost plugin validate

# Calculate projected costs
pulumicost cost projected --pulumi-json plan.json

# Get actual costs
pulumicost cost actual --pulumi-json plan.json --from 2025-01-01
`+"```"+`

## Development

### Prerequisites

- Go 1.21+
- PulumiCost Core development environment
- Cloud provider credentials (for actual cost retrieval)

### Building

`+"```bash"+`
# Build the plugin
make build

# Run tests
make test

# Run linters
make lint

# Build with debug info
make build-debug
`+"```"+`

### Project Structure

`+"```"+`
%s/
├── cmd/plugin/main.go          # Plugin entry point
├── internal/
│   ├── pricing/               # Pricing calculation logic
│   │   ├── calculator.go      # Main plugin implementation
│   │   └── data.go           # Pricing data structures
│   └── client/               # Cloud provider client
│       └── client.go         # API client implementation
├── examples/                 # Usage examples and test data
├── manifest.yaml            # Plugin manifest
├── Makefile                # Build scripts
└── README.md               # This file
`+"```"+`

### Implementation Guide

#### Projected Cost Calculation

Edit `+"`internal/pricing/calculator.go`"+` to implement your pricing logic:

`+"```go"+`
func (c *Calculator) GetProjectedCost(ctx context.Context, req *pbc.GetProjectedCostRequest) (*pbc.GetProjectedCostResponse, error) {
    // 1. Check if resource is supported
    if !c.Matcher().Supports(req.Resource) {
        return nil, pluginsdk.NotSupportedError(req.Resource)
    }

    // 2. Extract resource properties
    resourceType := req.Resource.ResourceType
    properties := req.Resource.Tags

    // 3. Calculate pricing based on resource type and properties
    unitPrice := c.calculateResourceCost(resourceType, properties)

    // 4. Return response
    return c.Calculator().CreateProjectedCostResponse("USD", unitPrice, "description"), nil
}
`+"```"+`

#### Actual Cost Retrieval

Edit `+"`internal/client/client.go`"+` to implement cloud provider API integration:

`+"```go"+`
func (c *Client) GetResourceCost(ctx context.Context, resourceID string, startTime, endTime int64) (float64, error) {
    // 1. Call cloud provider billing API
    // 2. Parse response and calculate total cost
    // 3. Return cost value
    return totalCost, nil
}
`+"```"+`

### Testing

The project includes testing utilities from the PulumiCost SDK:

`+"```go"+`
func TestPluginName(t *testing.T) {
    plugin := pricing.NewCalculator()
    testPlugin := pluginsdk.NewTestPlugin(t, plugin)
    testPlugin.TestName("%s")
}
`+"```"+`

### Adding Pricing Data

1. Update pricing data structures in `+"`internal/pricing/data.go`"+`
2. Implement pricing lookups in `+"`internal/pricing/calculator.go`"+`
3. Add test cases for new resource types

### Configuration

The plugin supports the following configuration options:

- Environment variables for cloud provider credentials
- Pricing data files for offline pricing calculations
- Regional pricing variations

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run `+"`make lint test`"+`
6. Submit a pull request

## License

[Add your license information here]

## Support

[Add support contact information here]
`, g.name, providersStr, providersStr, providersStr, g.name, g.name, g.name)

	return g.writeFile("README.md", content)
}

func (g *projectGenerator) generateTests() error {
	content := fmt.Sprintf(`package pricing

import (
	"testing"

	pbc "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
	"github.com/rshade/pulumicost-core/pkg/pluginsdk"
)

func TestCalculatorName(t *testing.T) {
	plugin := NewCalculator()
	testPlugin := pluginsdk.NewTestPlugin(t, plugin)
	testPlugin.TestName("%s")
}

func TestProjectedCostSupported(t *testing.T) {
	plugin := NewCalculator()
	testPlugin := pluginsdk.NewTestPlugin(t, plugin)

	// Test supported resource
	resource := pluginsdk.CreateTestResource("aws", "aws:ec2:Instance", map[string]string{
		"instanceType": "t3.micro",
		"region":       "us-east-1",
	})

	resp := testPlugin.TestProjectedCost(resource, false)
	if resp == nil {
		t.Fatal("Expected response, got nil")
	}

	if resp.Currency != "USD" {
		t.Errorf("Expected currency USD, got %%s", resp.Currency)
	}

	if resp.UnitPrice <= 0 {
		t.Errorf("Expected positive unit price, got %%f", resp.UnitPrice)
	}
}

func TestProjectedCostUnsupported(t *testing.T) {
	plugin := NewCalculator()
	testPlugin := pluginsdk.NewTestPlugin(t, plugin)

	// Test unsupported resource
	resource := pluginsdk.CreateTestResource("unsupported", "unsupported:resource:Type", nil)
	testPlugin.TestProjectedCost(resource, true) // Expect error
}

func TestActualCost(t *testing.T) {
	plugin := NewCalculator()
	testPlugin := pluginsdk.NewTestPlugin(t, plugin)

	// Test actual cost (should return error since not implemented)
	testPlugin.TestActualCost("resource-id-123", 1640995200, 1641081600, true) // Expect error
}

// Example of more specific test cases
func TestEC2InstancePricing(t *testing.T) {
	calculator := NewCalculator()

	testCases := []struct {
		name         string
		instanceType string
		expectedCost float64
	}{
		{"t3.micro", "t3.micro", 0.0104},
		{"t3.small", "t3.small", 0.0208},
		{"t3.medium", "t3.medium", 0.0416},
		{"unknown", "unknown-type", 0.0104}, // fallback
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource := &pbc.ResourceDescriptor{
				Provider:     "aws",
				ResourceType: "aws:ec2:Instance",
				Tags: map[string]string{
					"instanceType": tc.instanceType,
				},
			}

			cost := calculator.calculateEC2InstanceCost(resource)
			if cost != tc.expectedCost {
				t.Errorf("Expected cost %%f for %%s, got %%f", tc.expectedCost, tc.instanceType, cost)
			}
		})
	}
}
`, g.name)

	return g.writeFile("internal/pricing/calculator_test.go", content)
}

func (g *projectGenerator) writeFile(relativePath, content string) error {
	fullPath := filepath.Join(g.projectDir, relativePath)

	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing file %s: %w", relativePath, err)
	}

	return nil
}

// createProjectDirectory ensures a directory exists at the given path, creating it and any
// necessary parent directories with 0755 permissions.
// If the path already exists and force is false, it returns an error indicating the directory
// already exists. Returns an error on failure to create the directory, or nil on success.
func createProjectDirectory(path string, force bool) error {
	if _, err := os.Stat(path); err == nil {
		if !force {
			return fmt.Errorf("directory already exists: %s (use --force to overwrite)", path)
		}
	}

	return os.MkdirAll(path, 0755)
}

// isValidPluginName reports whether the provided name satisfies the plugin naming rules.
// The name must be between 2 and 50 characters, contain only lowercase letters (`a`–`z`),
// digits (`0`–`9`) or hyphens (`-`), and must not start or end with a hyphen.
// It returns true when all conditions are met, false otherwise.
func isValidPluginName(name string) bool {
	if len(name) < 2 || len(name) > 50 {
		return false
	}

	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
			return false
		}
	}

	// Cannot start or end with hyphen
	return name[0] != '-' && name[len(name)-1] != '-'
}
