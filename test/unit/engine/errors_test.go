package engine_test

import (
	"testing"
	"time"

	"github.com/rshade/pulumicost-core/internal/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateCrossProviderAggregation_EmptyResults tests error handling for empty results.
func TestCreateCrossProviderAggregation_EmptyResults(t *testing.T) {
	results := []engine.CostResult{}

	_, err := engine.CreateCrossProviderAggregation(results, engine.GroupByDaily)

	require.Error(t, err)
	assert.ErrorIs(t, err, engine.ErrEmptyResults)
}

// TestCreateCrossProviderAggregation_InvalidGroupBy tests error handling for invalid grouping.
func TestCreateCrossProviderAggregation_InvalidGroupBy(t *testing.T) {
	results := []engine.CostResult{
		{
			ResourceType: "aws:ec2:Instance",
			TotalCost:    100.0,
			Currency:     "USD",
			StartDate:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:      time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		},
	}

	// Non-time-based grouping should fail
	_, err := engine.CreateCrossProviderAggregation(results, engine.GroupByResource)

	require.Error(t, err)
	assert.ErrorIs(t, err, engine.ErrInvalidGroupBy)
}

// TestCreateCrossProviderAggregation_MixedCurrencies tests error handling for mixed currencies.
func TestCreateCrossProviderAggregation_MixedCurrencies(t *testing.T) {
	results := []engine.CostResult{
		{
			ResourceType: "aws:ec2:Instance",
			TotalCost:    100.0,
			Currency:     "USD",
			StartDate:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:      time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			ResourceType: "azure:compute:VirtualMachine",
			TotalCost:    150.0,
			Currency:     "EUR", // Different currency
			StartDate:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:      time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		},
	}

	_, err := engine.CreateCrossProviderAggregation(results, engine.GroupByDaily)

	require.Error(t, err)
	assert.ErrorIs(t, err, engine.ErrMixedCurrencies)
	assert.Contains(t, err.Error(), "USD")
	assert.Contains(t, err.Error(), "EUR")
}

// TestCreateCrossProviderAggregation_InvalidDateRange tests error handling for invalid date ranges.
func TestCreateCrossProviderAggregation_InvalidDateRange(t *testing.T) {
	results := []engine.CostResult{
		{
			ResourceType: "aws:ec2:Instance",
			TotalCost:    100.0,
			Currency:     "USD",
			StartDate:    time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			EndDate:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), // EndDate before StartDate
		},
	}

	_, err := engine.CreateCrossProviderAggregation(results, engine.GroupByDaily)

	require.Error(t, err)
	assert.ErrorIs(t, err, engine.ErrInvalidDateRange)
}

// TestGroupBy_IsValid tests GroupBy validation.
func TestGroupBy_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		groupBy  engine.GroupBy
		expected bool
	}{
		{"Resource grouping", engine.GroupByResource, true},
		{"Type grouping", engine.GroupByType, true},
		{"Provider grouping", engine.GroupByProvider, true},
		{"Daily grouping", engine.GroupByDaily, true},
		{"Monthly grouping", engine.GroupByMonthly, true},
		{"Date grouping", engine.GroupByDate, true},
		{"None grouping", engine.GroupByNone, true},
		{"Invalid grouping", engine.GroupBy("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.groupBy.IsValid())
		})
	}
}

// TestGroupBy_IsTimeBasedGrouping tests time-based grouping detection.
func TestGroupBy_IsTimeBasedGrouping(t *testing.T) {
	tests := []struct {
		name     string
		groupBy  engine.GroupBy
		expected bool
	}{
		{"Daily is time-based", engine.GroupByDaily, true},
		{"Monthly is time-based", engine.GroupByMonthly, true},
		{"Date is not time-based", engine.GroupByDate, false}, // Deprecated, non time-based
		{"Resource is not time-based", engine.GroupByResource, false},
		{"Type is not time-based", engine.GroupByType, false},
		{"Provider is not time-based", engine.GroupByProvider, false},
		{"None is not time-based", engine.GroupByNone, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.groupBy.IsTimeBasedGrouping())
		})
	}
}

// TestGroupBy_String tests string representation.
func TestGroupBy_String(t *testing.T) {
	tests := []struct {
		name     string
		groupBy  engine.GroupBy
		expected string
	}{
		{"resource", engine.GroupByResource, "resource"},
		{"type", engine.GroupByType, "type"},
		{"provider", engine.GroupByProvider, "provider"},
		{"daily", engine.GroupByDaily, "daily"},
		{"monthly", engine.GroupByMonthly, "monthly"},
		{"date", engine.GroupByDate, "date"},
		{"none", engine.GroupByNone, ""}, // GroupByNone is empty string
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.groupBy.String())
		})
	}
}

