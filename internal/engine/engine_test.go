package engine

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			result := formatPeriod(tt.from, tt.to)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchesTags(t *testing.T) {
	tests := []struct {
		name     string
		resource ResourceDescriptor
		tags     map[string]string
		expected bool
	}{
		{
			name:     "empty tags match any resource",
			resource: ResourceDescriptor{Properties: map[string]interface{}{"env": "prod"}},
			tags:     map[string]string{},
			expected: true,
		},
		{
			name: "matching tag",
			resource: ResourceDescriptor{
				Properties: map[string]interface{}{"env": "prod", "team": "backend"},
			},
			tags:     map[string]string{"env": "prod"},
			expected: true,
		},
		{
			name: "multiple matching tags",
			resource: ResourceDescriptor{
				Properties: map[string]interface{}{"env": "prod", "team": "backend"},
			},
			tags:     map[string]string{"env": "prod", "team": "backend"},
			expected: true,
		},
		{
			name: "non-matching tag",
			resource: ResourceDescriptor{
				Properties: map[string]interface{}{"env": "staging"},
			},
			tags:     map[string]string{"env": "prod"},
			expected: false,
		},
		{
			name: "missing tag",
			resource: ResourceDescriptor{
				Properties: map[string]interface{}{"team": "backend"},
			},
			tags:     map[string]string{"env": "prod"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesTags(tt.resource, tt.tags)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGroupResults(t *testing.T) {
	engine := &Engine{}
	
	results := []CostResult{
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
		groupBy          GroupBy
		expectedGroups   int
		expectedTotalSum float64
	}{
		{
			name:             "no grouping",
			groupBy:          GroupByNone,
			expectedGroups:   3,
			expectedTotalSum: 45.0,
		},
		{
			name:             "group by type",
			groupBy:          GroupByType,
			expectedGroups:   2, // aws:ec2 and gcp:compute grouped separately
			expectedTotalSum: 45.0,
		},
		{
			name:             "group by provider",
			groupBy:          GroupByProvider,
			expectedGroups:   2, // aws and gcp
			expectedTotalSum: 45.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grouped := engine.groupResults(results, tt.groupBy)
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

func TestAggregateResults(t *testing.T) {
	results := []CostResult{
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

	aggregated := aggregateResults(results, "test-group")
	
	assert.Equal(t, "test-group", aggregated.ResourceType)
	assert.Equal(t, "aggregated", aggregated.Adapter)
	assert.Equal(t, 25.0, aggregated.TotalCost)
	assert.Equal(t, 250.0, aggregated.Monthly)
	assert.Equal(t, 25.0, aggregated.Breakdown["compute"])
	assert.Contains(t, aggregated.Notes, "2 resources")
}

func TestGetActualCostWithOptions(t *testing.T) {
	engine := New(nil, nil) // No plugins for this test
	
	resources := []ResourceDescriptor{
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
		request         ActualCostRequest
		expectedResults int
	}{
		{
			name: "no filters",
			request: ActualCostRequest{
				Resources: resources,
				From:      from,
				To:        to,
			},
			expectedResults: 2,
		},
		{
			name: "filter by tag",
			request: ActualCostRequest{
				Resources: resources,
				From:      from,
				To:        to,
				Tags:      map[string]string{"env": "prod"},
			},
			expectedResults: 1,
		},
		{
			name: "group by provider",
			request: ActualCostRequest{
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
			results, err := engine.GetActualCostWithOptions(ctx, tt.request)
			
			require.NoError(t, err)
			assert.Len(t, results, tt.expectedResults)
			
			// All results should have placeholder data since no plugins
			for _, result := range results {
				assert.Equal(t, "none", result.Adapter)
				assert.Equal(t, "1 weeks", result.CostPeriod)
			}
		})
	}
}