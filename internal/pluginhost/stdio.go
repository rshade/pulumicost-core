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

type StdioLauncher struct {
	timeout time.Duration
}

func NewStdioLauncher() *StdioLauncher {
	return &StdioLauncher{
		timeout: 10 * time.Second,
	}
}

func (s *StdioLauncher) Start(ctx context.Context, path string, args ...string) (*grpc.ClientConn, func() error, error) {
	cmd := exec.CommandContext(ctx, path, append(args, "--stdio")...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("creating stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("creating stdout pipe: %w", err)
	}

	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("starting plugin: %w", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		cmd.Process.Kill()
		return nil, nil, fmt.Errorf("creating proxy listener: %w", err)
	}

	go s.proxy(listener, stdin, stdout)

	address := listener.Addr().String()

	connCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	conn, err := grpc.DialContext(connCtx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		cmd.Process.Kill()
		listener.Close()
		return nil, nil, fmt.Errorf("connecting to plugin: %w", err)
	}

	closeFn := func() error {
		conn.Close()
		listener.Close()
		if cmd.Process != nil {
			cmd.Process.Kill()
			cmd.Wait()
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

	go io.Copy(stdin, conn)
	io.Copy(conn, stdout)
}
