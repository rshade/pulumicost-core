package engine

import (
	"encoding/json"
	"io"
	"testing"
)

// TestRenderResults tests the main rendering dispatcher.
func TestRenderResults(t *testing.T) {
	tests := []struct {
		name    string
		format  OutputFormat
		results []CostResult
		wantErr bool
	}{
		{
			name:    "table format with empty results",
			format:  OutputTable,
			results: []CostResult{},
			wantErr: false,
		},
		{
			name:   "table format with single resource",
			format: OutputTable,
			results: []CostResult{
				{
					ResourceType: "aws:ec2:Instance",
					ResourceID:   "i-123",
					Adapter:      "kubecost",
					Currency:     "USD",
					Monthly:      73.00,
					Hourly:       0.10,
				},
			},
			wantErr: false,
		},
		{
			name:   "json format with results",
			format: OutputJSON,
			results: []CostResult{
				{ResourceType: "aws:ec2:Instance", Monthly: 50.00, Currency: "USD"},
			},
			wantErr: false,
		},
		{
			name:   "ndjson format with results",
			format: OutputNDJSON,
			results: []CostResult{
				{ResourceType: "aws:ec2:Instance", Monthly: 50.00, Currency: "USD"},
				{ResourceType: "aws:rds:Instance", Monthly: 100.00, Currency: "USD"},
			},
			wantErr: false,
		},
		{
			name:    "unsupported format",
			format:  OutputFormat("invalid"),
			results: []CostResult{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RenderResults(io.Discard, tt.format, tt.results)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderResults() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestRenderActualCostResults tests actual cost rendering.
func TestRenderActualCostResults(t *testing.T) {
	tests := []struct {
		name    string
		format  OutputFormat
		results []CostResult
		wantErr bool
	}{
		{
			name:    "table format with actual costs",
			format:  OutputTable,
			results: []CostResult{{TotalCost: 100.0, CostPeriod: "monthly", Currency: "USD"}},
			wantErr: false,
		},
		{
			name:    "json format with actual costs",
			format:  OutputJSON,
			results: []CostResult{{TotalCost: 100.0, Currency: "USD"}},
			wantErr: false,
		},
		{
			name:    "ndjson format with actual costs",
			format:  OutputNDJSON,
			results: []CostResult{{TotalCost: 100.0, Currency: "USD"}},
			wantErr: false,
		},
		{
			name:    "unsupported format",
			format:  OutputFormat("invalid"),
			results: []CostResult{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RenderActualCostResults(io.Discard, tt.format, tt.results)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderActualCostResults() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestRenderCrossProviderAggregation tests cross-provider rendering.
func TestRenderCrossProviderAggregation(t *testing.T) {
	aggregations := []CrossProviderAggregation{
		{
			Period:    "2025-01-01",
			Total:     100.0,
			Currency:  "USD",
			Providers: map[string]float64{"aws": 100.0},
		},
	}

	tests := []struct {
		name         string
		format       OutputFormat
		aggregations []CrossProviderAggregation
		groupBy      GroupBy
		wantErr      bool
	}{
		{
			name:         "table format daily",
			format:       OutputTable,
			aggregations: aggregations,
			groupBy:      GroupByDaily,
			wantErr:      false,
		},
		{
			name:         "table format monthly",
			format:       OutputTable,
			aggregations: aggregations,
			groupBy:      GroupByMonthly,
			wantErr:      false,
		},
		{
			name:         "json format",
			format:       OutputJSON,
			aggregations: aggregations,
			groupBy:      GroupByDaily,
			wantErr:      false,
		},
		{
			name:         "ndjson format",
			format:       OutputNDJSON,
			aggregations: aggregations,
			groupBy:      GroupByDaily,
			wantErr:      false,
		},
		{
			name:         "unsupported format",
			format:       OutputFormat("invalid"),
			aggregations: aggregations,
			groupBy:      GroupByDaily,
			wantErr:      true,
		},
		{
			name:         "empty aggregations",
			format:       OutputTable,
			aggregations: []CrossProviderAggregation{},
			groupBy:      GroupByDaily,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RenderCrossProviderAggregation(
				io.Discard,
				tt.format,
				tt.aggregations,
				tt.groupBy,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderCrossProviderAggregation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGetCurrencySymbol tests currency symbol mapping.
func TestGetCurrencySymbol(t *testing.T) {
	tests := []struct {
		currency string
		expected string
	}{
		{"USD", "$"},
		{"EUR", "€"},
		{"GBP", "£"},
		{"JPY", "¥"},
		{"CAD", "C$"},
		{"AUD", "A$"},
		{"UNKNOWN", "UNKNOWN"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.currency, func(t *testing.T) {
			got := getCurrencySymbol(tt.currency)
			if got != tt.expected {
				t.Errorf("getCurrencySymbol(%q) = %q, want %q", tt.currency, got, tt.expected)
			}
		})
	}
}

// TestOutputFormatConstants tests that output format constants are defined.
func TestOutputFormatConstants(t *testing.T) {
	tests := []struct {
		name   string
		format OutputFormat
		value  string
	}{
		{"table format", OutputTable, "table"},
		{"json format", OutputJSON, "json"},
		{"ndjson format", OutputNDJSON, "ndjson"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.format) != tt.value {
				t.Errorf("OutputFormat = %q, want %q", tt.format, tt.value)
			}
		})
	}
}

// TestJSONMarshaling tests that CostResult and CrossProviderAggregation marshal correctly.
func TestJSONMarshaling(t *testing.T) {
	t.Run("CostResult marshals to JSON", func(t *testing.T) {
		result := CostResult{
			ResourceType: "aws:ec2:Instance",
			ResourceID:   "i-123",
			Adapter:      "kubecost",
			Currency:     "USD",
			Monthly:      100.0,
			Hourly:       0.137,
		}

		data, err := json.Marshal(result)
		if err != nil {
			t.Fatalf("Failed to marshal CostResult: %v", err)
		}

		var unmarshaled CostResult
		if err = json.Unmarshal(data, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal CostResult: %v", err)
		}

		if unmarshaled.ResourceType != result.ResourceType {
			t.Errorf("ResourceType = %q, want %q", unmarshaled.ResourceType, result.ResourceType)
		}
		if unmarshaled.Monthly != result.Monthly {
			t.Errorf("Monthly = %f, want %f", unmarshaled.Monthly, result.Monthly)
		}
	})

	t.Run("CrossProviderAggregation marshals to JSON", func(t *testing.T) {
		agg := CrossProviderAggregation{
			Period:    "2025-01-01",
			Total:     150.0,
			Currency:  "USD",
			Providers: map[string]float64{"aws": 100.0, "azure": 50.0},
		}

		data, err := json.Marshal(agg)
		if err != nil {
			t.Fatalf("Failed to marshal CrossProviderAggregation: %v", err)
		}

		var unmarshaled CrossProviderAggregation
		if err = json.Unmarshal(data, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal CrossProviderAggregation: %v", err)
		}

		if unmarshaled.Period != agg.Period {
			t.Errorf("Period = %q, want %q", unmarshaled.Period, agg.Period)
		}
		if unmarshaled.Total != agg.Total {
			t.Errorf("Total = %f, want %f", unmarshaled.Total, agg.Total)
		}
		if len(unmarshaled.Providers) != len(agg.Providers) {
			t.Errorf(
				"Providers length = %d, want %d",
				len(unmarshaled.Providers),
				len(agg.Providers),
			)
		}
	})

	t.Run("AggregatedResults marshals to JSON", func(t *testing.T) {
		aggregated := &AggregatedResults{
			Summary: CostSummary{
				TotalMonthly: 100.0,
				TotalHourly:  0.137,
				Currency:     "USD",
				ByProvider:   map[string]float64{"aws": 100.0},
				ByService:    map[string]float64{"ec2": 100.0},
				ByAdapter:    map[string]float64{"kubecost": 100.0},
			},
			Resources: []CostResult{
				{ResourceType: "aws:ec2:Instance", Monthly: 100.0},
			},
		}

		data, err := json.Marshal(aggregated)
		if err != nil {
			t.Fatalf("Failed to marshal AggregatedResults: %v", err)
		}

		var unmarshaled AggregatedResults
		if err = json.Unmarshal(data, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal AggregatedResults: %v", err)
		}

		if unmarshaled.Summary.TotalMonthly != aggregated.Summary.TotalMonthly {
			t.Errorf(
				"TotalMonthly = %f, want %f",
				unmarshaled.Summary.TotalMonthly,
				aggregated.Summary.TotalMonthly,
			)
		}
	})
}

// TestRenderingWithNilResults tests edge cases with nil results.
func TestRenderingWithNilResults(t *testing.T) {
	t.Run("RenderResults with nil", func(t *testing.T) {
		err := RenderResults(io.Discard, OutputJSON, nil)
		if err != nil {
			t.Errorf("RenderResults with nil should not error: %v", err)
		}
	})

	t.Run("RenderActualCostResults with nil", func(t *testing.T) {
		err := RenderActualCostResults(io.Discard, OutputJSON, nil)
		if err != nil {
			t.Errorf("RenderActualCostResults with nil should not error: %v", err)
		}
	})

	t.Run("RenderCrossProviderAggregation with nil", func(t *testing.T) {
		err := RenderCrossProviderAggregation(io.Discard, OutputJSON, nil, GroupByDaily)
		if err != nil {
			t.Errorf("RenderCrossProviderAggregation with nil should not error: %v", err)
		}
	})
}

// TestRenderingWithEmptyResults tests edge cases with empty results.
func TestRenderingWithEmptyResults(t *testing.T) {
	t.Run("RenderResults with empty slice", func(t *testing.T) {
		err := RenderResults(io.Discard, OutputTable, []CostResult{})
		if err != nil {
			t.Errorf("RenderResults with empty should not error: %v", err)
		}
	})

	t.Run("RenderActualCostResults with empty slice", func(t *testing.T) {
		err := RenderActualCostResults(io.Discard, OutputTable, []CostResult{})
		if err != nil {
			t.Errorf("RenderActualCostResults with empty should not error: %v", err)
		}
	})

	t.Run("RenderCrossProviderAggregation with empty slice", func(t *testing.T) {
		err := RenderCrossProviderAggregation(
			io.Discard,
			OutputTable,
			[]CrossProviderAggregation{},
			GroupByDaily,
		)
		if err != nil {
			t.Errorf("RenderCrossProviderAggregation with empty should not error: %v", err)
		}
	})
}

// TestMultiProviderRendering tests rendering with multiple providers.
func TestMultiProviderRendering(t *testing.T) {
	results := []CostResult{
		{
			ResourceType: "aws:ec2:Instance",
			ResourceID:   "i-123",
			Adapter:      "kubecost",
			Currency:     "USD",
			Monthly:      50.0,
		},
		{
			ResourceType: "azure:compute:VirtualMachine",
			ResourceID:   "vm-456",
			Adapter:      "kubecost",
			Currency:     "USD",
			Monthly:      75.0,
		},
		{
			ResourceType: "gcp:compute:Instance",
			ResourceID:   "instance-789",
			Adapter:      "kubecost",
			Currency:     "USD",
			Monthly:      100.0,
		},
	}

	t.Run("multi-provider table rendering", func(t *testing.T) {
		err := RenderResults(io.Discard, OutputTable, results)
		if err != nil {
			t.Errorf("Multi-provider table rendering failed: %v", err)
		}
	})

	t.Run("multi-provider json rendering", func(t *testing.T) {
		err := RenderResults(io.Discard, OutputJSON, results)
		if err != nil {
			t.Errorf("Multi-provider JSON rendering failed: %v", err)
		}
	})

	t.Run("multi-provider ndjson rendering", func(t *testing.T) {
		err := RenderResults(io.Discard, OutputNDJSON, results)
		if err != nil {
			t.Errorf("Multi-provider NDJSON rendering failed: %v", err)
		}
	})
}

// TestLongResourceNameHandling tests truncation behavior.
func TestLongResourceNameHandling(t *testing.T) {
	longID := "i-" + string(make([]byte, 100))
	for i := range longID[2:] {
		longID = longID[:2+i] + "a" + longID[2+i+1:]
	}

	results := []CostResult{
		{
			ResourceType: "aws:ec2:Instance",
			ResourceID:   longID,
			Adapter:      "kubecost",
			Currency:     "USD",
			Monthly:      10.0,
		},
	}

	err := RenderResults(io.Discard, OutputTable, results)
	if err != nil {
		t.Errorf("Long resource name rendering failed: %v", err)
	}
}
