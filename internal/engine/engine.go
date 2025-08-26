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
	hoursPerDay                 = 24
	hoursPerMonth               = 730
	minProviderServiceParts     = 2    // For "provider:service"
	minProviderServiceTypeParts = 3    // For "provider:service:type"
	filterKeyValueParts         = 2    // For "key=value"
	defaultDatabaseMonthlyCost  = 50.0 // Default monthly cost for database resources
	defaultStorageMonthlyCost   = 5.0  // Default monthly cost for storage resources
	defaultComputeMonthlyCost   = 20.0 // Default monthly cost for compute resources
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
			// Single spec fallback per resource
			if e.loader != nil {
				if specRes := e.getProjectedCostFromSpec(resource); specRes != nil {
					results = append(results, *specRes)
					continue
				}
			}

			// Final fallback: no cost data available
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

	return nil, ErrNoCostData
}

func (e *Engine) getProjectedCostFromSpec(resource ResourceDescriptor) *CostResult {
	service := extractService(resource.Type)
	sku := extractSKU(resource)

	spec := e.loadSpecWithFallback(resource.Provider, service, sku)
	if spec == nil {
		return nil
	}

	monthly, hourly := calculateCostsFromSpec(spec, resource)
	return e.createSpecBasedResult(resource, spec, monthly, hourly)
}

func (e *Engine) loadSpecWithFallback(provider, service, sku string) *PricingSpec {
	// Try provider-service-sku pattern first
	if sku != "" {
		if spec := e.tryLoadSpec(provider, service, sku); spec != nil {
			return spec
		}
	}

	// Fallback to provider-service-default pattern
	if spec := e.tryLoadSpec(provider, service, "default"); spec != nil {
		return spec
	}

	// Last resort: try common resource patterns
	commonSKUs := []string{"standard", "basic"}
	for _, commonSKU := range commonSKUs {
		if spec := e.tryLoadSpec(provider, service, commonSKU); spec != nil {
			return spec
		}
	}

	return nil
}

func (e *Engine) tryLoadSpec(provider, service, sku string) *PricingSpec {
	specData, err := e.loader.LoadSpec(provider, service, sku)
	if err != nil {
		return nil
	}
	if spec, ok := specData.(*PricingSpec); ok {
		return spec
	}
	return nil
}

func (e *Engine) createSpecBasedResult(
	resource ResourceDescriptor,
	spec *PricingSpec,
	monthly, hourly float64,
) *CostResult {
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
	if len(parts) >= minProviderServiceParts {
		return parts[1]
	}
	return "default"
}

func extractSKU(resource ResourceDescriptor) string {
	// Try to extract SKU from resource properties first
	if sku := extractSKUFromProperties(resource.Properties); sku != "" {
		return sku
	}

	// Fallback to extracting from resource type
	return extractSKUFromType(resource.Type)
}

func extractSKUFromProperties(properties map[string]interface{}) string {
	if properties == nil {
		return ""
	}

	skuKeys := []string{"instanceType", "sku", "size", "type"}
	for _, key := range skuKeys {
		if skuStr, found := getStringProperty(properties, key); found {
			return skuStr
		}
	}
	return ""
}

func extractSKUFromType(resourceType string) string {
	parts := strings.Split(resourceType, ":")
	if len(parts) >= minProviderServiceTypeParts {
		return strings.ToLower(parts[2])
	}
	return ""
}

func getStringProperty(properties map[string]interface{}, key string) (string, bool) {
	if sku, ok := properties[key]; ok {
		if skuStr, isStr := sku.(string); isStr {
			return skuStr, true
		}
	}
	return "", false
}

func calculateCostsFromSpec(spec *PricingSpec, resource ResourceDescriptor) (float64, float64) {
	// Try to extract cost information from spec pricing
	if spec.Pricing != nil {
		if monthlyRate, hourlyRate, found := tryExtractCostsFromPricing(spec.Pricing, resource); found {
			return monthlyRate, hourlyRate
		}
	}

	// Ultimate fallback - conservative estimate based on resource type
	monthly := getDefaultMonthlyByType(resource.Type)
	hourly := monthly / hoursPerMonth
	return monthly, hourly
}

