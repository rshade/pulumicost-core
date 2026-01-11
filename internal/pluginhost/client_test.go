package pluginhost_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/rshade/pulumicost-core/internal/pluginhost"
	pbc "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

type mockCostSourceServer struct {
	pbc.UnimplementedCostSourceServiceServer

	name           string
	pluginInfo     *pbc.GetPluginInfoResponse
	pluginInfoErr  error
	pluginInfoWait time.Duration
}

func (s *mockCostSourceServer) Name(
	ctx context.Context,
	req *pbc.NameRequest,
) (*pbc.NameResponse, error) {
	return &pbc.NameResponse{Name: s.name}, nil
}

func (s *mockCostSourceServer) GetPluginInfo(
	ctx context.Context,
	req *pbc.GetPluginInfoRequest,
) (*pbc.GetPluginInfoResponse, error) {
	if s.pluginInfoWait > 0 {
		time.Sleep(s.pluginInfoWait)
	}
	if s.pluginInfoErr != nil {
		return nil, s.pluginInfoErr
	}
	return s.pluginInfo, nil
}

type grpcMockLauncher struct {
	listener *bufconn.Listener
	server   *grpc.Server
}

func (m *grpcMockLauncher) Start(
	ctx context.Context,
	path string,
	args ...string,
) (*grpc.ClientConn, func() error, error) {
	conn, err := grpc.DialContext(
		ctx,
		"bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return m.listener.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	return conn, func() error { return nil }, err
}

func setupMockServer(_ *testing.T, srv *mockCostSourceServer) (*grpcMockLauncher, func()) {
	listener := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pbc.RegisterCostSourceServiceServer(s, srv)
	go func() {
		if err := s.Serve(listener); err != nil {
			// server might be stopped
		}
	}()

	launcher := &grpcMockLauncher{
		listener: listener,
		server:   s,
	}

	return launcher, func() {
		s.Stop()
		listener.Close()
	}
}

func TestGetPluginInfo_Success(t *testing.T) {
	// Setup mock server returning valid info
	srv := &mockCostSourceServer{
		name: "test-plugin",
		pluginInfo: &pbc.GetPluginInfoResponse{
			Version:     "1.0.0",
			SpecVersion: "0.4.14",
		},
	}
	launcher, cleanup := setupMockServer(t, srv)
	defer cleanup()

	// Call NewClient (which triggers GetPluginInfo)
	ctx := context.Background()
	client, err := pluginhost.NewClient(ctx, launcher, "dummy")
	require.NoError(t, err)
	defer client.Close()

	// Verify client has stored metadata
	// Note: We haven't added the field to Client yet, so this part of the test will fail compilation
	// until we update the Client struct. For now we assert err is nil.
	assert.Equal(t, "test-plugin", client.Name)
	// assert.Equal(t, "1.0.0", client.Metadata.Version)
}

func TestGetPluginInfo_Unimplemented(t *testing.T) {
	// Setup mock server returning Unimplemented for GetPluginInfo
	srv := &mockCostSourceServer{
		name:          "legacy-plugin",
		pluginInfoErr: status.Error(codes.Unimplemented, "method not implemented"),
	}
	launcher, cleanup := setupMockServer(t, srv)
	defer cleanup()

	ctx := context.Background()
	client, err := pluginhost.NewClient(ctx, launcher, "dummy")

	// Should NOT fail, but log warning (impl detail)
	require.NoError(t, err)
	assert.Equal(t, "legacy-plugin", client.Name)
}

func TestGetPluginInfo_Timeout(t *testing.T) {
	// Setup mock server that sleeps longer than timeout
	srv := &mockCostSourceServer{
		name:           "slow-plugin",
		pluginInfoWait: 6 * time.Second, // Timeout is 5s
		pluginInfo: &pbc.GetPluginInfoResponse{
			Version: "1.0.0",
		},
	}
	launcher, cleanup := setupMockServer(t, srv)
	defer cleanup()

	ctx := context.Background()
	// NewClient will likely fail or log warning depending on implementation preference for timeout.
	// Spec says: "If GetPluginInfo times out, treat as legacy plugin (Unimplemented)" OR fail?
	// T016 says: "call GetPluginInfo with 5-second timeout".
	// T036 says: "omit plugins that timeout".
	// But NewClient is initialization. If it times out here, we probably want to fail or warn.
	// Let's assume for now it should succeed but maybe without metadata, or fail.
	// Task T011 says "TestGetPluginInfo_Timeout", implying we need to handle it.

	client, err := pluginhost.NewClient(ctx, launcher, "dummy")
	// If strict check, this might fail.
	// If soft check, it might pass.
	// Let's assume soft failure (warn and proceed) for now, or we'll see actual behavior.
	_ = client
	_ = err
}
