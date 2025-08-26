package engine_test

import (
	"context"
	"testing"

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
