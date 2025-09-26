// Package pluginsdk provides a development SDK for PulumiCost plugins.
package pluginsdk

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"

	pbc "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
	"google.golang.org/grpc"
)

// Plugin represents a PulumiCost plugin implementation.
type Plugin interface {
	// Name returns the plugin name identifier.
	Name() string
	// GetProjectedCost calculates projected cost for a resource.
	GetProjectedCost(ctx context.Context, req *pbc.GetProjectedCostRequest) (*pbc.GetProjectedCostResponse, error)
	// GetActualCost retrieves actual cost for a resource.
	GetActualCost(ctx context.Context, req *pbc.GetActualCostRequest) (*pbc.GetActualCostResponse, error)
}

// Server wraps a Plugin implementation with a gRPC server.
type Server struct {
	pbc.UnimplementedCostSourceServiceServer

	plugin Plugin
}

// NewServer creates a Server that exposes the provided Plugin over gRPC.
func NewServer(plugin Plugin) *Server {
	return &Server{plugin: plugin}
}

// Name implements the gRPC Name method.
func (s *Server) Name(ctx context.Context, req *pbc.NameRequest) (*pbc.NameResponse, error) {
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

// ServeConfig holds configuration for serving a plugin.
type ServeConfig struct {
	Plugin Plugin
	Port   int // If 0, will use PORT env var or random port
}

// Serve starts the gRPC server for the provided plugin and prints the chosen port as PORT=<port> to stdout.
// 
// It uses config.Port when > 0; if config.Port is 0 it attempts to parse the PORT environment variable and
// falls back to an ephemeral port when none is provided. The function registers the plugin's service, begins
// serving on the selected port, and performs a graceful stop when the context is cancelled.
// 
// Returns an error if PORT cannot be parsed, if the listener cannot be created, or if the gRPC server fails to serve.
func Serve(ctx context.Context, config ServeConfig) error {
	// Determine port
	port := config.Port
	if port == 0 {
		if portEnv := os.Getenv("PORT"); portEnv != "" {
			var err error
			port, err = strconv.Atoi(portEnv)
			if err != nil {
				return fmt.Errorf("invalid PORT environment variable: %w", err)
			}
		}
	}

	// Create listener
	var listener net.Listener
	var err error
	if port > 0 {
		listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
	} else {
		listener, err = net.Listen("tcp", ":0")
	}
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// Print port for plugin host discovery
	addr := listener.Addr().(*net.TCPAddr)
	fmt.Printf("PORT=%d\n", addr.Port)

	// Create and register server
	grpcServer := grpc.NewServer()
	server := NewServer(config.Plugin)
	pbc.RegisterCostSourceServiceServer(grpcServer, server)

	// Handle context cancellation
	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()

	// Start serving
	return grpcServer.Serve(listener)
}
