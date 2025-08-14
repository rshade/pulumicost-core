package pluginhost

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ProcessLauncher struct {
	timeout time.Duration
}

func NewProcessLauncher() *ProcessLauncher {
	return &ProcessLauncher{
		timeout: 10 * time.Second,
	}
}

func (p *ProcessLauncher) Start(ctx context.Context, path string, args ...string) (*grpc.ClientConn, func() error, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, nil, fmt.Errorf("creating listener: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	cmd := exec.CommandContext(ctx, path, append(args, fmt.Sprintf("--port=%d", port))...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PULUMICOST_PLUGIN_PORT=%d", port))
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("starting plugin: %w", err)
	}

	address := fmt.Sprintf("127.0.0.1:%d", port)
	
	connCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	var conn *grpc.ClientConn
	for {
		conn, err = grpc.DialContext(connCtx, address,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock())
		if err == nil {
			break
		}
		if connCtx.Err() != nil {
			cmd.Process.Kill()
			return nil, nil, fmt.Errorf("timeout connecting to plugin: %w", err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	closeFn := func() error {
		conn.Close()
		if cmd.Process != nil {
			cmd.Process.Kill()
			cmd.Wait()
		}
		return nil
	}

	return conn, closeFn, nil
}