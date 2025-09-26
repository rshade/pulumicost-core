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

// AWSExamplePlugin demonstrates a basic AWS cost calculation plugin.
type AWSExamplePlugin struct {
	*pluginsdk.BasePlugin
}

// "aws:ec2:Instance", "aws:s3:Bucket", and "aws:rds:Instance".
func NewAWSExamplePlugin() *AWSExamplePlugin {
	base := pluginsdk.NewBasePlugin("aws-example")

	// Configure supported AWS provider
	base.Matcher().AddProvider("aws")

	// Add supported resource types
	base.Matcher().AddResourceType("aws:ec2:Instance")
	base.Matcher().AddResourceType("aws:s3:Bucket")
	base.Matcher().AddResourceType("aws:rds:Instance")

	return &AWSExamplePlugin{
		BasePlugin: base,
	}
}

// GetProjectedCost calculates projected costs for AWS resources.
func (p *AWSExamplePlugin) GetProjectedCost(ctx context.Context, req *pbc.GetProjectedCostRequest) (*pbc.GetProjectedCostResponse, error) {
	if !p.Matcher().Supports(req.Resource) {
		return nil, pluginsdk.NotSupportedError(req.Resource)
	}

	var unitPrice float64
	var billingDetail string

	switch req.Resource.ResourceType {
	case "aws:ec2:Instance":
		unitPrice = p.calculateEC2Cost(req.Resource)
		billingDetail = "EC2 instance hourly cost"
	case "aws:s3:Bucket":
		unitPrice = p.calculateS3Cost(req.Resource)
		billingDetail = "S3 storage monthly cost per GB"
	case "aws:rds:Instance":
		unitPrice = p.calculateRDSCost(req.Resource)
		billingDetail = "RDS instance hourly cost"
	default:
		return nil, pluginsdk.NotSupportedError(req.Resource)
	}

	return p.Calculator().CreateProjectedCostResponse("USD", unitPrice, billingDetail), nil
}

// GetActualCost retrieves actual historical costs (placeholder implementation).
func (p *AWSExamplePlugin) GetActualCost(ctx context.Context, req *pbc.GetActualCostRequest) (*pbc.GetActualCostResponse, error) {
	// In a real implementation, this would call AWS Cost Explorer API
	return nil, pluginsdk.NoDataError(req.ResourceId)
}

// calculateEC2Cost calculates EC2 instance cost based on instance type and region.
func (p *AWSExamplePlugin) calculateEC2Cost(resource *pbc.ResourceDescriptor) float64 {
	instanceType := resource.Tags["instanceType"]
	region := resource.Tags["region"]

	// Default to t3.micro if not specified
	if instanceType == "" {
		instanceType = "t3.micro"
	}
	if region == "" {
		region = "us-east-1"
	}

	// Simplified pricing - in reality, this would come from AWS Pricing API
	basePrices := map[string]float64{
		"t3.micro":   0.0104,
		"t3.small":   0.0208,
		"t3.medium":  0.0416,
		"t3.large":   0.0832,
		"t3.xlarge":  0.1664,
		"t3.2xlarge": 0.3328,
		"m5.large":   0.096,
		"m5.xlarge":  0.192,
		"c5.large":   0.085,
		"c5.xlarge":  0.17,
	}

	price, exists := basePrices[instanceType]
	if !exists {
		price = 0.0104 // fallback to t3.micro
	}

	// Apply regional multiplier (simplified)
	regionalMultiplier := map[string]float64{
		"us-east-1":      1.0,
		"us-west-2":      1.0,
		"eu-west-1":      1.1,
		"ap-southeast-1": 1.2,
	}

	if multiplier, exists := regionalMultiplier[region]; exists {
		price *= multiplier
	}

	return price
}

// calculateS3Cost calculates S3 storage cost (per GB monthly).
func (p *AWSExamplePlugin) calculateS3Cost(resource *pbc.ResourceDescriptor) float64 {
	storageClass := resource.Tags["storageClass"]
	region := resource.Tags["region"]

	if storageClass == "" {
		storageClass = "STANDARD"
	}
	if region == "" {
		region = "us-east-1"
	}

	// S3 pricing per GB per month
	storagePrices := map[string]float64{
		"STANDARD":     0.023,
		"STANDARD_IA":  0.0125,
		"ONEZONE_IA":   0.01,
		"GLACIER":      0.004,
		"DEEP_ARCHIVE": 0.00099,
	}

	price, exists := storagePrices[storageClass]
	if !exists {
		price = 0.023 // fallback to STANDARD
	}

	return price
}

// calculateRDSCost calculates RDS instance cost based on instance class and engine.
func (p *AWSExamplePlugin) calculateRDSCost(resource *pbc.ResourceDescriptor) float64 {
	instanceClass := resource.Tags["instanceClass"]
	engine := resource.Tags["engine"]
	region := resource.Tags["region"]

	if instanceClass == "" {
		instanceClass = "db.t3.micro"
	}
	if engine == "" {
		engine = "mysql"
	}
	if region == "" {
		region = "us-east-1"
	}

	// RDS pricing (simplified)
	basePrices := map[string]float64{
		"db.t3.micro":  0.017,
		"db.t3.small":  0.034,
		"db.t3.medium": 0.068,
		"db.m5.large":  0.192,
		"db.m5.xlarge": 0.384,
	}

	price, exists := basePrices[instanceClass]
	if !exists {
		price = 0.017 // fallback to db.t3.micro
	}

	// Engine multiplier
	engineMultipliers := map[string]float64{
		"mysql":     1.0,
		"postgres":  1.0,
		"oracle":    2.0,
		"sqlserver": 1.5,
	}

	if multiplier, exists := engineMultipliers[engine]; exists {
		price *= multiplier
	}

	return price
}

// main starts the AWS example plugin, sets up signal-based graceful shutdown, and serves it on an OS-assigned port.
// It registers handlers for SIGINT and SIGTERM to cancel the server context and exits on serve errors.
func main() {
	// Create the plugin implementation
	plugin := NewAWSExamplePlugin()

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

	log.Printf("Starting %s plugin...", plugin.Name())
	if err := pluginsdk.Serve(ctx, config); err != nil {
		log.Fatalf("Failed to serve plugin: %v", err)
	}
}
