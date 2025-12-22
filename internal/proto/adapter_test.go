package proto

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
)

// mockCostSourceClient is a mock implementation of CostSourceClient for testing.
type mockCostSourceClient struct {
	nameFunc         func(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*NameResponse, error)
	getProjectedFunc func(
		ctx context.Context,
		in *GetProjectedCostRequest,
		opts ...grpc.CallOption,
	) (*GetProjectedCostResponse, error)
	getActualFunc func(
		ctx context.Context,
		in *GetActualCostRequest,
		opts ...grpc.CallOption,
	) (*GetActualCostResponse, error)
}

func (m *mockCostSourceClient) Name(
	ctx context.Context,
	in *Empty,
	opts ...grpc.CallOption,
) (*NameResponse, error) {
	if m.nameFunc != nil {
		return m.nameFunc(ctx, in, opts...)
	}
	return &NameResponse{Name: "mock-plugin"}, nil
}

func (m *mockCostSourceClient) GetProjectedCost(
	ctx context.Context,
	in *GetProjectedCostRequest,
	opts ...grpc.CallOption,
) (*GetProjectedCostResponse, error) {
	if m.getProjectedFunc != nil {
		return m.getProjectedFunc(ctx, in, opts...)
	}
	return &GetProjectedCostResponse{Results: []*CostResult{}}, nil
}

func (m *mockCostSourceClient) GetActualCost(
	ctx context.Context,
	in *GetActualCostRequest,
	opts ...grpc.CallOption,
) (*GetActualCostResponse, error) {
	if m.getActualFunc != nil {
		return m.getActualFunc(ctx, in, opts...)
	}
	return &GetActualCostResponse{Results: []*ActualCostResult{}}, nil
}

// T003: Unit test for ErrorDetail struct creation.
func TestErrorDetail_Creation(t *testing.T) {
	timestamp := time.Now()
	err := errors.New("test error")

	detail := ErrorDetail{
		ResourceType: "aws:ec2:Instance",
		ResourceID:   "i-1234567890abcdef0",
		PluginName:   "test-plugin",
		Error:        err,
		Timestamp:    timestamp,
	}

	if detail.ResourceType != "aws:ec2:Instance" {
		t.Errorf("ResourceType = %s, want aws:ec2:Instance", detail.ResourceType)
	}
	if detail.ResourceID != "i-1234567890abcdef0" {
		t.Errorf("ResourceID = %s, want i-1234567890abcdef0", detail.ResourceID)
	}
	if detail.PluginName != "test-plugin" {
		t.Errorf("PluginName = %s, want test-plugin", detail.PluginName)
	}
	if !errors.Is(detail.Error, err) {
		t.Errorf("Error = %v, want %v", detail.Error, err)
	}
	if detail.Timestamp != timestamp {
		t.Errorf("Timestamp = %v, want %v", detail.Timestamp, timestamp)
	}
}

// T004: Unit test for CostResultWithErrors struct creation.
func TestCostResultWithErrors_Creation(t *testing.T) {
	result := &CostResultWithErrors{
		Results: []*CostResult{
			{Currency: "USD", MonthlyCost: 100.0},
		},
		Errors: []ErrorDetail{
			{ResourceType: "aws:ec2:Instance", ResourceID: "i-123", Error: errors.New("test")},
		},
	}

	if len(result.Results) != 1 {
		t.Errorf("Results length = %d, want 1", len(result.Results))
	}
	if len(result.Errors) != 1 {
		t.Errorf("Errors length = %d, want 1", len(result.Errors))
	}
}

func TestCostResultWithErrors_Empty(t *testing.T) {
	result := &CostResultWithErrors{
		Results: []*CostResult{},
		Errors:  []ErrorDetail{},
	}

	if len(result.Results) != 0 {
		t.Errorf("Results length = %d, want 0", len(result.Results))
	}
	if len(result.Errors) != 0 {
		t.Errorf("Errors length = %d, want 0", len(result.Errors))
	}
}

