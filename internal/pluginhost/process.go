package pluginhost

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/rshade/pulumicost-core/internal/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultTimeout    = 10 * time.Second
	connectionDelay   = 100 * time.Millisecond
	connectionTimeout = 100 * time.Millisecond
	processWaitDelay  = 100 * time.Millisecond // Time to wait for I/O after killing process
)

// ProcessLauncher launches plugins as separate TCP server processes.
type ProcessLauncher struct {
	timeout time.Duration
}

// NewProcessLauncher creates a new TCP process-based plugin launcher.
func NewProcessLauncher() *ProcessLauncher {
	return &ProcessLauncher{
		timeout: defaultTimeout,
	}
}

// Start launches a plugin process with TCP communication and returns the gRPC connection.
func (p *ProcessLauncher) Start(
	ctx context.Context,
	path string,
	args ...string,
) (*grpc.ClientConn, func() error, error) {
	log := logging.FromContext(ctx)
	log.Debug().
		Ctx(ctx).
		Str("component", "pluginhost").
		Str("operation", "start_plugin").
		Str("plugin_path", path).
		Msg("starting plugin process")

	port, err := p.allocatePort(ctx)
	if err != nil {
		log.Error().
			Ctx(ctx).
			Str("component", "pluginhost").
			Err(err).
			Msg("failed to allocate port for plugin")
		return nil, nil, err
	}

	log.Debug().
		Ctx(ctx).
		Str("component", "pluginhost").
		Int("port", port).
		Msg("allocated port for plugin")

	cmd, err := p.startPlugin(ctx, path, port, args)
	if err != nil {
		log.Error().
			Ctx(ctx).
			Str("component", "pluginhost").
			Err(err).
			Str("plugin_path", path).
			Int("port", port).
			Msg("failed to start plugin process")
		return nil, nil, err
	}

	log.Debug().
		Ctx(ctx).
		Str("component", "pluginhost").
		Int("pid", cmd.Process.Pid).
		Msg("plugin process started")

	conn, err := p.connectToPlugin(ctx, fmt.Sprintf("127.0.0.1:%d", port), cmd)
	if err != nil {
		log.Error().
			Ctx(ctx).
			Str("component", "pluginhost").
			Err(err).
			Str("address", fmt.Sprintf("127.0.0.1:%d", port)).
			Msg("failed to connect to plugin")
		return nil, nil, err
	}

	log.Info().
		Ctx(ctx).
		Str("component", "pluginhost").
		Str("plugin_path", path).
		Int("port", port).
		Int("pid", cmd.Process.Pid).
		Msg("plugin connected successfully")

	closeFn := p.createCloseFn(ctx, conn, cmd)
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
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(TraceInterceptor()))
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
		cmd.WaitDelay = processWaitDelay
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}
}

func (p *ProcessLauncher) createCloseFn(ctx context.Context, conn *grpc.ClientConn, cmd *exec.Cmd) func() error {
	return func() error {
		log := logging.FromContext(ctx)
		log.Debug().
			Ctx(ctx).
			Str("component", "pluginhost").
			Str("operation", "close_plugin").
			Msg("closing plugin connection")

		if err := conn.Close(); err != nil {
			log.Warn().
				Ctx(ctx).
				Str("component", "pluginhost").
				Err(err).
				Msg("error closing gRPC connection")
			return fmt.Errorf("closing connection: %w", err)
		}
		if cmd.Process != nil {
			pid := cmd.Process.Pid
			cmd.WaitDelay = processWaitDelay
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
			log.Debug().
				Ctx(ctx).
				Str("component", "pluginhost").
				Int("pid", pid).
				Msg("plugin process terminated")
		}
		return nil
	}
}
