package engine_test

import (
	"testing"
	"time"

	"github.com/rshade/pulumicost-core/internal/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateCrossProviderAggregation_DailyGrouping tests daily aggregation.
func TestCreateCrossProviderAggregation_DailyGrouping(t *testing.T) {
	jan1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	jan2 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	results := []engine.CostResult{
		{
			ResourceType: "aws:ec2:Instance",
			TotalCost:    100.0,
			Currency:     "USD",
			StartDate:    jan1,
			EndDate:      jan2,
		},
		{
			ResourceType: "azure:compute:VirtualMachine",
			TotalCost:    150.0,
			Currency:     "USD",
			StartDate:    jan1,
			EndDate:      jan2,
		},
	}

	aggregations, err := engine.CreateCrossProviderAggregation(results, engine.GroupByDaily)

	require.NoError(t, err)
	require.Len(t, aggregations, 1)

	agg := aggregations[0]
	assert.Equal(t, "2024-01-01", agg.Period)
	assert.Equal(t, 250.0, agg.Total) // 100 + 150
	assert.Equal(t, "USD", agg.Currency)
	assert.Contains(t, agg.Providers, "aws")
	assert.Contains(t, agg.Providers, "azure")
	assert.Equal(t, 100.0, agg.Providers["aws"])
	assert.Equal(t, 150.0, agg.Providers["azure"])
}

// TestCreateCrossProviderAggregation_MonthlyGrouping tests monthly aggregation.
func TestCreateCrossProviderAggregation_MonthlyGrouping(t *testing.T) {
	jan1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	jan31 := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
	feb1 := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	feb29 := time.Date(2024, 2, 29, 23, 59, 59, 0, time.UTC)

	results := []engine.CostResult{
		{
			ResourceType: "aws:ec2:Instance",
			TotalCost:    3100.0,
			Currency:     "USD",
			StartDate:    jan1,
			EndDate:      jan31,
		},
		{
			ResourceType: "azure:compute:VirtualMachine",
			TotalCost:    4650.0,
			Currency:     "USD",
			StartDate:    feb1,
			EndDate:      feb29,
		},
	}

	aggregations, err := engine.CreateCrossProviderAggregation(results, engine.GroupByMonthly)

	require.NoError(t, err)
	require.Len(t, aggregations, 2)

	// Verify January
	jan := aggregations[0]
	assert.Equal(t, "2024-01", jan.Period)
	assert.Equal(t, 3100.0, jan.Total)
	assert.Contains(t, jan.Providers, "aws")

	// Verify February
	feb := aggregations[1]
	assert.Equal(t, "2024-02", feb.Period)
	assert.Equal(t, 4650.0, feb.Total)
	assert.Contains(t, feb.Providers, "azure")
}

// TestCreateCrossProviderAggregation_MultipleProviders tests aggregation with multiple providers.
func TestCreateCrossProviderAggregation_MultipleProviders(t *testing.T) {
	jan1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	jan2 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	results := []engine.CostResult{
		{
			ResourceType: "aws:ec2:Instance",
			TotalCost:    100.0,
			Currency:     "USD",
			StartDate:    jan1,
			EndDate:      jan2,
		},
		{
			ResourceType: "azure:compute:VirtualMachine",
			TotalCost:    150.0,
			Currency:     "USD",
			StartDate:    jan1,
			EndDate:      jan2,
		},
		{
			ResourceType: "gcp:compute:Instance",
			TotalCost:    75.0,
			Currency:     "USD",
			StartDate:    jan1,
			EndDate:      jan2,
		},
	}

	aggregations, err := engine.CreateCrossProviderAggregation(results, engine.GroupByDaily)

	require.NoError(t, err)
	require.Len(t, aggregations, 1)

	agg := aggregations[0]
	assert.Equal(t, 325.0, agg.Total) // 100 + 150 + 75
	assert.Len(t, agg.Providers, 3)
	assert.Equal(t, 100.0, agg.Providers["aws"])
	assert.Equal(t, 150.0, agg.Providers["azure"])
	assert.Equal(t, 75.0, agg.Providers["gcp"])
}

// TestCreateCrossProviderAggregation_MultipleDays tests aggregation across multiple days.
func TestCreateCrossProviderAggregation_MultipleDays(t *testing.T) {
	jan1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	jan2 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	jan3 := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
	jan4 := time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC)

	results := []engine.CostResult{
		{
			ResourceType: "aws:ec2:Instance",
			TotalCost:    100.0,
			Currency:     "USD",
			StartDate:    jan1,
			EndDate:      jan2,
		},
		{
			ResourceType: "aws:ec2:Instance",
			TotalCost:    110.0,
			Currency:     "USD",
			StartDate:    jan3,
			EndDate:      jan4,
		},
		{
			ResourceType: "azure:compute:VirtualMachine",
			TotalCost:    200.0,
			Currency:     "USD",
			StartDate:    jan1,
			EndDate:      jan2,
		},
	}

	aggregations, err := engine.CreateCrossProviderAggregation(results, engine.GroupByDaily)

	require.NoError(t, err)
	require.Len(t, aggregations, 2)

	// Verify sorting by period (chronological order)
	assert.Equal(t, "2024-01-01", aggregations[0].Period)
	assert.Equal(t, "2024-01-03", aggregations[1].Period)

	// Verify Day 1 totals
	day1 := aggregations[0]
	assert.Equal(t, 300.0, day1.Total) // 100 + 200
	assert.Equal(t, 100.0, day1.Providers["aws"])
	assert.Equal(t, 200.0, day1.Providers["azure"])

	// Verify Day 3 totals
	day3 := aggregations[1]
	assert.Equal(t, 110.0, day3.Total)
	assert.Equal(t, 110.0, day3.Providers["aws"])
}

