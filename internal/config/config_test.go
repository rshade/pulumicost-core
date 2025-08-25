package config

import (
	"os"
	"path/filepath"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	
	assert.Equal(t, "table", cfg.Output.DefaultFormat)
	assert.Equal(t, 2, cfg.Output.Precision)
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.NotEmpty(t, cfg.PluginDir)
	assert.NotEmpty(t, cfg.SpecDir)
	assert.NotEmpty(t, cfg.ConfigDir)
	assert.NotEmpty(t, cfg.ConfigFile)
}

func TestConfigSetGet(t *testing.T) {
	cfg := DefaultConfig()
	
	tests := []struct {
		key      string
		value    string
		expected interface{}
	}{
		{"output.default_format", "json", "json"},
		{"output.precision", "4", 4},
		{"logging.level", "debug", "debug"},
		{"logging.file", "/tmp/test.log", "/tmp/test.log"},
		{"plugins.aws.region", "us-west-2", "us-west-2"},
		{"plugins.aws.profile", "production", "production"},
		{"plugins.azure.subscription_id", "sub-123", "sub-123"},
	}
	
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			err := cfg.Set(tt.key, tt.value)
			require.NoError(t, err)
			
			value, err := cfg.Get(tt.key)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, value)
		})
	}
}

func TestConfigValidation(t *testing.T) {
	cfg := DefaultConfig()
	
	// Valid configuration should pass
	err := cfg.Validate()
	assert.NoError(t, err)
	
	// Invalid output format
	cfg.Output.DefaultFormat = "invalid"
	err = cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	
	// Reset to valid
	cfg.Output.DefaultFormat = "table"
	
	// Invalid precision
	cfg.Output.Precision = -1
	err = cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "precision must be between 0 and 10")
	
	cfg.Output.Precision = 11
	err = cfg.Validate()
	assert.Error(t, err)
	
	// Reset to valid
	cfg.Output.Precision = 2
	
	// Invalid log level
	cfg.Logging.Level = "invalid"
	err = cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid log level")
}

func TestConfigSaveLoad(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")
	
	cfg := DefaultConfig()
	cfg.ConfigFile = configFile
	cfg.ConfigDir = tempDir
	cfg.Output.DefaultFormat = "json"
	cfg.Output.Precision = 4
	cfg.Logging.Level = "debug"
	
	// Set plugin configuration
	err := cfg.Set("plugins.aws.region", "us-west-2")
	require.NoError(t, err)
	
	// Save configuration
	err = cfg.Save()
	require.NoError(t, err)
	
	// Load configuration
	newCfg := DefaultConfig()
	newCfg.ConfigFile = configFile
	err = newCfg.LoadFromFile(configFile)
	require.NoError(t, err)
	
	// Verify loaded values
	assert.Equal(t, "json", newCfg.Output.DefaultFormat)
	assert.Equal(t, 4, newCfg.Output.Precision)
	assert.Equal(t, "debug", newCfg.Logging.Level)
	
	// Verify plugin configuration
	awsConfig, exists := newCfg.Plugins["aws"]
	assert.True(t, exists)
	assert.Equal(t, "us-west-2", awsConfig.Region)
}

func TestCredentialEncryption(t *testing.T) {
	cfg := DefaultConfig()
	
	pluginName := "aws"
	credKey := "access_key"
	credValue := "AKIAIOSFODNN7EXAMPLE"
	
	// Set encrypted credential
	err := cfg.SetCredential(pluginName, credKey, credValue)
	require.NoError(t, err)
	
	// Retrieve and verify credential
	retrieved, err := cfg.GetCredential(pluginName, credKey)
	require.NoError(t, err)
	assert.Equal(t, credValue, retrieved)
	
	// Verify the stored value is encrypted (different from original)
	storedValue := cfg.Plugins[pluginName].Credentials[credKey]
	assert.NotEqual(t, credValue, storedValue)
	assert.NotEmpty(t, storedValue)
}

func TestEnvOverrides(t *testing.T) {
	// Set environment variables
	os.Setenv("PULUMICOST_OUTPUT_FORMAT", "json")
	os.Setenv("PULUMICOST_OUTPUT_PRECISION", "5")
	os.Setenv("PULUMICOST_LOG_LEVEL", "debug")
	os.Setenv("PULUMICOST_LOG_FILE", "/tmp/override.log")
	defer func() {
		os.Unsetenv("PULUMICOST_OUTPUT_FORMAT")
		os.Unsetenv("PULUMICOST_OUTPUT_PRECISION")
		os.Unsetenv("PULUMICOST_LOG_LEVEL")
		os.Unsetenv("PULUMICOST_LOG_FILE")
	}()
	
	cfg := DefaultConfig()
	cfg.applyEnvOverrides()
	
	assert.Equal(t, "json", cfg.Output.DefaultFormat)
	assert.Equal(t, 5, cfg.Output.Precision)
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, "/tmp/override.log", cfg.Logging.File)
}

func TestListAll(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Output.DefaultFormat = "json"
	cfg.Output.Precision = 3
	
	// Add plugin configuration
	err := cfg.Set("plugins.aws.region", "us-east-1")
	require.NoError(t, err)
	
	// Add credential
	err = cfg.SetCredential("aws", "access_key", "secret")
	require.NoError(t, err)
	
	configMap := cfg.ListAll()
	
	assert.Equal(t, "json", configMap["output.default_format"])
	assert.Equal(t, 3, configMap["output.precision"])
	assert.Equal(t, "us-east-1", configMap["plugins.aws.region"])
	assert.Equal(t, "<encrypted>", configMap["plugins.aws.credentials.access_key"])
}

func TestSetInvalidKeys(t *testing.T) {
	cfg := DefaultConfig()
	
	tests := []struct {
		key   string
		value string
	}{
		{"invalid", "value"},
		{"output.invalid", "value"},
		{"logging.invalid", "value"},
		{"output.default_format", "invalid"},
		{"output.precision", "invalid"},
		{"output.precision", "-1"},
		{"output.precision", "11"},
		{"logging.level", "invalid"},
	}
	
	for _, tt := range tests {
		t.Run(tt.key+"="+tt.value, func(t *testing.T) {
			err := cfg.Set(tt.key, tt.value)
			assert.Error(t, err)
		})
	}
}

func TestGetNonexistentKeys(t *testing.T) {
	cfg := DefaultConfig()
	
	tests := []string{
		"invalid",
		"output.invalid", 
		"logging.invalid",
		"plugins.nonexistent.setting",
	}
	
	for _, key := range tests {
		t.Run(key, func(t *testing.T) {
			_, err := cfg.Get(key)
			assert.Error(t, err)
		})
	}
}

func TestInitConfig(t *testing.T) {
	// Override default config path for testing
	originalHome := os.Getenv("HOME")
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)
	
	err := InitConfig()
	assert.NoError(t, err)
	
	// Verify config file was created
	cfg := DefaultConfig()
	_, err = os.Stat(cfg.ConfigFile)
	assert.NoError(t, err)
	
	// Test that init fails if config already exists
	err = InitConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestPluginCustomSettings(t *testing.T) {
	cfg := DefaultConfig()
	
	// Set custom plugin setting
	err := cfg.Set("plugins.custom.endpoint", "https://api.example.com")
	require.NoError(t, err)
	
	// Verify it's stored in Settings
	pluginConfig := cfg.Plugins["custom"]
	assert.Equal(t, "https://api.example.com", pluginConfig.Settings["endpoint"])
	
	// Verify we can retrieve it
	value, err := cfg.Get("plugins.custom.endpoint")
	require.NoError(t, err)
	assert.Equal(t, "https://api.example.com", value)
}