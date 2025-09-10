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
	daysPerWeek                 = 7     // Days in a week
	daysPerMonth                = 30    // Approximate days per month
	avgDaysPerMonth             = 30.44 // Accurate average days per month for conversions
	minProviderServiceParts     = 2     // For "provider:service"
	minProviderServiceTypeParts = 3     // For "provider:service:type"
	filterKeyValueParts         = 2     // For "key=value"
	defaultDatabaseMonthlyCost  = 50.0  // Default monthly cost for database resources
	defaultStorageMonthlyCost   = 5.0   // Default monthly cost for storage resources
	defaultComputeMonthlyCost   = 20.0  // Default monthly cost for compute resources
	defaultServiceName          = "default"
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
	return e.GetActualCostWithOptions(ctx, ActualCostRequest{
		Resources: resources,
		From:      from,
		To:        to,
	})
}

func (e *Engine) GetActualCostWithOptions(
	ctx context.Context,
	request ActualCostRequest,
) ([]CostResult, error) {
	var results []CostResult
	var partialErrors []error

	for _, resource := range request.Resources {
		// Filter by tags if specified
		if len(request.Tags) > 0 && !MatchesTags(resource, request.Tags) {
			continue
		}

		var resourceResult *CostResult
		for _, client := range e.clients {
			if request.Adapter != "" && client.Name != request.Adapter {
				continue
			}

			result, err := e.getActualCostFromPlugin(ctx, client, resource, request.From, request.To)
			if err != nil {
				partialErrors = append(partialErrors, fmt.Errorf("plugin %s: %w", client.Name, err))
				continue
			}
			if result != nil {
				resourceResult = result
				break
			}
		}

		// If no plugin provided data, create a placeholder result
		if resourceResult == nil {
			resourceResult = &CostResult{
				ResourceType: resource.Type,
				ResourceID:   resource.ID,
				Adapter:      "none",
				Currency:     "USD",
				TotalCost:    0,
				Notes:        "No actual cost data available",
				StartDate:    request.From,
				EndDate:      request.To,
				CostPeriod:   FormatPeriod(request.From, request.To),
			}
		}

		results = append(results, *resourceResult)
	}

	// Group results if requested
	if request.GroupBy != "" {
		results = e.GroupResults(results, GroupBy(request.GroupBy))
	}

	// Log partial errors but don't fail the entire operation
	_ = partialErrors // Errors collected but not currently logged

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
	if spec := e.tryLoadSpec(provider, service, defaultServiceName); spec != nil {
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
	totalHours := to.Sub(from).Hours()
	totalDays := int(totalHours / hoursPerDay)

	// Calculate daily costs if we have breakdown data
	var dailyCosts []float64
	if totalDays > 0 {
		dailyCosts = make([]float64, totalDays)
		avgDaily := result.TotalCost / float64(totalDays)
		for i := range totalDays {
			dailyCosts[i] = avgDaily
		}
	}

	return &CostResult{
		ResourceType: resource.Type,
		ResourceID:   resource.ID,
		Adapter:      client.Name,
		Currency:     result.Currency,
		Monthly:      result.TotalCost * avgDaysPerMonth / float64(totalDays), // More accurate monthly projection
		Hourly:       result.TotalCost / totalHours,
		TotalCost:    result.TotalCost,
		DailyCosts:   dailyCosts,
		Notes:        fmt.Sprintf("Actual cost from %s to %s", from.Format("2006-01-02"), to.Format("2006-01-02")),
		Breakdown:    result.CostBreakdown,
		StartDate:    from,
		EndDate:      to,
		CostPeriod:   FormatPeriod(from, to),
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
	return defaultServiceName
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

func MatchesTags(resource ResourceDescriptor, tags map[string]string) bool {
	if len(tags) == 0 {
		return true
	}

	for key, expectedValue := range tags {
		if prop, exists := resource.Properties[key]; exists {
			if propStr := fmt.Sprintf("%v", prop); propStr == expectedValue {
				continue
			}
		}
		return false
	}
	return true
}

func FormatPeriod(from, to time.Time) string {
	days := int(to.Sub(from).Hours() / hoursPerDay)
	if days == 1 {
		return "1 day"
	}
	if days < daysPerWeek {
		return fmt.Sprintf("%d days", days)
	}
	if days < daysPerMonth {
		weeks := days / daysPerWeek
		return fmt.Sprintf("%d weeks", weeks)
	}
	months := days / daysPerMonth
	return fmt.Sprintf("%d months", months)
}

func (e *Engine) GroupResults(results []CostResult, groupBy GroupBy) []CostResult {
	if groupBy == GroupByNone {
		return results
	}

	groups := make(map[string][]CostResult)

	for _, result := range results {
		var key string
		switch groupBy {
		case GroupByNone:
			// Should not reach here since we return early if GroupByNone
			key = defaultServiceName
		case GroupByResource:
			key = fmt.Sprintf("%s/%s", result.ResourceType, result.ResourceID)
		case GroupByType:
			key = result.ResourceType
		case GroupByProvider:
			// Extract provider from resource type (e.g., "aws:ec2/instance:Instance" -> "aws")
			if parts := strings.Split(result.ResourceType, ":"); len(parts) > 0 {
				key = parts[0]
			} else {
				key = "unknown"
			}
		case GroupByDate:
			key = result.StartDate.Format("2006-01-02")
		default:
			key = defaultServiceName
		}

		groups[key] = append(groups[key], result)
	}

	var grouped []CostResult
	for groupKey, groupResults := range groups {
		if len(groupResults) == 1 {
			grouped = append(grouped, groupResults[0])
		} else {
			// Aggregate multiple results into one
			aggregated := AggregateResultsInternal(groupResults, groupKey)
			grouped = append(grouped, aggregated)
		}
	}

	return grouped
}

func AggregateResultsInternal(results []CostResult, groupName string) CostResult {
	if len(results) == 0 {
		return CostResult{}
	}

	first := results[0]
	aggregated := CostResult{
		ResourceType: groupName,
		ResourceID:   fmt.Sprintf("aggregated-%d-resources", len(results)),
		Adapter:      "aggregated",
		Currency:     first.Currency,
		StartDate:    first.StartDate,
		EndDate:      first.EndDate,
		CostPeriod:   first.CostPeriod,
		Breakdown:    make(map[string]float64),
	}

	for _, result := range results {
		aggregated.Monthly += result.Monthly
		aggregated.Hourly += result.Hourly
		aggregated.TotalCost += result.TotalCost

		// Merge breakdowns
		for key, value := range result.Breakdown {
			aggregated.Breakdown[key] += value
		}

		// Extend daily costs
		if len(result.DailyCosts) > 0 {
			if len(aggregated.DailyCosts) < len(result.DailyCosts) {
				// Extend slice to match the longest daily costs
				extended := make([]float64, len(result.DailyCosts))
				copy(extended, aggregated.DailyCosts)
				aggregated.DailyCosts = extended
			}
			for i, dailyCost := range result.DailyCosts {
				if i < len(aggregated.DailyCosts) {
					aggregated.DailyCosts[i] += dailyCost
				}
			}
		}
	}

	aggregated.Notes = fmt.Sprintf("Aggregated costs from %d resources", len(results))
	return aggregated
}
