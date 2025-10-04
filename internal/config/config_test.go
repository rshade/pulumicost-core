//nolint:testifylint,usetesting // Test style preferences are acceptable
package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubHome sets up an isolated HOME directory for testing to prevent
// tests from reading/writing the real ~/.pulumicost directory.
func stubHome(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir) // Windows
}

func TestConfig_NewAndDefaults(t *testing.T) {
	stubHome(t)
	cfg := New()

	assert.Equal(t, "table", cfg.Output.DefaultFormat)
	assert.Equal(t, 2, cfg.Output.Precision)
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.NotEmpty(t, cfg.Logging.File)
	assert.NotNil(t, cfg.Plugins)
	assert.NotEmpty(t, cfg.PluginDir)
	assert.NotEmpty(t, cfg.SpecDir)
}

func TestConfig_SetGetValues(t *testing.T) {
	stubHome(t)
	cfg := New()

	// Test output values
	err := cfg.Set("output.default_format", "json")
	require.NoError(t, err)

	value, err := cfg.Get("output.default_format")
	require.NoError(t, err)
	assert.Equal(t, "json", value)

	err = cfg.Set("output.precision", "4")
	require.NoError(t, err)

	value, err = cfg.Get("output.precision")
	require.NoError(t, err)
	assert.Equal(t, 4, value)

	// Test plugin values
	err = cfg.Set("plugins.aws.region", "us-west-2")
	require.NoError(t, err)

	value, err = cfg.Get("plugins.aws.region")
	require.NoError(t, err)
	assert.Equal(t, "us-west-2", value)

	// Test logging values
	err = cfg.Set("logging.level", "debug")
	require.NoError(t, err)

	value, err = cfg.Get("logging.level")
	require.NoError(t, err)
	assert.Equal(t, "debug", value)
}

func TestConfig_SetErrors(t *testing.T) {
	stubHome(t)
	cfg := New()

	// Invalid section
	err := cfg.Set("invalid.key", "value")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown configuration section")

	// Invalid output key
	err = cfg.Set("output.invalid", "value")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown output setting")

	// Invalid precision value
	err = cfg.Set("output.precision", "invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "precision must be a number")

	// Invalid plugin key format
	err = cfg.Set("plugins.aws", "value")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin key must be in format")
}

func TestConfig_GetErrors(t *testing.T) {
	stubHome(t)
	cfg := New()

	// Unknown section
	_, err := cfg.Get("invalid.key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown configuration section")

	// Unknown output key
	_, err = cfg.Get("output.invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown output setting")

	// Unknown plugin
	_, err = cfg.Get("plugins.nonexistent.key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin not found")
}

func TestConfig_Validation(t *testing.T) {
	stubHome(t)
	cfg := New()

	// Valid configuration should pass
	err := cfg.Validate()
	assert.NoError(t, err)

	// Invalid output format
	cfg.Output.DefaultFormat = "invalid"
	err = cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")

	// Reset and test invalid precision
	cfg.Output.DefaultFormat = "table"
	cfg.Output.Precision = -1
	err = cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid precision")

	// Reset and test invalid log level
	cfg.Output.Precision = 2
	cfg.Logging.Level = "invalid"
	err = cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid log level")
}

func TestConfig_EncryptDecrypt(t *testing.T) {
	stubHome(t)
	cfg := New()

	original := "secret-value-123"

	encrypted, err := cfg.EncryptValue(original)
	require.NoError(t, err)
	assert.NotEqual(t, original, encrypted)
	assert.NotEmpty(t, encrypted)

	decrypted, err := cfg.DecryptValue(encrypted)
	require.NoError(t, err)
	assert.Equal(t, original, decrypted)
}

func TestConfig_EncryptDecryptErrors(t *testing.T) {
	stubHome(t)
	cfg := New()

	// Invalid base64 for decryption
	_, err := cfg.DecryptValue("invalid-base64")
	assert.Error(t, err)

	// Invalid encrypted data (too short)
	_, err = cfg.DecryptValue("dGVzdA==") // "test" in base64, too short for GCM
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ciphertext too short")
}

func TestConfig_SaveLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pulumicost-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	cfg := &Config{
		Output: OutputConfig{
			DefaultFormat: "json",
			Precision:     4,
		},
		Plugins: map[string]PluginConfig{
			"aws": {Config: map[string]interface{}{"region": "us-west-2"}},
		},
		Logging: LoggingConfig{
			Level: "debug",
			File:  filepath.Join(t.TempDir(), "test.log"),
		},
		configPath: filepath.Join(tmpDir, "config.yaml"),
		encKey:     deriveKey(),
	}

	// Save configuration
	err = cfg.Save()
	require.NoError(t, err)

	// Load configuration
	cfg2 := &Config{
		configPath: cfg.configPath,
		encKey:     deriveKey(),
	}
	err = cfg2.Load()
	require.NoError(t, err)

	assert.Equal(t, cfg.Output.DefaultFormat, cfg2.Output.DefaultFormat)
	assert.Equal(t, cfg.Output.Precision, cfg2.Output.Precision)
	assert.Equal(t, cfg.Logging.Level, cfg2.Logging.Level)
	assert.Equal(t, cfg.Logging.File, cfg2.Logging.File)
	assert.Equal(t, len(cfg.Plugins), len(cfg2.Plugins))

	awsConfig, exists := cfg2.Plugins["aws"]
	assert.True(t, exists)
	assert.Equal(t, "us-west-2", awsConfig.Config["region"])
}