func tryExtractCostsFromPricing(
	pricing map[string]interface{},
	resource ResourceDescriptor,
) (float64, float64, bool) {
	// Try direct monthly estimate first
	if monthly, hourly, found := tryMonthlyEstimate(pricing); found {
		return monthly, hourly, true
	}

	// Try hourly rates
	if monthly, hourly, found := tryHourlyRates(pricing); found {
		return monthly, hourly, true
	}

	// Try storage-based pricing
	if monthly, hourly, found := tryStoragePricing(pricing, resource); found {
		return monthly, hourly, true
	}

	// Fallback to any numeric value
	return tryFallbackNumericValue(pricing)
}

func tryMonthlyEstimate(pricing map[string]interface{}) (float64, float64, bool) {
	if monthlyFloat, ok := getFloatFromPricing(pricing, "monthlyEstimate"); ok {
		monthly := monthlyFloat
		hourly := monthly / hoursPerMonth
		return monthly, hourly, true
	}
	return 0, 0, false
}

func tryHourlyRates(pricing map[string]interface{}) (float64, float64, bool) {
	hourlyKeys := []string{"onDemandHourly", "hourlyRate"}
	for _, key := range hourlyKeys {
		if hourlyFloat, ok := getFloatFromPricing(pricing, key); ok {
			hourly := hourlyFloat
			monthly := hourly * hoursPerMonth
			return monthly, hourly, true
		}
	}
	return 0, 0, false
}

func tryStoragePricing(pricing map[string]interface{}, resource ResourceDescriptor) (float64, float64, bool) {
	sizeGB, hasSize := getStorageSize(resource)
	if !hasSize {
		return 0, 0, false
	}

	if priceFloat, ok := getFloatFromPricing(pricing, "pricePerGBMonth"); ok {
		monthly := sizeGB * priceFloat
		hourly := monthly / hoursPerMonth
		return monthly, hourly, true
	}
	return 0, 0, false
}

func tryFallbackNumericValue(pricing map[string]interface{}) (float64, float64, bool) {
	for _, value := range pricing {
		if floatValue, ok := value.(float64); ok && floatValue > 0 {
			monthly := floatValue * hoursPerMonth // Assume it's hourly
			hourly := floatValue
			return monthly, hourly, true
		}
	}
	return 0, 0, false
}

func getFloatFromPricing(pricing map[string]interface{}, key string) (float64, bool) {
	if value, ok := pricing[key]; ok {
		if floatValue, isFloat := value.(float64); isFloat {
			return floatValue, true
		}
	}
	return 0, false
}

func getDefaultMonthlyByType(resourceType string) float64 {
	resourceTypeLower := strings.ToLower(resourceType)
	switch {
	case strings.Contains(resourceTypeLower, "database") ||
		strings.Contains(resourceTypeLower, "rds"):
		return defaultDatabaseMonthlyCost
	case strings.Contains(resourceTypeLower, "storage") ||
		strings.Contains(resourceTypeLower, "s3"):
		return defaultStorageMonthlyCost
	default:
		return defaultComputeMonthlyCost
	}
}

func getStorageSize(resource ResourceDescriptor) (float64, bool) {
	if resource.Properties == nil {
		return 0, false
	}

	// Check common size property names
	sizeKeys := []string{"size", "sizeGb", "volumeSize", "allocatedStorage"}
	for _, key := range sizeKeys {
		if value, ok := resource.Properties[key]; ok {
			if size, parsed := parseFloatValue(value); parsed {
				return size, true
			}
		}
	}

	return 0, false
}

func parseFloatValue(value interface{}) (float64, bool) {
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
	parts := strings.SplitN(filter, "=", filterKeyValueParts)
	if len(parts) != filterKeyValueParts {
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
		// Check properties (case-insensitive key matching)
		if resource.Properties != nil {
			for propKey, propValue := range resource.Properties {
				if strings.ToLower(propKey) == key {
					return strings.Contains(strings.ToLower(fmt.Sprintf("%v", propValue)), value)
				}
			}
		}
		return false
	}
}
