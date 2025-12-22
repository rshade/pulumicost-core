package proto

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rshade/pulumicost-spec/sdk/go/pluginsdk/mapping"
	pbc "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// maxErrorsToDisplay is the maximum number of errors to show in summary before truncating.
	maxErrorsToDisplay = 5
)

// ErrorDetail captures information about a failed resource cost calculation.
type ErrorDetail struct {
	ResourceType string
	ResourceID   string
	PluginName   string
	Error        error
	Timestamp    time.Time
}

// CostResultWithErrors wraps results and any errors encountered during cost calculation.
type CostResultWithErrors struct {
	Results []*CostResult
	Errors  []ErrorDetail
}

// HasErrors returns true if any errors were encountered during cost calculation.
func (c *CostResultWithErrors) HasErrors() bool {
	return len(c.Errors) > 0
}

// ErrorSummary returns a human-readable summary of errors.
// Truncates the output after 5 errors to keep it readable.
func (c *CostResultWithErrors) ErrorSummary() string {
	if !c.HasErrors() {
		return ""
	}

	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("%d resource(s) failed:\n", len(c.Errors)))

	for i, err := range c.Errors {
		if i >= maxErrorsToDisplay {
			summary.WriteString(
				fmt.Sprintf("  ... and %d more errors\n", len(c.Errors)-maxErrorsToDisplay),
			)
			break
		}
		summary.WriteString(
			fmt.Sprintf("  - %s (%s): %v\n", err.ResourceType, err.ResourceID, err.Error),
		)
	}

	return summary.String()
}

// GetProjectedCostWithErrors calculates projected costs for resources with error tracking.
// GetProjectedCostWithErrors queries the provided CostSourceClient for projected costs for each resource
// and aggregates both successful CostResult entries and per-resource error details.
//
// GetProjectedCostWithErrors calls the client's GetProjectedCost once per resource in `resources`.
// For each resource, a successful response appends its returned results to the aggregated Results slice.
// If a per-resource RPC fails, an ErrorDetail is recorded in Errors (including timestamp and pluginName)
// and a placeholder CostResult with an error note is appended to Results. If a call succeeds but returns
// no results, a zero-cost placeholder CostResult is appended.
//
// Parameters:
//   - ctx: request context passed to the client calls.
//   - client: the CostSourceClient used to fetch projected cost data.
//   - pluginName: the name of the plugin making the requests; recorded on ErrorDetail entries.
//   - resources: slice of ResourceDescriptor values to query.
//
// Returns:
//
//	A pointer to a CostResultWithErrors containing a Results slice with one or more CostResult entries
//	(including placeholders for failures or empty responses) and an Errors slice with one ErrorDetail per
//	resource that experienced an RPC error.
func GetProjectedCostWithErrors(
	ctx context.Context,
	client CostSourceClient,
	pluginName string,
	resources []*ResourceDescriptor,
) *CostResultWithErrors {
	result := &CostResultWithErrors{
		Results: []*CostResult{},
		Errors:  []ErrorDetail{},
	}

	for _, resource := range resources {
		req := &GetProjectedCostRequest{
			Resources: []*ResourceDescriptor{resource},
		}

		resp, err := client.GetProjectedCost(ctx, req)
		if err != nil {
			// Track error instead of silent failure
			result.Errors = append(result.Errors, ErrorDetail{
				ResourceType: resource.Type,
				ResourceID:   resource.Type, // Use type as ID for now
				PluginName:   pluginName,
				Error:        fmt.Errorf("plugin call failed: %w", err),
				Timestamp:    time.Now(),
			})

			// Add placeholder result with error note
			result.Results = append(result.Results, &CostResult{
				Currency:    "USD",
				MonthlyCost: 0,
				HourlyCost:  0,
				Notes:       fmt.Sprintf("ERROR: %v", err),
			})
			continue
		}

		// Add successful results
		if len(resp.Results) > 0 {
			result.Results = append(result.Results, resp.Results...)
		} else {
			// Add empty result if no results returned
			result.Results = append(result.Results, &CostResult{
				Currency:    "USD",
				MonthlyCost: 0,
				HourlyCost:  0,
			})
		}
	}

	return result
}

