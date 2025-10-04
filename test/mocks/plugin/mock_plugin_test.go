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

func TestMockPlugin_Basic(t *testing.T) {
	mockPlugin := NewMockPlugin("test-plugin")
	require.NotNil(t, mockPlugin)

	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	assert.Greater(t, mockPlugin.GetPort(), 0)
	assert.Contains(t, mockPlugin.GetAddress(), "localhost:")
}

func TestMockPlugin_Name(t *testing.T) {
	mockPlugin := NewMockPlugin("test-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	// Connect to the mock plugin
	conn, err := grpc.Dial(mockPlugin.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewCostSourceServiceClient(conn)

	resp, err := client.Name(context.Background(), &pb.NameRequest{})
	require.NoError(t, err)
	assert.Equal(t, "test-plugin", resp.Name)
	assert.Equal(t, 1, mockPlugin.GetCallCount("Name"))
}

func TestMockPlugin_GetProjectedCost_Default(t *testing.T) {
	mockPlugin := NewMockPlugin("test-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	conn, err := grpc.Dial(mockPlugin.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
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

	assert.Equal(t, "USD", resp.Currency)
	assert.Equal(t, 10.0, resp.CostPerMonth)
	assert.Equal(t, 0.014, resp.UnitPrice)
	assert.Contains(t, resp.BillingDetail, "Mock cost")
	assert.Equal(t, 1, mockPlugin.GetCallCount("GetProjectedCost"))
}

func TestMockPlugin_GetProjectedCost_CustomResponse(t *testing.T) {
	mockPlugin := NewMockPlugin("test-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	// Set custom response for aws_instance
	customResponse := CreateProjectedCostResponse(
		"aws_instance", "USD", 50.0, 0.068, "Custom mock response")
	mockPlugin.SetProjectedCostResponse("aws_instance", customResponse)

	conn, err := grpc.Dial(mockPlugin.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
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

	assert.Equal(t, 50.0, resp.CostPerMonth)
	assert.Equal(t, 0.068, resp.UnitPrice)
	assert.Contains(t, resp.BillingDetail, "Custom mock response")
}

func TestMockPlugin_GetActualCost_Default(t *testing.T) {
	mockPlugin := NewMockPlugin("test-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	conn, err := grpc.Dial(mockPlugin.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
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
	require.Len(t, resp.Results, 1)

	result := resp.Results[0]
	assert.Equal(t, 25.50, result.Cost)
	assert.Equal(t, "mock-source", result.Source)
	assert.Equal(t, 1, mockPlugin.GetCallCount("GetActualCost"))
}

func TestMockPlugin_GetActualCost_CustomResponse(t *testing.T) {
	mockPlugin := NewMockPlugin("test-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	// Set custom response for specific resource
	customResponse := CreateActualCostResponse(
		"i-1234567890abcdef0", "USD", 75.25)
	mockPlugin.SetActualCostResponse("i-1234567890abcdef0", customResponse)

	conn, err := grpc.Dial(mockPlugin.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
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
	require.Len(t, resp.Results, 1)

	result := resp.Results[0]
	assert.Equal(t, "i-1234567890abcdef0", result.Source)
	assert.Equal(t, 75.25, result.Cost)
}

func TestMockPlugin_ErrorInjection(t *testing.T) {
	mockPlugin := NewMockPlugin("test-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	// Inject error for GetProjectedCost
	testError := errors.New("simulated plugin error")
	mockPlugin.SetError("GetProjectedCost", testError)

	conn, err := grpc.Dial(mockPlugin.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewCostSourceServiceClient(conn)

	req := &pb.GetProjectedCostRequest{
		Resource: &pb.ResourceDescriptor{ResourceType: "aws_instance"},
	}

	_, err = client.GetProjectedCost(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "simulated plugin error")
	assert.Equal(t, 1, mockPlugin.GetCallCount("GetProjectedCost"))
}

func TestMockPlugin_DelayInjection(t *testing.T) {
	mockPlugin := NewMockPlugin("test-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	// Inject delay for GetProjectedCost
	delay := 100 * time.Millisecond
	mockPlugin.SetDelay("GetProjectedCost", delay)

	conn, err := grpc.Dial(mockPlugin.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
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

func TestMockPlugin_CallCounting(t *testing.T) {
	mockPlugin := NewMockPlugin("test-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	conn, err := grpc.Dial(mockPlugin.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewCostSourceServiceClient(conn)

	// Make multiple calls
	for i := 0; i < 3; i++ {
		_, err := client.Name(context.Background(), &pb.NameRequest{})
		require.NoError(t, err)
	}

	for i := 0; i < 2; i++ {
		req := &pb.GetProjectedCostRequest{
			Resource: &pb.ResourceDescriptor{ResourceType: "aws_instance"},
		}
		_, err := client.GetProjectedCost(context.Background(), req)
		require.NoError(t, err)
	}

	assert.Equal(t, 3, mockPlugin.GetCallCount("Name"))
	assert.Equal(t, 2, mockPlugin.GetCallCount("GetProjectedCost"))
	assert.Equal(t, 0, mockPlugin.GetCallCount("GetActualCost"))

	// Reset counters
	mockPlugin.ResetCallCounts()
	assert.Equal(t, 0, mockPlugin.GetCallCount("Name"))
	assert.Equal(t, 0, mockPlugin.GetCallCount("GetProjectedCost"))
}

func TestMockPlugin_MultipleResources(t *testing.T) {
	mockPlugin := NewMockPlugin("test-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	conn, err := grpc.Dial(mockPlugin.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewCostSourceServiceClient(conn)

	// Test multiple resources by making separate calls
	resources := []string{"aws_instance", "aws_s3_bucket", "aws_rds_instance"}
	var responses []*pb.GetProjectedCostResponse

	for _, resourceType := range resources {
		req := &pb.GetProjectedCostRequest{
			Resource: &pb.ResourceDescriptor{ResourceType: resourceType, Provider: "aws"},
		}
		resp, err := client.GetProjectedCost(context.Background(), req)
		require.NoError(t, err)
		responses = append(responses, resp)
	}

	assert.Len(t, responses, 3)
	// Verify all responses have cost data
	for _, resp := range responses {
		assert.NotEmpty(t, resp.Currency)
		assert.Greater(t, resp.CostPerMonth, 0.0)
	}
}
