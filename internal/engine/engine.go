package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/rshade/pulumicost-core/internal/pluginhost"
	"github.com/rshade/pulumicost-core/internal/proto"
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

func (e *Engine) GetActualCost(ctx context.Context, resources []ResourceDescriptor, from, to time.Time) ([]CostResult, error) {
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

func (e *Engine) getProjectedCostFromPlugin(ctx context.Context, client *pluginhost.Client, resource ResourceDescriptor) (*CostResult, error) {
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
		specData, err := e.loader.LoadSpec(resource.Provider, extractService(resource.Type), "default")
		if err != nil {
			return nil, err
		}

		if specData != nil {
			// Type assert the spec data
			if spec, ok := specData.(*PricingSpec); ok && spec != nil {
				return &CostResult{
					ResourceType: resource.Type,
					ResourceID:   resource.ID,
					Adapter:      "local-spec",
					Currency:     spec.Currency,
					Monthly:      100.0, // Default estimate
					Hourly:       100.0 / 730,
					Notes:        "Estimated from local spec",
				}, nil
			}
		}
	}

	return nil, nil
}

func (e *Engine) getActualCostFromPlugin(ctx context.Context, client *pluginhost.Client, resource ResourceDescriptor, from, to time.Time) (*CostResult, error) {
	req := &proto.GetActualCostRequest{
		ResourceIds: []string{resource.ID},
		StartTime:   from.Unix(),
		EndTime:     to.Unix(),
	}

	resp, err := client.API.GetActualCost(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(resp.Results) == 0 {
		return nil, nil
	}

	result := resp.Results[0]
	return &CostResult{
		ResourceType: resource.Type,
		ResourceID:   resource.ID,
		Adapter:      client.Name,
		Currency:     result.Currency,
		Monthly:      result.TotalCost * 30.44 / float64(to.Sub(from).Hours()/24), // Approximate monthly
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
	return "default"
}
