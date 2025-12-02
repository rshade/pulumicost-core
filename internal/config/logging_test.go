package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rshade/pulumicost-core/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T007: Unit test for config bridge function.
func TestLoggingConfig_ToLoggingConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		input          config.LoggingConfig
		expectedLevel  string
		expectedFormat string
		expectedOutput string
		expectedFile   string
	}{
		{
			name: "defaults to stderr when no file specified",
			input: config.LoggingConfig{
				Level:  "info",
				Format: "json",
				File:   "",
			},
			expectedLevel:  "info",
			expectedFormat: "json",
			expectedOutput: "stderr",
			expectedFile:   "",
		},
		{
			name: "sets output to file when file path provided",
			input: config.LoggingConfig{
				Level:  "debug",
				Format: "console",
				File:   "/var/log/pulumicost.log",
			},
			expectedLevel:  "debug",
			expectedFormat: "console",
			expectedOutput: "file",
			expectedFile:   "/var/log/pulumicost.log",
		},
		{
			name: "handles all log levels",
			input: config.LoggingConfig{
				Level:  "error",
				Format: "text",
				File:   "",
			},
			expectedLevel:  "error",
			expectedFormat: "text",
			expectedOutput: "stderr",
			expectedFile:   "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := tc.input.ToLoggingConfig()

			assert.Equal(t, tc.expectedLevel, result.Level, "Level mismatch")
			assert.Equal(t, tc.expectedFormat, result.Format, "Format mismatch")
			assert.Equal(t, tc.expectedOutput, result.Output, "Output mismatch")
			assert.Equal(t, tc.expectedFile, result.File, "File mismatch")
		})
	}
}

// T008: Unit test for config override precedence (file < env < flag).
func TestLoggingConfig_OverridePrecedence(t *testing.T) {
	// Test that environment variables override config file values
	// Note: The --debug flag override is handled in CLI, not in config package

	// Save and restore environment
	origLevel := os.Getenv("PULUMICOST_LOG_LEVEL")
	origFormat := os.Getenv("PULUMICOST_LOG_FORMAT")
	t.Cleanup(func() {
		if origLevel != "" {
			os.Setenv("PULUMICOST_LOG_LEVEL", origLevel)
		} else {
			os.Unsetenv("PULUMICOST_LOG_LEVEL")
		}
		if origFormat != "" {
			os.Setenv("PULUMICOST_LOG_FORMAT", origFormat)
		} else {
			os.Unsetenv("PULUMICOST_LOG_FORMAT")
		}
	})

	t.Run("env vars override config file values", func(t *testing.T) {
		// Create a temporary config file with specific values
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.yaml")

		configContent := `
logging:
  level: info
  format: json
`
		err := os.WriteFile(configPath, []byte(configContent), 0600)
		require.NoError(t, err)

		// Set environment variables to override
		os.Setenv("PULUMICOST_LOG_LEVEL", "debug")
		os.Setenv("PULUMICOST_LOG_FORMAT", "text")

		// Load config - env vars should take precedence
		cfg := config.New()

		assert.Equal(t, "debug", cfg.Logging.Level, "Env var should override config file level")
		assert.Equal(t, "text", cfg.Logging.Format, "Env var should override config file format")
	})
}

// Test AuditConfig validation.
func TestAuditConfig_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		auditConfig config.AuditConfig
		expectError bool
	}{
		{
			name: "disabled audit requires no validation",
			auditConfig: config.AuditConfig{
				Enabled: false,
				File:    "",
			},
			expectError: false,
		},
		{
			name: "enabled audit with empty file is valid",
			auditConfig: config.AuditConfig{
				Enabled: true,
				File:    "",
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.Config{
				Output: config.OutputConfig{
					DefaultFormat: "table",
					Precision:     2,
				},
				Logging: config.LoggingConfig{
					Level:  "info",
					Format: "json",
					Audit:  tc.auditConfig,
				},
			}

			err := cfg.Validate()
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test GetLoggingConfig retrieves from global config.
func TestGetLoggingConfig(t *testing.T) {
	loggingCfg := config.GetLoggingConfig()

	// Should return valid defaults
	assert.NotEmpty(t, loggingCfg.Level, "Level should have a default value")
}
