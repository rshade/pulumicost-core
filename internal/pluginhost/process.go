package pluginhost

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultTimeout    = 10 * time.Second
	connectionDelay   = 100 * time.Millisecond
	connectionTimeout = 100 * time.Millisecond
)

type ProcessLauncher struct {
	timeout time.Duration
}

func NewProcessLauncher() *ProcessLauncher {
	return &ProcessLauncher{
		timeout: defaultTimeout,
	}
}

func (p *ProcessLauncher) Start(
	ctx context.Context,
	path string,
	args ...string,
) (*grpc.ClientConn, func() error, error) {
	port, err := p.allocatePort(ctx)
	if err != nil {
		return nil, nil, err
	}

	cmd, err := p.startPlugin(ctx, path, port, args)
	if err != nil {
		return nil, nil, err
	}

	conn, err := p.connectToPlugin(ctx, fmt.Sprintf("127.0.0.1:%d", port), cmd)
	if err != nil {
		return nil, nil, err
	}

	closeFn := p.createCloseFn(conn, cmd)
	return conn, closeFn, nil
}

func (p *ProcessLauncher) allocatePort(ctx context.Context) (int, error) {
	lc := &net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("creating listener: %w", err)
	}
	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return 0, errors.New("listener is not TCP address")
	}
	port := tcpAddr.Port
	if closeErr := listener.Close(); closeErr != nil {
		return 0, fmt.Errorf("closing listener: %w", closeErr)
	}
	return port, nil
}

func (p *ProcessLauncher) startPlugin(ctx context.Context, path string, port int, args []string) (*exec.Cmd, error) {
	//nolint:gosec // Plugin path is validated before execution
	cmd := exec.CommandContext(
		ctx,
		path,
		append(args, fmt.Sprintf("--port=%d", port))...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PULUMICOST_PLUGIN_PORT=%d", port))
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting plugin: %w", err)
	}
	return cmd, nil
}

func (p *ProcessLauncher) connectToPlugin(
	ctx context.Context,
	address string,
	cmd *exec.Cmd,
) (*grpc.ClientConn, error) {
	connCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	for {
		if connCtx.Err() != nil {
			p.killProcess(cmd)
			return nil, fmt.Errorf("timeout connecting to plugin: %w", connCtx.Err())
		}

		conn, err := p.tryConnect(address)
		if err != nil {
			time.Sleep(connectionDelay)
			continue
		}

		if p.isConnectionReady(connCtx, conn) {
			return conn, nil
		}

		_ = conn.Close()
		time.Sleep(connectionDelay)
	}
}

func (p *ProcessLauncher) tryConnect(address string) (*grpc.ClientConn, error) {
	return grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
}

func (p *ProcessLauncher) isConnectionReady(ctx context.Context, conn *grpc.ClientConn) bool {
	testCtx, testCancel := context.WithTimeout(ctx, connectionTimeout)
	defer testCancel()

	state := conn.GetState()
	if state == connectivity.Ready || state == connectivity.Idle {
		return true
	}

	conn.WaitForStateChange(testCtx, state)
	newState := conn.GetState()
	return newState == connectivity.Ready
}

func (p *ProcessLauncher) killProcess(cmd *exec.Cmd) {
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}

func (p *ProcessLauncher) createCloseFn(conn *grpc.ClientConn, cmd *exec.Cmd) func() error {
	return func() error {
		if err := conn.Close(); err != nil {
			return fmt.Errorf("closing connection: %w", err)
		}
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
		return nil
	}
}
