package pluginsdk

import (
	"context"
	"errors"
	"testing"

	pbc "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mockPlugin implements Plugin interface for testing.
type mockPlugin struct {
	name string
}

func (m *mockPlugin) Name() string { return m.name }

func (m *mockPlugin) GetProjectedCost(
	_ context.Context,
	_ *pbc.GetProjectedCostRequest,
) (*pbc.GetProjectedCostResponse, error) {
	return &pbc.GetProjectedCostResponse{}, nil
}

func (m *mockPlugin) GetActualCost(
	_ context.Context,
	_ *pbc.GetActualCostRequest,
) (*pbc.GetActualCostResponse, error) {
	return &pbc.GetActualCostResponse{}, nil
}

// mockSupportsPlugin implements both Plugin and SupportsProvider interfaces.
type mockSupportsPlugin struct {
	mockPlugin

	supported bool
	reason    string
	err       error
}

func (m *mockSupportsPlugin) Supports(
	_ context.Context,
	_ *pbc.SupportsRequest,
) (*pbc.SupportsResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &pbc.SupportsResponse{
		Supported: m.supported,
		Reason:    m.reason,
	}, nil
}

// mockRegistry implements RegistryLookup for testing.
type mockRegistry struct {
	plugins map[string]string // key: "provider:region", value: plugin name
}

func (m *mockRegistry) FindPlugin(provider, region string) string {
	key := provider + ":" + region
	return m.plugins[key]
}

// TestSupports_PluginImplementsAndReturnsSupported tests T005: plugin implements SupportsProvider and returns supported.
func TestSupports_PluginImplementsAndReturnsSupported(t *testing.T) {
	plugin := &mockSupportsPlugin{
		mockPlugin: mockPlugin{name: "test-plugin"},
		supported:  true,
	}
	registry := &mockRegistry{
		plugins: map[string]string{"aws:us-east-1": "test-plugin"},
	}
	server := NewServerWithRegistry(plugin, registry)

	req := &pbc.SupportsRequest{
		Resource: &pbc.ResourceDescriptor{
			Provider:     "aws",
			ResourceType: "ec2",
			Region:       "us-east-1",
		},
	}

	resp, err := server.Supports(context.Background(), req)

	require.NoError(t, err)
	assert.True(t, resp.GetSupported())
	assert.Empty(t, resp.GetReason())
}

// TestSupports_PluginImplementsAndReturnsNotSupported tests T006: plugin returns not-supported with reason.
func TestSupports_PluginImplementsAndReturnsNotSupported(t *testing.T) {
	plugin := &mockSupportsPlugin{
		mockPlugin: mockPlugin{name: "test-plugin"},
		supported:  false,
		reason:     "Resource type not supported by this plugin",
	}
	registry := &mockRegistry{
		plugins: map[string]string{"aws:us-east-1": "test-plugin"},
	}
	server := NewServerWithRegistry(plugin, registry)

	req := &pbc.SupportsRequest{
		Resource: &pbc.ResourceDescriptor{
			Provider:     "aws",
			ResourceType: "gamelift",
			Region:       "us-east-1",
		},
	}

	resp, err := server.Supports(context.Background(), req)

	require.NoError(t, err)
	assert.False(t, resp.GetSupported())
	assert.Equal(t, "Resource type not supported by this plugin", resp.GetReason())
}

// TestSupports_InvalidProviderRegionReturnsInvalidArgument tests T007: invalid provider/region returns InvalidArgument.
func TestSupports_InvalidProviderRegionReturnsInvalidArgument(t *testing.T) {
	plugin := &mockSupportsPlugin{
		mockPlugin: mockPlugin{name: "test-plugin"},
		supported:  true,
	}
	registry := &mockRegistry{
		plugins: map[string]string{"aws:us-east-1": "test-plugin"},
	}
	server := NewServerWithRegistry(plugin, registry)

	req := &pbc.SupportsRequest{
		Resource: &pbc.ResourceDescriptor{
			Provider:     "unknown-provider",
			ResourceType: "ec2",
			Region:       "us-east-1",
		},
	}

	_, err := server.Supports(context.Background(), req)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "no plugin registered")
}

// TestSupports_NoPluginRegisteredReturnsInvalidArgument tests T008: no plugin registered returns InvalidArgument.
func TestSupports_NoPluginRegisteredReturnsInvalidArgument(t *testing.T) {
	plugin := &mockSupportsPlugin{
		mockPlugin: mockPlugin{name: "test-plugin"},
		supported:  true,
	}
	// Empty registry - no plugins registered
	registry := &mockRegistry{
		plugins: map[string]string{},
	}
	server := NewServerWithRegistry(plugin, registry)

	req := &pbc.SupportsRequest{
		Resource: &pbc.ResourceDescriptor{
			Provider:     "aws",
			ResourceType: "ec2",
			Region:       "us-east-1",
		},
	}

	_, err := server.Supports(context.Background(), req)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "no plugin registered")
}

