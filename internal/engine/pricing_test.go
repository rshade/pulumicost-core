package engine

import (
	"context"
	"testing"
	"time"

	"github.com/rshade/pulumicost-core/internal/pluginhost"
	"github.com/stretchr/testify/assert"
)

func TestGetDefaultMonthlyByType(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		expected     float64
	}{
		{"database_rds", "aws:rds/instance:Instance", defaultDatabaseMonthlyCost},
		{"database_generic", "azure:db:Sql", defaultDatabaseMonthlyCost},
		{"storage_s3", "aws:s3/bucket:Bucket", defaultStorageMonthlyCost},
		{"storage_generic", "gcp:storage:Bucket", defaultStorageMonthlyCost},
		{"compute_ec2", "aws:ec2/instance:Instance", defaultComputeMonthlyCost},
		{"compute_generic", "unknown:resource:Type", defaultComputeMonthlyCost},
		{"case_insensitive", "AWS:RDS:INSTANCE", defaultDatabaseMonthlyCost},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDefaultMonthlyByType(tt.resourceType)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestParseFloatValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		wantVal  float64
		wantOk   bool
	}{
		{"float64", 10.5, 10.5, true},
		{"int", 100, 100.0, true},
		{"string_float", "12.34", 12.34, true},
		{"string_int", "50", 50.0, true},
		{"string_invalid", "abc", 0, false},
		{"nil", nil, 0, false},
		{"bool", true, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ok := parseFloatValue(tt.input)
			assert.Equal(t, tt.wantOk, ok)
			if ok {
				assert.Equal(t, tt.wantVal, val)
			}
		})
	}
}

func TestGetStorageSize(t *testing.T) {
	tests := []struct {
		name       string
		properties map[string]interface{}
		wantSize   float64
		wantFound  bool
	}{
		{"size_prop", map[string]interface{}{"size": 100}, 100.0, true},
		{"sizeGb_prop", map[string]interface{}{"sizeGb": "200"}, 200.0, true},
		{"volumeSize_prop", map[string]interface{}{"volumeSize": 50.5}, 50.5, true},
		{"allocatedStorage_prop", map[string]interface{}{"allocatedStorage": 500}, 500.0, true},
		{"no_size_prop", map[string]interface{}{"name": "test"}, 0, false},
		{"nil_props", nil, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := ResourceDescriptor{Properties: tt.properties}
			size, found := getStorageSize(res)
			assert.Equal(t, tt.wantFound, found)
			if found {
				assert.Equal(t, tt.wantSize, size)
			}
		})
	}
}

func TestTryStoragePricing(t *testing.T) {
	tests := []struct {
		name        string
		pricing     map[string]interface{}
		properties  map[string]interface{}
		wantMonthly float64
		wantHourly  float64
		wantFound   bool
	}{
		{
			name:        "valid_storage_pricing",
			pricing:     map[string]interface{}{"pricePerGBMonth": 0.10},
			properties:  map[string]interface{}{"size": 100},
			wantMonthly: 10.0, // 100 * 0.10
			wantHourly:  10.0 / hoursPerMonth,
			wantFound:   true,
		},
		{
			name:       "missing_size",
			pricing:    map[string]interface{}{"pricePerGBMonth": 0.10},
			properties: map[string]interface{}{"name": "test"},
			wantFound:  false,
		},
		{
			name:       "missing_pricing_key",
			pricing:    map[string]interface{}{"otherKey": 0.10},
			properties: map[string]interface{}{"size": 100},
			wantFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := ResourceDescriptor{Properties: tt.properties}
			m, h, found := tryStoragePricing(tt.pricing, res)
			assert.Equal(t, tt.wantFound, found)
			if found {
				assert.InDelta(t, tt.wantMonthly, m, 0.0001)
				assert.InDelta(t, tt.wantHourly, h, 0.0001)
			}
		})
	}
}

func TestTryFallbackNumericValue(t *testing.T) {
	tests := []struct {
		name        string
		pricing     map[string]interface{}
		wantMonthly float64
		wantHourly  float64
		wantFound   bool
	}{
		{
			name:        "single_numeric_value",
			pricing:     map[string]interface{}{"flatRate": 0.5},
			wantMonthly: 0.5 * hoursPerMonth,
			wantHourly:  0.5,
			wantFound:   true,
		},
		{
			name:      "no_numeric_values",
			pricing:   map[string]interface{}{"desc": "text only"},
			wantFound: false,
		},
		{
			name:      "zero_value_ignored",
			pricing:   map[string]interface{}{"free": 0.0},
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, h, found := tryFallbackNumericValue(tt.pricing)
			assert.Equal(t, tt.wantFound, found)
			if found {
				assert.InDelta(t, tt.wantMonthly, m, 0.0001)
				assert.InDelta(t, tt.wantHourly, h, 0.0001)
			}
		})
	}
}

func TestGetActualCost_Wrapper(t *testing.T) {
	// Test the simple wrapper function GetActualCost
	// We can't easily mock the full engine here without setting up plugins, 
	// but we can test that it calls GetActualCostWithOptions correctly.
	// Since we are in the engine package, we can just create an engine with no clients/loader
	// and expect it to return results (even if empty or placeholders).
	
	e := New([]*pluginhost.Client{}, nil)
	
	resources := []ResourceDescriptor{
		{Type: "aws:s3:Bucket", ID: "bucket-1", Provider: "aws"},
	}
	from := time.Now().Add(-24 * time.Hour)
	to := time.Now()
	
	results, err := e.GetActualCost(context.Background(), resources, from, to)
	
	// Should return a result (fallback placeholder) and no error
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "none", results[0].Adapter)
	assert.Equal(t, 0.0, results[0].TotalCost)
}
