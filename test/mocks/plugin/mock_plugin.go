// Package plugin provides mock implementations for testing plugin communication.
package plugin

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	pb "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
	"google.golang.org/grpc"
)

const (
	// Default mock cost values for testing.
	defaultMockCostPerMonth = 10.0
	defaultMockUnitPrice    = 0.014 // ~10/730 hours
	defaultMockActualCost   = 25.50
)

// MockPlugin implements a configurable plugin server for testing.
type MockPlugin struct {
	pb.UnimplementedCostSourceServiceServer

	name       string
	responses  map[string]*MockResponse
	errors     map[string]error
	delays     map[string]time.Duration
	callCounts map[string]int
	mu         sync.RWMutex
	server     *grpc.Server
	listener   net.Listener
	port       int
}

// MockResponse defines configurable responses for different methods.
type MockResponse struct {
	ProjectedCost *pb.GetProjectedCostResponse
	ActualCost    *pb.GetActualCostResponse
	Name          string
}

// NewMockPlugin creates a new mock plugin server.
func NewMockPlugin(name string) *MockPlugin {
	return &MockPlugin{
		name:       name,
		responses:  make(map[string]*MockResponse),
		errors:     make(map[string]error),
		delays:     make(map[string]time.Duration),
		callCounts: make(map[string]int),
	}
}

// Start starts the mock plugin server on an available port.
func (m *MockPlugin) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	lc := &net.ListenConfig{}
	lis, err := lc.Listen(
		context.Background(),
		"tcp",
		"127.0.0.1:0",
	) // Use port 0 for auto-assignment
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	m.listener = lis
	addr, ok := lis.Addr().(*net.TCPAddr)
	if !ok {
		_ = lis.Close()
		return fmt.Errorf("listener address is not TCP: %T", lis.Addr())
	}
	m.port = addr.Port

	m.server = grpc.NewServer()
	pb.RegisterCostSourceServiceServer(m.server, m)

	go func() {
		if serveErr := m.server.Serve(lis); serveErr != nil {
			slog.Error("Mock plugin server error", "error", serveErr)
		}
	}()

	return nil
}

// Stop stops the mock plugin server.
func (m *MockPlugin) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.server != nil {
		m.server.GracefulStop()
	}
	if m.listener != nil {
		_ = m.listener.Close()
	}
}

// GetPort returns the port the server is listening on.
func (m *MockPlugin) GetPort() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.port
}

// GetAddress returns the server address.
func (m *MockPlugin) GetAddress() string {
	return fmt.Sprintf("127.0.0.1:%d", m.GetPort())
}

// SetProjectedCostResponse configures the response for GetProjectedCost.
func (m *MockPlugin) SetProjectedCostResponse(key string, response *pb.GetProjectedCostResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.responses[key] == nil {
		m.responses[key] = &MockResponse{}
	}
	m.responses[key].ProjectedCost = response
}

// SetActualCostResponse configures the response for GetActualCost.
func (m *MockPlugin) SetActualCostResponse(key string, response *pb.GetActualCostResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.responses[key] == nil {
		m.responses[key] = &MockResponse{}
	}
	m.responses[key].ActualCost = response
}

// SetError configures an error to be returned for a specific method.
func (m *MockPlugin) SetError(method string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[method] = err
}

// SetDelay configures a delay for a specific method.
func (m *MockPlugin) SetDelay(method string, delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delays[method] = delay
}

// GetCallCount returns the number of times a method was called.
func (m *MockPlugin) GetCallCount(method string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callCounts[method]
}

// getError returns the configured error for a method under read lock.
func (m *MockPlugin) getError(method string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.errors[method]
}

// getDelay returns the configured delay for a method under read lock.
func (m *MockPlugin) getDelay(method string) time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delays[method]
}

// ResetCallCounts resets all call counters.
func (m *MockPlugin) ResetCallCounts() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCounts = make(map[string]int)
}

// Name implements the gRPC service.
func (m *MockPlugin) Name(ctx context.Context, _ *pb.NameRequest) (*pb.NameResponse, error) {
	m.mu.Lock()
	m.callCounts["Name"]++
	m.mu.Unlock()

	if err := m.getError("Name"); err != nil {
		return nil, err
	}

	if d := m.getDelay("Name"); d > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(d):
		}
	}

	return &pb.NameResponse{Name: m.name}, nil
}

// GetProjectedCost implements the gRPC service.
func (m *MockPlugin) GetProjectedCost(
	ctx context.Context,
	req *pb.GetProjectedCostRequest,
) (*pb.GetProjectedCostResponse, error) {
	m.mu.Lock()
	m.callCounts["GetProjectedCost"]++
	m.mu.Unlock()

	if err := m.getError("GetProjectedCost"); err != nil {
		return nil, err
	}

	if d := m.getDelay("GetProjectedCost"); d > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(d):
		}
	}

	// Find response based on resource type
	key := "default"
	if req.GetResource() != nil {
		key = req.GetResource().GetResourceType()
	}

	m.mu.RLock()
	response := m.responses[key]
	m.mu.RUnlock()

	if response != nil && response.ProjectedCost != nil {
		return response.ProjectedCost, nil
	}

	// Default response
	resourceType := "unknown"
	if req != nil && req.GetResource() != nil {
		resourceType = req.GetResource().GetResourceType()
	}
	return &pb.GetProjectedCostResponse{
		Currency:      "USD",
		CostPerMonth:  defaultMockCostPerMonth,
		UnitPrice:     defaultMockUnitPrice,
		BillingDetail: fmt.Sprintf("Mock cost for %s", resourceType),
	}, nil
}

// GetActualCost implements the gRPC service.
func (m *MockPlugin) GetActualCost(
	ctx context.Context,
	req *pb.GetActualCostRequest,
) (*pb.GetActualCostResponse, error) {
	m.mu.Lock()
	m.callCounts["GetActualCost"]++
	m.mu.Unlock()

	if err := m.getError("GetActualCost"); err != nil {
		return nil, err
	}

	if d := m.getDelay("GetActualCost"); d > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(d):
		}
	}

	// Find response based on resource ID
	key := "default"
	if req.GetResourceId() != "" {
		key = req.GetResourceId()
	}

	m.mu.RLock()
	response := m.responses[key]
	m.mu.RUnlock()

	if response != nil && response.ActualCost != nil {
		return response.ActualCost, nil
	}

	// Default response
	result := &pb.ActualCostResult{
		Source: req.GetResourceId(),
		Cost:   defaultMockActualCost,
	}

	return &pb.GetActualCostResponse{
		Results: []*pb.ActualCostResult{result},
	}, nil
}

// CreateProjectedCostResponse creates a standard projected cost response.
func CreateProjectedCostResponse(
	_ string,
	currency string,
	monthlyCost, hourlyCost float64,
	notes string,
) *pb.GetProjectedCostResponse {
	return &pb.GetProjectedCostResponse{
		Currency:      currency,
		CostPerMonth:  monthlyCost,
		UnitPrice:     hourlyCost,
		BillingDetail: notes,
	}
}

// CreateActualCostResponse creates a standard actual cost response.
func CreateActualCostResponse(
	resourceID string,
	_ string,
	totalCost float64,
) *pb.GetActualCostResponse {
	return &pb.GetActualCostResponse{
		Results: []*pb.ActualCostResult{
			{
				Source: resourceID,
				Cost:   totalCost,
			},
		},
	}
}
