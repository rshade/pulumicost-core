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
		Resources: []*pb.ResourceDescriptor{
			{
				Type:     "aws_instance",
				Provider: "aws",
				Properties: map[string]string{
					"instance_type": "t3.micro",
				},
			},
		},
	}
	
	resp, err := client.GetProjectedCost(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	
	result := resp.Results[0]
	assert.Equal(t, "aws_instance", result.ResourceType)
	assert.Equal(t, "USD", result.Currency)
	assert.Equal(t, 10.0, result.MonthlyCost)
	assert.Contains(t, result.Notes, "Mock cost")
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
		Resources: []*pb.ResourceDescriptor{
			{
				Type:     "aws_instance",
				Provider: "aws",
			},
		},
	}
	
	resp, err := client.GetProjectedCost(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	
	result := resp.Results[0]
	assert.Equal(t, "aws_instance", result.ResourceType)
	assert.Equal(t, 50.0, result.MonthlyCost)
	assert.Equal(t, 0.068, result.HourlyCost)
	assert.Equal(t, "Custom mock response", result.Notes)
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
		ResourceIDs: []string{"i-1234567890abcdef0"},
		StartTime:   1640995200, // 2022-01-01
		EndTime:     1643673600, // 2022-02-01
	}
	
	resp, err := client.GetActualCost(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	
	result := resp.Results[0]
	assert.Equal(t, "i-1234567890abcdef0", result.ResourceID)
	assert.Equal(t, "USD", result.Currency)
	assert.Equal(t, 25.50, result.TotalCost)
	assert.Equal(t, 1, mockPlugin.GetCallCount("GetActualCost"))
}

func TestMockPlugin_GetActualCost_CustomResponse(t *testing.T) {
	mockPlugin := NewMockPlugin("test-plugin")
	err := mockPlugin.Start()
	require.NoError(t, err)
	defer mockPlugin.Stop()

	// Set custom response for specific resource
	customResponse := CreateActualCostResponse(
		"i-1234567890abcdef0", "USD", 75.25, 1640995200, 1643673600)
	mockPlugin.SetActualCostResponse("i-1234567890abcdef0", customResponse)

	conn, err := grpc.Dial(mockPlugin.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewCostSourceServiceClient(conn)
	
	req := &pb.GetActualCostRequest{
		ResourceIDs: []string{"i-1234567890abcdef0"},
		StartTime:   1640995200,
		EndTime:     1643673600,
	}
	
	resp, err := client.GetActualCost(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	
	result := resp.Results[0]
	assert.Equal(t, "i-1234567890abcdef0", result.ResourceID)
	assert.Equal(t, 75.25, result.TotalCost)
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
		Resources: []*pb.ResourceDescriptor{{Type: "aws_instance"}},
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
		Resources: []*pb.ResourceDescriptor{{Type: "aws_instance"}},
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
			Resources: []*pb.ResourceDescriptor{{Type: "aws_instance"}},
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
	
	req := &pb.GetProjectedCostRequest{
		Resources: []*pb.ResourceDescriptor{
			{Type: "aws_instance", Provider: "aws"},
			{Type: "aws_s3_bucket", Provider: "aws"},
			{Type: "aws_rds_instance", Provider: "aws"},
		},
	}
	
	resp, err := client.GetProjectedCost(context.Background(), req)
	require.NoError(t, err)
	assert.Len(t, resp.Results, 3)
	
	// Should have results for all resources
	resourceTypes := make([]string, len(resp.Results))
	for i, result := range resp.Results {
		resourceTypes[i] = result.ResourceType
	}
	
	assert.Contains(t, resourceTypes, "aws_instance")
	assert.Contains(t, resourceTypes, "aws_s3_bucket")
	assert.Contains(t, resourceTypes, "aws_rds_instance")
}