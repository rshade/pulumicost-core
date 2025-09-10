package engine_test

import (
	"context"
	"testing"
	"time"

	"github.com/rshade/pulumicost-core/internal/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAggregateResults(t *testing.T) {
	results := []engine.CostResult{
		{
			ResourceType: "aws:ec2:Instance",
			ResourceID:   "i-123",
			Adapter:      "aws-plugin",
			Currency:     "USD",
			Monthly:      10.0,
			Hourly:       0.014,
		},
		{
			ResourceType: "aws:rds:Instance",
			ResourceID:   "db-456",
			Adapter:      "aws-plugin",
			Currency:     "USD",
			Monthly:      50.0,
			Hourly:       0.068,
		},
		{
			ResourceType: "azure:compute:VirtualMachine",
			ResourceID:   "vm-789",
			Adapter:      "local-spec",
			Currency:     "USD",
			Monthly:      25.0,
			Hourly:       0.034,
		},
	}

	aggregated := engine.AggregateResults(results)

	// Check summary totals
	assert.InDelta(t, 85.0, aggregated.Summary.TotalMonthly, 0.01)
	assert.InDelta(t, 0.116, aggregated.Summary.TotalHourly, 0.001)
	assert.Equal(t, "USD", aggregated.Summary.Currency)

	// Check provider breakdown
	assert.InDelta(t, 60.0, aggregated.Summary.ByProvider["aws"], 0.01)
	assert.InDelta(t, 25.0, aggregated.Summary.ByProvider["azure"], 0.01)

	// Check service breakdown
	assert.InDelta(t, 10.0, aggregated.Summary.ByService["ec2"], 0.01)
	assert.InDelta(t, 50.0, aggregated.Summary.ByService["rds"], 0.01)
	assert.InDelta(t, 25.0, aggregated.Summary.ByService["compute"], 0.01)

	// Check adapter breakdown
	assert.InDelta(t, 60.0, aggregated.Summary.ByAdapter["aws-plugin"], 0.01)
	assert.InDelta(t, 25.0, aggregated.Summary.ByAdapter["local-spec"], 0.01)

	// Check resources are preserved
	assert.Len(t, aggregated.Resources, 3)
}

func TestFilterResources(t *testing.T) {
	resources := []engine.ResourceDescriptor{
		{
			Type:     "aws:ec2:Instance",
			ID:       "i-123",
			Provider: "aws",
			Properties: map[string]interface{}{
				"instanceType": "t3.micro",
			},
		},
		{
			Type:     "aws:rds:Instance",
			ID:       "db-456",
			Provider: "aws",
		},
		{
			Type:     "azure:compute:VirtualMachine",
			ID:       "vm-789",
			Provider: "azure",
		},
	}

	tests := []struct {
		name     string
		filter   string
		expected int
	}{
		{
			name:     "no filter",
			filter:   "",
			expected: 3,
		},
		{
			name:     "filter by provider",
			filter:   "provider=aws",
			expected: 2,
		},
		{
			name:     "filter by type",
			filter:   "type=ec2",
			expected: 1,
		},
		{
			name:     "filter by service",
			filter:   "service=rds",
			expected: 1,
		},
		{
			name:     "filter by property",
			filter:   "instanceType=t3.micro",
			expected: 1,
		},
		{
			name:     "invalid filter",
			filter:   "invalid",
			expected: 3, // Should include all on invalid filter
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := engine.FilterResources(resources, tt.filter)
			assert.Len(t, filtered, tt.expected)
		})
	}
}

func TestGetProjectedCostEmpty(t *testing.T) {
	// Test with no clients and no loader
	eng := engine.New(nil, nil)

	results, err := eng.GetProjectedCost(context.Background(), []engine.ResourceDescriptor{
		{
			Type: "aws:ec2:Instance",
			ID:   "i-123",
		},
	})

	require.NoError(t, err)
	require.Len(t, results, 1)

	result := results[0]
	assert.Equal(t, "aws:ec2:Instance", result.ResourceType)
	assert.Equal(t, "i-123", result.ResourceID)
	assert.Equal(t, "none", result.Adapter)
	assert.Equal(t, "USD", result.Currency)
	assert.InDelta(t, 0.0, result.Monthly, 0.01)
	assert.InDelta(t, 0.0, result.Hourly, 0.01)
	assert.Contains(t, result.Notes, "No pricing information available")
}

