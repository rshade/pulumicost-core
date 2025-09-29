package pluginsdk

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	pbc "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const defaultTestTimeout = 5 * time.Second

// TestServer provides utilities for testing plugins.
type TestServer struct {
	t        *testing.T
	server   *grpc.Server
	listener net.Listener
	client   pbc.CostSourceServiceClient
	conn     *grpc.ClientConn
}

// NewTestServer spins up a temporary gRPC server for a plugin and returns a TestServer
// wrapper that exposes a typed client and cleanup helpers. The cleanup is registered
// with the provided testing.T via TestPlugin and reported failures stop the test.
func NewTestServer(t *testing.T, plugin Plugin) *TestServer {
	t.Helper()

	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pluginServer := NewServer(plugin)
	pbc.RegisterCostSourceServiceServer(server, pluginServer)

	go func() {
		if serveErr := server.Serve(listener); serveErr != nil && !errors.Is(serveErr, grpc.ErrServerStopped) {
			var netErr net.Error
			if errors.As(serveErr, &netErr) && !netErr.Timeout() {
				t.Logf("test server error: %v", serveErr)
			}
		}
	}()

	conn, err := grpc.NewClient(
		listener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		server.GracefulStop()
		if closeErr := listener.Close(); closeErr != nil {
			t.Logf("closing listener after dial setup failure: %v", closeErr)
		}
		t.Fatalf("Failed to create client: %v", err)
	}

	conn.Connect()
	waitCtx, waitCancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer waitCancel()
	for state := conn.GetState(); state != connectivity.Ready; state = conn.GetState() {
		if !conn.WaitForStateChange(waitCtx, state) {
			server.GracefulStop()
			if closeErr := listener.Close(); closeErr != nil {
				t.Logf("closing listener after connect timeout: %v", closeErr)
			}
			t.Fatalf("Failed to connect: %v", waitCtx.Err())
		}
	}

	client := pbc.NewCostSourceServiceClient(conn)

	return &TestServer{
		t:        t,
		server:   server,
		listener: listener,
		client:   client,
		conn:     conn,
	}
}

// Client returns the gRPC client for testing.
func (ts *TestServer) Client() pbc.CostSourceServiceClient {
	return ts.client
}

// Close stops the test server and cleans up resources.
func (ts *TestServer) Close() {
	if ts == nil {
		return
	}
	var errs error
	if ts.conn != nil {
		if err := ts.conn.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			errs = errors.Join(errs, fmt.Errorf("closing client connection: %w", err))
		}
	}
	if ts.server != nil {
		ts.server.GracefulStop()
	}
	if ts.listener != nil {
		if err := ts.listener.Close(); err != nil {
			errs = errors.Join(errs, fmt.Errorf("closing listener: %w", err))
		}
	}
	if errs != nil && ts.t != nil {
		ts.t.Helper()
		ts.t.Fatalf("failed closing test server: %v", errs)
	}
}

// TestPlugin provides test utilities for plugin implementations.
type TestPlugin struct {
	*testing.T

	server *TestServer
	client pbc.CostSourceServiceClient
}

// NewTestPlugin creates a TestPlugin backed by an in-process gRPC TestServer for the
// provided plugin and registers cleanup to stop the server when the test finishes.
//
// The returned TestPlugin contains the testing.T, the created TestServer, and a
// CostSourceServiceClient connected to that server.
func NewTestPlugin(t *testing.T, plugin Plugin) *TestPlugin {
	t.Helper()

	server := NewTestServer(t, plugin)

	t.Cleanup(func() {
		server.Close()
	})

	return &TestPlugin{
		T:      t,
		server: server,
		client: server.Client(),
	}
}

// TestName tests the plugin's Name method.
func (tp *TestPlugin) TestName(expectedName string) {
	tp.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()

	resp, err := tp.client.Name(ctx, &pbc.NameRequest{})
	if err != nil {
		tp.Fatalf("Name() failed: %v", err)
	}

	if resp.GetName() != expectedName {
		tp.Errorf("Expected name %q, got %q", expectedName, resp.GetName())
	}
}

// TestProjectedCost tests a projected cost calculation.
func (tp *TestPlugin) TestProjectedCost(
	resource *pbc.ResourceDescriptor,
	expectError bool,
) *pbc.GetProjectedCostResponse {
	tp.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()

	req := &pbc.GetProjectedCostRequest{Resource: resource}
	resp, err := tp.client.GetProjectedCost(ctx, req)

	if expectError {
		if err == nil {
			tp.Errorf("Expected error for resource %v, but got none", resource)
		}
		return nil
	}

	if err != nil {
		tp.Fatalf("GetProjectedCost() failed: %v", err)
	}

	// Basic validation
	if resp.GetCurrency() == "" {
		tp.Errorf("Response missing currency")
	}
	if resp.GetUnitPrice() < 0 {
		tp.Errorf("Negative unit price: %f", resp.GetUnitPrice())
	}

	return resp
}

// TestActualCost tests an actual cost retrieval.
func (tp *TestPlugin) TestActualCost(
	resourceID string,
	startTime, endTime int64,
	expectError bool,
) *pbc.GetActualCostResponse {
	tp.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()

	req := &pbc.GetActualCostRequest{
		ResourceId: resourceID,
		Start:      timestamppb.New(time.Unix(startTime, 0)),
		End:        timestamppb.New(time.Unix(endTime, 0)),
		Tags:       make(map[string]string),
	}

	resp, err := tp.client.GetActualCost(ctx, req)

	if expectError {
		if err == nil {
			tp.Errorf("Expected error for resource ID %s, but got none", resourceID)
		}
		return nil
	}

	if err != nil {
		tp.Fatalf("GetActualCost() failed: %v", err)
	}

	// Basic validation
	results := resp.GetResults()
	if len(results) == 0 {
		tp.Errorf("No results returned")
	}

	for _, result := range results {
		if result.GetCost() < 0 {
			tp.Errorf("Negative cost in result: %f", result.GetCost())
		}
	}

	return resp
}

// CreateTestResource creates a ResourceDescriptor for tests with the given provider and resource type.
// If properties is nil, an empty tag map is created and assigned to the descriptor's Tags field.
func CreateTestResource(provider, resourceType string, properties map[string]string) *pbc.ResourceDescriptor {
	if properties == nil {
		properties = make(map[string]string)
	}

	return &pbc.ResourceDescriptor{
		Provider:     provider,
		ResourceType: resourceType,
		Tags:         properties,
	}
}