// TestCreateCrossProviderAggregation_WithDailyCosts tests aggregation with daily cost breakdown.
func TestCreateCrossProviderAggregation_WithDailyCosts(t *testing.T) {
	jan1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	jan4 := time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC)

	results := []engine.CostResult{
		{
			ResourceType: "aws:ec2:Instance",
			Currency:     "USD",
			StartDate:    jan1,
			EndDate:      jan4,
			DailyCosts:   []float64{100.0, 110.0, 120.0}, // 3 days of costs
		},
	}

	aggregations, err := engine.CreateCrossProviderAggregation(results, engine.GroupByDaily)

	require.NoError(t, err)
	require.Len(t, aggregations, 3)

	// Verify daily breakdown
	assert.Equal(t, "2024-01-01", aggregations[0].Period)
	assert.Equal(t, 100.0, aggregations[0].Total)

	assert.Equal(t, "2024-01-02", aggregations[1].Period)
	assert.Equal(t, 110.0, aggregations[1].Total)

	assert.Equal(t, "2024-01-03", aggregations[2].Period)
	assert.Equal(t, 120.0, aggregations[2].Total)
}

// TestCreateCrossProviderAggregation_FallbackToMonthly tests fallback to monthly costs.
func TestCreateCrossProviderAggregation_FallbackToMonthly(t *testing.T) {
	jan1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	jan2 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	results := []engine.CostResult{
		{
			ResourceType: "aws:ec2:Instance",
			TotalCost:    0,      // No actual cost
			Monthly:      3100.0, // Fall back to monthly estimate
			Currency:     "USD",
			StartDate:    jan1,
			EndDate:      jan2,
		},
	}

	aggregations, err := engine.CreateCrossProviderAggregation(results, engine.GroupByDaily)

	require.NoError(t, err)
	require.Len(t, aggregations, 1)

	// Should convert monthly to daily (3100 / 30.44 ≈ 101.84)
	agg := aggregations[0]
	assert.InDelta(t, 101.84, agg.Total, 0.1)
}

// TestCreateCrossProviderAggregation_EmptyCurrencyDefaultsToUSD tests empty currency handling.
func TestCreateCrossProviderAggregation_EmptyCurrencyDefaultsToUSD(t *testing.T) {
	jan1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	jan2 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	results := []engine.CostResult{
		{
			ResourceType: "aws:ec2:Instance",
			TotalCost:    100.0,
			Currency:     "", // Empty currency
			StartDate:    jan1,
			EndDate:      jan2,
		},
		{
			ResourceType: "azure:compute:VirtualMachine",
			TotalCost:    150.0,
			Currency:     "USD",
			StartDate:    jan1,
			EndDate:      jan2,
		},
	}

	aggregations, err := engine.CreateCrossProviderAggregation(results, engine.GroupByDaily)

	require.NoError(t, err)
	require.Len(t, aggregations, 1)

	// Empty currency should default to USD and match
	agg := aggregations[0]
	assert.Equal(t, "USD", agg.Currency)
	assert.Equal(t, 250.0, agg.Total)
}

// TestCreateCrossProviderAggregation_ZeroDateHandling tests handling of zero dates.
func TestCreateCrossProviderAggregation_ZeroDateHandling(t *testing.T) {
	jan1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	results := []engine.CostResult{
		{
			ResourceType: "aws:ec2:Instance",
			TotalCost:    100.0,
			Currency:     "USD",
			StartDate:    jan1,
			EndDate:      time.Time{}, // Zero date (should be valid)
		},
	}

	aggregations, err := engine.CreateCrossProviderAggregation(results, engine.GroupByDaily)

	// Should succeed (zero EndDate is allowed as long as it's not before StartDate)
	require.NoError(t, err)
	require.Len(t, aggregations, 1)
}

