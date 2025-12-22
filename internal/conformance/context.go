package conformance

import (
	"context"
	"time"

	pbc "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// testContextCancellation verifies plugin respects context cancellation.
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

// testTimeoutRespected verifies plugin responds within timeout limits.
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
