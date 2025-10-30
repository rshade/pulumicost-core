package engine

import (
	"testing"
	"time"
)

// Test GroupBy validation.
func TestGroupBy_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		groupBy  GroupBy
		expected bool
	}{
		{"valid resource", GroupByResource, true},
		{"valid type", GroupByType, true},
		{"valid provider", GroupByProvider, true},
		{"valid date", GroupByDate, true},
		{"valid daily", GroupByDaily, true},
		{"valid monthly", GroupByMonthly, true},
		{"valid none", GroupByNone, true},
		{"invalid empty string not GroupByNone", GroupBy(""), true}, // Empty string is GroupByNone
		{"invalid random", GroupBy("random"), false},
		{"invalid uppercase", GroupBy("DAILY"), false},
		{"invalid mixed case", GroupBy("Daily"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.groupBy.IsValid()
			if got != tt.expected {
				t.Errorf("IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Test time-based grouping detection.
func TestGroupBy_IsTimeBasedGrouping(t *testing.T) {
	tests := []struct {
		name     string
		groupBy  GroupBy
		expected bool
	}{
		{"daily is time-based", GroupByDaily, true},
		{"monthly is time-based", GroupByMonthly, true},
		{"resource is not time-based", GroupByResource, false},
		{"type is not time-based", GroupByType, false},
		{"provider is not time-based", GroupByProvider, false},
		{"date is not time-based", GroupByDate, false},
		{"none is not time-based", GroupByNone, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.groupBy.IsTimeBasedGrouping()
			if got != tt.expected {
				t.Errorf("IsTimeBasedGrouping() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Test String() method.
func TestGroupBy_String(t *testing.T) {
	tests := []struct {
		name     string
		groupBy  GroupBy
		expected string
	}{
		{"resource", GroupByResource, "resource"},
		{"type", GroupByType, "type"},
		{"provider", GroupByProvider, "provider"},
		{"date", GroupByDate, "date"},
		{"daily", GroupByDaily, "daily"},
		{"monthly", GroupByMonthly, "monthly"},
		{"none", GroupByNone, ""},
		{"empty", GroupBy(""), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.groupBy.String()
			if got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// Test ResourceDescriptor creation.
func TestResourceDescriptor(t *testing.T) {
	rd := ResourceDescriptor{
		Type:     "aws:ec2:Instance",
		ID:       "i-123456",
		Provider: "aws",
		Properties: map[string]interface{}{
			"instanceType": "t3.micro",
			"region":       "us-east-1",
		},
	}

	if rd.Type != "aws:ec2:Instance" {
		t.Errorf("Type = %q, want %q", rd.Type, "aws:ec2:Instance")
	}
	if rd.ID != "i-123456" {
		t.Errorf("ID = %q, want %q", rd.ID, "i-123456")
	}
	if rd.Provider != "aws" {
		t.Errorf("Provider = %q, want %q", rd.Provider, "aws")
	}
	if len(rd.Properties) != 2 {
		t.Errorf("Properties length = %d, want 2", len(rd.Properties))
	}

	// Verify properties
	if instanceType, ok := rd.Properties["instanceType"]; !ok || instanceType != "t3.micro" {
		t.Errorf("Properties[instanceType] = %v, want t3.micro", instanceType)
	}
	if region, ok := rd.Properties["region"]; !ok || region != "us-east-1" {
		t.Errorf("Properties[region] = %v, want us-east-1", region)
	}
}

// Test CostResult creation and defaults.
func TestCostResult(t *testing.T) {
	now := time.Now()
	endDate := now.AddDate(0, 1, 0)

	cr := CostResult{
		ResourceType: "aws:ec2:Instance",
		ResourceID:   "i-123456",
		Adapter:      "kubecost",
		Currency:     "USD",
		Monthly:      100.50,
		Hourly:       0.1377,
		TotalCost:    100.50,
		Notes:        "Test cost result",
		StartDate:    now,
		EndDate:      endDate,
		Breakdown: map[string]float64{
			"compute": 80.00,
			"storage": 20.50,
		},
		DailyCosts: []float64{3.35, 3.35, 3.35},
		CostPeriod: "monthly",
	}

	// Verify all fields
	if cr.ResourceType != "aws:ec2:Instance" {
		t.Errorf("ResourceType = %q, want aws:ec2:Instance", cr.ResourceType)
	}
	if cr.ResourceID != "i-123456" {
		t.Errorf("ResourceID = %q, want i-123456", cr.ResourceID)
	}
	if cr.Adapter != "kubecost" {
		t.Errorf("Adapter = %q, want kubecost", cr.Adapter)
	}
	if cr.Currency != "USD" {
		t.Errorf("Currency = %q, want USD", cr.Currency)
	}
	if cr.Monthly != 100.50 {
		t.Errorf("Monthly = %f, want 100.50", cr.Monthly)
	}
	if cr.Hourly != 0.1377 {
		t.Errorf("Hourly = %f, want 0.1377", cr.Hourly)
	}
	if cr.TotalCost != 100.50 {
		t.Errorf("TotalCost = %f, want 100.50", cr.TotalCost)
	}
	if len(cr.Breakdown) != 2 {
		t.Errorf("Breakdown length = %d, want 2", len(cr.Breakdown))
	}
	if len(cr.DailyCosts) != 3 {
		t.Errorf("DailyCosts length = %d, want 3", len(cr.DailyCosts))
	}
	if cr.CostPeriod != "monthly" {
		t.Errorf("CostPeriod = %q, want monthly", cr.CostPeriod)
	}
	if cr.Notes != "Test cost result" {
		t.Errorf("Notes = %q, want 'Test cost result'", cr.Notes)
	}
	if cr.StartDate.IsZero() {
		t.Error("StartDate should not be zero")
	}
	if cr.EndDate.IsZero() {
		t.Error("EndDate should not be zero")
	}

	// Verify breakdown
	if cr.Breakdown["compute"] != 80.00 {
		t.Errorf("Breakdown[compute] = %f, want 80.00", cr.Breakdown["compute"])
	}
	if cr.Breakdown["storage"] != 20.50 {
		t.Errorf("Breakdown[storage] = %f, want 20.50", cr.Breakdown["storage"])
	}
}

// Test CrossProviderAggregation.
func TestCrossProviderAggregation(t *testing.T) {
	agg := CrossProviderAggregation{
		Period: "2024-01-15",
		Providers: map[string]float64{
			"aws":   250.00,
			"azure": 180.50,
			"gcp":   95.25,
		},
		Total:    525.75,
		Currency: "USD",
	}

	if agg.Period != "2024-01-15" {
		t.Errorf("Period = %q, want 2024-01-15", agg.Period)
	}
	if agg.Total != 525.75 {
		t.Errorf("Total = %f, want 525.75", agg.Total)
	}
	if agg.Currency != "USD" {
		t.Errorf("Currency = %q, want USD", agg.Currency)
	}
	if len(agg.Providers) != 3 {
		t.Errorf("Providers length = %d, want 3", len(agg.Providers))
	}

	// Verify provider costs
	if agg.Providers["aws"] != 250.00 {
		t.Errorf("Providers[aws] = %f, want 250.00", agg.Providers["aws"])
	}
	if agg.Providers["azure"] != 180.50 {
		t.Errorf("Providers[azure] = %f, want 180.50", agg.Providers["azure"])
	}
	if agg.Providers["gcp"] != 95.25 {
		t.Errorf("Providers[gcp] = %f, want 95.25", agg.Providers["gcp"])
	}

	// Verify total matches sum
	var sum float64
	for _, cost := range agg.Providers {
		sum += cost
	}
	if sum != agg.Total {
		t.Errorf("Provider sum %f != Total %f", sum, agg.Total)
	}
}

// Test error types.
func TestErrorTypes(t *testing.T) {
	errors := []struct {
		name string
		err  error
	}{
		{"ErrNoCostData", ErrNoCostData},
		{"ErrMixedCurrencies", ErrMixedCurrencies},
		{"ErrInvalidGroupBy", ErrInvalidGroupBy},
		{"ErrEmptyResults", ErrEmptyResults},
		{"ErrInvalidDateRange", ErrInvalidDateRange},
	}

	for _, tt := range errors {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Error("Error should not be nil")
			}
			if tt.err.Error() == "" {
				t.Errorf("Error message should not be empty for %v", tt.err)
			}
		})
	}
}
