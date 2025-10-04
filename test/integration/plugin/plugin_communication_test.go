package plugin

import (
	"context"
	"testing"
	"time"

	"github.com/rshade/pulumicost-core/internal/proto"
	"github.com/rshade/pulumicost-core/test/mocks/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
)

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

	client := proto.NewCostSourceClient(conn)

	// Test Name method
	nameResp, err := client.Name(context.Background(), &proto.Empty{})
	require.NoError(t, err)
	assert.Equal(t, "test-integration-plugin", nameResp.GetName())
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

	// Create direct gRPC connection
	conn, err := grpc.NewClient(mockPlugin.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := proto.NewCostSourceClient(conn)

	// Test projected cost calculation
	req := &proto.GetProjectedCostRequest{
		Resources: []*proto.ResourceDescriptor{
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

	resp, err := client.GetProjectedCost(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp)

	// Verify response fields
	assert.Equal(t, "USD", resp.Currency)
	assert.Equal(t, 73.0, resp.CostPerMonth)
	assert.Equal(t, 0.10, resp.UnitPrice)
	assert.Contains(t, resp.BillingDetail, "t3.micro")
}

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

	client := proto.NewCostSourceClient(conn)

	// Test actual cost retrieval
	req := &proto.GetActualCostRequest{
		ResourceIDs: []string{resourceID},
		StartTime:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
		EndTime:     time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC).Unix(),
	}

	resp, err := client.GetActualCost(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)

	result := resp.Results[0]
	assert.Equal(t, resourceID, result.Source)
	assert.Equal(t, 85.25, result.Cost)
}

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

	client := proto.NewCostSourceClient(conn)

	// Test error handling
	req := &proto.GetProjectedCostRequest{
		Resources: []*proto.ResourceDescriptor{
			{Type: "aws_instance", Provider: "aws"},
		},
	}

	_, err = client.GetProjectedCost(context.Background(), req)
	assert.Error(t, err)

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

	// Create direct gRPC connection
	conn, err := grpc.NewClient(mockPlugin.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := proto.NewCostSourceClient(conn)

	// Test with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	req := &proto.GetProjectedCostRequest{
		Resources: []*proto.ResourceDescriptor{
			{Type: "aws_instance", Provider: "aws"},
		},
	}

	start := time.Now()
	_, err = client.GetProjectedCost(ctx, req)
	duration := time.Since(start)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
	assert.Less(t, duration, 1*time.Second) // Should timeout quickly
}