// TestSupports_PluginErrorReturnsInternalStatus tests T009: plugin error returns Internal status.
func TestSupports_PluginErrorReturnsInternalStatus(t *testing.T) {
	plugin := &mockSupportsPlugin{
		mockPlugin: mockPlugin{name: "test-plugin"},
		err:        errors.New("database connection failed"),
	}
	registry := &mockRegistry{
		plugins: map[string]string{"aws:us-east-1": "test-plugin"},
	}
	server := NewServerWithRegistry(plugin, registry)

	req := &pbc.SupportsRequest{
		Resource: &pbc.ResourceDescriptor{
			Provider:     "aws",
			ResourceType: "ec2",
			Region:       "us-east-1",
		},
	}

	_, err := server.Supports(context.Background(), req)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	// Error message should be generic, not expose internal details
	assert.Contains(t, st.Message(), "plugin failed to execute")
}

// TestSupports_PluginNotImplementsReturnsDefaultResponse tests T014: plugin without SupportsProvider returns default.
func TestSupports_PluginNotImplementsReturnsDefaultResponse(t *testing.T) {
	// Use mockPlugin which does NOT implement SupportsProvider
	plugin := &mockPlugin{name: "legacy-plugin"}
	registry := &mockRegistry{
		plugins: map[string]string{"aws:us-east-1": "legacy-plugin"},
	}
	server := NewServerWithRegistry(plugin, registry)

	req := &pbc.SupportsRequest{
		Resource: &pbc.ResourceDescriptor{
			Provider:     "aws",
			ResourceType: "ec2",
			Region:       "us-east-1",
		},
	}

	resp, err := server.Supports(context.Background(), req)

	require.NoError(t, err)
	assert.False(t, resp.GetSupported())
	assert.NotEmpty(t, resp.GetReason())
}

// TestSupports_DefaultResponseIncludesReason tests T015: default response includes reason explaining not implemented.
func TestSupports_DefaultResponseIncludesReason(t *testing.T) {
	plugin := &mockPlugin{name: "legacy-plugin"}
	registry := &mockRegistry{
		plugins: map[string]string{"aws:us-east-1": "legacy-plugin"},
	}
	server := NewServerWithRegistry(plugin, registry)

	req := &pbc.SupportsRequest{
		Resource: &pbc.ResourceDescriptor{
			Provider:     "aws",
			ResourceType: "ec2",
			Region:       "us-east-1",
		},
	}

	resp, err := server.Supports(context.Background(), req)

	require.NoError(t, err)
	assert.False(t, resp.GetSupported())
	assert.Contains(t, resp.GetReason(), "not implemented")
}

// TestSupports_NilResourceReturnsInvalidArgument tests that a nil resource descriptor returns InvalidArgument.
func TestSupports_NilResourceReturnsInvalidArgument(t *testing.T) {
	plugin := &mockSupportsPlugin{
		mockPlugin: mockPlugin{name: "test-plugin"},
		supported:  true,
	}
	registry := &mockRegistry{
		plugins: map[string]string{"aws:us-east-1": "test-plugin"},
	}
	server := NewServerWithRegistry(plugin, registry)

	// Request with nil resource
	req := &pbc.SupportsRequest{Resource: nil}

	_, err := server.Supports(context.Background(), req)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "resource descriptor is required")
}

// TestSupports_NewServerUsesDefaultRegistry tests that NewServer uses DefaultRegistryLookup.
func TestSupports_NewServerUsesDefaultRegistry(t *testing.T) {
	plugin := &mockSupportsPlugin{
		mockPlugin: mockPlugin{name: "test-plugin"},
		supported:  true,
	}
	// Use NewServer which uses DefaultRegistryLookup (always returns empty)
	server := NewServer(plugin)

	req := &pbc.SupportsRequest{
		Resource: &pbc.ResourceDescriptor{
			Provider:     "aws",
			ResourceType: "ec2",
			Region:       "us-east-1",
		},
	}

	_, err := server.Supports(context.Background(), req)

	// DefaultRegistryLookup returns "" for all lookups, so this should fail with InvalidArgument
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "no plugin registered")
}

// TestNewServerWithRegistry_NilRegistryUsesDefault tests that nil registry falls back to default.
func TestNewServerWithRegistry_NilRegistryUsesDefault(t *testing.T) {
	plugin := &mockSupportsPlugin{
		mockPlugin: mockPlugin{name: "test-plugin"},
		supported:  true,
	}
	// Pass nil registry - should fall back to DefaultRegistryLookup
	server := NewServerWithRegistry(plugin, nil)

	req := &pbc.SupportsRequest{
		Resource: &pbc.ResourceDescriptor{
			Provider:     "aws",
			ResourceType: "ec2",
			Region:       "us-east-1",
		},
	}

	_, err := server.Supports(context.Background(), req)

	// Should behave same as NewServer - DefaultRegistryLookup returns ""
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "no plugin registered")
}
