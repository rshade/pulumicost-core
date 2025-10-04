// Package engine provides the core cost calculation and aggregation functionality for PulumiCost.
//
// This package orchestrates cost calculations between multiple cloud provider plugins,
// handles fallback to local pricing specifications, and provides comprehensive
// cross-provider cost aggregation and analysis capabilities.
//
// Key Features:
//   - Multi-provider cost calculation with plugin architecture
//   - Cross-provider cost aggregation with currency validation
//   - Time-based grouping (daily/monthly) with intelligent cost conversion
//   - Flexible resource filtering and tag-based matching
//   - Multiple output formats (table, JSON, NDJSON) with rich cost breakdowns
//   - Graceful fallback from plugins to local YAML specifications
//
// Architecture Overview:
//
//	Pulumi Resources → Engine → Plugins (gRPC) → Cost Results
//	                      ↓ (fallback)
//	                 Spec Loader (YAML) → Cost Results
//	                      ↓
//	              Cross-Provider Aggregation → Formatted Output
//
// Thread Safety:
//   - Engine instances are safe for concurrent use
//   - Individual cost calculations are independent
//   - Aggregation functions are pure and stateless
//
// Performance:
//   - Designed for datasets up to 10,000 cost results
//   - Memory usage scales linearly with resource count
//   - Plugin communication uses connection pooling
package engine

import (
	"time"

	"github.com/rshade/pulumicost-core/internal/spec"
)

type ResourceDescriptor struct {
	Type       string
	ID         string
	Provider   string
	Properties map[string]interface{}
}

type CostResult struct {
	ResourceType string             `json:"resourceType"`
	ResourceID   string             `json:"resourceId"`
	Adapter      string             `json:"adapter"`
	Currency     string             `json:"currency"`
	Monthly      float64            `json:"monthly"`
	Hourly       float64            `json:"hourly"`
	Notes        string             `json:"notes"`
	Breakdown    map[string]float64 `json:"breakdown"`
	// Actual cost specific fields
	TotalCost  float64   `json:"totalCost,omitempty"`
	DailyCosts []float64 `json:"dailyCosts,omitempty"`
	CostPeriod string    `json:"costPeriod,omitempty"`
	StartDate  time.Time `json:"startDate,omitempty"`
	EndDate    time.Time `json:"endDate,omitempty"`
}

type ActualCostRequest struct {
	Resources []ResourceDescriptor
	From      time.Time
	To        time.Time
	Adapter   string
	GroupBy   string
	Tags      map[string]string
}

// CrossProviderAggregation represents daily/monthly cost aggregation across providers.
//
// This type enables comprehensive cost analysis across multiple cloud providers
// (AWS, Azure, GCP, etc.) with time-based aggregation for trend analysis and
// cost optimization insights.
//
// Usage Examples:
//
//	// Daily aggregation showing multi-provider costs
//	agg := CrossProviderAggregation{
//		Period: "2024-01-15",
//		Providers: map[string]float64{
//			"aws":   245.67,
//			"azure": 156.23,
//			"gcp":   89.45,
//		},
//		Total:    491.35,
//		Currency: "USD",
//	}
//
//	// Monthly aggregation for cost trending
//	agg := CrossProviderAggregation{
//		Period: "2024-01",
//		Providers: map[string]float64{
//			"aws":   7620.15,
//			"azure": 4843.67,
//		},
//		Total:    12463.82,
//		Currency: "USD",
//	}
type CrossProviderAggregation struct {
	Period    string             `json:"period"`    // Date (2024-01-01) or Month (2024-01)
	Providers map[string]float64 `json:"providers"` // Provider name -> cost
	Total     float64            `json:"total"`     // Total cost for this period
	Currency  string             `json:"currency"`  // Currency for all costs
}

// GroupBy defines the available grouping strategies for cost result aggregation.
//
// This type enables flexible cost analysis by grouping results along different
// dimensions such as resources, providers, or time periods. Each grouping type
// provides unique insights for cost optimization and analysis.
type GroupBy string

