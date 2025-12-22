//go:build e2e
// +build e2e

package e2e

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_Output_Table(t *testing.T) {
	binary := findPulumicostBinary()
	require.NotEmpty(t, binary)

	planPath, _ := filepath.Abs("../fixtures/plans/aws/simple.json")

	// Run with table output (default)
	cmd := exec.Command(binary, "cost", "projected", "--pulumi-json", planPath, "--output", "table")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err)

	outStr := string(output)
	assert.Contains(t, outStr, "COST SUMMARY")
	assert.Contains(t, outStr, "RESOURCE DETAILS")
}
