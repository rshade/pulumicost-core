package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCostProjectedCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "missing required flag",
			args:        []string{},
			expectError: true,
			errorMsg:    "required flag(s) \"pulumi-json\" not set",
		},
		{
			name:        "help flag",
			args:        []string{"--help"},
			expectError: false,
		},
		{
			name:        "with all flags",
			args:        []string{
				"--pulumi-json", "test.json",
				"--spec-dir", "/tmp/specs",
				"--adapter", "test-adapter",
				"--output", "json",
				"--filter", "type=aws:ec2/instance",
			},
			expectError: true, // Will fail because file doesn't exist
			errorMsg:    "loading Pulumi plan",
		},
		{
			name:        "with only required flag",
			args:        []string{"--pulumi-json", "test.json"},
			expectError: true, // Will fail because file doesn't exist
			errorMsg:    "loading Pulumi plan",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := newCostProjectedCmd()
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCostProjectedCmdFlags(t *testing.T) {
	cmd := newCostProjectedCmd()

	// Check required flags
	pulumiJsonFlag := cmd.Flags().Lookup("pulumi-json")
	assert.NotNil(t, pulumiJsonFlag)
	assert.Equal(t, "string", pulumiJsonFlag.Value.Type())
	assert.Equal(t, "", pulumiJsonFlag.DefValue)

	// Check optional flags
	specDirFlag := cmd.Flags().Lookup("spec-dir")
	assert.NotNil(t, specDirFlag)
	assert.Equal(t, "string", specDirFlag.Value.Type())

	adapterFlag := cmd.Flags().Lookup("adapter")
	assert.NotNil(t, adapterFlag)
	assert.Equal(t, "string", adapterFlag.Value.Type())

	outputFlag := cmd.Flags().Lookup("output")
	assert.NotNil(t, outputFlag)
	assert.Equal(t, "string", outputFlag.Value.Type())
	assert.Equal(t, "table", outputFlag.DefValue)

	filterFlag := cmd.Flags().Lookup("filter")
	assert.NotNil(t, filterFlag)
	assert.Equal(t, "string", filterFlag.Value.Type())
	assert.Equal(t, "", filterFlag.DefValue)
}

func TestCostProjectedCmdHelp(t *testing.T) {
	var buf bytes.Buffer
	cmd := newCostProjectedCmd()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Calculate projected costs by analyzing a Pulumi preview JSON output")
	assert.Contains(t, output, "Calculate projected costs by analyzing a Pulumi preview JSON output")
	assert.Contains(t, output, "--pulumi-json")
	assert.Contains(t, output, "--spec-dir")
	assert.Contains(t, output, "--adapter")
	assert.Contains(t, output, "--output")
	assert.Contains(t, output, "--filter")
	assert.Contains(t, output, "Resource filter expressions")
}

func TestCostProjectedCmdExamples(t *testing.T) {
	cmd := newCostProjectedCmd()
	
	// Check that examples are present
	assert.NotEmpty(t, cmd.Example)
	assert.Contains(t, cmd.Example, "pulumicost cost projected --pulumi-json plan.json")
	assert.Contains(t, cmd.Example, "--filter \"type=aws:ec2/instance\"")
	assert.Contains(t, cmd.Example, "--output json")
	assert.Contains(t, cmd.Example, "--adapter aws-plugin")
	assert.Contains(t, cmd.Example, "--spec-dir ./custom-specs")
}