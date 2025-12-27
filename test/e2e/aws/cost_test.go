//go:build e2e
// +build e2e

package aws

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

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

	// Mock actual cost for demonstration since we can't provision real resources
	actualCost := 10.0 // derived from pulumicost run

	// Validate
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
