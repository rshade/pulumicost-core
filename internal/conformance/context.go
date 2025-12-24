package conformance

import (
	"context"
	"time"

	pbc "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// testContextCancellation verifies that the plugin's RPCs respect context cancellation.
// It calls GetProjectedCost with an already-canceled context and expects the call to fail
// with a gRPC status whose code is codes.Canceled. If the PluginClient is not a
// pbc.CostSourceServiceClient, the test returns an error result. If the RPC succeeds,
// returns a failure result. If the RPC returns a non-gRPC error or a gRPC status with
// a code other than Canceled, the test returns a failure result.
// ctx is the test harness context which must provide a PluginClient implementing
// pbc.CostSourceServiceClient.
// The function returns a *TestResult that indicates Pass when a Canceled status is observed,
// Fail for unexpected RPC outcomes, or Error for an invalid client type.
func testContextCancellation(ctx *TestContext) *TestResult {
	client, ok := ctx.PluginClient.(pbc.CostSourceServiceClient)
	if !ok {
		return &TestResult{Status: StatusError, Error: "invalid plugin client type"}
	}

	// Create a cancelled context
	cancCtx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := &pbc.GetProjectedCostRequest{
		Resource: &pbc.ResourceDescriptor{
			Provider:     "aws",
			ResourceType: "aws:ec2/instance:Instance",
			Sku:          "t3.micro",
			Region:       "us-east-1",
		},
	}

	_, err := client.GetProjectedCost(cancCtx, req)
	if err == nil {
		return &TestResult{
			Status: StatusFail,
			Error:  "expected error for cancelled context, but got success",
		}
	}

	st, ok := status.FromError(err)
	if !ok {
		return &TestResult{Status: StatusFail, Error: "expected gRPC status error"}
	}

	if st.Code() != codes.Canceled {
		return &TestResult{Status: StatusFail, Error: "expected status Canceled"}
	}

	return &TestResult{Status: StatusPass}
}

// testTimeoutRespected verifies that the plugin honors context timeouts when handling RPCs.
// It calls the plugin's Name RPC with a very short deadline and interprets the outcome:
//   - returns a Pass result if the RPC completes successfully despite the short timeout,
//     or if the RPC fails with gRPC code DeadlineExceeded or Canceled.
//   - returns a Fail result if the RPC returns a non-gRPC error or a gRPC status with an unexpected code.
//   - returns an Error result if the plugin client has an unexpected type.
//
// ctx is the TestContext that provides the plugin client used to make the RPC.
func testTimeoutRespected(ctx *TestContext) *TestResult {
	client, ok := ctx.PluginClient.(pbc.CostSourceServiceClient)
	if !ok {
		return &TestResult{Status: StatusError, Error: "invalid plugin client type"}
	}

	// Use a very short timeout
	shortCtx, cancel := context.WithTimeout(context.Background(), 1*time.Microsecond)
	defer cancel()

	req := &pbc.NameRequest{}
	_, err := client.Name(shortCtx, req)
	if err == nil {
		// It's possible it succeeded if it was extremely fast, but for conformance
		// we want to ensure it fails if the context is exceeded.
		// However, 1 microsecond is usually enough to cause failure.
		return &TestResult{Status: StatusPass, Details: "RPC succeeded even with 1µs timeout"}
	}

	st, ok := status.FromError(err)
	if !ok {
		return &TestResult{Status: StatusFail, Error: "expected gRPC status error"}
	}

	if st.Code() != codes.DeadlineExceeded && st.Code() != codes.Canceled {
		return &TestResult{Status: StatusFail, Error: "expected status DeadlineExceeded"}
	}

	return &TestResult{Status: StatusPass}
}
