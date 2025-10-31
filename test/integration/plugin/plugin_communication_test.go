// Package plugin_test provides integration tests for plugin host communication.
package plugin_test

import (
	"context"
	"testing"
	"time"

	"github.com/rshade/pulumicost-core/test/mocks/plugin"
	pb "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TestPluginCommunication_BasicConnection tests basic gRPC connection and Name method.
func TestPluginCommunication_BasicConnection(t *testing.T) {
	// Start mock plugin server
	mockPlugin := plugin.NewMockPlugin("test-integration-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	// Create direct gRPC connection
	conn, err := grpc.NewClient(mockPlugin.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewCostSourceServiceClient(conn)

	// Test Name method
	nameResp, err := client.Name(context.Background(), &pb.NameRequest{})
	require.NoError(t, err)
	assert.Equal(t, "test-integration-plugin", nameResp.GetName())
}

// TestPluginCommunication_ProjectedCostFlow tests projected cost calculation flow with custom responses.
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

	// Create direct gRPC connection
	conn, err := grpc.NewClient(mockPlugin.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewCostSourceServiceClient(conn)

	// Test projected cost calculation
	req := &pb.GetProjectedCostRequest{
		Resource: &pb.ResourceDescriptor{
			ResourceType: "aws_instance",
			Provider:     "aws",
			Tags: map[string]string{
				"instance_type": "t3.micro",
				"region":        "us-east-1",
			},
		},
	}

	resp, err := client.GetProjectedCost(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp)

	// Verify response fields
	assert.Equal(t, "USD", resp.GetCurrency())
	assert.InDelta(t, 73.0, resp.GetCostPerMonth(), 0.01)
	assert.InDelta(t, 0.10, resp.GetUnitPrice(), 0.01)
	assert.Contains(t, resp.GetBillingDetail(), "t3.micro")
}

// TestPluginCommunication_ActualCostFlow tests actual cost retrieval with custom responses.
func TestPluginCommunication_ActualCostFlow(t *testing.T) {
	// Start mock plugin server
	mockPlugin := plugin.NewMockPlugin("actual-cost-provider")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	// Configure custom actual cost response
	resourceID := "i-1234567890abcdef0"
	customResponse := plugin.CreateActualCostResponse(resourceID, "USD", 85.25)
	mockPlugin.SetActualCostResponse(resourceID, customResponse)

	// Create direct gRPC connection
	conn, err := grpc.NewClient(mockPlugin.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewCostSourceServiceClient(conn)

	// Test actual cost retrieval
	req := &pb.GetActualCostRequest{
		ResourceId: resourceID,
		Start:      timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		End:        timestamppb.New(time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)),
		Tags:       make(map[string]string),
	}

	resp, err := client.GetActualCost(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, resp.GetResults(), 1)

	result := resp.GetResults()[0]
	assert.Equal(t, resourceID, result.GetSource())
	assert.InDelta(t, 85.25, result.GetCost(), 0.01)
}

// TestPluginCommunication_ErrorHandling tests error injection and handling in plugin communication.
func TestPluginCommunication_ErrorHandling(t *testing.T) {
	// Start mock plugin server
	mockPlugin := plugin.NewMockPlugin("error-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	// Configure error responses
	mockPlugin.SetError("GetProjectedCost", assert.AnError)

	// Create direct gRPC connection
	conn, err := grpc.NewClient(mockPlugin.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewCostSourceServiceClient(conn)

	// Test error handling
	req := &pb.GetProjectedCostRequest{
		Resource: &pb.ResourceDescriptor{ResourceType: "aws_instance", Provider: "aws"},
	}

	_, err = client.GetProjectedCost(context.Background(), req)
	require.Error(t, err)

	// Verify error call was counted
	assert.Equal(t, 1, mockPlugin.GetCallCount("GetProjectedCost"))
}

// TestPluginCommunication_Timeout tests context timeout handling with delayed responses.
func TestPluginCommunication_Timeout(t *testing.T) {
	// Start mock plugin server with delay
	mockPlugin := plugin.NewMockPlugin("slow-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	// Configure slow response (longer than context timeout)
	mockPlugin.SetDelay("GetProjectedCost", 2*time.Second)

	// Create direct gRPC connection
	conn, err := grpc.NewClient(mockPlugin.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewCostSourceServiceClient(conn)

	// Test with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	req := &pb.GetProjectedCostRequest{
		Resource: &pb.ResourceDescriptor{ResourceType: "aws_instance", Provider: "aws"},
	}

	start := time.Now()
	_, err = client.GetProjectedCost(ctx, req)
	duration := time.Since(start)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
	assert.Less(t, duration, 1*time.Second) // Should timeout quickly
}
