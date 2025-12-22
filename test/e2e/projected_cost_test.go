//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_ProjectedCost tests the projected cost workflow using the compiled binary.
func TestE2E_ProjectedCost(t *testing.T) {
	binary := findPulumicostBinary()
	require.NotEmpty(t, binary, "pulumicost binary not found")

	// Use fixture plan
	planPath, _ := filepath.Abs("../fixtures/plans/aws/simple.json")
	require.FileExists(t, planPath)

	// Run command
	cmd := exec.Command(binary, "cost", "projected", "--pulumi-json", planPath, "--output", "json")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Command failed: %s", string(output))

	// Verify JSON output
	var result map[string]interface{}
	err = json.Unmarshal(output, &result)
	require.NoError(t, err, "Failed to parse JSON output: %s", string(output))

	// Verify structure
	assert.Contains(t, result, "summary")
	assert.Contains(t, result, "resources")
}
