package engine

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rshade/pulumicost-core/internal/pluginhost"
	"github.com/rshade/pulumicost-core/internal/proto"
)

const (
	hoursPerDay = 24
)

var (
	ErrNoCostData = errors.New("no cost data available")
)

type SpecLoader interface {
	LoadSpec(provider, service, sku string) (interface{}, error)
}

type Engine struct {
	clients []*pluginhost.Client
	loader  SpecLoader
}

func New(clients []*pluginhost.Client, loader SpecLoader) *Engine {
	return &Engine{
		clients: clients,
		loader:  loader,
	}
}

func (e *Engine) GetProjectedCost(ctx context.Context, resources []ResourceDescriptor) ([]CostResult, error) {
	var results []CostResult

	for _, resource := range resources {
		var resourceResults []CostResult

		for _, client := range e.clients {
			result, err := e.getProjectedCostFromPlugin(ctx, client, resource)
			if err != nil {
				continue
			}
			if result != nil {
				resourceResults = append(resourceResults, *result)
			}
		}

		if len(resourceResults) == 0 {
			results = append(results, CostResult{
				ResourceType: resource.Type,
				ResourceID:   resource.ID,
				Adapter:      "none",
				Currency:     "USD",
				Monthly:      0,
				Hourly:       0,
				Notes:        "No pricing information available",
			})
		} else {
			results = append(results, resourceResults...)
		}
	}

	return results, nil
}

func (e *Engine) GetActualCost(
	ctx context.Context,
	resources []ResourceDescriptor,
	from, to time.Time,
) ([]CostResult, error) {
	var results []CostResult

	for _, resource := range resources {
		for _, client := range e.clients {
			result, err := e.getActualCostFromPlugin(ctx, client, resource, from, to)
			if err != nil {
				continue
			}
			if result != nil {
				results = append(results, *result)
				break
			}
		}
	}

	return results, nil
}

func (e *Engine) getProjectedCostFromPlugin(
	ctx context.Context,
	client *pluginhost.Client,
	resource ResourceDescriptor,
) (*CostResult, error) {
	const defaultEstimate = 100.0
	const hoursPerMonth = 730
	// Try to get pricing from plugin first
	req := &proto.GetProjectedCostRequest{
		Resources: []*proto.ResourceDescriptor{
			{
				Type:       resource.Type,
				Provider:   resource.Provider,
				Properties: convertToProto(resource.Properties),
			},
		},
	}

	resp, err := client.API.GetProjectedCost(ctx, req)
	if err == nil && len(resp.Results) > 0 {
		result := resp.Results[0]
		return &CostResult{
			ResourceType: resource.Type,
			ResourceID:   resource.ID,
			Adapter:      client.Name,
			Currency:     result.Currency,
			Monthly:      result.MonthlyCost,
			Hourly:       result.HourlyCost,
			Notes:        result.Notes,
			Breakdown:    result.CostBreakdown,
		}, nil
	}

	// Fallback to local spec if available
	if e.loader != nil {
		if result := e.getProjectedCostFromSpec(resource); result != nil {
			return result, nil
		}
	}

	return nil, ErrNoCostData
}

func (e *Engine) getProjectedCostFromSpec(resource ResourceDescriptor) *CostResult {
	const hoursPerMonth = 730

	// Extract service and SKU from resource type
	service := extractService(resource.Type)
	sku := extractSKU(resource)

	// Try multiple fallback patterns for spec loading
	var spec *PricingSpec
	var specErr error

	// 1. Try provider-service-sku pattern first
	if sku != "" {
		if specData, err := e.loader.LoadSpec(resource.Provider, service, sku); err == nil {
			if s, ok := specData.(*PricingSpec); ok {
				spec = s
			}
		}
	}

	// 2. Fallback to provider-service-default pattern
	if spec == nil {
		if specData, err := e.loader.LoadSpec(resource.Provider, service, "default"); err == nil {
			if s, ok := specData.(*PricingSpec); ok {
				spec = s
			}
		}
	}

	// 3. Last resort: try common resource patterns
	if spec == nil {
		commonSKUs := []string{"standard", "basic", "default"}
		for _, commonSKU := range commonSKUs {
			if specData, err := e.loader.LoadSpec(resource.Provider, service, commonSKU); err == nil {
				if s, ok := specData.(*PricingSpec); ok {
					spec = s
					break
				}
			}
		}
	}

	if spec == nil {
		return nil
	}

	// Calculate costs from spec pricing data
	monthly, hourly := calculateCostsFromSpec(spec, resource)

	return &CostResult{
		ResourceType: resource.Type,
		ResourceID:   resource.ID,
		Adapter:      "local-spec",
		Currency:     spec.Currency,
		Monthly:      monthly,
		Hourly:       hourly,
		Notes:        fmt.Sprintf("Calculated from local spec: %s-%s-%s", spec.Provider, spec.Service, spec.SKU),
		Breakdown: map[string]float64{
			"base_cost": monthly,
		},
	}
}

