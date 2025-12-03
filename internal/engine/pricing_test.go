package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDefaultMonthlyByType(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		expected     float64
	}{
		{"database", "aws:rds:dbInstance", defaultDatabaseMonthlyCost},
		{"database_case_insensitive", "aws:RDS:dbInstance", defaultDatabaseMonthlyCost},
		{"storage", "aws:s3:Bucket", defaultStorageMonthlyCost},
		{"storage_generic", "azure:storage:account", defaultStorageMonthlyCost},
		{"compute_default", "aws:ec2:Instance", defaultComputeMonthlyCost},
		{"unknown", "gcp:unknown:resource", defaultComputeMonthlyCost},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := getDefaultMonthlyByType(tt.resourceType)
			assert.Equal(t, tt.expected, cost)
		})
	}
}

func TestTryStoragePricing(t *testing.T) {
	tests := []struct {
		name            string
		pricing         map[string]interface{}
		resource        ResourceDescriptor
		expectedMonthly float64
		expectedHourly  float64
		expectedOk      bool
	}{
		{
			name:    "valid_pricing_with_size",
			pricing: map[string]interface{}{"pricePerGBMonth": 0.1},
			resource: ResourceDescriptor{
				Properties: map[string]interface{}{"size": 100.0},
			},
			expectedMonthly: 10.0,         // 100 * 0.1
			expectedHourly:  10.0 / 730.0, // 10 / 730
			expectedOk:      true,
		},
		{
			name:    "missing_size",
			pricing: map[string]interface{}{"pricePerGBMonth": 0.1},
			resource: ResourceDescriptor{
				Properties: map[string]interface{}{},
			},
			expectedMonthly: 0,
			expectedHourly:  0,
			expectedOk:      false,
		},
		{
			name:    "missing_pricing",
			pricing: map[string]interface{}{"other": 0.1},
			resource: ResourceDescriptor{
				Properties: map[string]interface{}{"size": 100.0},
			},
			expectedMonthly: 0,
			expectedHourly:  0,
			expectedOk:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monthly, hourly, ok := tryStoragePricing(tt.pricing, tt.resource)
			assert.Equal(t, tt.expectedOk, ok)
			if tt.expectedOk {
				assert.InDelta(t, tt.expectedMonthly, monthly, 0.0001)
				assert.InDelta(t, tt.expectedHourly, hourly, 0.0001)
			}
		})
	}
}

// TestParseFloatValue tests type conversion for various input types.
// This covers the parseFloatValue function which accepts float64, int, or numeric strings.
func TestParseFloatValue(t *testing.T) {
	tests := []struct {
		name       string
		input      interface{}
		expectedOk bool
		expected   float64
	}{
		// Direct float64 passthrough
		{"float64_positive", 42.5, true, 42.5},
		{"float64_zero", 0.0, true, 0.0},
		{"float64_negative", -123.45, true, -123.45},

		// Integer conversion
		{"int_positive", 100, true, 100.0},
		{"int_zero", 0, true, 0.0},
		{"int_negative", -50, true, -50.0},

		// String parsing
		{"string_integer", "42", true, 42.0},
		{"string_float", "3.14159", true, 3.14159},
		{"string_negative", "-99.9", true, -99.9},
		{"string_scientific", "1.5e10", true, 1.5e10},

		// Invalid inputs return false
		{"string_non_numeric", "not-a-number", false, 0},
		{"string_empty", "", false, 0},
		{"nil_value", nil, false, 0},
		{"bool_true", true, false, 0},
		{"bool_false", false, false, 0},
		{"struct_type", struct{}{}, false, 0},
		{"slice_type", []int{1, 2, 3}, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := parseFloatValue(tt.input)
			assert.Equal(t, tt.expectedOk, ok, "unexpected ok value")
			if tt.expectedOk {
				assert.InDelta(t, tt.expected, result, 0.0001, "unexpected result value")
			}
		})
	}
}
