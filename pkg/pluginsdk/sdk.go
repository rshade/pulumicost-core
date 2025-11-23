// Package pluginsdk provides a development SDK for PulumiCost plugins.
package pluginsdk

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"

	pbc "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DefaultSupportsNotImplementedReason is the standardized message returned when
// a plugin does not implement the SupportsProvider interface.
const DefaultSupportsNotImplementedReason = "Supports capability not implemented by this plugin"

// Plugin represents a PulumiCost plugin implementation.
type Plugin interface {
	// Name returns the plugin name identifier.
	Name() string
	// GetProjectedCost calculates projected cost for a resource.
	GetProjectedCost(ctx context.Context, req *pbc.GetProjectedCostRequest) (*pbc.GetProjectedCostResponse, error)
	// GetActualCost retrieves actual cost for a resource.
	GetActualCost(ctx context.Context, req *pbc.GetActualCostRequest) (*pbc.GetActualCostResponse, error)
}

// SupportsProvider is an optional interface that plugins can implement to indicate
// whether they support pricing for specific resource types. Plugins that do not
// implement this interface will receive a default "not supported" response.
type SupportsProvider interface {
	// Supports checks if the plugin supports pricing for the given resource.
	Supports(ctx context.Context, req *pbc.SupportsRequest) (*pbc.SupportsResponse, error)
}

// RegistryLookup defines the interface for looking up plugins by provider and region.
// This is used to validate incoming Supports requests against registered plugins.
type RegistryLookup interface {
	// FindPlugin returns the plugin name for the given provider and region.
	// Returns empty string if no plugin is registered for the combination.
	FindPlugin(provider, region string) string
}

// DefaultRegistryLookup provides a no-op registry lookup that always returns empty.
// This causes all Supports() calls to return InvalidArgument since no plugin
// can be found. Use a real RegistryLookup implementation in production.
type DefaultRegistryLookup struct{}

// FindPlugin always returns empty string indicating no plugin is registered.
func (d *DefaultRegistryLookup) FindPlugin(_, _ string) string {
	return ""
}

// Server wraps a Plugin implementation with a gRPC server.
type Server struct {
	pbc.UnimplementedCostSourceServiceServer

	plugin   Plugin
	registry RegistryLookup
	logger   *slog.Logger
}

// NewServer creates a Server that exposes the provided Plugin over gRPC.
// Uses DefaultRegistryLookup which returns empty for all lookups, causing
// Supports() calls to return InvalidArgument. Use NewServerWithRegistry
// to provide a real registry for production use.
func NewServer(plugin Plugin) *Server {
	return &Server{
		plugin:   plugin,
		registry: &DefaultRegistryLookup{},
		logger:   slog.Default(),
	}
}

// NewServerWithRegistry creates a Server with a custom registry lookup.
// If registry is nil, DefaultRegistryLookup is used.
func NewServerWithRegistry(plugin Plugin, registry RegistryLookup) *Server {
	if registry == nil {
		registry = &DefaultRegistryLookup{}
	}
	return &Server{
		plugin:   plugin,
		registry: registry,
		logger:   slog.Default(),
	}
}

// Name implements the gRPC Name method.
func (s *Server) Name(_ context.Context, _ *pbc.NameRequest) (*pbc.NameResponse, error) {
	return &pbc.NameResponse{Name: s.plugin.Name()}, nil
}

// GetProjectedCost implements the gRPC GetProjectedCost method.
func (s *Server) GetProjectedCost(
	ctx context.Context,
	req *pbc.GetProjectedCostRequest,
) (*pbc.GetProjectedCostResponse, error) {
	return s.plugin.GetProjectedCost(ctx, req)
}

// GetActualCost implements the gRPC GetActualCost method.
func (s *Server) GetActualCost(ctx context.Context, req *pbc.GetActualCostRequest) (*pbc.GetActualCostResponse, error) {
	return s.plugin.GetActualCost(ctx, req)
}