// GetActualCostWithErrors calculates actual costs for resource IDs with error tracking.
// It queries the provided CostSourceClient for actual costs for each resource ID in `req.ResourceIDs`
// and aggregates both successful ActualCostResult entries and per-resource error details.
//
// Parameters:
//   - ctx: request context passed to the client calls.
//   - client: the CostSourceClient used to fetch actual cost data.
//   - pluginName: the name of the plugin making the requests; recorded on ErrorDetail entries.
//   - req: the GetActualCostRequest containing resource IDs and time range.
//
// Returns:
//
//	A pointer to a CostResultWithErrors containing a Results slice (converted from ActualCostResult)
//	and an Errors slice with one ErrorDetail per resource ID that experienced an RPC error.
func GetActualCostWithErrors(
	ctx context.Context,
	client CostSourceClient,
	pluginName string,
	req *GetActualCostRequest,
) *CostResultWithErrors {
	result := &CostResultWithErrors{
		Results: []*CostResult{},
		Errors:  []ErrorDetail{},
	}

	for _, resourceID := range req.ResourceIDs {
		singleReq := &GetActualCostRequest{
			ResourceIDs: []string{resourceID},
			StartTime:   req.StartTime,
			EndTime:     req.EndTime,
		}

		resp, err := client.GetActualCost(ctx, singleReq)
		if err != nil {
			result.Errors = append(result.Errors, ErrorDetail{
				ResourceID: resourceID,
				PluginName: pluginName,
				Error:      fmt.Errorf("plugin call failed: %w", err),
				Timestamp:  time.Now(),
			})

			result.Results = append(result.Results, &CostResult{
				Currency:    "USD",
				MonthlyCost: 0,
				HourlyCost:  0,
				Notes:       fmt.Sprintf("ERROR: %v", err),
			})
			continue
		}

		// Aggregate total cost from results and convert to CostResult
		if len(resp.Results) > 0 {
			for _, actual := range resp.Results {
				costResult := &CostResult{
					Currency:       actual.Currency,
					MonthlyCost:    actual.TotalCost, // Total cost for the period
					HourlyCost:     0,
					CostBreakdown:  actual.CostBreakdown,
					Sustainability: make(map[string]SustainabilityMetric),
				}

				// Map impact metrics
				for k, v := range actual.Sustainability {
					costResult.Sustainability[k] = SustainabilityMetric{
						Value: v.Value,
						Unit:  v.Unit,
					}
				}
				result.Results = append(result.Results, costResult)
			}
		} else {
			result.Results = append(result.Results, &CostResult{
				Currency:    "USD",
				MonthlyCost: 0,
				HourlyCost:  0,
			})
		}
	}

	return result
}

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
	Currency       string
	MonthlyCost    float64
	HourlyCost     float64
	Notes          string
	CostBreakdown  map[string]float64
	Sustainability map[string]SustainabilityMetric
}

