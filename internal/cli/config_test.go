package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/rshade/pulumicost-core/internal/cli"
	"github.com/rshade/pulumicost-core/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestConfig(t *testing.T) (string, func()) {
	t.Helper()
	testHome := t.TempDir()
	t.Setenv("HOME", testHome)
	// If a reset helper exists, call it here; otherwise noop.
	cleanup := func() {}
	return testHome, cleanup
}

func TestConfigInitCmd(t *testing.T) {
	testHome, cleanup := setupTestConfig(t)
	defer cleanup()
	
	cmd := cli.NewConfigInitCmd()
	cmd.SetOutput(os.Stdout)
	
	var output bytes.Buffer
	cmd.SetOutput(&output)
	
	// Test successful init
	err := cmd.Execute()
	require.NoError(t, err)
	
	// Check that config file was created
	configPath := filepath.Join(testHome, ".pulumicost", "config.yaml")
	_, err = os.Stat(configPath)
	assert.NoError(t, err)
	
	// Check output message
	assert.Contains(t, output.String(), "Configuration initialized successfully")
}

func TestConfigInitCmdForce(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()
	
	// Create existing config
	cfg := config.New()
	err := cfg.Save()
	require.NoError(t, err)
	
	cmd := cli.NewConfigInitCmd()
	var output bytes.Buffer
	cmd.SetOutput(&output)
	
	// Test without force flag should fail
	err = cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	
	// Test with force flag should succeed
	output.Reset()
	cmd.SetArgs([]string{"--force"})
	err = cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, output.String(), "Configuration initialized successfully")
}