// MockSpecLoader for testing.
type MockSpecLoader struct {
	specs map[string]*engine.PricingSpec
}

func (m *MockSpecLoader) LoadSpec(provider, service, sku string) (interface{}, error) {
	key := provider + "-" + service + "-" + sku
	if spec, ok := m.specs[key]; ok {
		return spec, nil
	}
	return nil, engine.ErrNoCostData
}

func TestGetProjectedCostWithSpecLoader(t *testing.T) {
	loader := &MockSpecLoader{
		specs: map[string]*engine.PricingSpec{
			"aws-ec2-t3.micro": {
				Provider: "aws",
				Service:  "ec2",
				SKU:      "t3.micro",
				Currency: "USD",
				Pricing: map[string]interface{}{
					"monthlyEstimate": 7.59,
				},
			},
		},
	}

	eng := engine.New(nil, loader)

	results, err := eng.GetProjectedCost(context.Background(), []engine.ResourceDescriptor{
		{
			Type:     "aws:ec2:Instance",
			ID:       "i-123",
			Provider: "aws",
			Properties: map[string]interface{}{
				"instanceType": "t3.micro",
			},
		},
	})

	require.NoError(t, err)
	require.Len(t, results, 1)

	result := results[0]
	assert.Equal(t, "aws:ec2:Instance", result.ResourceType)
	assert.Equal(t, "i-123", result.ResourceID)
	assert.Equal(t, "local-spec", result.Adapter)
	assert.Equal(t, "USD", result.Currency)
	assert.InDelta(t, 7.59, result.Monthly, 0.01)
	assert.InDelta(t, 7.59/730, result.Hourly, 0.001)
	assert.Contains(t, result.Notes, "local spec")
}

