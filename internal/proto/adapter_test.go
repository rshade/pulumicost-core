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

func (m *mockCostSourceClient) Name(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*NameResponse, error) {
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
				{ResourceType: "aws:rds:Instance", ResourceID: "db-456", Error: errors.New("test2")},
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

		result := GetProjectedCostWithErrors(context.Background(), mockClient, "test-plugin", resources)

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
			t.Errorf("Error ResourceType = %s, want aws:rds:Instance", result.Errors[0].ResourceType)
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

	t.Run("all resources succeed", func(t *testing.T) {
		mockClient := &mockCostSourceClient{
			getProjectedFunc: func(ctx context.Context, in *GetProjectedCostRequest, opts ...grpc.CallOption) (*GetProjectedCostResponse, error) {
				return &GetProjectedCostResponse{
					Results: []*CostResult{
						{Currency: "USD", MonthlyCost: 100.0},
					},
				}, nil
			},
		}

		resources := []*ResourceDescriptor{
			{Type: "aws:ec2:Instance", Provider: "aws", Properties: map[string]string{}},
		}

		result := GetProjectedCostWithErrors(context.Background(), mockClient, "test-plugin", resources)

		if result.HasErrors() {
			t.Error("HasErrors() should return false when all resources succeed")
		}
		if len(result.Results) != 1 {
			t.Errorf("Results length = %d, want 1", len(result.Results))
		}
	})

	t.Run("all resources fail", func(t *testing.T) {
		mockClient := &mockCostSourceClient{
			getProjectedFunc: func(ctx context.Context, in *GetProjectedCostRequest, opts ...grpc.CallOption) (*GetProjectedCostResponse, error) {
				return nil, errors.New("plugin unavailable")
			},
		}

		resources := []*ResourceDescriptor{
			{Type: "aws:ec2:Instance", Provider: "aws", Properties: map[string]string{}},
			{Type: "aws:rds:Instance", Provider: "aws", Properties: map[string]string{}},
		}

		result := GetProjectedCostWithErrors(context.Background(), mockClient, "test-plugin", resources)

		// Should have 2 placeholder results
		if len(result.Results) != 2 {
			t.Errorf("Results length = %d, want 2", len(result.Results))
		}

		// Should have 2 errors
		if len(result.Errors) != 2 {
			t.Errorf("Errors length = %d, want 2", len(result.Errors))
		}
	})
}
