package conformance

import (
	"context"
	"fmt"

	pbc "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// testGetProjectedCostPermissionDenied verifies that GetProjectedCost returns PermissionDenied
// when requesting a forbidden resource.
func testGetProjectedCostPermissionDenied(ctx *TestContext) *TestResult {
	return testGetProjectedCostErrorCode(ctx, "forbidden:resource", codes.PermissionDenied)
}

// testGetProjectedCostInternal verifies that GetProjectedCost returns Internal
// when an internal error occurs.
func testGetProjectedCostInternal(ctx *TestContext) *TestResult {
	return testGetProjectedCostErrorCode(ctx, "error:internal", codes.Internal)
}

// testGetProjectedCostUnavailable verifies that GetProjectedCost returns Unavailable
// when the service is unavailable.
func testGetProjectedCostUnavailable(ctx *TestContext) *TestResult {
	return testGetProjectedCostErrorCode(ctx, "error:unavailable", codes.Unavailable)
}

// testGetProjectedCostErrorCode is a helper that checks for a specific gRPC error code.
func testGetProjectedCostErrorCode(ctx *TestContext, resourceType string, expectedCode codes.Code) *TestResult {
	client, ok := ctx.PluginClient.(pbc.CostSourceServiceClient)
	if !ok {
		return &TestResult{Status: StatusError, Error: "invalid plugin client type"}
	}

	req := &pbc.GetProjectedCostRequest{
		Resource: &pbc.ResourceDescriptor{
			Provider:     "aws",
			ResourceType: resourceType,
			Sku:          "any",
			Region:       "us-east-1",
		},
	}

	rpcCtx, cancel := context.WithTimeout(context.Background(), ctx.Timeout)
	defer cancel()

	_, err := client.GetProjectedCost(rpcCtx, req)
	if err == nil {
		return &TestResult{
			Status: StatusFail,
			Error:  fmt.Sprintf("expected error %s for resource %s, but got success", expectedCode, resourceType),
		}
	}

	st, ok := status.FromError(err)
	if !ok {
		return &TestResult{
			Status: StatusFail,
			Error:  fmt.Sprintf("expected gRPC status error, got: %v", err),
		}
	}

	if st.Code() != expectedCode {
		return &TestResult{
			Status: StatusFail,
			Error:  fmt.Sprintf("expected %s, got: %s", expectedCode, st.Code()),
		}
	}

	return &TestResult{
		Status:  StatusPass,
		Details: fmt.Sprintf("Received expected error: %s", st.Code()),
	}
}
