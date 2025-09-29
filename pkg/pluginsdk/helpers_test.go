package pluginsdk_test

import (
	"context"
	"testing"

	"github.com/rshade/pulumicost-core/pkg/pluginsdk"
	pbc "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
)

func TestResourceMatcher(t *testing.T) {
	matcher := pluginsdk.NewResourceMatcher()

	// Add supported providers and resource types
	matcher.AddProvider("aws")
	matcher.AddResourceType("aws:ec2:Instance")

	testCases := []struct {
		name     string
		resource *pbc.ResourceDescriptor
		expected bool
	}{
		{
			name: "supported provider and type",
			resource: &pbc.ResourceDescriptor{
				Provider:     "aws",
				ResourceType: "aws:ec2:Instance",
			},
			expected: true,
		},
		{
			name: "unsupported provider",
			resource: &pbc.ResourceDescriptor{
				Provider:     "azure",
				ResourceType: "azure:compute:VirtualMachine",
			},
			expected: false,
		},
		{
			name: "supported provider, unsupported type",
			resource: &pbc.ResourceDescriptor{
				Provider:     "aws",
				ResourceType: "aws:s3:Bucket",
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := matcher.Supports(tc.resource)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for resource %v", tc.expected, result, tc.resource)
			}
		})
	}
}

func TestResourceMatcherNoFilters(t *testing.T) {
	// Empty matcher should support everything
	matcher := pluginsdk.NewResourceMatcher()

	resource := &pbc.ResourceDescriptor{
		Provider:     "any",
		ResourceType: "any:resource:Type",
	}

	if !matcher.Supports(resource) {
		t.Errorf("Empty matcher should support all resources")
	}
}

func TestCostCalculator(t *testing.T) {
	calc := pluginsdk.NewCostCalculator()

	// Test hourly to monthly conversion
	hourly := 0.10
	monthly := calc.HourlyToMonthly(hourly)
	expected := 73.0 // 0.10 * 730
	if monthly != expected {
		t.Errorf("Expected monthly cost %f, got %f", expected, monthly)
	}

	// Test monthly to hourly conversion
	monthlyInput := 146.0
	hourlyResult := calc.MonthlyToHourly(monthlyInput)
	expectedHourly := 0.2 // 146.0 / 730
	if hourlyResult != expectedHourly {
		t.Errorf("Expected hourly cost %f, got %f", expectedHourly, hourlyResult)
	}
}

func TestCostCalculatorResponses(t *testing.T) {
	calc := pluginsdk.NewCostCalculator()

	// Test projected cost response
	resp := calc.CreateProjectedCostResponse("USD", 0.05, "Test billing detail")

	if resp.GetCurrency() != "USD" {
		t.Errorf("Expected currency USD, got %s", resp.GetCurrency())
	}

	if resp.GetUnitPrice() != 0.05 {
		t.Errorf("Expected unit price 0.05, got %f", resp.GetUnitPrice())
	}

	if resp.GetCostPerMonth() != 36.5 { // 0.05 * 730
		t.Errorf("Expected cost per month 36.5, got %f", resp.GetCostPerMonth())
	}

	if resp.GetBillingDetail() != "Test billing detail" {
		t.Errorf("Expected billing detail 'Test billing detail', got %s", resp.GetBillingDetail())
	}

	// Test actual cost response
	results := []*pbc.ActualCostResult{
		{Cost: 10.0, Source: "test"},
	}
	actualResp := calc.CreateActualCostResponse(results)

	if len(actualResp.GetResults()) != 1 {
		t.Errorf("Expected 1 result, got %d", len(actualResp.GetResults()))
	}

	if actualResp.GetResults()[0].GetCost() != 10.0 {
		t.Errorf("Expected cost 10.0, got %f", actualResp.GetResults()[0].GetCost())
	}
}

func TestErrorFunctions(t *testing.T) {
	resource := &pbc.ResourceDescriptor{
		Provider:     "test",
		ResourceType: "test:resource:Type",
	}

	// Test NotSupportedError
	err := pluginsdk.NotSupportedError(resource)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	expectedMsg := "resource type test:resource:Type from provider test is not supported"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}

	// Test NoDataError
	err = pluginsdk.NoDataError("test-resource-id")
	if err == nil {
		t.Error("Expected error, got nil")
	}

	expectedMsg = "no cost data available for resource test-resource-id"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestBasePlugin(t *testing.T) {
	plugin := pluginsdk.NewBasePlugin("test-plugin")

	// Test name
	if plugin.Name() != "test-plugin" {
		t.Errorf("Expected name 'test-plugin', got %s", plugin.Name())
	}

	// Test matcher access
	matcher := plugin.Matcher()
	if matcher == nil {
		t.Error("Expected matcher, got nil")
	}

	// Test calculator access
	calc := plugin.Calculator()
	if calc == nil {
		t.Error("Expected calculator, got nil")
	}

	// Test default implementations
	ctx := context.Background()

	// Test default GetProjectedCost (should return not supported error)
	req := &pbc.GetProjectedCostRequest{
		Resource: &pbc.ResourceDescriptor{
			Provider:     "test",
			ResourceType: "test:resource:Type",
		},
	}

	_, err := plugin.GetProjectedCost(ctx, req)
	if err == nil {
		t.Error("Expected not supported error, got nil")
	}

	// Test default GetActualCost (should return no data error)
	actualReq := &pbc.GetActualCostRequest{
		ResourceId: "test-resource-id",
	}

	_, err = plugin.GetActualCost(ctx, actualReq)
	if err == nil {
		t.Error("Expected no data error, got nil")
	}
}
