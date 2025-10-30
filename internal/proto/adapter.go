// Package proto provides adapter types and client wrappers for interfacing with the pulumicost-spec gRPC protocol.
// This package bridges the engine's internal representation with the protobuf-generated types,
// enabling seamless integration with the CostSourceService defined in pulumicost-spec.
package proto

import (
	"context"
	"time"

	pbc "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Empty represents an empty request/response for compatibility with existing engine code.
type Empty struct{}

// ResourceDescriptor describes a cloud resource for cost calculation requests.
// It contains the resource type, provider, and properties needed for pricing lookups.
type ResourceDescriptor struct {
	Type       string
	Provider   string
	Properties map[string]string
}

// GetProjectedCostRequest contains resources for which projected costs should be calculated.
type GetProjectedCostRequest struct {
	Resources []*ResourceDescriptor
}

// CostResult represents the calculated cost information for a single resource.
// It includes monthly and hourly costs, currency, and detailed cost breakdowns.
type CostResult struct {
	Currency      string
	MonthlyCost   float64
	HourlyCost    float64
	Notes         string
	CostBreakdown map[string]float64
}

// GetProjectedCostResponse contains the results of projected cost calculations.
type GetProjectedCostResponse struct {
	Results []*CostResult
}

// GetActualCostRequest contains parameters for querying historical actual costs.
// It includes resource IDs and a time range for cost data retrieval.
type GetActualCostRequest struct {
	ResourceIDs []string
	StartTime   int64
	EndTime     int64
}

// ActualCostResult represents the calculated actual cost data retrieved from cloud providers.
// It includes the total cost and detailed breakdowns by service or resource.
type ActualCostResult struct {
	Currency      string
	TotalCost     float64
	CostBreakdown map[string]float64
}

// GetActualCostResponse contains the results of actual cost queries.
type GetActualCostResponse struct {
	Results []*ActualCostResult
}

// NameResponse contains the plugin name returned by the Name RPC call.
type NameResponse struct {
	Name string
}

// GetName returns the plugin name from the response.
func (n *NameResponse) GetName() string {
	return n.Name
}

// CostSourceClient wraps the generated gRPC client from pulumicost-spec.
type CostSourceClient interface {
	Name(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*NameResponse, error)
	GetProjectedCost(
		ctx context.Context,
		in *GetProjectedCostRequest,
		opts ...grpc.CallOption,
	) (*GetProjectedCostResponse, error)
	GetActualCost(
		ctx context.Context,
		in *GetActualCostRequest,
		opts ...grpc.CallOption,
	) (*GetActualCostResponse, error)
}

// NewCostSourceClient creates a new cost source client using the real proto client.
func NewCostSourceClient(conn *grpc.ClientConn) CostSourceClient {
	return &clientAdapter{
		client: pbc.NewCostSourceServiceClient(conn),
	}
}

// clientAdapter adapts the generated client to our internal interface.
type clientAdapter struct {
	client pbc.CostSourceServiceClient
}

func (c *clientAdapter) Name(ctx context.Context, _ *Empty, opts ...grpc.CallOption) (*NameResponse, error) {
	resp, err := c.client.Name(ctx, &pbc.NameRequest{}, opts...)
	if err != nil {
		return nil, err
	}
	return &NameResponse{Name: resp.GetName()}, nil
}

func (c *clientAdapter) GetProjectedCost(
	ctx context.Context,
	in *GetProjectedCostRequest,
	opts ...grpc.CallOption,
) (*GetProjectedCostResponse, error) {
	// Convert internal request to proto request
	var results []*CostResult

	for _, resource := range in.Resources {
		req := &pbc.GetProjectedCostRequest{
			Resource: &pbc.ResourceDescriptor{
				Provider:     resource.Provider,
				ResourceType: resource.Type,
				Sku:          "", // Will be filled from properties if available
				Region:       "", // Will be filled from properties if available
				Tags:         resource.Properties,
			},
		}

		// Extract SKU and region from properties if available
		if sku, ok := resource.Properties["sku"]; ok {
			req.Resource.Sku = sku
		}
		if region, ok := resource.Properties["region"]; ok {
			req.Resource.Region = region
		}

		resp, err := c.client.GetProjectedCost(ctx, req, opts...)
		if err != nil {
			// Continue to next resource on error
			continue
		}

		result := &CostResult{
			Currency:    resp.GetCurrency(),
			MonthlyCost: resp.GetCostPerMonth(),
			HourlyCost:  resp.GetUnitPrice(), // Assuming hourly for now
			Notes:       resp.GetBillingDetail(),
			CostBreakdown: map[string]float64{
				"unit_price": resp.GetUnitPrice(),
			},
		}
		results = append(results, result)
	}

	return &GetProjectedCostResponse{Results: results}, nil
}

func (c *clientAdapter) GetActualCost(
	ctx context.Context,
	in *GetActualCostRequest,
	opts ...grpc.CallOption,
) (*GetActualCostResponse, error) {
	// Convert internal request to proto request
	var results []*ActualCostResult

	for _, resourceID := range in.ResourceIDs {
		req := &pbc.GetActualCostRequest{
			ResourceId: resourceID,
			Start:      timestamppb.New(time.Unix(in.StartTime, 0)),
			End:        timestamppb.New(time.Unix(in.EndTime, 0)),
			Tags:       make(map[string]string), // Empty tags for now
		}

		resp, err := c.client.GetActualCost(ctx, req, opts...)
		if err != nil {
			// Continue to next resource on error
			continue
		}

		// Aggregate total cost from results
		totalCost := 0.0
		breakdown := make(map[string]float64)

		for _, result := range resp.GetResults() {
			totalCost += result.GetCost()
			if result.GetSource() != "" {
				breakdown[result.GetSource()] = result.GetCost()
			}
		}

		result := &ActualCostResult{
			Currency:      "USD", // Default to USD if not specified
			TotalCost:     totalCost,
			CostBreakdown: breakdown,
		}
		results = append(results, result)
	}

	return &GetActualCostResponse{Results: results}, nil
}
