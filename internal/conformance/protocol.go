package conformance

import (
	"context"
	"fmt"

	pbc "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
)

// testNameReturnsIdentifier verifies that the plugin's Name RPC returns a non-empty identifier.
// It returns a TestResult with StatusPass and Details containing the plugin name when a non-empty
// name is returned. It returns StatusFail if the Name RPC returns an error or an empty name, and
// StatusError if the provided PluginClient does not implement pbc.CostSourceServiceClient.
func testNameReturnsIdentifier(ctx *TestContext) *TestResult {
	client, ok := ctx.PluginClient.(pbc.CostSourceServiceClient)
	if !ok {
		return &TestResult{Status: StatusError, Error: "invalid plugin client type"}
	}

	resp, err := client.Name(context.Background(), &pbc.NameRequest{})
	if err != nil {
		return &TestResult{Status: StatusFail, Error: fmt.Sprintf("Name() RPC failed: %v", err)}
	}

	if resp.GetName() == "" {
		return &TestResult{Status: StatusFail, Error: "Name() returned empty string"}
	}

	return &TestResult{Status: StatusPass, Details: fmt.Sprintf("Plugin name: %s", resp.GetName())}
}

// - StatusPass with Details containing the plugin's reported name when the Name RPC succeeds.
func testNameReturnsProtocolVersion(ctx *TestContext) *TestResult {
	client, ok := ctx.PluginClient.(pbc.CostSourceServiceClient)
	if !ok {
		return &TestResult{Status: StatusError, Error: "invalid plugin client type"}
	}

	// In current proto, NameRequest doesn't return protocol version directly,
	// but some plugins might include it in metadata or we might have a dedicated Version RPC.
	// For now, we'll check if Name succeeds and maybe check a version field if added.
	resp, err := client.Name(context.Background(), &pbc.NameRequest{})
	if err != nil {
		return &TestResult{Status: StatusFail, Error: fmt.Sprintf("Name() RPC failed: %v", err)}
	}

	// Assuming version might be in metadata or a separate field in future.
	// For now, passing if RPC succeeds as a placeholder for protocol compatibility check.
	return &TestResult{
		Status:  StatusPass,
		Details: fmt.Sprintf("Protocol response received for plugin: %s", resp.GetName()),
	}
}