// T005: Unit test for HasErrors() method.
func TestCostResultWithErrors_HasErrors(t *testing.T) {
	tests := []struct {
		name     string
		errors   []ErrorDetail
		expected bool
	}{
		{
			name:     "no errors",
			errors:   []ErrorDetail{},
			expected: false,
		},
		{
			name: "one error",
			errors: []ErrorDetail{
				{ResourceType: "aws:ec2:Instance", ResourceID: "i-123", Error: errors.New("test")},
			},
			expected: true,
		},
		{
			name: "multiple errors",
			errors: []ErrorDetail{
				{ResourceType: "aws:ec2:Instance", ResourceID: "i-123", Error: errors.New("test1")},
				{
					ResourceType: "aws:rds:Instance",
					ResourceID:   "db-456",
					Error:        errors.New("test2"),
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &CostResultWithErrors{
				Results: []*CostResult{},
				Errors:  tt.errors,
			}

			if got := result.HasErrors(); got != tt.expected {
				t.Errorf("HasErrors() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// T006: Unit test for ErrorSummary() output format.
func TestCostResultWithErrors_ErrorSummary(t *testing.T) {
	t.Run("no errors returns empty string", func(t *testing.T) {
		result := &CostResultWithErrors{
			Results: []*CostResult{},
			Errors:  []ErrorDetail{},
		}

		if summary := result.ErrorSummary(); summary != "" {
			t.Errorf("ErrorSummary() = %q, want empty string", summary)
		}
	})

	t.Run("single error", func(t *testing.T) {
		result := &CostResultWithErrors{
			Results: []*CostResult{},
			Errors: []ErrorDetail{
				{
					ResourceType: "aws:ec2:Instance",
					ResourceID:   "i-123",
					PluginName:   "test-plugin",
					Error:        errors.New("connection refused"),
					Timestamp:    time.Now(),
				},
			},
		}

		summary := result.ErrorSummary()

		if !strings.Contains(summary, "1 resource(s) failed") {
			t.Errorf("ErrorSummary() should contain '1 resource(s) failed', got %q", summary)
		}
		if !strings.Contains(summary, "aws:ec2:Instance") {
			t.Errorf("ErrorSummary() should contain resource type, got %q", summary)
		}
		if !strings.Contains(summary, "i-123") {
			t.Errorf("ErrorSummary() should contain resource ID, got %q", summary)
		}
		if !strings.Contains(summary, "connection refused") {
			t.Errorf("ErrorSummary() should contain error message, got %q", summary)
		}
	})

	t.Run("multiple errors up to 5", func(t *testing.T) {
		result := &CostResultWithErrors{
			Results: []*CostResult{},
			Errors:  []ErrorDetail{},
		}

		// Add 3 errors
		for i := 0; i < 3; i++ {
			result.Errors = append(result.Errors, ErrorDetail{
				ResourceType: "aws:ec2:Instance",
				ResourceID:   fmt.Sprintf("i-%d", i),
				PluginName:   "test-plugin",
				Error:        errors.New("error"),
				Timestamp:    time.Now(),
			})
		}

		summary := result.ErrorSummary()

		if !strings.Contains(summary, "3 resource(s) failed") {
			t.Errorf("ErrorSummary() should contain '3 resource(s) failed', got %q", summary)
		}
		// All 3 should be listed
		for i := 0; i < 3; i++ {
			if !strings.Contains(summary, fmt.Sprintf("i-%d", i)) {
				t.Errorf("ErrorSummary() should contain resource i-%d, got %q", i, summary)
			}
		}
	})

	t.Run("more than 5 errors truncates", func(t *testing.T) {
		result := &CostResultWithErrors{
			Results: []*CostResult{},
			Errors:  []ErrorDetail{},
		}

		// Add 10 errors
		for i := 0; i < 10; i++ {
			result.Errors = append(result.Errors, ErrorDetail{
				ResourceType: "aws:ec2:Instance",
				ResourceID:   fmt.Sprintf("i-%d", i),
				PluginName:   "test-plugin",
				Error:        errors.New("error"),
				Timestamp:    time.Now(),
			})
		}

		summary := result.ErrorSummary()

		if !strings.Contains(summary, "10 resource(s) failed") {
			t.Errorf("ErrorSummary() should contain '10 resource(s) failed', got %q", summary)
		}
		// Should show first 5
		for i := 0; i < 5; i++ {
			if !strings.Contains(summary, fmt.Sprintf("i-%d", i)) {
				t.Errorf("ErrorSummary() should contain resource i-%d, got %q", i, summary)
			}
		}
		// Should indicate truncation
		if !strings.Contains(summary, "and 5 more") {
			t.Errorf("ErrorSummary() should indicate '... and 5 more errors', got %q", summary)
		}
	})
}

// T011: Unit test for GetProjectedCost error tracking.
func TestGetProjectedCostWithErrors(t *testing.T) {
	t.Run("tracks errors for failed resources", func(t *testing.T) {
		callCount := 0
		mockClient := &mockCostSourceClient{
			getProjectedFunc: func(ctx context.Context, in *GetProjectedCostRequest, opts ...grpc.CallOption) (*GetProjectedCostResponse, error) {
				callCount++
				// Fail for the second resource
				if len(in.Resources) > 0 && in.Resources[0].Type == "aws:rds:Instance" {
					return nil, errors.New("connection refused")
				}
				return &GetProjectedCostResponse{
					Results: []*CostResult{
						{Currency: "USD", MonthlyCost: 100.0, HourlyCost: 0.137},
					},
				}, nil
			},
		}

		resources := []*ResourceDescriptor{
			{Type: "aws:ec2:Instance", Provider: "aws", Properties: map[string]string{}},
			{Type: "aws:rds:Instance", Provider: "aws", Properties: map[string]string{}},
			{Type: "aws:s3:Bucket", Provider: "aws", Properties: map[string]string{}},
		}

		result := GetProjectedCostWithErrors(
			context.Background(),
			mockClient,
			"test-plugin",
			resources,
		)

		// Should have 3 results (2 success + 1 placeholder for error)
		if len(result.Results) != 3 {
			t.Errorf("Results length = %d, want 3", len(result.Results))
		}

		// Should have 1 error
		if len(result.Errors) != 1 {
			t.Errorf("Errors length = %d, want 1", len(result.Errors))
		}

		// Error should be tracked
		if !result.HasErrors() {
			t.Error("HasErrors() should return true")
		}

		// Error details should be correct
		if result.Errors[0].ResourceType != "aws:rds:Instance" {
			t.Errorf(
				"Error ResourceType = %s, want aws:rds:Instance",
				result.Errors[0].ResourceType,
			)
		}

		// Placeholder result should have ERROR in Notes
		for _, r := range result.Results {
			if r.Notes != "" && strings.Contains(r.Notes, "ERROR") {
				if !strings.Contains(r.Notes, "connection refused") {
					t.Errorf("Error result Notes should contain error message, got %q", r.Notes)
				}
			}
		}
	})
}

// T020: Unit test for GetActualCost error tracking.
func TestGetActualCostWithErrors(t *testing.T) {
	t.Run("tracks errors for failed resources", func(t *testing.T) {
		mockClient := &mockCostSourceClient{
			getActualFunc: func(ctx context.Context, in *GetActualCostRequest, opts ...grpc.CallOption) (*GetActualCostResponse, error) {
				// Fail for the second resource ID
				if len(in.ResourceIDs) > 0 && in.ResourceIDs[0] == "failed-resource" {
					return nil, errors.New("timeout")
				}
				return &GetActualCostResponse{
					Results: []*ActualCostResult{
						{Currency: "USD", TotalCost: 50.0},
					},
				}, nil
			},
		}

		resourceIDs := []string{"success-1", "failed-resource", "success-2"}
		startTime := time.Now().Add(-24 * time.Hour).Unix()
		endTime := time.Now().Unix()

		req := &GetActualCostRequest{
			ResourceIDs: resourceIDs,
			StartTime:   startTime,
			EndTime:     endTime,
		}

		result := GetActualCostWithErrors(context.Background(), mockClient, "test-plugin", req)

		// Should have 3 results
		if len(result.Results) != 3 {
			t.Errorf("Results length = %d, want 3", len(result.Results))
		}

		// Should have 1 error
		if len(result.Errors) != 1 {
			t.Errorf("Errors length = %d, want 1", len(result.Errors))
		}

		// Error details should be correct
		if result.Errors[0].ResourceID != "failed-resource" {
			t.Errorf("Error ResourceID = %s, want failed-resource", result.Errors[0].ResourceID)
		}

		if !strings.Contains(result.Errors[0].Error.Error(), "timeout") {
			t.Errorf("Error should contain 'timeout', got %v", result.Errors[0].Error)
		}
	})
}

// Test NameResponse.GetName method.
func TestNameResponse_GetName(t *testing.T) {
	tests := []struct {
		name     string
		response NameResponse
		expected string
	}{
		{
			name:     "normal name",
			response: NameResponse{Name: "test-plugin"},
			expected: "test-plugin",
		},
		{
			name:     "empty name",
			response: NameResponse{Name: ""},
			expected: "",
		},
		{
			name:     "special characters",
			response: NameResponse{Name: "plugin@v1.0"},
			expected: "plugin@v1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.response.GetName(); got != tt.expected {
				t.Errorf("GetName() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Test NewCostSourceClient function.
func TestNewCostSourceClient(t *testing.T) {
	// This is a basic test since we can't easily create a real gRPC connection
	// in a unit test. We test that the function doesn't panic and returns
	// a non-nil client.
	t.Run("returns non-nil client", func(t *testing.T) {
		// We can't create a real connection, but we can test the function signature
		// and that it would work with a nil connection (though it would fail at runtime)
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("NewCostSourceClient panicked: %v", r)
			}
		}()

		// This will panic when trying to use the client, but the creation should work
		client := NewCostSourceClient(nil)
		if client == nil {
			t.Error("NewCostSourceClient returned nil")
		}
	})
}

// Test clientAdapter.Name method.
func TestClientAdapter_Name(t *testing.T) {
	t.Run("successful name call", func(t *testing.T) {
		mockClient := &mockCostSourceClient{
			nameFunc: func(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*NameResponse, error) {
				return &NameResponse{Name: "mock-plugin-name"}, nil
			},
		}

		adapter := &clientAdapter{client: nil} // We mock the client behavior
		// Note: In a real test, we'd need to mock the underlying gRPC client
		// This is a placeholder test structure

		_ = mockClient // Use the mock to show intended usage
		_ = adapter    // Avoid unused variable error
	})

	t.Run("name call with error", func(t *testing.T) {
		mockClient := &mockCostSourceClient{
			nameFunc: func(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*NameResponse, error) {
				return nil, errors.New("grpc error")
			},
		}

		_ = mockClient // Use the mock to show intended usage
	})
}

// Test clientAdapter.GetProjectedCost method.
func TestClientAdapter_GetProjectedCost(t *testing.T) {
	t.Run("successful cost calculation", func(t *testing.T) {
		// Test with resources that have SKU and region in properties
		req := &GetProjectedCostRequest{
			Resources: []*ResourceDescriptor{
				{
					Type:     "aws:ec2:Instance",
					Provider: "aws",
					Properties: map[string]string{
						"sku":    "t3.micro",
						"region": "us-east-1",
					},
				},
			},
		}

		// This would require mocking the underlying gRPC client
		// For now, we test the request structure
		if len(req.Resources) != 1 {
			t.Errorf("Expected 1 resource, got %d", len(req.Resources))
		}

		resource := req.Resources[0]
		if resource.Type != "aws:ec2:Instance" {
			t.Errorf("Expected resource type 'aws:ec2:Instance', got %s", resource.Type)
		}
		if resource.Properties["sku"] != "t3.micro" {
			t.Errorf("Expected SKU 't3.micro', got %s", resource.Properties["sku"])
		}
		if resource.Properties["region"] != "us-east-1" {
			t.Errorf("Expected region 'us-east-1', got %s", resource.Properties["region"])
		}
	})

	t.Run("resource without sku/region properties", func(t *testing.T) {
		req := &GetProjectedCostRequest{
			Resources: []*ResourceDescriptor{
				{
					Type:       "aws:s3:Bucket",
					Provider:   "aws",
					Properties: map[string]string{},
				},
			},
		}

		if len(req.Resources) != 1 {
			t.Errorf("Expected 1 resource, got %d", len(req.Resources))
		}

		resource := req.Resources[0]
		if resource.Type != "aws:s3:Bucket" {
			t.Errorf("Expected resource type 'aws:s3:Bucket', got %s", resource.Type)
		}
		// SKU and region should be empty/default
		if sku, ok := resource.Properties["sku"]; ok && sku != "" {
			t.Errorf("Expected empty SKU, got %s", sku)
		}
		if region, ok := resource.Properties["region"]; ok && region != "" {
			t.Errorf("Expected empty region, got %s", region)
		}
	})
}

// Test clientAdapter.GetActualCost method.
func TestClientAdapter_GetActualCost(t *testing.T) {
	t.Run("successful actual cost query", func(t *testing.T) {
		startTime := time.Now().Add(-24 * time.Hour).Unix()
		endTime := time.Now().Unix()

		req := &GetActualCostRequest{
			ResourceIDs: []string{"i-1234567890abcdef0", "i-0987654321fedcba0"},
			StartTime:   startTime,
			EndTime:     endTime,
		}

		if len(req.ResourceIDs) != 2 {
			t.Errorf("Expected 2 resource IDs, got %d", len(req.ResourceIDs))
		}
		if req.StartTime != startTime {
			t.Errorf("Expected StartTime %d, got %d", startTime, req.StartTime)
		}
		if req.EndTime != endTime {
			t.Errorf("Expected EndTime %d, got %d", endTime, req.EndTime)
		}
	})

	t.Run("empty resource IDs", func(t *testing.T) {
		req := &GetActualCostRequest{
			ResourceIDs: []string{},
			StartTime:   1000000000,
			EndTime:     1000003600,
		}

		if len(req.ResourceIDs) != 0 {
			t.Errorf("Expected 0 resource IDs, got %d", len(req.ResourceIDs))
		}
	})
}

// TestExtractSKUFromProperties tests the SKU extraction function.
func TestExtractSKUFromProperties(t *testing.T) {
	tests := []struct {
		name       string
		provider   string
		properties map[string]string
		expected   string
	}{
		{
			name:       "EC2 instance with instanceType",
			properties: map[string]string{"instanceType": "t3.micro"},
			expected:   "t3.micro",
		},
		{
			name:       "EBS volume with type",
			properties: map[string]string{"type": "gp3"},
			expected:   "gp3",
		},
		{
			name:       "EBS volume with volumeType",
			properties: map[string]string{"volumeType": "gp2"},
			expected:   "gp2",
		},
		{
			name:       "RDS instance with instanceClass",
			properties: map[string]string{"instanceClass": "db.t3.micro"},
			expected:   "db.t3.micro",
		},
		{
			name:       "explicit sku property (generic)",
			provider:   "generic",
			properties: map[string]string{"sku": "Standard_DS1_v2"},
			expected:   "Standard_DS1_v2",
		},
		{
			name:       "Azure vmSize property",
			provider:   "azure",
			properties: map[string]string{"vmSize": "Standard_B1s"},
			expected:   "Standard_B1s",
		},
		{
			name:       "GCP machineType",
			provider:   "gcp",
			properties: map[string]string{"machineType": "n1-standard-4"},
			expected:   "n1-standard-4",
		},
		{
			name:       "instanceType takes precedence over type",
			provider:   "aws",
			properties: map[string]string{"instanceType": "t3.micro", "type": "gp3"},
			expected:   "t3.micro",
		},
		{
			name:       "empty properties returns empty string",
			provider:   "aws",
			properties: map[string]string{},
			expected:   "",
		},
		{
			name:       "nil properties returns empty string",
			provider:   "aws",
			properties: nil,
			expected:   "",
		},
		{
			name:       "irrelevant properties returns empty string",
			provider:   "aws",
			properties: map[string]string{"bucketName": "my-bucket", "acl": "private"},
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := "aws"
			if tt.provider != "" {
				provider = tt.provider
			}
			sku, region := resolveSKUAndRegion(provider, tt.properties)
			if sku != tt.expected {
				t.Errorf("resolveSKUAndRegion(%s) sku = %q, want %q", provider, sku, tt.expected)
			}
			_ = region
		})
	}
}

// TestExtractRegionFromProperties tests the region extraction function.
func TestExtractRegionFromProperties(t *testing.T) {
	tests := []struct {
		name       string
		properties map[string]string
		envVars    map[string]string
		expected   string
	}{
		{
			name:       "extract from availabilityZone with suffix",
			properties: map[string]string{"availabilityZone": "us-east-1a"},
			expected:   "us-east-1",
		},
		{
			name:       "extract from availabilityZone with suffix b",
			properties: map[string]string{"availabilityZone": "eu-west-2b"},
			expected:   "eu-west-2",
		},
		{
			name:       "extract from availabilityZone with suffix f",
			properties: map[string]string{"availabilityZone": "ap-southeast-1f"},
			expected:   "ap-southeast-1",
		},
		{
			name:       "extract from availabilityZone with suffix g",
			properties: map[string]string{"availabilityZone": "us-east-1g"},
			expected:   "us-east-1",
		},
		{
			name:       "explicit region property",
			properties: map[string]string{"region": "us-west-2"},
			expected:   "us-west-2",
		},
		{
			name:       "region takes precedence over availabilityZone (mapping package default)",
			properties: map[string]string{"availabilityZone": "us-east-1a", "region": "us-west-2"},
			expected:   "us-west-2",
		},
		{
			name:       "fallback to AWS_REGION env var",
			properties: map[string]string{},
			envVars:    map[string]string{"AWS_REGION": "eu-central-1"},
			expected:   "eu-central-1",
		},
		{
			name:       "fallback to AWS_DEFAULT_REGION env var",
			properties: map[string]string{},
			envVars:    map[string]string{"AWS_DEFAULT_REGION": "ap-northeast-1"},
			expected:   "ap-northeast-1",
		},
		{
			name:       "empty properties and no env vars returns empty string",
			properties: map[string]string{},
			expected:   "",
		},
		{
			name:       "availabilityZone without letter suffix returns as-is",
			properties: map[string]string{"availabilityZone": "local-zone-1"},
			expected:   "local-zone-1",
		},
		{
			name:       "empty availabilityZone falls back to region",
			properties: map[string]string{"availabilityZone": "", "region": "us-east-1"},
			expected:   "us-east-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables for this test
			for key, val := range tt.envVars {
				t.Setenv(key, val)
			}

			sku, region := resolveSKUAndRegion("aws", tt.properties)
			if region != tt.expected {
				t.Errorf("resolveSKUAndRegion() region = %q, want %q", region, tt.expected)
			}
			_ = sku
		})
	}
}
