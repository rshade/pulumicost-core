package conformance

import (
	"context"
	"fmt"

	pbc "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// testGetProjectedCostValid verifies GetProjectedCost returns cost for a valid resource.
func testGetProjectedCostValid(ctx *TestContext) *TestResult {
	client, ok := ctx.PluginClient.(pbc.CostSourceServiceClient)
	if !ok {
		return &TestResult{Status: StatusError, Error: "invalid plugin client type"}
	}

	// Create a standard test resource (e.g., t3.micro EC2)
	req := &pbc.GetProjectedCostRequest{
		Resource: &pbc.ResourceDescriptor{
			Provider:     "aws",
			ResourceType: "aws:ec2/instance:Instance",
			Sku:          "t3.micro",
			Region:       "us-east-1",
		},
	}

	resp, err := client.GetProjectedCost(context.Background(), req)
	if err != nil {
		return &TestResult{
			Status: StatusFail,
			Error:  fmt.Sprintf("GetProjectedCost() RPC failed: %v", err),
		}
	}

	if resp.GetCurrency() == "" {
		return &TestResult{Status: StatusFail, Error: "response missing currency"}
	}

	return &TestResult{
		Status:  StatusPass,
		Details: fmt.Sprintf("Estimated cost: %f %s", resp.GetCostPerMonth(), resp.GetCurrency()),
	}
}

// testGetProjectedCostInvalid verifies GetProjectedCost returns NotFound for unsupported resource.
func testGetProjectedCostInvalid(ctx *TestContext) *TestResult {
	client, ok := ctx.PluginClient.(pbc.CostSourceServiceClient)
	if !ok {
		return &TestResult{Status: StatusError, Error: "invalid plugin client type"}
	}

	// Create an invalid/unsupported resource
	req := &pbc.GetProjectedCostRequest{
		Resource: &pbc.ResourceDescriptor{
			Provider:     "aws",
			ResourceType: "invalid:resource",
			Sku:          "non-existent",
			Region:       "us-east-1",
		},
	}

	_, err := client.GetProjectedCost(context.Background(), req)
	if err == nil {
		return &TestResult{
			Status: StatusFail,
			Error:  "expected error for invalid resource, but got success",
		}
	}

	// Check if it's a gRPC status error
	st, ok := status.FromError(err)
	if !ok {
		return &TestResult{
			Status: StatusFail,
			Error:  fmt.Sprintf("expected gRPC status error, got: %v", err),
		}
	}

	if st.Code() != codes.NotFound && st.Code() != codes.InvalidArgument {
		return &TestResult{
			Status: StatusFail,
			Error:  fmt.Sprintf("expected NotFound or InvalidArgument, got: %s", st.Code()),
		}
	}

	return &TestResult{
		Status:  StatusPass,
		Details: fmt.Sprintf("Received expected error: %s", st.Code()),
	}
}
