// Package plugin_test provides integration tests for plugin host communication.
package plugin_test

import (
	"context"
	"strings"
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
	// Start mock plugin server on TCP
	server, err := plugin.StartMockServerTCP()
	require.NoError(t, err)
	defer server.Stop()

	// Create direct gRPC connection
	conn, err := grpc.NewClient(server.Address(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewCostSourceServiceClient(conn)

	// Test Name method
	nameResp, err := client.Name(context.Background(), &pb.NameRequest{})
	require.NoError(t, err)
	assert.Equal(t, "mock-plugin", nameResp.GetName())
}

// TestPluginCommunication_ProjectedCostFlow tests projected cost calculation flow with custom responses.
func TestPluginCommunication_ProjectedCostFlow(t *testing.T) {
	// Start mock plugin server on TCP
	server, err := plugin.StartMockServerTCP()
	require.NoError(t, err)
	defer server.Stop()

	// Configure custom response for AWS instance
	customResponse := plugin.QuickResponse("USD", 73.0, 0.10)
	customResponse.Notes = "EC2 t3.micro instance in us-east-1"
	server.Plugin.SetProjectedCostResponse("aws_instance", customResponse)

	// Create direct gRPC connection
	conn, err := grpc.NewClient(server.Address(), grpc.WithTransportCredentials(insecure.NewCredentials()))
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
	// Start mock plugin server on TCP
	server, err := plugin.StartMockServerTCP()
	require.NoError(t, err)
	defer server.Stop()

	// Configure custom actual cost response
	resourceID := "i-1234567890abcdef0"
	customResponse := plugin.QuickActualResponse("USD", 85.25)
	server.Plugin.SetActualCostResponse(resourceID, customResponse)

	// Create direct gRPC connection
	conn, err := grpc.NewClient(server.Address(), grpc.WithTransportCredentials(insecure.NewCredentials()))
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
	assert.Equal(t, "total", result.GetSource()) // Source is breakdown category, not resource ID
	assert.InDelta(t, 85.25, result.GetCost(), 0.01)
}

// TestPluginCommunication_ErrorHandling tests error injection and handling in plugin communication.
func TestPluginCommunication_ErrorHandling(t *testing.T) {
	// Start mock plugin server on TCP
	server, err := plugin.StartMockServerTCP()
	require.NoError(t, err)
	defer server.Stop()

	// Configure error responses
	server.Plugin.SetError("GetProjectedCost", plugin.ErrorProtocol)

	// Create direct gRPC connection
	conn, err := grpc.NewClient(server.Address(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewCostSourceServiceClient(conn)

	// Test error handling
	req := &pb.GetProjectedCostRequest{
		Resource: &pb.ResourceDescriptor{ResourceType: "aws_instance", Provider: "aws"},
	}

	_, err = client.GetProjectedCost(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "protocol")
}

// TestPluginCommunication_Timeout tests context timeout handling with delayed responses.
func TestPluginCommunication_Timeout(t *testing.T) {
	// Start mock plugin server on TCP
	server, err := plugin.StartMockServerTCP()
	require.NoError(t, err)
	defer server.Stop()

	// Configure slow response (2000ms latency, longer than 500ms context timeout)
	server.Plugin.SetLatency(2000)

	// Create direct gRPC connection
	conn, err := grpc.NewClient(server.Address(), grpc.WithTransportCredentials(insecure.NewCredentials()))
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
	// gRPC may return either "context deadline exceeded" or "DeadlineExceeded" error code
	errStr := err.Error()
	assert.True(t, containsAny(errStr, "context deadline exceeded", "DeadlineExceeded"),
		"expected timeout error, got: %s", errStr)
	assert.Less(t, duration, 1*time.Second) // Should timeout quickly
}

// containsAny returns true if s contains any of the substrings.
func containsAny(s string, substrings ...string) bool {
	for _, substr := range substrings {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}
