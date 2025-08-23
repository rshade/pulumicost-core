package engine

import (
	"context"
	"errors"
	"fmt"
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
		if len(request.Tags) > 0 && !matchesTags(resource, request.Tags) {
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
				CostPeriod:   formatPeriod(request.From, request.To),
			}
		}

		results = append(results, *resourceResult)
	}

	// Group results if requested
	if request.GroupBy != "" {
		results = e.groupResults(results, GroupBy(request.GroupBy))
	}

	// Log partial errors but don't fail the entire operation
	if len(partialErrors) > 0 {
		// Could log these errors or include them in response metadata
		// For now, we'll continue with partial results
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
		specData, specErr := e.loader.LoadSpec(resource.Provider, extractService(resource.Type), "default")
		if specErr != nil {
			return nil, specErr
		}

		if specData != nil {
			// Type assert the spec data
			if spec, ok := specData.(*PricingSpec); ok && spec != nil {
				return &CostResult{
					ResourceType: resource.Type,
					ResourceID:   resource.ID,
					Adapter:      "local-spec",
					Currency:     spec.Currency,
					Monthly:      defaultEstimate, // Default estimate
					Hourly:       defaultEstimate / hoursPerMonth,
					Notes:        "Estimated from local spec",
				}, nil
			}
		}
	}

	return nil, ErrNoCostData
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
		for i := 0; i < totalDays; i++ {
			dailyCosts[i] = avgDaily
		}
	}

	return &CostResult{
		ResourceType: resource.Type,
		ResourceID:   resource.ID,
		Adapter:      client.Name,
		Currency:     result.Currency,
		Monthly:      result.TotalCost * 30.44 / float64(totalDays), // More accurate monthly projection
		Hourly:       result.TotalCost / totalHours,
		TotalCost:    result.TotalCost,
		DailyCosts:   dailyCosts,
		Notes:        fmt.Sprintf("Actual cost from %s to %s", from.Format("2006-01-02"), to.Format("2006-01-02")),
		Breakdown:    result.CostBreakdown,
		StartDate:    from,
		EndDate:      to,
		CostPeriod:   formatPeriod(from, to),
	}, nil
}

func convertToProto(properties map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range properties {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
}

func extractService(_ string) string {
	return "default"
}

func matchesTags(resource ResourceDescriptor, tags map[string]string) bool {
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

func formatPeriod(from, to time.Time) string {
	days := int(to.Sub(from).Hours() / hoursPerDay)
	if days == 1 {
		return "1 day"
	}
	if days < 7 {
		return fmt.Sprintf("%d days", days)
	}
	if days < 30 {
		weeks := days / 7
		return fmt.Sprintf("%d weeks", weeks)
	}
	months := days / 30
	return fmt.Sprintf("%d months", months)
}

func (e *Engine) groupResults(results []CostResult, groupBy GroupBy) []CostResult {
	if groupBy == GroupByNone {
		return results
	}

	groups := make(map[string][]CostResult)
	
	for _, result := range results {
		var key string
		switch groupBy {
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
			key = "default"
		}
		
		groups[key] = append(groups[key], result)
	}

	var grouped []CostResult
	for groupKey, groupResults := range groups {
		if len(groupResults) == 1 {
			grouped = append(grouped, groupResults[0])
		} else {
			// Aggregate multiple results into one
			aggregated := aggregateResults(groupResults, groupKey)
			grouped = append(grouped, aggregated)
		}
	}

	return grouped
}

func aggregateResults(results []CostResult, groupName string) CostResult {
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
