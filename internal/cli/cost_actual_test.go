package cli_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/rshade/pulumicost-core/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCostActualCmd(t *testing.T) {
	// Set log level to error to avoid cluttering test output with debug logs
	t.Setenv("PULUMICOST_LOG_LEVEL", "error")

	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "missing required flags",
			args:        []string{},
			expectError: true,
			errorMsg:    "required flag(s)",
		},
		{
			name:        "missing from flag",
			args:        []string{"--pulumi-json", "test.json"},
			expectError: true,
			errorMsg:    "required flag(s) \"from\" not set",
		},
		{
			name:        "help flag",
			args:        []string{"--help"},
			expectError: false,
		},
		{
			name: "with all flags",
			args: []string{
				"--pulumi-json", "test.json",
				"--from", "2025-01-01",
				"--to", "2025-01-31",
				"--adapter", "test-adapter",
				"--output", "json",
				"--group-by", "type",
			},
			expectError: true, // Will fail because file doesn't exist
			errorMsg:    "loading Pulumi plan",
		},
		{
			name: "with required flags only (to defaults to now)",
			args: []string{
				"--pulumi-json", "test.json",
				"--from", "2025-01-01",
			},
			expectError: true, // Will fail because file doesn't exist
			errorMsg:    "loading Pulumi plan",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := cli.NewCostActualCmd()
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

func TestCostActualCmdFlags(t *testing.T) {
	// Set log level to error to avoid cluttering test output with debug logs
	t.Setenv("PULUMICOST_LOG_LEVEL", "error")
	cmd := cli.NewCostActualCmd()

	// Check required flags
	pulumiJSONFlag := cmd.Flags().Lookup("pulumi-json")
	assert.NotNil(t, pulumiJSONFlag)
	assert.Equal(t, "string", pulumiJSONFlag.Value.Type())

	fromFlag := cmd.Flags().Lookup("from")
	assert.NotNil(t, fromFlag)
	assert.Equal(t, "string", fromFlag.Value.Type())

	// Check optional flags
	toFlag := cmd.Flags().Lookup("to")
	assert.NotNil(t, toFlag)
	assert.Equal(t, "string", toFlag.Value.Type())
	assert.Contains(t, toFlag.Usage, "defaults to now")

	adapterFlag := cmd.Flags().Lookup("adapter")
	assert.NotNil(t, adapterFlag)
	assert.Equal(t, "string", adapterFlag.Value.Type())

	outputFlag := cmd.Flags().Lookup("output")
	assert.NotNil(t, outputFlag)
	assert.Equal(t, "string", outputFlag.Value.Type())
	assert.Equal(t, "table", outputFlag.DefValue)

	groupByFlag := cmd.Flags().Lookup("group-by")
	assert.NotNil(t, groupByFlag)
	assert.Equal(t, "string", groupByFlag.Value.Type())
	assert.Contains(t, groupByFlag.Usage, "resource, type, provider")
}

func TestCostActualCmdHelp(t *testing.T) {
	// Set log level to error to avoid cluttering test output with debug logs
	t.Setenv("PULUMICOST_LOG_LEVEL", "error")
	var buf bytes.Buffer
	cmd := cli.NewCostActualCmd()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Fetch actual historical costs")
	assert.Contains(t, output, "cloud provider billing APIs")
	assert.Contains(t, output, "--pulumi-json")
	assert.Contains(t, output, "--from")
	assert.Contains(t, output, "--to")
	assert.Contains(t, output, "--group-by")
	assert.Contains(t, output, "defaults to now")
}

func TestCostActualCmdExamples(t *testing.T) {
	// Set log level to error to avoid cluttering test output with debug logs
	t.Setenv("PULUMICOST_LOG_LEVEL", "error")
	cmd := cli.NewCostActualCmd()

	// Check that examples are present
	assert.NotEmpty(t, cmd.Example)
	assert.Contains(t, cmd.Example, "pulumicost cost actual --pulumi-json plan.json --from")
	assert.Contains(t, cmd.Example, "to defaults to now")
	assert.Contains(t, cmd.Example, "--group-by type")
	assert.Contains(t, cmd.Example, "--group-by provider")
	assert.Contains(t, cmd.Example, "RFC3339 timestamps")
}

func TestParseTimeRange(t *testing.T) {
	// Set log level to error to avoid cluttering test output with debug logs
	t.Setenv("PULUMICOST_LOG_LEVEL", "error")
	tests := []struct {
		name        string
		fromStr     string
		toStr       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid date range",
			fromStr:     "2025-01-01",
			toStr:       "2025-01-31",
			expectError: false,
		},
		{
			name:        "valid RFC3339 range",
			fromStr:     "2025-01-01T00:00:00Z",
			toStr:       "2025-01-31T23:59:59Z",
			expectError: false,
		},
		{
			name:        "to before from",
			fromStr:     "2025-01-31",
			toStr:       "2025-01-01",
			expectError: true,
			errorMsg:    "'to' date must be after 'from' date",
		},
		{
			name:        "invalid from date",
			fromStr:     "invalid",
			toStr:       "2025-01-31",
			expectError: true,
			errorMsg:    "parsing 'from' date",
		},
		{
			name:        "invalid to date",
			fromStr:     "2025-01-01",
			toStr:       "invalid",
			expectError: true,
			errorMsg:    "parsing 'to' date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from, to, err := cli.ParseTimeRange(tt.fromStr, tt.toStr)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
				assert.True(t, to.After(from) || to.Equal(from))
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	// Set log level to error to avoid cluttering test output with debug logs
	t.Setenv("PULUMICOST_LOG_LEVEL", "error")
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "YYYY-MM-DD format",
			input:       "2025-01-15",
			expectError: false,
		},
		{
			name:        "RFC3339 format",
			input:       time.RFC3339,
			expectError: true, // RFC3339 is a constant, not a valid date
		},
		{
			name:        "RFC3339 actual date",
			input:       "2025-01-15T10:30:00Z",
			expectError: false,
		},
		{
			name:        "invalid format",
			input:       "01/15/2025",
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cli.ParseTime(tt.input)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "unable to parse date")
			} else {
				require.NoError(t, err)
				assert.False(t, result.IsZero())
			}
		})
	}
}