// Supports implements the gRPC Supports method.
// It performs two-step validation: first checks registry for plugin by provider/region,
// then delegates to the plugin's Supports method if implemented.
func (s *Server) Supports(ctx context.Context, req *pbc.SupportsRequest) (*pbc.SupportsResponse, error) {
	// Validate request has resource descriptor
	if req.GetResource() == nil {
		return nil, status.Error(codes.InvalidArgument, "resource descriptor is required")
	}

	resource := req.GetResource()
	provider := resource.GetProvider()
	region := resource.GetRegion()

	// Step 1: Registry lookup - validate provider/region combination
	pluginName := s.registry.FindPlugin(provider, region)
	if pluginName == "" {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"no plugin registered for provider %q and region %q",
			provider,
			region,
		)
	}

	// Step 2: Check if plugin implements SupportsProvider
	supportsProvider, ok := s.plugin.(SupportsProvider)
	if !ok {
		// Plugin does not implement SupportsProvider - return default response
		return &pbc.SupportsResponse{
			Supported: false,
			Reason:    DefaultSupportsNotImplementedReason,
		}, nil
	}

	// Delegate to plugin's Supports method
	resp, err := supportsProvider.Supports(ctx, req)
	if err != nil {
		// Log the detailed error server-side for debugging
		s.logger.ErrorContext(ctx, "Supports handler error",
			"resource_type", resource.GetResourceType(),
			"provider", provider,
			"region", region,
			"error", err,
		)
		// Return generic message to client (internal error details not exposed)
		return nil, status.Error(codes.Internal, "plugin failed to execute")
	}

	return resp, nil
}

// ServeConfig holds configuration for serving a plugin.
type ServeConfig struct {
	Plugin   Plugin
	Port     int            // If 0, will use PORT env var or random port
	Registry RegistryLookup // Optional; if nil, DefaultRegistryLookup is used
}

func resolvePort(requested int) (int, error) {
	if requested > 0 {
		return requested, nil
	}
	portEnv := os.Getenv("PORT")
	if portEnv == "" {
		return 0, nil
	}
	value, err := strconv.Atoi(portEnv)
	if err != nil {
		if _, writeErr := fmt.Fprintf(
			os.Stderr,
			"Ignoring invalid PORT %q: %v; falling back to ephemeral port\n",
			portEnv,
			err,
		); writeErr != nil {
			return 0, fmt.Errorf("writing invalid port warning: %w", writeErr)
		}
		return 0, nil
	}
	return value, nil
}

func listenOnLoopback(ctx context.Context, port int) (net.Listener, *net.TCPAddr, error) {
	address := "127.0.0.1:0"
	if port > 0 {
		address = net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	}
	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", address)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen: %w", err)
	}

	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		closeErr := listener.Close()
		if closeErr != nil {
			return nil, nil, errors.Join(
				errors.New("listener address is not TCP"),
				fmt.Errorf("closing listener: %w", closeErr),
			)
		}
		return nil, nil, errors.New("listener address is not TCP")
	}

	return listener, tcpAddr, nil
}

func announcePort(listener net.Listener, addr *net.TCPAddr) error {
	if _, err := fmt.Fprintf(os.Stdout, "PORT=%d\n", addr.Port); err != nil {
		closeErr := listener.Close()
		if closeErr != nil {
			return errors.Join(
				fmt.Errorf("writing port: %w", err),
				fmt.Errorf("closing listener: %w", closeErr),
			)
		}
		return fmt.Errorf("writing port: %w", err)
	}
	return nil
}

// Serve starts the gRPC server for the provided plugin and prints the chosen port as PORT=<port> to stdout.
//
// It uses config.Port when > 0; if config.Port is 0 it attempts to parse the PORT environment variable and
// falls back to an ephemeral port when none is provided. The function registers the plugin's service, begins
// serving on the selected port, and performs a graceful stop when the context is cancelled.
//
// Returns an error if PORT cannot be parsed, if the listener cannot be created, or if the gRPC server fails to serve.

// Serve starts the plugin server and handles incoming gRPC requests.
func Serve(ctx context.Context, config ServeConfig) error {
	port, err := resolvePort(config.Port)
	if err != nil {
		return err
	}

	listener, tcpAddr, err := listenOnLoopback(ctx, port)
	if err != nil {
		return err
	}
	if announceErr := announcePort(listener, tcpAddr); announceErr != nil {
		return announceErr
	}

	// Create and register server
	grpcServer := grpc.NewServer()
	server := NewServerWithRegistry(config.Plugin, config.Registry)
	pbc.RegisterCostSourceServiceServer(grpcServer, server)

	// Handle context cancellation
	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()

	// Start serving
	err = grpcServer.Serve(listener)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return err
	}
	return nil
}