// TestMatchesTags tests tag matching logic.
func TestMatchesTags(t *testing.T) {
	tests := []struct {
		name        string
		resource    engine.ResourceDescriptor
		tags        map[string]string
		shouldMatch bool
	}{
		{
			name: "Empty tags always match",
			resource: engine.ResourceDescriptor{
				Properties: map[string]interface{}{
					"env": "prod",
				},
			},
			tags:        map[string]string{},
			shouldMatch: true,
		},
		{
			name: "Exact tag match",
			resource: engine.ResourceDescriptor{
				Properties: map[string]interface{}{
					"env":  "prod",
					"team": "backend",
				},
			},
			tags: map[string]string{
				"env": "prod",
			},
			shouldMatch: true,
		},
		{
			name: "Multiple tag match",
			resource: engine.ResourceDescriptor{
				Properties: map[string]interface{}{
					"env":  "prod",
					"team": "backend",
				},
			},
			tags: map[string]string{
				"env":  "prod",
				"team": "backend",
			},
			shouldMatch: true,
		},
		{
			name: "Tag mismatch",
			resource: engine.ResourceDescriptor{
				Properties: map[string]interface{}{
					"env": "prod",
				},
			},
			tags: map[string]string{
				"env": "dev",
			},
			shouldMatch: false,
		},
		{
			name: "Missing tag",
			resource: engine.ResourceDescriptor{
				Properties: map[string]interface{}{
					"env": "prod",
				},
			},
			tags: map[string]string{
				"team": "backend",
			},
			shouldMatch: false,
		},
		{
			name:     "Nil properties",
			resource: engine.ResourceDescriptor{},
			tags: map[string]string{
				"env": "prod",
			},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.MatchesTags(tt.resource, tt.tags)
			assert.Equal(t, tt.shouldMatch, result)
		})
	}
}

// TestFilterResources tests resource filtering logic.
func TestFilterResources(t *testing.T) {
	resources := []engine.ResourceDescriptor{
		{
			Type:     "aws:ec2/instance:Instance",
			ID:       "i-001",
			Provider: "aws",
			Properties: map[string]interface{}{
				"instanceType": "t3.micro",
			},
		},
		{
			Type:     "azure:compute/virtualMachine:VirtualMachine",
			ID:       "vm-001",
			Provider: "azure",
			Properties: map[string]interface{}{
				"size": "Standard_B1s",
			},
		},
		{
			Type:     "aws:s3/bucket:Bucket",
			ID:       "bucket-001",
			Provider: "aws",
			Properties: map[string]interface{}{
				"region": "us-east-1",
			},
		},
	}

	tests := []struct {
		name          string
		filter        string
		expectedCount int
		expectedIDs   []string
	}{
		{
			name:          "No filter returns all",
			filter:        "",
			expectedCount: 3,
			expectedIDs:   []string{"i-001", "vm-001", "bucket-001"},
		},
		{
			name:          "Filter by provider",
			filter:        "provider=aws",
			expectedCount: 2,
			expectedIDs:   []string{"i-001", "bucket-001"},
		},
		{
			name:          "Filter by type",
			filter:        "type=ec2",
			expectedCount: 1,
			expectedIDs:   []string{"i-001"},
		},
		{
			name:          "Filter by service",
			filter:        "service=s3",
			expectedCount: 1,
			expectedIDs:   []string{"bucket-001"},
		},
		{
			name:          "Filter by ID",
			filter:        "id=vm-001",
			expectedCount: 1,
			expectedIDs:   []string{"vm-001"},
		},
		{
			name:          "Filter by property",
			filter:        "instanceType=t3.micro",
			expectedCount: 1,
			expectedIDs:   []string{"i-001"},
		},
		{
			name:          "No matches",
			filter:        "provider=gcp",
			expectedCount: 0,
			expectedIDs:   []string{},
		},
		{
			name:          "Invalid filter format",
			filter:        "invalidfilter",
			expectedCount: 3, // Invalid filter includes all
			expectedIDs:   []string{"i-001", "vm-001", "bucket-001"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := engine.FilterResources(resources, tt.filter)
			assert.Len(t, filtered, tt.expectedCount)

			if len(tt.expectedIDs) > 0 {
				actualIDs := make([]string, len(filtered))
				for i, r := range filtered {
					actualIDs[i] = r.ID
				}
				assert.ElementsMatch(t, tt.expectedIDs, actualIDs)
			}
		})
	}
}

// TestFormatPeriod tests period formatting.
func TestFormatPeriod(t *testing.T) {
	tests := []struct {
		name     string
		from     time.Time
		to       time.Time
		expected string
	}{
		{
			name:     "One day",
			from:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			to:       time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			expected: "1 day",
		},
		{
			name:     "Multiple days",
			from:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			to:       time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
			expected: "4 days",
		},
		{
			name:     "Weeks",
			from:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			to:       time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			expected: "2 weeks",
		},
		{
			name:     "Months",
			from:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			to:       time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
			expected: "2 months",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.FormatPeriod(tt.from, tt.to)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestAggregateResults_EmptyInput tests aggregation with empty results.
func TestAggregateResults_EmptyInput(t *testing.T) {
	results := []engine.CostResult{}

	aggregated := engine.AggregateResults(results)

	require.NotNil(t, aggregated)
	assert.Equal(t, 0.0, aggregated.Summary.TotalMonthly)
	assert.Equal(t, 0.0, aggregated.Summary.TotalHourly)
	assert.Equal(t, "USD", aggregated.Summary.Currency) // Default currency
	assert.Empty(t, aggregated.Resources)
}

// TestAggregateResults_NilBreakdown tests handling of nil breakdown maps.
func TestAggregateResults_NilBreakdown(t *testing.T) {
	results := []engine.CostResult{
		{
			ResourceType: "aws:ec2/instance:Instance",
			Monthly:      10.0,
			Breakdown:    nil, // Nil breakdown
		},
	}

	aggregated := engine.AggregateResults(results)

	require.NotNil(t, aggregated)
	assert.Equal(t, 10.0, aggregated.Summary.TotalMonthly)
	// Should not panic on nil breakdown
}
