package conformance

import (
	"context"
	"fmt"

	pbc "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
)

// testBatchHandling verifies plugin handles multiple resources in a single RPC call if supported.
// Note: Currently GetProjectedCost handles single resource per RPC in v1 proto,
// but we might add a batch RPC later. For now, we'll verify it handles multiple
// testBatchHandling verifies that the plugin can process a sequence of GetProjectedCost
// requests without failing and returns a TestResult summarizing the outcome.
// 
// It expects ctx.PluginClient to implement pbc.CostSourceServiceClient; if the assertion
// fails it returns a TestResult with StatusError. The function issues a fixed number of
// GetProjectedCost requests for a representative AWS EC2 resource, counts successful RPCs,
// and returns a TestResult with StatusPass when all requests succeed or StatusFail when
// one or more requests fail.
// 
// Parameters:
//   - ctx: the TestContext providing the plugin client used to perform RPC calls.
// 
// Returns:
//   - *TestResult describing whether the batch of sequential calls all succeeded, or an
//     error result if the plugin client is of the wrong type.
func testBatchHandling(ctx *TestContext) *TestResult {
	client, ok := ctx.PluginClient.(pbc.CostSourceServiceClient)
	if !ok {
		return &TestResult{Status: StatusError, Error: "invalid plugin client type"}
	}

	const batchSize = 5
	successCount := 0

	for range batchSize {
		req := &pbc.GetProjectedCostRequest{
			Resource: &pbc.ResourceDescriptor{
				Provider:     "aws",
				ResourceType: "aws:ec2/instance:Instance",
				Sku:          "t3.micro",
				Region:       "us-east-1",
			},
		}

		_, err := client.GetProjectedCost(context.Background(), req)
		if err == nil {
			successCount++
		}
	}

	if successCount != batchSize {
		return &TestResult{
			Status: StatusFail,
			Error: fmt.Sprintf(
				"Batch handling failed: %d/%d calls succeeded",
				successCount,
				batchSize,
			),
		}
	}

	return &TestResult{
		Status:  StatusPass,
		Details: fmt.Sprintf("Successfully processed %d sequential calls", batchSize),
	}
}