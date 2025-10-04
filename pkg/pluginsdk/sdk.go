// Package pluginsdk provides a development SDK for PulumiCost plugins.
package pluginsdk

import (
	"context"
	"errors"
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

// ServeConfig holds configuration for serving a plugin.
type ServeConfig struct {
	Plugin Plugin
	Port   int // If 0, will use PORT env var or random port
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
	server := NewServer(config.Plugin)
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
