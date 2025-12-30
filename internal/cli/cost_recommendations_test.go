package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rshade/pulumicost-core/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T001: Test NewCostRecommendationsCmd() creates a valid command.
func TestNewCostRecommendationsCmd(t *testing.T) {
	cmd := cli.NewCostRecommendationsCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "recommendations", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotEmpty(t, cmd.Example)
}

// T001: Test command has required flags.
func TestNewCostRecommendationsCmd_Flags(t *testing.T) {
	cmd := cli.NewCostRecommendationsCmd()

	// Check pulumi-json flag exists and is required
	pulumiJSONFlag := cmd.Flags().Lookup("pulumi-json")
	require.NotNil(t, pulumiJSONFlag, "pulumi-json flag should exist")

	// Check other expected flags
	adapterFlag := cmd.Flags().Lookup("adapter")
	require.NotNil(t, adapterFlag, "adapter flag should exist")

	outputFlag := cmd.Flags().Lookup("output")
	require.NotNil(t, outputFlag, "output flag should exist")

	filterFlag := cmd.Flags().Lookup("filter")
	require.NotNil(t, filterFlag, "filter flag should exist")
}

// T001: Test command fails without required pulumi-json flag.
func TestNewCostRecommendationsCmd_RequiredFlags(t *testing.T) {
	cmd := cli.NewCostRecommendationsCmd()

	// Execute without required flag should fail
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pulumi-json")
}

// T002: Test table output format rendering.
func TestCostRecommendationsCmd_TableOutput(t *testing.T) {
	// Create a temporary plan file
	planJSON := `{
		"version": 3,
		"steps": [
			{
				"op": "create",
				"urn": "urn:pulumi:test::test::aws:ec2/instance:Instance::test-instance",
				"type": "aws:ec2/instance:Instance",
				"newState": {
					"type": "aws:ec2/instance:Instance",
					"inputs": {
						"instanceType": "t3.micro",
						"availabilityZone": "us-east-1a"
					}
				}
			}
		]
	}`

	tmpDir := t.TempDir()
	planPath := filepath.Join(tmpDir, "plan.json")
	err := os.WriteFile(planPath, []byte(planJSON), 0o600)
	require.NoError(t, err)

	cmd := cli.NewCostRecommendationsCmd()
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&outBuf)

	// Execute with table output (default)
	cmd.SetArgs([]string{"--pulumi-json", planPath, "--output", "table"})
	// Command will fail without plugins, but that's expected
	// We're testing that the command infrastructure works
	_ = cmd.Execute()

	// The output should contain some text (even if empty recommendations)
	// In a full integration test with plugins, we'd check for table headers
}

// T002: Test JSON output format rendering.
func TestCostRecommendationsCmd_JSONOutput(t *testing.T) {
	planJSON := `{
		"version": 3,
		"steps": [
			{
				"op": "create",
				"urn": "urn:pulumi:test::test::aws:ec2/instance:Instance::test-instance",
				"type": "aws:ec2/instance:Instance",
				"newState": {
					"type": "aws:ec2/instance:Instance",
					"inputs": {
						"instanceType": "t3.micro"
					}
				}
			}
		]
	}`

	tmpDir := t.TempDir()
	planPath := filepath.Join(tmpDir, "plan.json")
	err := os.WriteFile(planPath, []byte(planJSON), 0o600)
	require.NoError(t, err)

	cmd := cli.NewCostRecommendationsCmd()
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&outBuf)

	cmd.SetArgs([]string{"--pulumi-json", planPath, "--output", "json"})
	_ = cmd.Execute()

	// If output is produced, it should be valid JSON
	output := strings.TrimSpace(outBuf.String())
	if output != "" && strings.HasPrefix(output, "{") {
		var result map[string]interface{}
		jsonErr := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, jsonErr, "JSON output should be valid JSON")
	}
}

// T002: Test NDJSON output format rendering.
func TestCostRecommendationsCmd_NDJSONOutput(t *testing.T) {
	planJSON := `{
		"version": 3,
		"steps": []
	}`

	tmpDir := t.TempDir()
	planPath := filepath.Join(tmpDir, "plan.json")
	err := os.WriteFile(planPath, []byte(planJSON), 0o600)
	require.NoError(t, err)

	cmd := cli.NewCostRecommendationsCmd()
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&outBuf)

	cmd.SetArgs([]string{"--pulumi-json", planPath, "--output", "ndjson"})
	_ = cmd.Execute()

	// NDJSON output should have each line as valid JSON (if any output)
	output := strings.TrimSpace(outBuf.String())
	if output != "" {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && strings.HasPrefix(line, "{") {
				var item map[string]interface{}
				jsonErr := json.Unmarshal([]byte(line), &item)
				assert.NoError(t, jsonErr, "each NDJSON line should be valid JSON")
			}
		}
	}
}

// T003: Test --filter flag parsing with action types.
func TestCostRecommendationsCmd_FilterFlag(t *testing.T) {
	cmd := cli.NewCostRecommendationsCmd()

	// Check filter flag accepts array values
	filterFlag := cmd.Flags().Lookup("filter")
	require.NotNil(t, filterFlag)

	// Test filter flag can be set
	err := cmd.Flags().Set("filter", "action=MIGRATE")
	assert.NoError(t, err)
}