func (e *Engine) getActualCostFromPlugin(
	ctx context.Context,
	client *pluginhost.Client,
	resource ResourceDescriptor,
	from, to time.Time,
) (*CostResult, error) {
	req := &proto.GetActualCostRequest{
		ResourceIDs: []string{resource.ID},
		StartTime:   from.Unix(),
		EndTime:     to.Unix(),
	}

	resp, err := client.API.GetActualCost(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(resp.Results) == 0 {
		return nil, ErrNoCostData
	}

	result := resp.Results[0]
	return &CostResult{
		ResourceType: resource.Type,
		ResourceID:   resource.ID,
		Adapter:      client.Name,
		Currency:     result.Currency,
		Monthly:      result.TotalCost * 30.44 / float64(to.Sub(from).Hours()/hoursPerDay), // Approximate monthly
		Hourly:       result.TotalCost / float64(to.Sub(from).Hours()),
		Notes:        fmt.Sprintf("Actual cost from %s to %s", from.Format("2006-01-02"), to.Format("2006-01-02")),
		Breakdown:    result.CostBreakdown,
	}, nil
}

func convertToProto(properties map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range properties {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
}

func extractService(resourceType string) string {
	// Extract service from resource type like "aws:ec2:Instance" -> "ec2"
	parts := strings.Split(resourceType, ":")
	if len(parts) >= 2 {
		return parts[1]
	}
	return "default"
}

func extractSKU(resource ResourceDescriptor) string {
	// Try to extract SKU from various property keys
	if properties := resource.Properties; properties != nil {
		// Check common SKU property names
		if sku, ok := properties["instanceType"]; ok {
			if skuStr, ok := sku.(string); ok {
				return skuStr
			}
		}
		if sku, ok := properties["sku"]; ok {
			if skuStr, ok := sku.(string); ok {
				return skuStr
			}
		}
		if sku, ok := properties["size"]; ok {
			if skuStr, ok := sku.(string); ok {
				return skuStr
			}
		}
		if sku, ok := properties["type"]; ok {
			if skuStr, ok := sku.(string); ok {
				return skuStr
			}
		}
	}

	// Extract from resource type itself, e.g., "aws:ec2:Instance" -> "Instance"
	parts := strings.Split(resource.Type, ":")
	if len(parts) >= 3 {
		return strings.ToLower(parts[2])
	}

	return ""
}

func calculateCostsFromSpec(spec *PricingSpec, resource ResourceDescriptor) (monthly, hourly float64) {
	const hoursPerMonth = 730

	// Try to extract cost information from spec pricing
	if spec.Pricing != nil {
		// Look for common pricing fields
		if monthlyValue, ok := spec.Pricing["monthlyEstimate"]; ok {
			if monthlyFloat, ok := monthlyValue.(float64); ok {
				monthly = monthlyFloat
				hourly = monthly / hoursPerMonth
				return
			}
		}

		if hourlyValue, ok := spec.Pricing["onDemandHourly"]; ok {
			if hourlyFloat, ok := hourlyValue.(float64); ok {
				hourly = hourlyFloat
				monthly = hourly * hoursPerMonth
				return
			}
		}

		if hourlyValue, ok := spec.Pricing["hourlyRate"]; ok {
			if hourlyFloat, ok := hourlyValue.(float64); ok {
				hourly = hourlyFloat
				monthly = hourly * hoursPerMonth
				return
			}
		}

		// For storage resources, calculate based on size
		if sizeGB, hasSize := getStorageSize(resource); hasSize {
			if pricePerGB, ok := spec.Pricing["pricePerGBMonth"]; ok {
				if priceFloat, ok := pricePerGB.(float64); ok {
					monthly = sizeGB * priceFloat
					hourly = monthly / hoursPerMonth
					return
				}
			}
		}

		// Fallback to any numeric value in pricing
		for _, value := range spec.Pricing {
			if floatValue, ok := value.(float64); ok && floatValue > 0 {
				monthly = floatValue * hoursPerMonth // Assume it's hourly
				hourly = floatValue
				return
			}
		}
	}

	// Ultimate fallback - conservative estimate based on resource type
	if strings.Contains(strings.ToLower(resource.Type), "database") ||
		strings.Contains(strings.ToLower(resource.Type), "rds") {
		monthly = 50.0 // Database default
	} else if strings.Contains(strings.ToLower(resource.Type), "storage") ||
		strings.Contains(strings.ToLower(resource.Type), "s3") {
		monthly = 5.0 // Storage default
	} else {
		monthly = 20.0 // Compute default
	}

	hourly = monthly / hoursPerMonth
	return
}

func getStorageSize(resource ResourceDescriptor) (float64, bool) {
	if resource.Properties == nil {
		return 0, false
	}

	// Check common size property names
	sizeKeys := []string{"size", "sizeGb", "volumeSize", "allocatedStorage"}
	for _, key := range sizeKeys {
		if value, ok := resource.Properties[key]; ok {
			if sizeFloat, ok := value.(float64); ok {
				return sizeFloat, true
			}
			if sizeInt, ok := value.(int); ok {
				return float64(sizeInt), true
			}
			if sizeStr, ok := value.(string); ok {
				// Try to parse size from string
				if parsed, err := strconv.ParseFloat(sizeStr, 64); err == nil {
					return parsed, true
				}
			}
		}
	}

	return 0, false
}

func AggregateResults(results []CostResult) *AggregatedResults {
	if len(results) == 0 {
		return &AggregatedResults{
			Summary: CostSummary{
				Currency:   "USD",
				ByProvider: make(map[string]float64),
				ByService:  make(map[string]float64),
				ByAdapter:  make(map[string]float64),
				Resources:  []CostResult{},
			},
			Resources: []CostResult{},
		}
	}

	summary := CostSummary{
		Currency:   results[0].Currency, // Use currency from first result
		ByProvider: make(map[string]float64),
		ByService:  make(map[string]float64),
		ByAdapter:  make(map[string]float64),
		Resources:  results,
	}

	for _, result := range results {
		// Aggregate totals
		summary.TotalMonthly += result.Monthly
		summary.TotalHourly += result.Hourly

		// Aggregate by provider
		provider := extractProviderFromType(result.ResourceType)
		summary.ByProvider[provider] += result.Monthly

		// Aggregate by service
		service := extractService(result.ResourceType)
		summary.ByService[service] += result.Monthly

		// Aggregate by adapter
		summary.ByAdapter[result.Adapter] += result.Monthly
	}

	return &AggregatedResults{
		Summary:   summary,
		Resources: results,
	}
}

func extractProviderFromType(resourceType string) string {
	// Extract provider from resource type like "aws:ec2:Instance" -> "aws"
	parts := strings.Split(resourceType, ":")
	if len(parts) >= 1 {
		return parts[0]
	}
	return "unknown"
}

func FilterResources(resources []ResourceDescriptor, filter string) []ResourceDescriptor {
	if filter == "" {
		return resources
	}

	var filtered []ResourceDescriptor
	for _, resource := range resources {
		if matchesFilter(resource, filter) {
			filtered = append(filtered, resource)
		}
	}
	return filtered
}

func matchesFilter(resource ResourceDescriptor, filter string) bool {
	// Parse filter format like "type=aws:ec2/instance" or "provider=aws"
	parts := strings.SplitN(filter, "=", 2)
	if len(parts) != 2 {
		return true // Invalid filter, include all
	}

	key := strings.ToLower(strings.TrimSpace(parts[0]))
	value := strings.ToLower(strings.TrimSpace(parts[1]))

	switch key {
	case "type":
		return strings.Contains(strings.ToLower(resource.Type), value)
	case "provider":
		provider := extractProviderFromType(resource.Type)
		return strings.Contains(strings.ToLower(provider), value)
	case "service":
		service := extractService(resource.Type)
		return strings.Contains(strings.ToLower(service), value)
	case "id":
		return strings.Contains(strings.ToLower(resource.ID), value)
	default:
		// Check properties
		if resource.Properties != nil {
			if propValue, ok := resource.Properties[key]; ok {
				return strings.Contains(strings.ToLower(fmt.Sprintf("%v", propValue)), value)
			}
		}
		return false
	}
}
