package plugin

import (
	"context"
	"testing"
	"time"

	"github.com/rshade/pulumicost-core/internal/pluginhost"
	"github.com/rshade/pulumicost-core/test/mocks/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
)

func TestPluginCommunication_BasicConnection(t *testing.T) {
	// Start mock plugin server
	mockPlugin := plugin.NewMockPlugin("test-integration-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	// Create plugin client
	client, err := pluginhost.NewClient("test-plugin", mockPlugin.GetAddress())
	require.NoError(t, err)
	defer client.Close()

	// Test Name method
	nameResp, err := client.API.Name(context.Background(), &pb.NameRequest{})
	require.NoError(t, err)
	assert.Equal(t, "test-integration-plugin", nameResp.Name)
}

func TestPluginCommunication_ProjectedCostFlow(t *testing.T) {
	// Start mock plugin server
	mockPlugin := plugin.NewMockPlugin("cost-calculator")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	// Configure custom response for AWS instance
	customResponse := plugin.CreateProjectedCostResponse(
		"aws_instance", "USD", 73.0, 0.10, "EC2 t3.micro instance in us-east-1")
	mockPlugin.SetProjectedCostResponse("aws_instance", customResponse)

	// Create plugin client
	client, err := pluginhost.NewClient("cost-calculator", mockPlugin.GetAddress())
	require.NoError(t, err)
	defer client.Close()

	// Test projected cost calculation
	req := &pb.GetProjectedCostRequest{
		Resources: []*pb.ResourceDescriptor{
			{
				Type:     "aws_instance",
				Provider: "aws",
				Properties: map[string]string{
					"instance_type": "t3.micro",
					"region":        "us-east-1",
				},
			},
		},
	}

	resp, err := client.API.GetProjectedCost(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)

	result := resp.Results[0]
	assert.Equal(t, "aws_instance", result.ResourceType)
	assert.Equal(t, "USD", result.Currency)
	assert.Equal(t, 73.0, result.MonthlyCost)
	assert.Equal(t, 0.10, result.HourlyCost)
	assert.Contains(t, result.Notes, "t3.micro")
	assert.NotEmpty(t, result.CostBreakdown)
}

func TestPluginCommunication_ActualCostFlow(t *testing.T) {
	// Start mock plugin server
	mockPlugin := plugin.NewMockPlugin("actual-cost-provider")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	// Configure custom actual cost response
	resourceID := "i-1234567890abcdef0"
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	endTime := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC).Unix()
	
	customResponse := plugin.CreateActualCostResponse(
		resourceID, "USD", 85.25, startTime, endTime)
	mockPlugin.SetActualCostResponse(resourceID, customResponse)

	// Create plugin client
	client, err := pluginhost.NewClient("actual-cost-provider", mockPlugin.GetAddress())
	require.NoError(t, err)
	defer client.Close()

	// Test actual cost retrieval
	req := &pb.GetActualCostRequest{
		ResourceIDs: []string{resourceID},
		StartTime:   startTime,
		EndTime:     endTime,
	}

	resp, err := client.API.GetActualCost(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)

	result := resp.Results[0]
	assert.Equal(t, resourceID, result.ResourceID)
	assert.Equal(t, "USD", result.Currency)
	assert.Equal(t, 85.25, result.TotalCost)
	assert.Equal(t, startTime, result.StartTime)
	assert.Equal(t, endTime, result.EndTime)
	assert.NotEmpty(t, result.CostBreakdown)
}

