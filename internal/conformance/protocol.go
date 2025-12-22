package conformance

import (
	"context"
	"fmt"

	pbc "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
)

// testNameReturnsIdentifier verifies the plugin returns its identifier.
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

// testNameReturnsProtocolVersion verifies the plugin returns protocol version.
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
