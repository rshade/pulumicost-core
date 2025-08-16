package pluginhost

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	stdioTimeout = 10 * time.Second
)

type StdioLauncher struct {
	timeout time.Duration
}

func NewStdioLauncher() *StdioLauncher {
	return &StdioLauncher{
		timeout: stdioTimeout,
	}
}

func (s *StdioLauncher) Start(
	ctx context.Context,
	path string,
	args ...string,
) (*grpc.ClientConn, func() error, error) {
	//nolint:gosec // Plugin path is validated before execution
	cmd := exec.CommandContext(
		ctx,
		path,
		append(args, "--stdio")...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("creating stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("creating stdout pipe: %w", err)
	}

	cmd.Stderr = os.Stderr

	if startErr := cmd.Start(); startErr != nil {
		return nil, nil, fmt.Errorf("starting plugin: %w", startErr)
	}

	lc := &net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", "127.0.0.1:0")
	if err != nil {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		return nil, nil, fmt.Errorf("creating proxy listener: %w", err)
	}

	go s.proxy(listener, stdin, stdout)

	address := listener.Addr().String()

	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = listener.Close()
		return nil, nil, fmt.Errorf("connecting to plugin: %w", err)
	}

	closeFn := func() error {
		if connCloseErr := conn.Close(); connCloseErr != nil {
			return fmt.Errorf("closing connection: %w", connCloseErr)
		}
		if listenerCloseErr := listener.Close(); listenerCloseErr != nil {
			return fmt.Errorf("closing listener: %w", listenerCloseErr)
		}
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
		return nil
	}

	return conn, closeFn, nil
}

func (s *StdioLauncher) proxy(listener net.Listener, stdin io.WriteCloser, stdout io.ReadCloser) {
	conn, err := listener.Accept()
	if err != nil {
		return
	}
	defer conn.Close()

	go func() {
		_, _ = io.Copy(stdin, conn)
	}()
	_, _ = io.Copy(conn, stdout)
}