func TestPluginCommunication_ErrorHandling(t *testing.T) {
	// Start mock plugin server
	mockPlugin := plugin.NewMockPlugin("error-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	// Configure error responses
	mockPlugin.SetError("GetProjectedCost", assert.AnError)

	// Create plugin client
	client, err := pluginhost.NewClient("error-plugin", mockPlugin.GetAddress())
	require.NoError(t, err)
	defer client.Close()

	// Test error handling
	req := &pb.GetProjectedCostRequest{
		Resources: []*pb.ResourceDescriptor{
			{Type: "aws_instance", Provider: "aws"},
		},
	}

	_, err = client.API.GetProjectedCost(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "assert.AnError")

	// Verify error call was counted
	assert.Equal(t, 1, mockPlugin.GetCallCount("GetProjectedCost"))
}

func TestPluginCommunication_Timeout(t *testing.T) {
	// Start mock plugin server with delay
	mockPlugin := plugin.NewMockPlugin("slow-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	// Configure slow response (longer than context timeout)
	mockPlugin.SetDelay("GetProjectedCost", 2*time.Second)

	// Create plugin client
	client, err := pluginhost.NewClient("slow-plugin", mockPlugin.GetAddress())
	require.NoError(t, err)
	defer client.Close()

	// Test with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	req := &pb.GetProjectedCostRequest{
		Resources: []*pb.ResourceDescriptor{
			{Type: "aws_instance", Provider: "aws"},
		},
	}

	start := time.Now()
	_, err = client.API.GetProjectedCost(ctx, req)
	duration := time.Since(start)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
	assert.Less(t, duration, 1*time.Second) // Should timeout quickly
}

func TestPluginCommunication_MultipleClients(t *testing.T) {
	// Start mock plugin server
	mockPlugin := plugin.NewMockPlugin("multi-client-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	// Create multiple clients
	client1, err := pluginhost.NewClient("client1", mockPlugin.GetAddress())
	require.NoError(t, err)
	defer client1.Close()

	client2, err := pluginhost.NewClient("client2", mockPlugin.GetAddress())
	require.NoError(t, err)
	defer client2.Close()

	// Both clients should be able to communicate
	nameResp1, err := client1.API.Name(context.Background(), &pb.NameRequest{})
	require.NoError(t, err)
	assert.Equal(t, "multi-client-plugin", nameResp1.Name)

	nameResp2, err := client2.API.Name(context.Background(), &pb.NameRequest{})
	require.NoError(t, err)
	assert.Equal(t, "multi-client-plugin", nameResp2.Name)

	// Verify both calls were counted
	assert.Equal(t, 2, mockPlugin.GetCallCount("Name"))
}

func TestPluginCommunication_ResourceBatching(t *testing.T) {
	// Start mock plugin server
	mockPlugin := plugin.NewMockPlugin("batch-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	// Create plugin client
	client, err := pluginhost.NewClient("batch-plugin", mockPlugin.GetAddress())
	require.NoError(t, err)
	defer client.Close()

	// Test with multiple resources in single request
	req := &pb.GetProjectedCostRequest{
		Resources: []*pb.ResourceDescriptor{
			{Type: "aws_instance", Provider: "aws"},
			{Type: "aws_s3_bucket", Provider: "aws"},
			{Type: "aws_rds_instance", Provider: "aws"},
			{Type: "aws_lambda_function", Provider: "aws"},
		},
	}

	resp, err := client.API.GetProjectedCost(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, resp.Results, 4)

	// Verify all resources got results
	resourceTypes := make(map[string]bool)
	for _, result := range resp.Results {
		resourceTypes[result.ResourceType] = true
		assert.Equal(t, "USD", result.Currency)
		assert.Greater(t, result.MonthlyCost, 0.0)
		assert.NotEmpty(t, result.Notes)
	}

	assert.True(t, resourceTypes["aws_instance"])
	assert.True(t, resourceTypes["aws_s3_bucket"])
	assert.True(t, resourceTypes["aws_rds_instance"])
	assert.True(t, resourceTypes["aws_lambda_function"])

	// Should be only one call to GetProjectedCost
	assert.Equal(t, 1, mockPlugin.GetCallCount("GetProjectedCost"))
}

func TestPluginCommunication_ConnectionRecovery(t *testing.T) {
	// Start mock plugin server
	mockPlugin := plugin.NewMockPlugin("recovery-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)

	// Create plugin client
	client, err := pluginhost.NewClient("recovery-plugin", mockPlugin.GetAddress())
	require.NoError(t, err)
	defer client.Close()

	// Test initial connection
	_, err = client.API.Name(context.Background(), &pb.NameRequest{})
	require.NoError(t, err)

	// Stop the plugin server
	mockPlugin.Stop()

	// This should fail
	_, err = client.API.Name(context.Background(), &pb.NameRequest{})
	assert.Error(t, err)

	// Restart the plugin server on the same address
	mockPlugin2 := plugin.NewMockPlugin("recovery-plugin-v2")
	err = mockPlugin2.Start()
	require.NoError(t, err)
	defer mockPlugin2.Stop()

	// Note: In a real scenario, the client would need reconnection logic
	// This test demonstrates the behavior when connection is lost
}

func TestPluginCommunication_ProtocolValidation(t *testing.T) {
	// Start mock plugin server
	mockPlugin := plugin.NewMockPlugin("protocol-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	// Create plugin client
	client, err := pluginhost.NewClient("protocol-plugin", mockPlugin.GetAddress())
	require.NoError(t, err)
	defer client.Close()

	// Test with invalid/empty request
	emptyReq := &pb.GetProjectedCostRequest{Resources: nil}
	resp, err := client.API.GetProjectedCost(context.Background(), emptyReq)
	require.NoError(t, err)
	assert.Empty(t, resp.Results) // Should handle empty resources gracefully

	// Test with malformed resource descriptor
	malformedReq := &pb.GetProjectedCostRequest{
		Resources: []*pb.ResourceDescriptor{
			{Type: "", Provider: "", Properties: nil}, // Empty fields
		},
	}
	resp, err = client.API.GetProjectedCost(context.Background(), malformedReq)
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	
	// Mock plugin should still return result even for empty resource
	result := resp.Results[0]
	assert.Equal(t, "", result.ResourceType) // Mirrors input
	assert.NotEmpty(t, result.Currency)      // Has defaults
}