// Package plugin provides a configurable mock plugin for testing PulumiCost's plugin communication.
// The mock plugin implements the CostSourceService gRPC interface with support for:
// - Configurable responses for different resource types
// - Error injection for testing failure scenarios
// - Performance simulation for load testing
// - State reset between tests for isolation
package plugin

import (
	"errors"

	"github.com/rshade/pulumicost-core/internal/proto"
)

// MockConfig defines configuration for mock plugin behavior.
type MockConfig struct {
	// ProjectedCostResponses maps resource types to their configured cost responses
	ProjectedCostResponses map[string]*proto.CostResult

	// ActualCostResponses maps resource IDs to their configured actual cost responses
	ActualCostResponses map[string]*proto.ActualCostResult

	// ErrorType specifies which error to inject (if any)
	ErrorType ErrorType

	// ErrorMethod specifies which method should return an error
	ErrorMethod string

	// LatencyMS specifies simulated latency in milliseconds
	LatencyMS int
}

// ErrorType represents different types of errors the mock can simulate.
type ErrorType string

const (
	// ErrorNone indicates no error should be returned.
	ErrorNone ErrorType = ""

	// ErrorTimeout simulates a timeout error.
	ErrorTimeout ErrorType = "timeout"

	// ErrorProtocol simulates a protocol/gRPC error.
	ErrorProtocol ErrorType = "protocol"

	// ErrorInvalidData simulates invalid data from the plugin.
	ErrorInvalidData ErrorType = "invalid_data"

	// ErrorUnavailable simulates the service being unavailable.
	ErrorUnavailable ErrorType = "unavailable"
)

var (
	// ErrMockTimeout is returned when timeout error is configured.
	ErrMockTimeout = errors.New("mock plugin: simulated timeout")

	// ErrMockProtocol is returned when protocol error is configured.
	ErrMockProtocol = errors.New("mock plugin: simulated protocol error")

	// ErrMockInvalidData is returned when invalid data error is configured.
	ErrMockInvalidData = errors.New("mock plugin: simulated invalid data")

	// ErrMockUnavailable is returned when unavailable error is configured.
	ErrMockUnavailable = errors.New("mock plugin: service unavailable")

	// ErrMockNotConfigured is returned when no response is configured for a request.
	ErrMockNotConfigured = errors.New("mock plugin: no response configured for resource")
)

// MockPlugin represents a configurable mock plugin server.
type MockPlugin struct {
	config MockConfig
}

// NewMockPlugin creates a new mock plugin with default configuration.
func NewMockPlugin() *MockPlugin {
	return &MockPlugin{
		config: MockConfig{
			ProjectedCostResponses: make(map[string]*proto.CostResult),
			ActualCostResponses:    make(map[string]*proto.ActualCostResult),
			ErrorType:              ErrorNone,
			ErrorMethod:            "",
			LatencyMS:              0,
		},
	}
}

// Configure sets the mock plugin configuration.
// This method can be called multiple times to update configuration.
func (m *MockPlugin) Configure(config MockConfig) {
	m.config = config
}

// SetProjectedCostResponse configures a response for a specific resource type.
// This is a convenience method for setting individual responses.
func (m *MockPlugin) SetProjectedCostResponse(resourceType string, response *proto.CostResult) {
	if m.config.ProjectedCostResponses == nil {
		m.config.ProjectedCostResponses = make(map[string]*proto.CostResult)
	}
	m.config.ProjectedCostResponses[resourceType] = response
}

// SetActualCostResponse configures a response for a specific resource ID.
// This is a convenience method for setting individual responses.
func (m *MockPlugin) SetActualCostResponse(resourceID string, response *proto.ActualCostResult) {
	if m.config.ActualCostResponses == nil {
		m.config.ActualCostResponses = make(map[string]*proto.ActualCostResult)
	}
	m.config.ActualCostResponses[resourceID] = response
}

// SetError configures the mock to return an error for a specific method.
// methodName should be "GetProjectedCost" or "GetActualCost".
// Set errorType to ErrorNone to clear error injection.
func (m *MockPlugin) SetError(methodName string, errorType ErrorType) {
	m.config.ErrorMethod = methodName
	m.config.ErrorType = errorType
}

// SetLatency configures simulated latency in milliseconds.
// Set to 0 to disable latency simulation.
func (m *MockPlugin) SetLatency(latencyMS int) {
	m.config.LatencyMS = latencyMS
}

// Reset clears all configuration and returns the mock to its default state.
// This should be called between tests to ensure isolation.
func (m *MockPlugin) Reset() {
	m.config = MockConfig{
		ProjectedCostResponses: make(map[string]*proto.CostResult),
		ActualCostResponses:    make(map[string]*proto.ActualCostResult),
		ErrorType:              ErrorNone,
		ErrorMethod:            "",
		LatencyMS:              0,
	}
}

// GetConfig returns the current mock configuration (for testing/debugging).
func (m *MockPlugin) GetConfig() MockConfig {
	return m.config
}

// shouldInjectError determines if an error should be injected for the given method.
func (m *MockPlugin) shouldInjectError(methodName string) error {
	if m.config.ErrorMethod != methodName || m.config.ErrorType == ErrorNone {
		return nil
	}

	switch m.config.ErrorType {
	case ErrorNone:
		return nil
	case ErrorTimeout:
		return ErrMockTimeout
	case ErrorProtocol:
		return ErrMockProtocol
	case ErrorInvalidData:
		return ErrMockInvalidData
	case ErrorUnavailable:
		return ErrMockUnavailable
	default:
		return nil
	}
}
