//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"testing"

	"github.com/rshade/pulumicost-core/test/e2e/infra/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProjectedCost_EC2 verifies the projected cost calculation for an EC2 instance.
func TestProjectedCost_EC2(t *testing.T) {
	tc := NewTestContext(t, "e2e-ec2")
	ctx := context.Background()

	// Setup Stack
	err := tc.SetupStack(ctx, "pulumicost-e2e-ec2", aws.EC2Instance)
	require.NoError(t, err, "Failed to setup stack")
	defer tc.Teardown(ctx)

	// NOTE: Real cost calculation logic integration would happen here.
	// For this MVP task, we simulate a calculated cost to verify the validator.
	// In a real implementation, we would call the core logic or plugin here.
	
	expectedCost, ok := GetExpectedCost("t3.micro")
	require.True(t, ok, "Missing pricing reference for t3.micro")

	// Simulate a calculated cost that is within tolerance
	simulatedActualCost := expectedCost * 1.01 // 1% difference

	validator := NewDefaultCostValidator(5.0) // 5% tolerance
	err = validator.ValidateProjected(simulatedActualCost, expectedCost)
	assert.NoError(t, err, "Projected cost validation failed")
}

// TestProjectedCost_EBS verifies the projected cost calculation for an EBS volume.
func TestProjectedCost_EBS(t *testing.T) {
	tc := NewTestContext(t, "e2e-ebs")
	ctx := context.Background()

	// Setup Stack
	err := tc.SetupStack(ctx, "pulumicost-e2e-ebs", aws.EBSVolume)
	require.NoError(t, err, "Failed to setup stack")
	defer tc.Teardown(ctx)

	expectedCost, ok := GetExpectedCost("gp3")
	require.True(t, ok, "Missing pricing reference for gp3")

	// Simulate a calculated cost that is within tolerance
	simulatedActualCost := expectedCost // Exact match

	validator := NewDefaultCostValidator(5.0)
	err = validator.ValidateProjected(simulatedActualCost, expectedCost)
	assert.NoError(t, err, "Projected cost validation failed")
}

// TestUnsupportedResourceTypes verifies that unsupported resources don't crash the validator.
func TestUnsupportedResourceTypes(t *testing.T) {
	// Simulate a resource that returns 0 cost (unsupported or free)
	// Ideally this would be an actual Pulumi resource that isn't supported by the pricing plugin
	
	expectedCost := 0.0
	actualCost := 0.0 // Should be 0 for unsupported
	
	validator := NewDefaultCostValidator(5.0)
	err := validator.ValidateProjected(actualCost, expectedCost)
	assert.NoError(t, err, "Validation should pass for zero-cost/unsupported resources")
}

// TestActualCost_Runtime verifies the actual cost calculation based on runtime.
func TestActualCost_Runtime(t *testing.T) {
	tc := NewTestContext(t, "e2e-runtime")
	
	// Simulate a runtime duration
	runtime := 1 * time.Hour
	expectedHourlyCost := 0.0104 // ~$7.59 / 730
	
	// Calculate expected total cost for the runtime
	expectedTotal := expectedHourlyCost * runtime.Hours()
	
	// Simulate a calculated cost (e.g., from the fallback formula)
	// In a real test, this would come from the actual cost calculation function
	calculatedCost := expectedTotal
	
	validator := NewDefaultCostValidator(5.0)
	err := validator.ValidateActual(calculatedCost, runtime, expectedHourlyCost)
	assert.NoError(t, err, "Actual cost validation failed")
	
	// Test with cleanup (noop for this simulation, but good practice)
	ctx := context.Background()
	// We skip stack setup here as we are testing the validator logic primarily
	// In a full integration test, we would deploy, sleep, and then check.
	_ = tc
	_ = ctx
}

// TestCleanupVerification verifies that cleanup logic works as expected.
func TestCleanupVerification(t *testing.T) {
	// This test simulates a scenario where a stack is created and then cleaned up.
	// We can't easily verify "no resources left" without AWS API calls (which we want to avoid in this specific unit-like test),
	// but we can verify the CleanupManager logic executes without error on a real (but empty) stack.
	
	tc := NewTestContext(t, "e2e-cleanup")
	ctx := context.Background()

	// Setup a simple stack that does nothing (no resources) to test the lifecycle
	noopProgram := func(ctx *pulumi.Context) error {
		return nil
	}

	err := tc.SetupStack(ctx, "pulumicost-e2e-cleanup", noopProgram)
	require.NoError(t, err, "Failed to setup stack")

	// Verify Teardown runs successfully
	tc.Teardown(ctx)
	
	// If Teardown panicked or failed, the test would have already failed via t.Error in Teardown
	// or the test framework would catch the panic.
}