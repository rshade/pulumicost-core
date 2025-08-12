package pluginhost

import (
	"context"
	"fmt"

	"github.com/rshade/pulumicost-core/internal/proto"
	"google.golang.org/grpc"
)

type Client struct {
	Name  string
	Conn  *grpc.ClientConn
	API   proto.CostSourceClient
	Close func() error
}

type Launcher interface {
	Start(ctx context.Context, path string, args ...string) (*grpc.ClientConn, func() error, error)
}

func NewClient(ctx context.Context, launcher Launcher, binPath string) (*Client, error) {
	conn, closeFn, err := launcher.Start(ctx, binPath)
	if err != nil {
		return nil, err
	}
	
	api := proto.NewCostSourceClient(conn)
	
	nameResp, err := api.Name(ctx, &proto.Empty{})
	if err != nil {
		closeFn()
		return nil, fmt.Errorf("getting plugin name: %w", err)
	}
	
	return &Client{
		Name:  nameResp.GetName(),
		Conn:  conn,
		API:   api,
		Close: closeFn,
	}, nil
}
