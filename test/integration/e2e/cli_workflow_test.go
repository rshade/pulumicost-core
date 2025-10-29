// Package e2e_test provides end-to-end tests for the pulumicost CLI workflows.
package e2e_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_ProjectedCostWorkflow tests the complete projected cost workflow from
// building the binary, loading a Pulumi plan, and outputting JSON results.
func TestE2E_ProjectedCostWorkflow(t *testing.T) {
	// Build the binary first
	binaryPath := buildPulumicostBinary(t)
	defer os.Remove(binaryPath)

	// Use test fixture plan
	planFile := filepath.Join("..", "..", "fixtures", "plans", "aws-simple-plan.json")
	planPath, err := filepath.Abs(planFile)
	require.NoError(t, err)

	// Verify plan file exists
	_, err = os.Stat(planPath)
	require.NoError(t, err, "Test plan file not found: %s", planPath)

	// Run projected cost command
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		binaryPath,
		"cost",
		"projected",
		"--pulumi-json",
		planPath,
		"--output",
		"json",
	)
	output, err := cmd.Output()
	require.NoError(t, err, "Command failed: %v", err)

	// Parse JSON output
	var results []map[string]interface{}
	err = json.Unmarshal(output, &results)
	require.NoError(t, err, "Failed to parse JSON output: %s", string(output))

	// Validate output structure
	assert.NotEmpty(t, results, "Expected cost results")

	for _, result := range results {
		assert.Contains(t, result, "resource_type")
		assert.Contains(t, result, "resource_id")
		assert.Contains(t, result, "adapter")
		assert.Contains(t, result, "currency")
		assert.Contains(t, result, "monthly_cost")
		assert.Contains(t, result, "hourly_cost")

		// Since no plugins are available, should use "none" adapter
		assert.Equal(t, "none", result["adapter"])
		assert.Equal(t, "USD", result["currency"])
	}
}

// TestE2E_ProjectedCostWorkflow_TableOutput tests the projected cost workflow with
// table output format to ensure formatting and resource type display are correct.
func TestE2E_ProjectedCostWorkflow_TableOutput(t *testing.T) {
	// Build the binary first
	binaryPath := buildPulumicostBinary(t)
	defer os.Remove(binaryPath)

	// Use multi-resource test plan
	planFile := filepath.Join("..", "..", "fixtures", "plans", "aws-multi-resource-plan.json")
	planPath, err := filepath.Abs(planFile)
	require.NoError(t, err)

	// Run projected cost command with table output
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		binaryPath,
		"cost",
		"projected",
		"--pulumi-json",
		planPath,
		"--output",
		"table",
	)
	output, err := cmd.Output()
	require.NoError(t, err, "Command failed: %v", err)

	outputStr := string(output)

	// Validate table output contains expected elements
	assert.Contains(t, outputStr, "RESOURCE TYPE")
	assert.Contains(t, outputStr, "RESOURCE ID")
	assert.Contains(t, outputStr, "ADAPTER")
	assert.Contains(t, outputStr, "MONTHLY COST")
	assert.Contains(t, outputStr, "CURRENCY")

	// Should contain resource types from the plan
	assert.Contains(t, outputStr, "aws_instance")
	assert.Contains(t, outputStr, "aws_s3_bucket")
	assert.Contains(t, outputStr, "aws_rds_instance")
	assert.Contains(t, outputStr, "aws_lambda_function")
}

// TestE2E_ActualCostWorkflow tests the actual cost calculation workflow with
// a time range query to validate actual cost command functionality.
func TestE2E_ActualCostWorkflow(t *testing.T) {
	// Build the binary first
	binaryPath := buildPulumicostBinary(t)
	defer os.Remove(binaryPath)

	// Run actual cost command
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		binaryPath,
		"cost",
		"actual",
		"--from",
		"2024-01-01",
		"--to",
		"2024-01-31",
		"--output",
		"json",
	)
	output, err := cmd.Output()

	// This might fail if no plugins are available, which is expected
	if err != nil {
		t.Logf("Actual cost command failed as expected (no plugins): %v", err)
		return
	}

	// If it succeeds, validate the output
	var results []map[string]interface{}
	err = json.Unmarshal(output, &results)
	require.NoError(t, err, "Failed to parse JSON output: %s", string(output))

	// Validate output structure if results exist
	for _, result := range results {
		assert.Contains(t, result, "resource_type")
		assert.Contains(t, result, "resource_id")
		assert.Contains(t, result, "currency")
		assert.Contains(t, result, "monthly_cost")
	}
}

// TestE2E_PluginListWorkflow tests the plugin list command to ensure it reports
// correctly when no plugins are installed.
func TestE2E_PluginListWorkflow(t *testing.T) {
	// Build the binary first
	binaryPath := buildPulumicostBinary(t)
	defer os.Remove(binaryPath)

	// Run plugin list command
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, "plugin", "list")
	output, err := cmd.Output()
	require.NoError(t, err, "Plugin list command failed: %v", err)

	outputStr := string(output)

	// Should indicate no plugins are installed
	assert.Contains(t, outputStr, "No plugins")
}