func TestFormatPeriod(t *testing.T) {
	tests := []struct {
		name     string
		from     time.Time
		to       time.Time
		expected string
	}{
		{
			name:     "1 day",
			from:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			to:       time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
			expected: "1 day",
		},
		{
			name:     "3 days",
			from:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			to:       time.Date(2025, 1, 4, 0, 0, 0, 0, time.UTC),
			expected: "3 days",
		},
		{
			name:     "1 week",
			from:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			to:       time.Date(2025, 1, 8, 0, 0, 0, 0, time.UTC),
			expected: "1 weeks",
		},
		{
			name:     "2 weeks",
			from:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			to:       time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			expected: "2 weeks",
		},
		{
			name:     "1 month",
			from:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			to:       time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
			expected: "1 months",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.FormatPeriod(tt.from, tt.to)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchesTags(t *testing.T) {
	tests := []struct {
		name     string
		resource engine.ResourceDescriptor
		tags     map[string]string
		expected bool
	}{
		{
			name:     "empty tags match any resource",
			resource: engine.ResourceDescriptor{Properties: map[string]interface{}{"env": "prod"}},
			tags:     map[string]string{},
			expected: true,
		},
		{
			name: "matching tag",
			resource: engine.ResourceDescriptor{
				Properties: map[string]interface{}{"env": "prod", "team": "backend"},
			},
			tags:     map[string]string{"env": "prod"},
			expected: true,
		},
		{
			name: "multiple matching tags",
			resource: engine.ResourceDescriptor{
				Properties: map[string]interface{}{"env": "prod", "team": "backend"},
			},
			tags:     map[string]string{"env": "prod", "team": "backend"},
			expected: true,
		},
		{
			name: "non-matching tag",
			resource: engine.ResourceDescriptor{
				Properties: map[string]interface{}{"env": "staging"},
			},
			tags:     map[string]string{"env": "prod"},
			expected: false,
		},
		{
			name: "missing tag",
			resource: engine.ResourceDescriptor{
				Properties: map[string]interface{}{"team": "backend"},
			},
			tags:     map[string]string{"env": "prod"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.MatchesTags(tt.resource, tt.tags)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGroupResults(t *testing.T) {
	eng := &engine.Engine{}

	results := []engine.CostResult{
		{
			ResourceType: "aws:ec2/instance:Instance",
			ResourceID:   "i-123",
			Adapter:      "aws",
			TotalCost:    10.0,
		},
		{
			ResourceType: "aws:ec2/instance:Instance",
			ResourceID:   "i-456",
			Adapter:      "aws",
			TotalCost:    15.0,
		},
		{
			ResourceType: "gcp:compute/instance:Instance",
			ResourceID:   "vm-789",
			Adapter:      "gcp",
			TotalCost:    20.0,
		},
	}

	tests := []struct {
		name             string
		groupBy          engine.GroupBy
		expectedGroups   int
		expectedTotalSum float64
	}{
		{
			name:             "no grouping",
			groupBy:          engine.GroupByNone,
			expectedGroups:   3,
			expectedTotalSum: 45.0,
		},
		{
			name:             "group by type",
			groupBy:          engine.GroupByType,
			expectedGroups:   2, // aws:ec2 and gcp:compute grouped separately
			expectedTotalSum: 45.0,
		},
		{
			name:             "group by provider",
			groupBy:          engine.GroupByProvider,
			expectedGroups:   2, // aws and gcp
			expectedTotalSum: 45.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grouped := eng.GroupResults(results, tt.groupBy)
			assert.Len(t, grouped, tt.expectedGroups)

			// Calculate total cost across all groups
			var totalCost float64
			for _, result := range grouped {
				totalCost += result.TotalCost
			}
			assert.InDelta(t, tt.expectedTotalSum, totalCost, 0.01)
		})
	}
}

func TestAggregateResultsFunction(t *testing.T) {
	results := []engine.CostResult{
		{
			ResourceType: "aws:ec2/instance:Instance",
			ResourceID:   "i-123",
			TotalCost:    10.0,
			Monthly:      100.0,
			Breakdown:    map[string]float64{"compute": 10.0},
		},
		{
			ResourceType: "aws:ec2/instance:Instance",
			ResourceID:   "i-456",
			TotalCost:    15.0,
			Monthly:      150.0,
			Breakdown:    map[string]float64{"compute": 15.0},
		},
	}

	aggregated := engine.AggregateResultsInternal(results, "test-group")

	assert.Equal(t, "test-group", aggregated.ResourceType)
	assert.Equal(t, "aggregated", aggregated.Adapter)
	assert.InDelta(t, 25.0, aggregated.TotalCost, 0.01)
	assert.InDelta(t, 250.0, aggregated.Monthly, 0.01)
	assert.InDelta(t, 25.0, aggregated.Breakdown["compute"], 0.01)
	assert.Contains(t, aggregated.Notes, "2 resources")
}

func TestGetActualCostWithOptions(t *testing.T) {
	eng := engine.New(nil, nil) // No plugins for this test

	resources := []engine.ResourceDescriptor{
		{
			Type:       "aws:ec2/instance:Instance",
			ID:         "i-123",
			Provider:   "aws",
			Properties: map[string]interface{}{"env": "prod"},
		},
		{
			Type:       "aws:s3/bucket:Bucket",
			ID:         "bucket-456",
			Provider:   "aws",
			Properties: map[string]interface{}{"env": "staging"},
		},
	}

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 1, 8, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name            string
		request         engine.ActualCostRequest
		expectedResults int
	}{
		{
			name: "no filters",
			request: engine.ActualCostRequest{
				Resources: resources,
				From:      from,
				To:        to,
			},
			expectedResults: 2,
		},
		{
			name: "filter by tag",
			request: engine.ActualCostRequest{
				Resources: resources,
				From:      from,
				To:        to,
				Tags:      map[string]string{"env": "prod"},
			},
			expectedResults: 1,
		},
		{
			name: "group by provider",
			request: engine.ActualCostRequest{
				Resources: resources,
				From:      from,
				To:        to,
				GroupBy:   "provider",
			},
			expectedResults: 1, // Both are AWS resources, should be grouped
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			results, err := eng.GetActualCostWithOptions(ctx, tt.request)

			require.NoError(t, err)
			assert.Len(t, results, tt.expectedResults)

			// Check results based on test case
			for _, result := range results {
				if tt.name == "group by provider" {
					// When grouped, adapter should be "aggregated"
					assert.Equal(t, "aggregated", result.Adapter)
				} else {
					// Individual results should have "none" adapter
					assert.Equal(t, "none", result.Adapter)
				}
				assert.Equal(t, "1 weeks", result.CostPeriod)
			}
		})
	}
}