func TestConfig_List(t *testing.T) {
	stubHome(t)
	cfg := New()
	cfg.Set("plugins.aws.region", "us-west-2")
	cfg.Set("output.default_format", "json")

	list := cfg.List()

	assert.Contains(t, list, "output")
	assert.Contains(t, list, "plugins")
	assert.Contains(t, list, "logging")

	output := list["output"].(OutputConfig)
	assert.Equal(t, "json", output.DefaultFormat)
}

func TestConfig_PluginMethods(t *testing.T) {
	stubHome(t)
	cfg := New()

	// Test GetPluginConfig for non-existent plugin
	config, err := cfg.GetPluginConfig("nonexistent")
	require.NoError(t, err)
	assert.Empty(t, config)

	// Test SetPluginConfig
	pluginConfig := map[string]interface{}{
		"region":  "us-east-1",
		"profile": "production",
	}
	cfg.SetPluginConfig("aws", pluginConfig)

	// Test GetPluginConfig for existing plugin
	retrievedConfig, err := cfg.GetPluginConfig("aws")
	require.NoError(t, err)
	assert.Equal(t, pluginConfig, retrievedConfig)
}

func TestConfig_EnvironmentOverrides(t *testing.T) {
	// Set environment variables
	customLogFile := filepath.Join(t.TempDir(), "custom.log")
	os.Setenv("PULUMICOST_OUTPUT_FORMAT", "json")
	os.Setenv("PULUMICOST_OUTPUT_PRECISION", "5")
	os.Setenv("PULUMICOST_LOG_LEVEL", "debug")
	os.Setenv("PULUMICOST_LOG_FILE", customLogFile)
	os.Setenv("PULUMICOST_PLUGIN_AWS_REGION", "eu-west-1")
	os.Setenv("PULUMICOST_PLUGIN_AWS_PROFILE", "test")

	defer func() {
		os.Unsetenv("PULUMICOST_OUTPUT_FORMAT")
		os.Unsetenv("PULUMICOST_OUTPUT_PRECISION")
		os.Unsetenv("PULUMICOST_LOG_LEVEL")
		os.Unsetenv("PULUMICOST_LOG_FILE")
		os.Unsetenv("PULUMICOST_PLUGIN_AWS_REGION")
		os.Unsetenv("PULUMICOST_PLUGIN_AWS_PROFILE")
	}()

	stubHome(t)
	cfg := New()

	assert.Equal(t, "json", cfg.Output.DefaultFormat)
	assert.Equal(t, 5, cfg.Output.Precision)
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, customLogFile, cfg.Logging.File)

	awsConfig, exists := cfg.Plugins["aws"]
	assert.True(t, exists)
	assert.Equal(t, "eu-west-1", awsConfig.Config["region"])
	assert.Equal(t, "test", awsConfig.Config["profile"])
}

func TestConfig_BackwardCompatibility(t *testing.T) {
	stubHome(t)
	cfg := New()

	// Test legacy methods still work
	assert.NotEmpty(t, cfg.PluginDir)
	assert.NotEmpty(t, cfg.SpecDir)

	pluginPath := cfg.PluginPath("test-plugin", "1.0.0")
	assert.Contains(t, pluginPath, "test-plugin")
	assert.Contains(t, pluginPath, "1.0.0")
}
