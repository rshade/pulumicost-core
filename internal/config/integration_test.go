package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOutputFormat(t *testing.T) {
	// Test CLI flag takes precedence
	result := GetOutputFormat("json")
	assert.Equal(t, "json", result)
	
	// Test environment variable fallback
	os.Setenv("PULUMICOST_OUTPUT_FORMAT", "ndjson")
	defer os.Unsetenv("PULUMICOST_OUTPUT_FORMAT")
	
	result = GetOutputFormat("")
	assert.Equal(t, "ndjson", result)
	
	// Test config file fallback
	os.Unsetenv("PULUMICOST_OUTPUT_FORMAT")
	tempDir := t.TempDir()
	setupTestConfigFile(t, tempDir, map[string]interface{}{
		"output": map[string]interface{}{
			"default_format": "json",
		},
	})
	
	result = GetOutputFormat("")
	assert.Equal(t, "json", result)
}

func TestGetOutputPrecision(t *testing.T) {
	// Test environment variable
	os.Setenv("PULUMICOST_OUTPUT_PRECISION", "5")
	defer os.Unsetenv("PULUMICOST_OUTPUT_PRECISION")
	
	result := GetOutputPrecision()
	assert.Equal(t, 5, result)
	
	// Test config file fallback
	os.Unsetenv("PULUMICOST_OUTPUT_PRECISION")
	tempDir := t.TempDir()
	setupTestConfigFile(t, tempDir, map[string]interface{}{
		"output": map[string]interface{}{
			"precision": 3,
		},
	})
	
	result = GetOutputPrecision()
	assert.Equal(t, 3, result)
}

func TestGetPluginConfig(t *testing.T) {
	tempDir := t.TempDir()
	setupTestConfigFile(t, tempDir, map[string]interface{}{
		"plugins": map[string]interface{}{
			"aws": map[string]interface{}{
				"region":  "us-west-2",
				"profile": "production",
			},
		},
	})
	
	// Test existing plugin
	pluginConfig, err := GetPluginConfig("aws")
	require.NoError(t, err)
	assert.Equal(t, "us-west-2", pluginConfig.Region)
	assert.Equal(t, "production", pluginConfig.Profile)
	
	// Test non-existing plugin returns empty config
	pluginConfig, err = GetPluginConfig("nonexistent")
	require.NoError(t, err)
	assert.Empty(t, pluginConfig.Region)
	assert.NotNil(t, pluginConfig.Credentials)
	assert.NotNil(t, pluginConfig.Settings)
}

func TestIsDebugMode(t *testing.T) {
	// Test environment variable
	os.Setenv("PULUMICOST_LOG_LEVEL", "debug")
	defer os.Unsetenv("PULUMICOST_LOG_LEVEL")
	
	result := IsDebugMode()
	assert.True(t, result)
	
	os.Setenv("PULUMICOST_LOG_LEVEL", "DEBUG")
	result = IsDebugMode()
	assert.True(t, result)
	
	os.Setenv("PULUMICOST_LOG_LEVEL", "info")
	result = IsDebugMode()
	assert.False(t, result)
	
	// Test config file fallback
	os.Unsetenv("PULUMICOST_LOG_LEVEL")
	tempDir := t.TempDir()
	setupTestConfigFile(t, tempDir, map[string]interface{}{
		"logging": map[string]interface{}{
			"level": "debug",
		},
	})
	
	result = IsDebugMode()
	assert.True(t, result)
}

func TestGetLogFile(t *testing.T) {
	// Test environment variable
	expectedPath := "/tmp/test.log"
	os.Setenv("PULUMICOST_LOG_FILE", expectedPath)
	defer os.Unsetenv("PULUMICOST_LOG_FILE")
	
	result := GetLogFile()
	assert.Equal(t, expectedPath, result)
	
	// Test config file fallback
	os.Unsetenv("PULUMICOST_LOG_FILE")
	tempDir := t.TempDir()
	configLogPath := "/tmp/config.log"
	setupTestConfigFile(t, tempDir, map[string]interface{}{
		"logging": map[string]interface{}{
			"file": configLogPath,
		},
	})
	
	result = GetLogFile()
	assert.Equal(t, configLogPath, result)
}

func TestParseIntSafe(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		hasError bool
	}{
		{"0", 0, false},
		{"123", 123, false},
		{"999", 999, false},
		{"abc", 0, true},
		{"12a", 0, true},
		{"-1", 0, true},
		{"", 0, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseIntSafe(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// setupTestConfigFile creates a test config file in the temp directory
func setupTestConfigFile(t *testing.T, tempDir string, configData map[string]interface{}) {
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() { os.Setenv("HOME", originalHome) })
	
	cfg := DefaultConfig()
	
	// Apply test configuration data
	if output, ok := configData["output"].(map[string]interface{}); ok {
		if format, ok := output["default_format"].(string); ok {
			cfg.Output.DefaultFormat = format
		}
		if precision, ok := output["precision"].(int); ok {
			cfg.Output.Precision = precision
		}
	}
	
	if logging, ok := configData["logging"].(map[string]interface{}); ok {
		if level, ok := logging["level"].(string); ok {
			cfg.Logging.Level = level
		}
		if file, ok := logging["file"].(string); ok {
			cfg.Logging.File = file
		}
	}
	
	if plugins, ok := configData["plugins"].(map[string]interface{}); ok {
		cfg.Plugins = make(map[string]PluginConfig)
		for pluginName, pluginData := range plugins {
			if pluginMap, ok := pluginData.(map[string]interface{}); ok {
				pluginConfig := PluginConfig{
					Credentials: make(map[string]string),
					Settings:    make(map[string]interface{}),
				}
				
				if region, ok := pluginMap["region"].(string); ok {
					pluginConfig.Region = region
				}
				if profile, ok := pluginMap["profile"].(string); ok {
					pluginConfig.Profile = profile
				}
				
				cfg.Plugins[pluginName] = pluginConfig
			}
		}
	}
	
	err := cfg.Save()
	require.NoError(t, err)
}