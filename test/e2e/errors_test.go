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
	
	// Should fail
	require.Error(t, err)
	assert.Contains(t, string(output), "no such file")
}

func TestE2E_Errors_InvalidFormat(t *testing.T) {
	binary := findPulumicostBinary()
	require.NotEmpty(t, binary)

	// Run with invalid format
	cmd := exec.Command(binary, "cost", "projected", "--output", "invalid")
	output, err := cmd.CombinedOutput()
	
	// Should fail
	require.Error(t, err)
	assert.Contains(t, string(output), "invalid")
}
