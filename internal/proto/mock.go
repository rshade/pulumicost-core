package proto

import (
	"context"

	"google.golang.org/grpc"
)

// Mock protobuf types until pulumicost-spec is fully implemented

type Empty struct{}

type ResourceDescriptor struct {
	Type       string
	Provider   string
	Properties map[string]string
}

type GetProjectedCostRequest struct {
	Resources []*ResourceDescriptor
}

type CostResult struct {
	Currency      string
	MonthlyCost   float64
	HourlyCost    float64
	Notes         string
	CostBreakdown map[string]float64
}

type GetProjectedCostResponse struct {
	Results []*CostResult
}

type GetActualCostRequest struct {
	ResourceIds []string
	StartTime   int64
	EndTime     int64
}

type ActualCostResult struct {
	Currency      string
	TotalCost     float64
	CostBreakdown map[string]float64
}

type GetActualCostResponse struct {
	Results []*ActualCostResult
}

type NameResponse struct {
	Name string
}

func (n *NameResponse) GetName() string {
	return n.Name
}

type CostSourceClient interface {
	Name(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*NameResponse, error)
	GetProjectedCost(ctx context.Context, in *GetProjectedCostRequest, opts ...grpc.CallOption) (*GetProjectedCostResponse, error)
	GetActualCost(ctx context.Context, in *GetActualCostRequest, opts ...grpc.CallOption) (*GetActualCostResponse, error)
}

func NewCostSourceClient(conn *grpc.ClientConn) CostSourceClient {
	return &mockClient{}
}

type mockClient struct{}

func (m *mockClient) Name(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*NameResponse, error) {
	return &NameResponse{Name: "mock-plugin"}, nil
}

func (m *mockClient) GetProjectedCost(ctx context.Context, in *GetProjectedCostRequest, opts ...grpc.CallOption) (*GetProjectedCostResponse, error) {
	return &GetProjectedCostResponse{Results: []*CostResult{}}, nil
}

func (m *mockClient) GetActualCost(ctx context.Context, in *GetActualCostRequest, opts ...grpc.CallOption) (*GetActualCostResponse, error) {
	return &GetActualCostResponse{Results: []*ActualCostResult{}}, nil
}