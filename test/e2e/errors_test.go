//go:build e2e
// +build e2e

package e2e

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_Errors_MissingFile(t *testing.T) {
	binary := findPulumicostBinary()
	require.NotEmpty(t, binary)

	// Run with non-existent file
	cmd := exec.Command(binary, "cost", "projected", "--pulumi-json", "nonexistent.json")
	output, err := cmd.CombinedOutput()

	// Should fail with exit code 1
	require.Error(t, err)
	if exitErr, ok := err.(*exec.ExitError); ok {
		assert.Equal(t, 1, exitErr.ExitCode(), "expected exit code 1 for missing file")
	}
	assert.Contains(t, string(output), "no such file")
}

func TestE2E_Errors_InvalidFormat(t *testing.T) {
	binary := findPulumicostBinary()
	require.NotEmpty(t, binary)

	// Run with invalid format
	cmd := exec.Command(binary, "cost", "projected", "--output", "invalid")
	output, err := cmd.CombinedOutput()

	// Should fail with exit code 1
	require.Error(t, err)
	if exitErr, ok := err.(*exec.ExitError); ok {
		assert.Equal(t, 1, exitErr.ExitCode(), "expected exit code 1 for invalid format")
	}
	assert.Contains(t, string(output), "invalid")
}