func TestConfigSetCmd(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()
	
	cmd := cli.NewConfigSetCmd()
	var output bytes.Buffer
	cmd.SetOutput(&output)
	
	// Test setting output format
	cmd.SetArgs([]string{"output.default_format", "json"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, output.String(), "Configuration updated: output.default_format = json")
	
	// Verify the value was set
	cfg := config.New()
	require.NoError(t, cfg.Load())
	value, err := cfg.Get("output.default_format")
	require.NoError(t, err)
	assert.Equal(t, "json", value)
}

func TestConfigSetCmdWithEncryption(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()
	
	cmd := cli.NewConfigSetCmd()
	var output bytes.Buffer
	cmd.SetOutput(&output)
	
	// Test setting encrypted value
	cmd.SetArgs([]string{"plugins.aws.secret_key", "my-secret", "--encrypt"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, output.String(), "Value encrypted before storage")
	assert.Contains(t, output.String(), "Configuration updated")
	
	// Verify the value was encrypted
	cfg := config.New()
	require.NoError(t, cfg.Load())
	value, err := cfg.Get("plugins.aws.secret_key")
	require.NoError(t, err)
	assert.NotEqual(t, "my-secret", value) // Should be encrypted
	assert.IsType(t, "", value)           // Should be string
}

func TestConfigSetCmdErrors(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()
	
	cmd := cli.NewConfigSetCmd()
	var output bytes.Buffer
	cmd.SetOutput(&output)
	
	// Test invalid key
	cmd.SetArgs([]string{"invalid.key", "value"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown configuration section")
	
	// Test invalid precision value
	cmd.SetArgs([]string{"output.precision", "invalid"})
	err = cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "precision must be a number")
}

func TestConfigGetCmd(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()
	
	// Set up some config values
	cfg := config.New()
	require.NoError(t, cfg.Set("output.default_format", "json"))
	require.NoError(t, cfg.Set("plugins.aws.region", "us-west-2"))
	require.NoError(t, cfg.Save())
	
	cmd := cli.NewConfigGetCmd()
	var output bytes.Buffer
	cmd.SetOutput(&output)
	
	// Test getting simple value
	cmd.SetArgs([]string{"output.default_format"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "json\n", output.String())
	
	// Test getting plugin value
	output.Reset()
	cmd.SetArgs([]string{"plugins.aws.region"})
	err = cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "us-west-2\n", output.String())
}

func TestConfigGetCmdWithDecryption(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()
	
	// Set up encrypted value
	cfg := config.New()
	encrypted, err := cfg.EncryptValue("secret-value")
	require.NoError(t, err)
	require.NoError(t, cfg.Set("plugins.aws.secret_key", encrypted))
	require.NoError(t, cfg.Save())
	
	cmd := cli.NewConfigGetCmd()
	var output bytes.Buffer
	cmd.SetOutput(&output)
	
	// Test getting encrypted value without decryption
	cmd.SetArgs([]string{"plugins.aws.secret_key"})
	err = cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, output.String(), encrypted) // Should show encrypted value
	assert.NotContains(t, output.String(), "secret-value")
	
	// Test getting encrypted value with decryption
	output.Reset()
	cmd.SetArgs([]string{"plugins.aws.secret_key", "--decrypt"})
	err = cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "secret-value\n", output.String())
}

func TestConfigGetCmdErrors(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()
	
	// Create initial config file
	cfg := config.New()
	require.NoError(t, cfg.Save())
	
	cmd := cli.NewConfigGetCmd()
	var output bytes.Buffer
	cmd.SetOutput(&output)
	
	// Test invalid key
	cmd.SetArgs([]string{"invalid.key"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown configuration section")
	
	// Test non-existent plugin
	cmd.SetArgs([]string{"plugins.nonexistent.key"})
	err = cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin not found")
}

func TestConfigListCmd(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()
	
	// Set up some config values
	cfg := config.New()
	require.NoError(t, cfg.Set("output.default_format", "json"))
	require.NoError(t, cfg.Set("plugins.aws.region", "us-west-2"))
	require.NoError(t, cfg.Save())
	
	cmd := cli.NewConfigListCmd()
	var output bytes.Buffer
	cmd.SetOutput(&output)
	
	// Test YAML output (default)
	err := cmd.Execute()
	require.NoError(t, err)
	
	yamlOutput := output.String()
	assert.Contains(t, yamlOutput, "output:")
	assert.Contains(t, yamlOutput, "default_format: json")
	assert.Contains(t, yamlOutput, "plugins:")
	assert.Contains(t, yamlOutput, "aws:")
	assert.Contains(t, yamlOutput, "region: us-west-2")
	
	// Test JSON output
	output.Reset()
	cmd.SetArgs([]string{"--format", "json"})
	err = cmd.Execute()
	require.NoError(t, err)
	
	jsonOutput := output.String()
	assert.Contains(t, jsonOutput, "\"output\":")
	assert.Contains(t, jsonOutput, "\"default_format\": \"json\"")
	assert.Contains(t, jsonOutput, "\"plugins\":")
}

func TestConfigListCmdErrors(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()
	
	// Create initial config file
	cfg := config.New()
	require.NoError(t, cfg.Save())
	
	cmd := cli.NewConfigListCmd()
	var output bytes.Buffer
	cmd.SetOutput(&output)
	
	// Test invalid format
	cmd.SetArgs([]string{"--format", "invalid"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestConfigValidateCmd(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()
	
	// Create initial config file
	cfg := config.New()
	require.NoError(t, cfg.Save())
	
	cmd := cli.NewConfigValidateCmd()
	var output bytes.Buffer
	cmd.SetOutput(&output)
	
	// Test valid configuration
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, output.String(), "✅ Configuration is valid")
	
	// Test with verbose flag
	output.Reset()
	cmd.SetArgs([]string{"--verbose"})
	err = cmd.Execute()
	require.NoError(t, err)
	
	verboseOutput := output.String()
	assert.Contains(t, verboseOutput, "✅ Configuration is valid")
	assert.Contains(t, verboseOutput, "Configuration details:")
	assert.Contains(t, verboseOutput, "Output format:")
	assert.Contains(t, verboseOutput, "Logging level:")
}

func TestConfigValidateCmdErrors(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()
	
	// Set invalid configuration
	cfg := config.New()
	// Ensure file exists with current defaults first.
	require.NoError(t, cfg.Save())
	// Reload, mutate, and save invalid field.
	cfg.Output.DefaultFormat = "invalid"
	require.NoError(t, cfg.Save())
	
	cmd := cli.NewConfigValidateCmd()
	var output bytes.Buffer
	cmd.SetOutput(&output)
	
	// Test invalid configuration
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
}

func TestConfigCommandsIntegration(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()
	
	var output bytes.Buffer
	
	// Test full workflow: init -> set -> get -> validate -> list
	
	// 1. Initialize config
	initCmd := cli.NewConfigInitCmd()
	initCmd.SetOutput(&output)
	err := initCmd.Execute()
	require.NoError(t, err)
	
	// 2. Set some values
	setCmd := cli.NewConfigSetCmd()
	setCmd.SetOutput(&output)
	setCmd.SetArgs([]string{"output.default_format", "json"})
	err = setCmd.Execute()
	require.NoError(t, err)
	
	setCmd2 := cli.NewConfigSetCmd()
	setCmd2.SetOutput(&output)
	setCmd2.SetArgs([]string{"plugins.aws.region", "eu-west-1"})
	err = setCmd2.Execute()
	require.NoError(t, err)
	
	// 3. Get values to verify
	getCmd := cli.NewConfigGetCmd()
	output.Reset()
	getCmd.SetOutput(&output)
	getCmd.SetArgs([]string{"output.default_format"})
	err = getCmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "json\n", output.String())
	
	// 4. Validate configuration
	validateCmd := cli.NewConfigValidateCmd()
	output.Reset()
	validateCmd.SetOutput(&output)
	err = validateCmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, output.String(), "✅ Configuration is valid")
	
	// 5. List all configuration
	listCmd := cli.NewConfigListCmd()
	output.Reset()
	listCmd.SetOutput(&output)
	err = listCmd.Execute()
	require.NoError(t, err)
	
	listOutput := output.String()
	assert.Contains(t, listOutput, "default_format: json")
	assert.Contains(t, listOutput, "region: eu-west-1")
}

func TestConfigCmdWrongArgs(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()
	
	// Test set command with wrong number of args
	setCmd := cli.NewConfigSetCmd()
	setCmd.SetArgs([]string{"only-one-arg"})
	err := setCmd.Execute()
	assert.Error(t, err)
	
	// Test get command with wrong number of args
	getCmd := cli.NewConfigGetCmd()
	getCmd.SetArgs([]string{})
	err = getCmd.Execute()
	assert.Error(t, err)
}