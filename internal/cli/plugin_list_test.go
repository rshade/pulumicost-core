package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPluginListCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "no flags",
			args:        []string{},
			expectError: false,
		},
		{
			name:        "verbose flag",
			args:        []string{"--verbose"},
			expectError: false,
		},
		{
			name:        "help flag",
			args:        []string{"--help"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := newPluginListCmd()
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPluginListCmdFlags(t *testing.T) {
	cmd := newPluginListCmd()

	// Check verbose flag
	verboseFlag := cmd.Flags().Lookup("verbose")
	assert.NotNil(t, verboseFlag)
	assert.Equal(t, "bool", verboseFlag.Value.Type())
	assert.Equal(t, "false", verboseFlag.DefValue)
	assert.Contains(t, verboseFlag.Usage, "Show detailed plugin information")
}

func TestPluginListCmdHelp(t *testing.T) {
	var buf bytes.Buffer
	cmd := newPluginListCmd()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "List all installed plugins with their versions and paths")
	assert.Contains(t, output, "List all installed plugins with their versions and paths")
	assert.Contains(t, output, "--verbose")
	assert.Contains(t, output, "Show detailed plugin information")
}

func TestPluginListCmdExamples(t *testing.T) {
	cmd := newPluginListCmd()

	// Check that examples are present
	assert.NotEmpty(t, cmd.Example)
	assert.Contains(t, cmd.Example, "pulumicost plugin list")
	assert.Contains(t, cmd.Example, "pulumicost plugin list --verbose")
	assert.Contains(t, cmd.Example, "List plugins with detailed information")
}

func TestPluginListCmdOutput(t *testing.T) {
	cmd := newPluginListCmd()

	// The command should execute without error even when no plugins exist
	err := cmd.RunE(cmd, []string{})
	require.NoError(t, err)
}