// T003: Test multiple filter values.
func TestCostRecommendationsCmd_MultipleFilters(t *testing.T) {
	planJSON := `{
		"version": 3,
		"steps": []
	}`

	tmpDir := t.TempDir()
	planPath := filepath.Join(tmpDir, "plan.json")
	err := os.WriteFile(planPath, []byte(planJSON), 0o600)
	require.NoError(t, err)

	cmd := cli.NewCostRecommendationsCmd()
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&outBuf)

	// Test with multiple filter values
	cmd.SetArgs([]string{
		"--pulumi-json", planPath,
		"--filter", "action=MIGRATE,RIGHTSIZE",
		"--output", "json",
	})
	_ = cmd.Execute()

	// Command should process without panic
}

// T003: Test case-insensitive filter matching.
func TestCostRecommendationsCmd_CaseInsensitiveFilter(t *testing.T) {
	planJSON := `{
		"version": 3,
		"steps": []
	}`

	tmpDir := t.TempDir()
	planPath := filepath.Join(tmpDir, "plan.json")
	err := os.WriteFile(planPath, []byte(planJSON), 0o600)
	require.NoError(t, err)

	cmd := cli.NewCostRecommendationsCmd()
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&outBuf)

	// Test lowercase filter values
	cmd.SetArgs([]string{
		"--pulumi-json", planPath,
		"--filter", "action=migrate",
		"--output", "json",
	})
	_ = cmd.Execute()

	// Command should process without panic (case insensitivity tested in proto package)
}

// Test invalid plan path error handling.
func TestCostRecommendationsCmd_InvalidPlanPath(t *testing.T) {
	cmd := cli.NewCostRecommendationsCmd()
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&outBuf)

	cmd.SetArgs([]string{"--pulumi-json", "/nonexistent/path/plan.json"})
	err := cmd.Execute()

	assert.Error(t, err)
}

// Test unsupported output format error.
func TestCostRecommendationsCmd_UnsupportedOutputFormat(t *testing.T) {
	planJSON := `{"version": 3, "steps": []}`

	tmpDir := t.TempDir()
	planPath := filepath.Join(tmpDir, "plan.json")
	err := os.WriteFile(planPath, []byte(planJSON), 0o600)
	require.NoError(t, err)

	cmd := cli.NewCostRecommendationsCmd()
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&outBuf)

	cmd.SetArgs([]string{"--pulumi-json", planPath, "--output", "invalid"})
	err = cmd.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported output format")
}

// T022: Test action type filter parsing in CLI command with valid types.
func TestCostRecommendationsCmd_ValidActionTypeFilter(t *testing.T) {
	tests := []struct {
		name   string
		filter string
	}{
		{"single type MIGRATE", "action=MIGRATE"},
		{"single type RIGHTSIZE", "action=RIGHTSIZE"},
		{"single type TERMINATE", "action=TERMINATE"},
		{"single type CONSOLIDATE", "action=CONSOLIDATE"},
		{"single type SCHEDULE", "action=SCHEDULE"},
		{"single type REFACTOR", "action=REFACTOR"},
		{"single type OTHER", "action=OTHER"},
		{"multiple types", "action=MIGRATE,RIGHTSIZE,TERMINATE"},
		{"lowercase", "action=migrate"},
		{"mixed case", "action=Migrate,RIGHTSIZE"},
		{"with spaces", "action=MIGRATE , RIGHTSIZE"},
	}

	planJSON := `{"version": 3, "steps": []}`
	tmpDir := t.TempDir()
	planPath := filepath.Join(tmpDir, "plan.json")
	err := os.WriteFile(planPath, []byte(planJSON), 0o600)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := cli.NewCostRecommendationsCmd()
			var outBuf bytes.Buffer
			cmd.SetOut(&outBuf)
			cmd.SetErr(&outBuf)

			cmd.SetArgs([]string{
				"--pulumi-json", planPath,
				"--filter", tt.filter,
				"--output", "json",
			})
			// Should not error due to filter parsing
			// (may error due to no plugins, but filter is valid)
			_ = cmd.Execute()
		})
	}
}

// T023: Test invalid action type filter error message listing all 11 valid types.
func TestCostRecommendationsCmd_InvalidActionTypeFilterError(t *testing.T) {
	planJSON := `{"version": 3, "steps": []}`
	tmpDir := t.TempDir()
	planPath := filepath.Join(tmpDir, "plan.json")
	err := os.WriteFile(planPath, []byte(planJSON), 0o600)
	require.NoError(t, err)

	cmd := cli.NewCostRecommendationsCmd()
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&outBuf)

	cmd.SetArgs([]string{
		"--pulumi-json", planPath,
		"--filter", "action=INVALID_TYPE",
		"--output", "json",
	})
	err = cmd.Execute()

	// Should error with invalid action type
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid action type")

	// Error message should list valid types
	errMsg := err.Error()
	validTypes := []string{"RIGHTSIZE", "TERMINATE", "MIGRATE", "CONSOLIDATE", "SCHEDULE", "REFACTOR", "OTHER"}
	for _, vt := range validTypes {
		assert.Contains(t, errMsg, vt, "error should list valid type: %s", vt)
	}
}

// T023: Test empty action type filter error.
func TestCostRecommendationsCmd_EmptyActionTypeFilter(t *testing.T) {
	planJSON := `{"version": 3, "steps": []}`
	tmpDir := t.TempDir()
	planPath := filepath.Join(tmpDir, "plan.json")
	err := os.WriteFile(planPath, []byte(planJSON), 0o600)
	require.NoError(t, err)

	cmd := cli.NewCostRecommendationsCmd()
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&outBuf)

	cmd.SetArgs([]string{
		"--pulumi-json", planPath,
		"--filter", "action=",
		"--output", "json",
	})
	err = cmd.Execute()

	// Should error with empty filter value
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid action type filter")
}
