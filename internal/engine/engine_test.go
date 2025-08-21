package engine

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractService(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		expected     string
	}{
		{
			name:         "aws ec2 instance",
			resourceType: "aws:ec2:Instance",
			expected:     "ec2",
		},
		{
			name:         "aws rds instance",
			resourceType: "aws:rds:Instance",
			expected:     "rds",
		},
		{
			name:         "azure compute vm",
			resourceType: "azure:compute:VirtualMachine",
			expected:     "compute",
		},
		{
			name:         "single part type",
			resourceType: "simple",
			expected:     "default",
		},
		{
			name:         "empty type",
			resourceType: "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractService(tt.resourceType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractSKU(t *testing.T) {
	tests := []struct {
		name     string
		resource ResourceDescriptor
		expected string
	}{
		{
			name: "instanceType property",
			resource: ResourceDescriptor{
				Type: "aws:ec2:Instance",
				Properties: map[string]interface{}{
					"instanceType": "t3.micro",
				},
			},
			expected: "t3.micro",
		},
		{
			name: "sku property",
			resource: ResourceDescriptor{
				Type: "aws:ec2:Instance",
				Properties: map[string]interface{}{
					"sku": "standard-b2s",
				},
			},
			expected: "standard-b2s",
		},
		{
			name: "from resource type",
			resource: ResourceDescriptor{
				Type:       "aws:ec2:Instance",
				Properties: nil,
			},
			expected: "instance",
		},
		{
			name: "no properties",
			resource: ResourceDescriptor{
				Type:       "aws:ec2:Instance",
				Properties: map[string]interface{}{},
			},
			expected: "instance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSKU(tt.resource)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateCostsFromSpec(t *testing.T) {
	tests := []struct {
		name             string
		spec             *PricingSpec
		resource         ResourceDescriptor
		expectedMonthly  float64
		expectedHourly   float64
	}{
		{
			name: "monthly estimate in spec",
			spec: &PricingSpec{
				Currency: "USD",
				Pricing: map[string]interface{}{
					"monthlyEstimate": 7.59,
				},
			},
			resource:        ResourceDescriptor{},
			expectedMonthly: 7.59,
			expectedHourly:  7.59 / 730,
		},
		{
			name: "hourly rate in spec",
			spec: &PricingSpec{
				Currency: "USD",
				Pricing: map[string]interface{}{
					"onDemandHourly": 0.0104,
				},
			},
			resource:        ResourceDescriptor{},
			expectedMonthly: 0.0104 * 730,
			expectedHourly:  0.0104,
		},
		{
			name: "storage with size",
			spec: &PricingSpec{
				Currency: "USD",
				Pricing: map[string]interface{}{
					"pricePerGBMonth": 0.10,
				},
			},
			resource: ResourceDescriptor{
				Properties: map[string]interface{}{
					"size": 100.0,
				},
			},
			expectedMonthly: 10.0, // 100 GB * $0.10/GB
			expectedHourly:  10.0 / 730,
		},
		{
			name: "database fallback",
			spec: &PricingSpec{
				Currency: "USD",
				Pricing:  map[string]interface{}{},
			},
			resource: ResourceDescriptor{
				Type: "aws:rds:Instance",
			},
			expectedMonthly: 50.0,
			expectedHourly:  50.0 / 730,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monthly, hourly := calculateCostsFromSpec(tt.spec, tt.resource)
			assert.InDelta(t, tt.expectedMonthly, monthly, 0.01)
			assert.InDelta(t, tt.expectedHourly, hourly, 0.001)
		})
	}
}

func TestAggregateResults(t *testing.T) {
	results := []CostResult{
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

	aggregated := AggregateResults(results)

	// Check summary totals
	assert.Equal(t, 85.0, aggregated.Summary.TotalMonthly)
	assert.InDelta(t, 0.116, aggregated.Summary.TotalHourly, 0.001)
	assert.Equal(t, "USD", aggregated.Summary.Currency)

	// Check provider breakdown
	assert.Equal(t, 60.0, aggregated.Summary.ByProvider["aws"])
	assert.Equal(t, 25.0, aggregated.Summary.ByProvider["azure"])

	// Check service breakdown
	assert.Equal(t, 10.0, aggregated.Summary.ByService["ec2"])
	assert.Equal(t, 50.0, aggregated.Summary.ByService["rds"])
	assert.Equal(t, 25.0, aggregated.Summary.ByService["compute"])

	// Check adapter breakdown
	assert.Equal(t, 60.0, aggregated.Summary.ByAdapter["aws-plugin"])
	assert.Equal(t, 25.0, aggregated.Summary.ByAdapter["local-spec"])

	// Check resources are preserved
	assert.Len(t, aggregated.Resources, 3)
}

func TestFilterResources(t *testing.T) {
	resources := []ResourceDescriptor{
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
			filtered := FilterResources(resources, tt.filter)
			assert.Len(t, filtered, tt.expected)
		})
	}
}

func TestGetProjectedCostEmpty(t *testing.T) {
	// Test with no clients and no loader
	engine := New(nil, nil)
	
	results, err := engine.GetProjectedCost(context.Background(), []ResourceDescriptor{
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
	assert.Equal(t, 0.0, result.Monthly)
	assert.Equal(t, 0.0, result.Hourly)
	assert.Contains(t, result.Notes, "No pricing information available")
}

// MockSpecLoader for testing
type MockSpecLoader struct {
	specs map[string]*PricingSpec
}

func (m *MockSpecLoader) LoadSpec(provider, service, sku string) (interface{}, error) {
	key := provider + "-" + service + "-" + sku
	if spec, ok := m.specs[key]; ok {
		return spec, nil
	}
	return nil, ErrNoCostData
}

func TestGetProjectedCostWithSpecLoader(t *testing.T) {
	loader := &MockSpecLoader{
		specs: map[string]*PricingSpec{
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

	engine := New(nil, loader)
	
	results, err := engine.GetProjectedCost(context.Background(), []ResourceDescriptor{
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
	assert.Equal(t, 7.59, result.Monthly)
	assert.InDelta(t, 7.59/730, result.Hourly, 0.001)
	assert.Contains(t, result.Notes, "local spec")
}