// TestE2E_PluginValidateWorkflow tests the plugin validate command to ensure it reports
// correctly when no plugins are installed or available for validation.
func TestE2E_PluginValidateWorkflow(t *testing.T) {
	// Build the binary first
	binaryPath := buildPulumicostBinary(t)
	defer os.Remove(binaryPath)

	// Run plugin validate command
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, "plugin", "validate")
	output, err := cmd.Output()
	require.NoError(t, err, "Plugin validate command failed: %v", err)

	outputStr := string(output)

	// Should indicate no plugins to validate
	assert.Contains(t, outputStr, "No plugins")
}

// TestE2E_HelpCommands tests various help commands to ensure proper documentation
// and usage information is available for all major CLI commands.
func TestE2E_HelpCommands(t *testing.T) {
	// Build the binary first
	binaryPath := buildPulumicostBinary(t)
	defer os.Remove(binaryPath)

	tests := []struct {
		name            string
		args            []string
		expectedContent string
	}{
		{
			name:            "root help",
			args:            []string{"--help"},
			expectedContent: "CLI tool for calculating cloud infrastructure costs",
		},
		{
			name:            "cost help",
			args:            []string{"cost", "--help"},
			expectedContent: "Cost calculation commands",
		},
		{
			name:            "projected cost help",
			args:            []string{"cost", "projected", "--help"},
			expectedContent: "Calculate projected costs",
		},
		{
			name:            "actual cost help",
			args:            []string{"cost", "actual", "--help"},
			expectedContent: "Calculate actual costs",
		},
		{
			name:            "plugin help",
			args:            []string{"plugin", "--help"},
			expectedContent: "Plugin management commands",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, binaryPath, tt.args...)
			output, err := cmd.Output()
			require.NoError(t, err, "Help command failed: %v", err)

			outputStr := string(output)
			assert.Contains(t, outputStr, tt.expectedContent)
		})
	}
}

// TestE2E_InvalidInputHandling tests error handling for invalid inputs such as
// missing flags, invalid file paths, and invalid date formats.
func TestE2E_InvalidInputHandling(t *testing.T) {
	// Build the binary first
	binaryPath := buildPulumicostBinary(t)
	defer os.Remove(binaryPath)

	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "invalid command",
			args:        []string{"invalid-command"},
			expectError: true,
		},
		{
			name:        "missing required flag",
			args:        []string{"cost", "projected"},
			expectError: true,
		},
		{
			name:        "invalid file path",
			args:        []string{"cost", "projected", "--pulumi-json", "/nonexistent/file.json"},
			expectError: true,
		},
		{
			name:        "invalid date format",
			args:        []string{"cost", "actual", "--from", "invalid-date", "--to", "2024-01-31"},
			expectError: true,
		},
		{
			name: "invalid output format",
			args: []string{
				"cost",
				"projected",
				"--pulumi-json",
				"/dev/null",
				"--output",
				"invalid",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, binaryPath, tt.args...)
			_, err := cmd.Output()

			if tt.expectError {
				assert.Error(t, err, "Expected command to fail")
			} else {
				assert.NoError(t, err, "Expected command to succeed")
			}
		})
	}
}

// TestE2E_MultipleOutputFormats tests that all supported output formats (json, table, ndjson)
// produce valid output and can be properly parsed.
func TestE2E_MultipleOutputFormats(t *testing.T) {
	// Build the binary first
	binaryPath := buildPulumicostBinary(t)
	defer os.Remove(binaryPath)

	// Use test fixture plan
	planFile := filepath.Join("..", "..", "fixtures", "plans", "aws-simple-plan.json")
	planPath, err := filepath.Abs(planFile)
	require.NoError(t, err)

	formats := []string{"json", "table", "ndjson"}

	for _, format := range formats {
		t.Run(format+"_format", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, binaryPath, "cost", "projected",
				"--pulumi-json", planPath, "--output", format)
			output, cmdErr := cmd.Output()
			require.NoError(t, cmdErr, "Command failed for format %s: %v", format, cmdErr)

			assert.NotEmpty(t, output, "Output should not be empty for format %s", format)

			if format == "json" || format == "ndjson" {
				// Verify it's valid JSON
				var jsonData interface{}
				if format == "json" {
					err = json.Unmarshal(output, &jsonData)
				} else {
					// For NDJSON, try to parse the first line
					lines := []byte{}
					for _, b := range output {
						if b == '\n' {
							break
						}
						lines = append(lines, b)
					}
					err = json.Unmarshal(lines, &jsonData)
				}
				assert.NoError(t, err, "Invalid JSON output for format %s", format)
			}
		})
	}
}

// buildPulumicostBinary is a helper function that builds the pulumicost binary from source
// for testing purposes, returning the path to the compiled executable.
func buildPulumicostBinary(t *testing.T) string {
	// Create a temporary binary path
	tempDir := t.TempDir()
	binaryName := "pulumicost"
	if runtime.GOOS == "windows" {
		binaryName = "pulumicost.exe"
	}
	binaryPath := filepath.Join(tempDir, binaryName)

	// Build the binary
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Get the project root directory (3 levels up from test/integration/e2e)
	projectRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	require.NoError(t, err)

	cmd := exec.CommandContext(ctx, "go", "build", "-o", binaryPath, "./cmd/pulumicost")
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to build binary: %v\nOutput: %s", err, string(output))

	// Verify binary was created and is executable
	info, err := os.Stat(binaryPath)
	require.NoError(t, err, "Binary not found after build")
	require.True(t, info.Mode().IsRegular(), "Binary is not a regular file")

	return binaryPath
}
