package plugin

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	pb "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
	"google.golang.org/grpc"
)

// MockPlugin implements a configurable plugin server for testing
type MockPlugin struct {
	pb.UnimplementedCostSourceServiceServer
	
	name        string
	responses   map[string]*MockResponse
	errors      map[string]error
	delays      map[string]time.Duration
	callCounts  map[string]int
	mu          sync.RWMutex
	server      *grpc.Server
	listener    net.Listener
	port        int
}

// MockResponse defines configurable responses for different methods
type MockResponse struct {
	ProjectedCost *pb.GetProjectedCostResponse
	ActualCost    *pb.GetActualCostResponse
	Name          string
}

// NewMockPlugin returns a new MockPlugin with the given name and its internal maps
// (responses, errors, delays, callCounts) initialized. The returned plugin is not
// started; call Start to open the listener and begin serving.
func NewMockPlugin(name string) *MockPlugin {
	return &MockPlugin{
		name:       name,
		responses:  make(map[string]*MockResponse),
		errors:     make(map[string]error),
		delays:     make(map[string]time.Duration),
		callCounts: make(map[string]int),
	}
}

// Start starts the mock plugin server on an available port
func (m *MockPlugin) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	lis, err := net.Listen("tcp", "127.0.0.1:0") // Use port 0 for auto-assignment
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	
	m.listener = lis
	m.port = lis.Addr().(*net.TCPAddr).Port
	
	m.server = grpc.NewServer()
	pb.RegisterCostSourceServiceServer(m.server, m)
	
	go func() {
		if err := m.server.Serve(lis); err != nil {
			log.Printf("Mock plugin server error: %v", err)
		}
	}()
	
	return nil
}

// Stop stops the mock plugin server
func (m *MockPlugin) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.server != nil {
		m.server.GracefulStop()
	}
	if m.listener != nil {
		m.listener.Close()
	}
}

// GetPort returns the port the server is listening on
func (m *MockPlugin) GetPort() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.port
}

// GetAddress returns the server address
func (m *MockPlugin) GetAddress() string {
	return fmt.Sprintf("localhost:%d", m.GetPort())
}

// SetProjectedCostResponse configures the response for GetProjectedCost
func (m *MockPlugin) SetProjectedCostResponse(key string, response *pb.GetProjectedCostResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.responses[key] == nil {
		m.responses[key] = &MockResponse{}
	}
	m.responses[key].ProjectedCost = response
}

// SetActualCostResponse configures the response for GetActualCost
func (m *MockPlugin) SetActualCostResponse(key string, response *pb.GetActualCostResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.responses[key] == nil {
		m.responses[key] = &MockResponse{}
	}
	m.responses[key].ActualCost = response
}

// SetError configures an error to be returned for a specific method
func (m *MockPlugin) SetError(method string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[method] = err
}

// SetDelay configures a delay for a specific method
func (m *MockPlugin) SetDelay(method string, delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delays[method] = delay
}

// GetCallCount returns the number of times a method was called
func (m *MockPlugin) GetCallCount(method string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callCounts[method]
}

// ResetCallCounts resets all call counters
func (m *MockPlugin) ResetCallCounts() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCounts = make(map[string]int)
}

// Name implements the gRPC service
func (m *MockPlugin) Name(ctx context.Context, req *pb.NameRequest) (*pb.NameResponse, error) {
	m.mu.Lock()
	m.callCounts["Name"]++
	m.mu.Unlock()
	
	if err := m.errors["Name"]; err != nil {
		return nil, err
	}
	
	if delay := m.delays["Name"]; delay > 0 {
		time.Sleep(delay)
	}
	
	return &pb.NameResponse{Name: m.name}, nil
}

// GetProjectedCost implements the gRPC service
func (m *MockPlugin) GetProjectedCost(ctx context.Context, req *pb.GetProjectedCostRequest) (*pb.GetProjectedCostResponse, error) {
	m.mu.Lock()
	m.callCounts["GetProjectedCost"]++
	m.mu.Unlock()
	
	if err := m.errors["GetProjectedCost"]; err != nil {
		return nil, err
	}
	
	if delay := m.delays["GetProjectedCost"]; delay > 0 {
		time.Sleep(delay)
	}
	
	// Find response based on resource type
	key := "default"
	if req.Resource != nil {
		key = req.Resource.ResourceType
	}
	
	m.mu.RLock()
	response := m.responses[key]
	m.mu.RUnlock()
	
	if response != nil && response.ProjectedCost != nil {
		return response.ProjectedCost, nil
	}
	
	// Default response
	return &pb.GetProjectedCostResponse{
		Currency:       "USD",
		CostPerMonth:   10.0, // Default mock cost
		UnitPrice:      0.014, // ~10/730 hours
		BillingDetail:  fmt.Sprintf("Mock cost for %s", req.Resource.ResourceType),
	}, nil
}

// GetActualCost implements the gRPC service
func (m *MockPlugin) GetActualCost(ctx context.Context, req *pb.GetActualCostRequest) (*pb.GetActualCostResponse, error) {
	m.mu.Lock()
	m.callCounts["GetActualCost"]++
	m.mu.Unlock()
	
	if err := m.errors["GetActualCost"]; err != nil {
		return nil, err
	}
	
	if delay := m.delays["GetActualCost"]; delay > 0 {
		time.Sleep(delay)
	}
	
	// Find response based on resource ID
	key := "default"
	if req.ResourceId != "" {
		key = req.ResourceId
	}
	
	m.mu.RLock()
	response := m.responses[key]
	m.mu.RUnlock()
	
	if response != nil && response.ActualCost != nil {
		return response.ActualCost, nil
	}
	
	// Default response
	result := &pb.ActualCostResult{
		Source: req.ResourceId,
		Cost:   25.50, // Default mock actual cost
	}
	
	return &pb.GetActualCostResponse{
		Results: []*pb.ActualCostResult{result},
	}, nil
}

// CreateProjectedCostResponse creates a GetProjectedCostResponse populated with the provided currency,
// monthly cost, hourly unit price and billing notes for the specified resource type.
func CreateProjectedCostResponse(resourceType, currency string, monthlyCost, hourlyCost float64, notes string) *pb.GetProjectedCostResponse {
	return &pb.GetProjectedCostResponse{
		Currency:       currency,
		CostPerMonth:   monthlyCost,
		UnitPrice:      hourlyCost,
		BillingDetail:  notes,
	}
}

// CreateActualCostResponse constructs a GetActualCostResponse containing a single ActualCostResult
// with Source set to resourceID and Cost set to totalCost. The currency parameter is accepted but not set on the returned result.
func CreateActualCostResponse(resourceID, currency string, totalCost float64) *pb.GetActualCostResponse {
	return &pb.GetActualCostResponse{
		Results: []*pb.ActualCostResult{
			{
				Source: resourceID,
				Cost:   totalCost,
			},
		},
	}
}