// SustainabilityMetric represents a single sustainability impact measurement.
type SustainabilityMetric struct {
	Value float64
	Unit  string
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
	Currency       string
	TotalCost      float64
	CostBreakdown  map[string]float64
	Sustainability map[string]SustainabilityMetric
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

func (c *clientAdapter) Name(
	ctx context.Context,
	_ *Empty,
	opts ...grpc.CallOption,
) (*NameResponse, error) {
	resp, err := c.client.Name(ctx, &pbc.NameRequest{}, opts...)
	if err != nil {
		return nil, err
	}
	return &NameResponse{Name: resp.GetName()}, nil
}

// resolveSKUAndRegion extracts the SKU and region from resource properties based on the cloud provider.
func resolveSKUAndRegion(provider string, properties map[string]string) (string, string) {
	var sku, region string
	switch strings.ToLower(provider) {
	case "aws":
		sku = mapping.ExtractAWSSKU(properties)
		if sku == "" {
			// Fallback for RDS and other AWS resources not covered by ExtractAWSSKU
			sku = mapping.ExtractSKU(properties, "dbInstanceClass", "sku", "type", "tier")
		}
		region = mapping.ExtractAWSRegion(properties)
	case "azure", "azure-native":
		sku = mapping.ExtractAzureSKU(properties)
		region = mapping.ExtractAzureRegion(properties)
	case "gcp", "google-native":
		sku = mapping.ExtractGCPSKU(properties)
		region = mapping.ExtractGCPRegion(properties)
	default:
		sku = mapping.ExtractSKU(properties)
		region = mapping.ExtractRegion(properties)
	}

	// Fallback to environment variables for region if still empty
	if region == "" {
		if envReg := os.Getenv("AWS_REGION"); envReg != "" {
			region = envReg
		} else {
			envReg = os.Getenv("AWS_DEFAULT_REGION")
			if envReg != "" {
				region = envReg
			}
		}
	}

	return sku, region
}

func (c *clientAdapter) GetProjectedCost(
	ctx context.Context,
	in *GetProjectedCostRequest,
	opts ...grpc.CallOption,
) (*GetProjectedCostResponse, error) {
	// Convert internal request to proto request
	var results []*CostResult

	for _, resource := range in.Resources {
		// Extract SKU and region from properties using intelligent mapping
		sku, region := resolveSKUAndRegion(resource.Provider, resource.Properties)

		req := &pbc.GetProjectedCostRequest{
			Resource: &pbc.ResourceDescriptor{
				Provider:     resource.Provider,
				ResourceType: resource.Type,
				Sku:          sku,
				Region:       region,
				Tags:         resource.Properties,
			},
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
			Sustainability: make(map[string]SustainabilityMetric),
		}

		// Map impact metrics
		for _, metric := range resp.GetImpactMetrics() {
			var key string
			switch metric.GetKind() {
			case pbc.MetricKind_METRIC_KIND_CARBON_FOOTPRINT:
				key = "carbon_footprint"
			case pbc.MetricKind_METRIC_KIND_ENERGY_CONSUMPTION:
				key = "energy_consumption"
			case pbc.MetricKind_METRIC_KIND_WATER_USAGE:
				key = "water_usage"
			case pbc.MetricKind_METRIC_KIND_UNSPECIFIED:
				key = "unspecified"
			default:
				key = strings.ToLower(metric.GetKind().String())
			}
			result.Sustainability[key] = SustainabilityMetric{
				Value: metric.GetValue(),
				Unit:  metric.GetUnit(),
			}
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
			Currency:       "USD", // Default to USD if not specified
			TotalCost:      totalCost,
			CostBreakdown:  breakdown,
			Sustainability: make(map[string]SustainabilityMetric),
		}

		// Aggregate impact metrics (summing values for same kind across results)
		for _, pbcResult := range resp.GetResults() {
			for _, metric := range pbcResult.GetImpactMetrics() {
				var key string
				switch metric.GetKind() {
				case pbc.MetricKind_METRIC_KIND_CARBON_FOOTPRINT:
					key = "carbon_footprint"
				case pbc.MetricKind_METRIC_KIND_ENERGY_CONSUMPTION:
					key = "energy_consumption"
				case pbc.MetricKind_METRIC_KIND_WATER_USAGE:
					key = "water_usage"
				case pbc.MetricKind_METRIC_KIND_UNSPECIFIED:
					key = "unspecified"
				default:
					key = strings.ToLower(metric.GetKind().String())
				}

				m := result.Sustainability[key]
				m.Value += metric.GetValue()
				m.Unit = metric.GetUnit()
				result.Sustainability[key] = m
			}
		}
		results = append(results, result)
	}

	return &GetActualCostResponse{Results: results}, nil
}
