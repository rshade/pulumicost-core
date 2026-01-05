//go:build e2e
// +build e2e

package aws

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rshade/pulumicost-core/test/e2e"
	"github.com/stretchr/testify/require"
)

func TestAWSCostValidation(t *testing.T) {
	skip, reason := e2e.ShouldSkip("aws")
	if skip {
		t.Skip(reason)
	}

	config := e2e.LoadConfig()
	require.NotNil(t, config)

	// Load expected costs
	expectedCosts := LoadExpectedCosts(t)

	t.Logf("Running AWS cost validation with tolerance %f", config.Tolerance)

	// Set up test context with timeout appropriate for infrastructure provisioning
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	tc := e2e.NewTestContext(t, "aws-cost")
	defer tc.Teardown(ctx)

	// Set up the EC2 project
	projectPath := filepath.Join("..", "projects", "ec2")
	err := tc.SetupProject(ctx, projectPath)
	require.NoError(t, err, "Failed to set up EC2 project")

	// Run pulumicost to get actual cost
	actualCost, err := tc.RunPulumicost(ctx)
	require.NoError(t, err, "Failed to run pulumicost")

	t.Logf("Actual cost from pulumicost: %.4f", actualCost)

	// Validate cost for t3.micro instance
	ValidateCost(t, "t3.micro", actualCost, expectedCosts["t3.micro"], config.Tolerance)
}

func LoadExpectedCosts(t *testing.T) map[string]float64 {
	path := filepath.Join("fixtures", "expected_costs.json")
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var costs map[string]float64
	err = json.Unmarshal(data, &costs)
	require.NoError(t, err)

	return costs
}
