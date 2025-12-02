package pluginhost

import (
	"context"
	"fmt"

	"github.com/rshade/pulumicost-core/internal/proto"
	"google.golang.org/grpc"
)

// Client wraps a gRPC connection to a plugin and provides the cost source API.
type Client struct {
	Name  string
	Conn  *grpc.ClientConn
	API   proto.CostSourceClient
	Close func() error
}

// Launcher is an interface for different plugin launching strategies (TCP or stdio).
type Launcher interface {
	Start(ctx context.Context, path string, args ...string) (*grpc.ClientConn, func() error, error)
}

// NewClient creates a new plugin client by launching the plugin and establishing a gRPC connection.
func NewClient(ctx context.Context, launcher Launcher, binPath string) (*Client, error) {
	conn, closeFn, err := launcher.Start(ctx, binPath)
	if err != nil {
		return nil, err
	}

	api := proto.NewCostSourceClient(conn)

	nameResp, err := api.Name(ctx, &proto.Empty{})
	if err != nil {
		if closeErr := closeFn(); closeErr != nil {
			return nil, fmt.Errorf("getting plugin name: %w (close error: %w)", err, closeErr)
		}
		return nil, fmt.Errorf("getting plugin name: %w", err)
	}

	return &Client{
		Name:  nameResp.GetName(),
		Conn:  conn,
		API:   api,
		Close: closeFn,
	}, nil
}