// TestCreateCrossProviderAggregation_SortingOrder tests chronological sorting.
func TestCreateCrossProviderAggregation_SortingOrder(t *testing.T) {
	jan1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	jan2 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	jan15 := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	jan16 := time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)
	dec31 := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	jan1Next := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	results := []engine.CostResult{
		{
			ResourceType: "aws:ec2:Instance",
			TotalCost:    100.0,
			Currency:     "USD",
			StartDate:    dec31, // Latest date
			EndDate:      jan1Next,
		},
		{
			ResourceType: "aws:ec2:Instance",
			TotalCost:    100.0,
			Currency:     "USD",
			StartDate:    jan1, // Earliest date
			EndDate:      jan2,
		},
		{
			ResourceType: "aws:ec2:Instance",
			TotalCost:    100.0,
			Currency:     "USD",
			StartDate:    jan15, // Middle date
			EndDate:      jan16,
		},
	}

	aggregations, err := engine.CreateCrossProviderAggregation(results, engine.GroupByDaily)

	require.NoError(t, err)
	require.Len(t, aggregations, 3)

	// Verify chronological order
	assert.Equal(t, "2024-01-01", aggregations[0].Period)
	assert.Equal(t, "2024-01-15", aggregations[1].Period)
	assert.Equal(t, "2024-12-31", aggregations[2].Period)
}

// TestAggregateResults_SingleResource tests aggregation with one resource.
func TestAggregateResults_SingleResource(t *testing.T) {
	results := []engine.CostResult{
		{
			ResourceType: "aws:ec2/instance:Instance",
			ResourceID:   "i-001",
			Monthly:      10.0,
			Hourly:       0.014,
			Currency:     "USD",
		},
	}

	aggregated := engine.AggregateResults(results)

	require.NotNil(t, aggregated)
	assert.Equal(t, 10.0, aggregated.Summary.TotalMonthly)
	assert.Equal(t, 0.014, aggregated.Summary.TotalHourly)
	assert.Equal(t, "USD", aggregated.Summary.Currency)
	assert.Len(t, aggregated.Resources, 1)
}

// TestAggregateResults_MultipleResources tests aggregation with multiple resources.
func TestAggregateResults_MultipleResources(t *testing.T) {
	results := []engine.CostResult{
		{
			ResourceType: "aws:ec2/instance:Instance",
			Monthly:      10.0,
			Hourly:       0.014,
			Currency:     "USD",
		},
		{
			ResourceType: "aws:s3/bucket:Bucket",
			Monthly:      5.0,
			Hourly:       0.007,
			Currency:     "USD",
		},
		{
			ResourceType: "aws:rds/instance:Instance",
			Monthly:      20.0,
			Hourly:       0.027,
			Currency:     "USD",
		},
	}

	aggregated := engine.AggregateResults(results)

	require.NotNil(t, aggregated)
	assert.Equal(t, 35.0, aggregated.Summary.TotalMonthly)
	assert.InDelta(t, 0.048, aggregated.Summary.TotalHourly, 0.001)
	assert.Len(t, aggregated.Resources, 3)
}

// TestAggregateResults_ByProvider tests provider-level aggregation.
func TestAggregateResults_ByProvider(t *testing.T) {
	results := []engine.CostResult{
		{
			ResourceType: "aws:ec2/instance:Instance",
			Monthly:      10.0,
			Currency:     "USD",
		},
		{
			ResourceType: "aws:s3/bucket:Bucket",
			Monthly:      5.0,
			Currency:     "USD",
		},
		{
			ResourceType: "azure:compute/virtualMachine:VirtualMachine",
			Monthly:      15.0,
			Currency:     "USD",
		},
	}

	aggregated := engine.AggregateResults(results)

	assert.Equal(t, 15.0, aggregated.Summary.ByProvider["aws"])
	assert.Equal(t, 15.0, aggregated.Summary.ByProvider["azure"])
}

// TestAggregateResults_ByService tests service-level aggregation.
func TestAggregateResults_ByService(t *testing.T) {
	results := []engine.CostResult{
		{
			ResourceType: "aws:ec2/instance:Instance",
			Monthly:      10.0,
			Currency:     "USD",
		},
		{
			ResourceType: "aws:ec2/volume:Volume",
			Monthly:      3.0,
			Currency:     "USD",
		},
		{
			ResourceType: "aws:s3/bucket:Bucket",
			Monthly:      5.0,
			Currency:     "USD",
		},
	}

	aggregated := engine.AggregateResults(results)

	assert.Equal(t, 13.0, aggregated.Summary.ByService["ec2"])
	assert.Equal(t, 5.0, aggregated.Summary.ByService["s3"])
}

// TestAggregateResults_ByAdapter tests adapter-level aggregation.
func TestAggregateResults_ByAdapter(t *testing.T) {
	results := []engine.CostResult{
		{
			ResourceType: "aws:ec2/instance:Instance",
			Adapter:      "plugin1",
			Monthly:      10.0,
			Currency:     "USD",
		},
		{
			ResourceType: "aws:s3/bucket:Bucket",
			Adapter:      "plugin2",
			Monthly:      5.0,
			Currency:     "USD",
		},
		{
			ResourceType: "aws:rds/instance:Instance",
			Adapter:      "local-spec",
			Monthly:      20.0,
			Currency:     "USD",
		},
	}

	aggregated := engine.AggregateResults(results)

	assert.Equal(t, 10.0, aggregated.Summary.ByAdapter["plugin1"])
	assert.Equal(t, 5.0, aggregated.Summary.ByAdapter["plugin2"])
	assert.Equal(t, 20.0, aggregated.Summary.ByAdapter["local-spec"])
}