// GroupBy constants define all supported aggregation strategies.
//
// Resource Groupings:
//   - GroupByResource: Groups by individual resource (ResourceType/ResourceID)
//   - GroupByType: Groups by resource type (e.g., "aws:ec2:Instance")
//   - GroupByProvider: Groups by cloud provider (e.g., "aws", "azure", "gcp")
//
// Time-Based Groupings:
//   - GroupByDaily: Groups by calendar date ("2006-01-02") for daily trends
//   - GroupByMonthly: Groups by month ("2006-01") for monthly analysis
//   - GroupByDate: Deprecated legacy date-key grouping (non time-based for cross-provider).
//     Prefer GroupByDaily for time-based aggregations.
//
// Special Values:
//   - GroupByNone: No grouping (empty string) - returns results as-is
//
// Usage Guidelines:
//   - Use resource groupings for cost attribution and resource optimization
//   - Use time-based groupings for trend analysis and forecasting
//   - Time-based groupings require results with valid StartDate/EndDate fields
//   - Cross-provider aggregation only supports time-based groupings
const (
	GroupByResource GroupBy = "resource"
	GroupByType     GroupBy = "type"
	GroupByProvider GroupBy = "provider"
	GroupByDate     GroupBy = "date" // Deprecated: use GroupByDaily
	GroupByDaily    GroupBy = "daily"
	GroupByMonthly  GroupBy = "monthly"
	GroupByNone     GroupBy = ""
)

// IsValid returns true if the GroupBy value is valid.
//
// This method provides compile-time safety for GroupBy values and enables
// validation before processing. All defined GroupBy constants are considered
// valid, including the empty string (GroupByNone).
//
// Returns:
//   - true: For all defined GroupBy constants
//   - false: For any other string values
//
// Usage Examples:
//
//	// Validate user input
//	groupBy := GroupBy(userInput)
//	if !groupBy.IsValid() {
//		return fmt.Errorf("invalid groupBy: %s", userInput)
//	}
//
//	// Safe to use in switch statements
//	switch groupBy {
//	case GroupByDaily, GroupByMonthly:
//		// Time-based processing
//	case GroupByResource, GroupByType, GroupByProvider:
//		// Resource-based processing
//	}
func (g GroupBy) IsValid() bool {
	switch g {
	case GroupByResource, GroupByType, GroupByProvider, GroupByDate, GroupByDaily, GroupByMonthly, GroupByNone:
		return true
	default:
		return false
	}
}

// IsTimeBasedGrouping returns true if the GroupBy requires time-based aggregation.
//
// Time-based groupings require cost results with valid StartDate and EndDate fields
// and are the only grouping types supported by cross-provider aggregation functions.
// This method is used internally to validate aggregation requests and determine
// processing strategies.
//
// Time-Based GroupBy Values:
//   - GroupByDaily: Requires daily cost data aggregation
//   - GroupByMonthly: Requires monthly cost data aggregation
//   - GroupByDate: Deprecated legacy date-key grouping (non time-based for cross-provider)
//
// Non-Time-Based GroupBy Values:
//   - GroupByResource: Groups by resource identity
//   - GroupByType: Groups by resource type
//   - GroupByProvider: Groups by cloud provider
//   - GroupByNone: No grouping applied
//
// Returns:
//   - true: For GroupByDaily and GroupByMonthly only
//   - false: For all other GroupBy values
//
// Usage Examples:
//
//	// Validate for cross-provider aggregation
//	if !groupBy.IsTimeBasedGrouping() {
//		return ErrInvalidGroupBy
//	}
//
//	// Route to appropriate processing
//	if groupBy.IsTimeBasedGrouping() {
//		return CreateCrossProviderAggregation(results, groupBy)
//	} else {
//		return engine.GroupResults(results, groupBy)
//	}
func (g GroupBy) IsTimeBasedGrouping() bool {
	return g == GroupByDaily || g == GroupByMonthly
}

// String returns the string representation of the GroupBy.
//
// This method implements the Stringer interface and provides a consistent
// string representation for logging, debugging, and serialization.
//
// Returns:
//   - string: The underlying string value of the GroupBy
//   - "": Empty string for GroupByNone
//
// Usage Examples:
//
//	// Logging and debugging
//	log.Printf("Processing with groupBy: %s", groupBy.String())
//
//	// CLI flag validation
//	if groupBy.String() == "" {
//		groupBy = GroupByResource // Default
//	}
//
//	// JSON serialization (automatic via json package)
//	type Request struct {
//		GroupBy GroupBy `json:"groupBy"`
//	}
func (g GroupBy) String() string {
	return string(g)
}

type ProjectedCostRequest struct {
	Resources []ResourceDescriptor
	SpecDir   string
	Adapter   string
}

// PricingSpec is an alias to the PricingSpec from the spec package to ensure type consistency.
type PricingSpec = spec.PricingSpec

type CostSummary struct {
	TotalMonthly float64            `json:"totalMonthly"`
	TotalHourly  float64            `json:"totalHourly"`
	Currency     string             `json:"currency"`
	ByProvider   map[string]float64 `json:"byProvider"`
	ByService    map[string]float64 `json:"byService"`
	ByAdapter    map[string]float64 `json:"byAdapter"`
	Resources    []CostResult       `json:"resources"`
}

type AggregatedResults struct {
	Summary   CostSummary  `json:"summary"`
	Resources []CostResult `json:"resources"`
}
