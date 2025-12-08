package cli_test

import (
	"bytes"
	"testing"

	"github.com/rshade/pulumicost-core/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRootCmd(t *testing.T) {
	// Set log level to error to avoid cluttering test output with debug logs
	t.Setenv("PULUMICOST_LOG_LEVEL", "error")
	tests := []struct {
		name        string
		args        []string
		expectError bool
		checkOutput func(t *testing.T, output string)
	}{
		{
			name:        "help flag",
			args:        []string{"--help"},
			expectError: false,
			checkOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "PulumiCost: Calculate projected and actual cloud costs via plugins")
				assert.Contains(t, output, "Available Commands:")
				assert.Contains(t, output, "cost")
				assert.Contains(t, output, "plugin")
			},
		},
		{
			name:        "version flag",
			args:        []string{"--version"},
			expectError: false,
			checkOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "pulumicost version")
			},
		},
		{
			name:        "debug flag",
			args:        []string{"--debug", "--help"},
			expectError: false,
			checkOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "PulumiCost")
			},
		},
		{
			name:        "invalid command",
			args:        []string{"invalid"},
			expectError: true,
		},
		{
			name:        "cost subcommand help",
			args:        []string{"cost", "--help"},
			expectError: false,
			checkOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "Cost calculation commands")
				assert.Contains(t, output, "projected")
				assert.Contains(t, output, "actual")
			},
		},
		{
			name:        "plugin subcommand help",
			args:        []string{"plugin", "--help"},
			expectError: false,
			checkOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "Plugin management commands")
				assert.Contains(t, output, "list")
				assert.Contains(t, output, "validate")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := cli.NewRootCmd("test-version")
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.checkOutput != nil {
					tt.checkOutput(t, buf.String())
				}
			}
		})
	}
}

func TestRootCmdExamples(t *testing.T) {
	// Set log level to error to avoid cluttering test output with debug logs
	t.Setenv("PULUMICOST_LOG_LEVEL", "error")
	cmd := cli.NewRootCmd("test-version")

	// Check that examples are present
	assert.NotEmpty(t, cmd.Example)
	assert.Contains(t, cmd.Example, "pulumicost cost projected")
	assert.Contains(t, cmd.Example, "pulumicost cost actual")
	assert.Contains(t, cmd.Example, "pulumicost plugin list")
	assert.Contains(t, cmd.Example, "pulumicost plugin validate")
}

func TestRootCmdStructure(t *testing.T) {
	// Set log level to error to avoid cluttering test output with debug logs
	t.Setenv("PULUMICOST_LOG_LEVEL", "error")
	cmd := cli.NewRootCmd("test-version")

	// Check that main subcommands exist
	costCmd, _, err := cmd.Find([]string{"cost"})
	require.NoError(t, err)
	assert.NotNil(t, costCmd)

	pluginCmd, _, err := cmd.Find([]string{"plugin"})
	require.NoError(t, err)
	assert.NotNil(t, pluginCmd)

	// Check that cost subcommands exist
	projectedCmd, _, err := cmd.Find([]string{"cost", "projected"})
	require.NoError(t, err)
	assert.NotNil(t, projectedCmd)

	actualCmd, _, err := cmd.Find([]string{"cost", "actual"})
	require.NoError(t, err)
	assert.NotNil(t, actualCmd)

	// Check that plugin subcommands exist
	listCmd, _, err := cmd.Find([]string{"plugin", "list"})
	require.NoError(t, err)
	assert.NotNil(t, listCmd)

	validateCmd, _, err := cmd.Find([]string{"plugin", "validate"})
	require.NoError(t, err)
	assert.NotNil(t, validateCmd)
}

func TestRootCmdFlags(t *testing.T) {
	// Set log level to error to avoid cluttering test output with debug logs
	t.Setenv("PULUMICOST_LOG_LEVEL", "error")
	cmd := cli.NewRootCmd("test-version")

	// Check persistent flags
	debugFlag := cmd.PersistentFlags().Lookup("debug")
	assert.NotNil(t, debugFlag)
	assert.Equal(t, "bool", debugFlag.Value.Type())
	assert.Equal(t, "false", debugFlag.DefValue)

	// Check version flag is available
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--version"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "test-version")
}
