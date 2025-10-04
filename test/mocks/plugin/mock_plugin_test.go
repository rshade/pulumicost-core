package plugin

import (
	"context"
	"errors"
	"testing"
	"time"

	pb "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TestMockPlugin_Basic tests basic mock plugin creation and server startup.
func TestMockPlugin_Basic(t *testing.T) {
	mockPlugin := NewMockPlugin("test-plugin")
	require.NotNil(t, mockPlugin)

	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	assert.Positive(t, mockPlugin.GetPort())
	assert.Contains(t, mockPlugin.GetAddress(), "localhost:")
}

// TestMockPlugin_Name tests the Name RPC method and call counting.
func TestMockPlugin_Name(t *testing.T) {
	mockPlugin := NewMockPlugin("test-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	// Connect to the mock plugin
	conn, err := grpc.NewClient(mockPlugin.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewCostSourceServiceClient(conn)

	resp, err := client.Name(context.Background(), &pb.NameRequest{})
	require.NoError(t, err)
	assert.Equal(t, "test-plugin", resp.GetName())
	assert.Equal(t, 1, mockPlugin.GetCallCount("Name"))
}

// TestMockPlugin_GetProjectedCost_Default tests default projected cost responses.
func TestMockPlugin_GetProjectedCost_Default(t *testing.T) {
	mockPlugin := NewMockPlugin("test-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	conn, err := grpc.NewClient(mockPlugin.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewCostSourceServiceClient(conn)

	req := &pb.GetProjectedCostRequest{
		Resource: &pb.ResourceDescriptor{
			ResourceType: "aws_instance",
			Provider:     "aws",
			Tags: map[string]string{
				"instance_type": "t3.micro",
			},
		},
	}

	resp, err := client.GetProjectedCost(context.Background(), req)
	require.NoError(t, err)

	assert.Equal(t, "USD", resp.GetCurrency())
	assert.InDelta(t, 10.0, resp.GetCostPerMonth(), 0.01)
	assert.InDelta(t, 0.014, resp.GetUnitPrice(), 0.01)
	assert.Contains(t, resp.GetBillingDetail(), "Mock cost")
	assert.Equal(t, 1, mockPlugin.GetCallCount("GetProjectedCost"))
}

// TestMockPlugin_GetProjectedCost_CustomResponse tests custom projected cost responses.
func TestMockPlugin_GetProjectedCost_CustomResponse(t *testing.T) {
	mockPlugin := NewMockPlugin("test-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	// Set custom response for aws_instance
	customResponse := CreateProjectedCostResponse(
		"aws_instance", "USD", 50.0, 0.068, "Custom mock response")
	mockPlugin.SetProjectedCostResponse("aws_instance", customResponse)

	conn, err := grpc.NewClient(mockPlugin.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewCostSourceServiceClient(conn)

	req := &pb.GetProjectedCostRequest{
		Resource: &pb.ResourceDescriptor{
			ResourceType: "aws_instance",
			Provider:     "aws",
		},
	}

	resp, err := client.GetProjectedCost(context.Background(), req)
	require.NoError(t, err)

	assert.InDelta(t, 50.0, resp.GetCostPerMonth(), 0.01)
	assert.InDelta(t, 0.068, resp.GetUnitPrice(), 0.01)
	assert.Contains(t, resp.GetBillingDetail(), "Custom mock response")
}

// TestMockPlugin_GetActualCost_Default tests default actual cost responses.
func TestMockPlugin_GetActualCost_Default(t *testing.T) {
	mockPlugin := NewMockPlugin("test-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	conn, err := grpc.NewClient(mockPlugin.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewCostSourceServiceClient(conn)

	req := &pb.GetActualCostRequest{
		ResourceId: "i-1234567890abcdef0",
		Start:      timestamppb.New(time.Unix(1640995200, 0)), // 2022-01-01
		End:        timestamppb.New(time.Unix(1643673600, 0)), // 2022-02-01
		Tags:       make(map[string]string),
	}

	resp, err := client.GetActualCost(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, resp.GetResults(), 1)

	result := resp.GetResults()[0]
	assert.InDelta(t, 25.50, result.GetCost(), 0.01)
	assert.Equal(t, "mock-source", result.GetSource())
	assert.Equal(t, 1, mockPlugin.GetCallCount("GetActualCost"))
}

// TestMockPlugin_GetActualCost_CustomResponse tests custom actual cost responses.
func TestMockPlugin_GetActualCost_CustomResponse(t *testing.T) {
	mockPlugin := NewMockPlugin("test-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	// Set custom response for specific resource
	customResponse := CreateActualCostResponse(
		"i-1234567890abcdef0", "USD", 75.25)
	mockPlugin.SetActualCostResponse("i-1234567890abcdef0", customResponse)

	conn, err := grpc.NewClient(mockPlugin.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewCostSourceServiceClient(conn)

	req := &pb.GetActualCostRequest{
		ResourceId: "i-1234567890abcdef0",
		Start:      timestamppb.New(time.Unix(1640995200, 0)),
		End:        timestamppb.New(time.Unix(1643673600, 0)),
		Tags:       make(map[string]string),
	}

	resp, err := client.GetActualCost(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, resp.GetResults(), 1)

	result := resp.GetResults()[0]
	assert.Equal(t, "i-1234567890abcdef0", result.GetSource())
	assert.InDelta(t, 75.25, result.GetCost(), 0.01)
}

// TestMockPlugin_ErrorInjection tests error injection for testing error paths.
func TestMockPlugin_ErrorInjection(t *testing.T) {
	mockPlugin := NewMockPlugin("test-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	// Inject error for GetProjectedCost
	testError := errors.New("simulated plugin error")
	mockPlugin.SetError("GetProjectedCost", testError)

	conn, err := grpc.NewClient(mockPlugin.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewCostSourceServiceClient(conn)

	req := &pb.GetProjectedCostRequest{
		Resource: &pb.ResourceDescriptor{ResourceType: "aws_instance"},
	}

	_, err = client.GetProjectedCost(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "simulated plugin error")
	assert.Equal(t, 1, mockPlugin.GetCallCount("GetProjectedCost"))
}

// TestMockPlugin_DelayInjection tests delay injection for timeout testing.
func TestMockPlugin_DelayInjection(t *testing.T) {
	mockPlugin := NewMockPlugin("test-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	// Inject delay for GetProjectedCost
	delay := 100 * time.Millisecond
	mockPlugin.SetDelay("GetProjectedCost", delay)

	conn, err := grpc.NewClient(mockPlugin.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewCostSourceServiceClient(conn)

	req := &pb.GetProjectedCostRequest{
		Resource: &pb.ResourceDescriptor{ResourceType: "aws_instance"},
	}

	start := time.Now()
	_, err = client.GetProjectedCost(context.Background(), req)
	duration := time.Since(start)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, duration, delay)
}

// TestMockPlugin_CallCounting tests call counting and reset functionality.
func TestMockPlugin_CallCounting(t *testing.T) {
	mockPlugin := NewMockPlugin("test-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	conn, dialErr := grpc.NewClient(mockPlugin.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, dialErr)
	defer conn.Close()

	client := pb.NewCostSourceServiceClient(conn)

	// Make multiple calls
	for range 3 {
		_, nameErr := client.Name(context.Background(), &pb.NameRequest{})
		require.NoError(t, nameErr)
	}

	for range 2 {
		req := &pb.GetProjectedCostRequest{
			Resource: &pb.ResourceDescriptor{ResourceType: "aws_instance"},
		}
		_, projectedErr := client.GetProjectedCost(context.Background(), req)
		require.NoError(t, projectedErr)
	}

	assert.Equal(t, 3, mockPlugin.GetCallCount("Name"))
	assert.Equal(t, 2, mockPlugin.GetCallCount("GetProjectedCost"))
	assert.Equal(t, 0, mockPlugin.GetCallCount("GetActualCost"))

	// Reset counters
	mockPlugin.ResetCallCounts()
	assert.Equal(t, 0, mockPlugin.GetCallCount("Name"))
	assert.Equal(t, 0, mockPlugin.GetCallCount("GetProjectedCost"))
}

// TestMockPlugin_MultipleResources tests handling multiple resource types.
func TestMockPlugin_MultipleResources(t *testing.T) {
	mockPlugin := NewMockPlugin("test-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	conn, dialErr := grpc.NewClient(mockPlugin.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, dialErr)
	defer conn.Close()

	client := pb.NewCostSourceServiceClient(conn)

	// Test multiple resources by making separate calls
	resources := []string{"aws_instance", "aws_s3_bucket", "aws_rds_instance"}
	var responses []*pb.GetProjectedCostResponse

	for _, resourceType := range resources {
		req := &pb.GetProjectedCostRequest{
			Resource: &pb.ResourceDescriptor{ResourceType: resourceType, Provider: "aws"},
		}
		resp, costErr := client.GetProjectedCost(context.Background(), req)
		require.NoError(t, costErr)
		responses = append(responses, resp)
	}

	assert.Len(t, responses, 3)
	// Verify all responses have cost data
	for _, resp := range responses {
		assert.NotEmpty(t, resp.GetCurrency())
		assert.Greater(t, resp.GetCostPerMonth(), 0.0)
	}
}
