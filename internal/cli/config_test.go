package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rshade/pulumicost-core/internal/cli"
	"github.com/rshade/pulumicost-core/internal/config"
)

func TestConfigInitCmd(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)
	
	cmd := cli.NewConfigInitCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	
	err := cmd.Execute()
	require.NoError(t, err)
	
	// Verify config file was created
	cfg := config.DefaultConfig()
	_, err = os.Stat(cfg.ConfigFile)
	assert.NoError(t, err)
	
	// Verify output message
	output := out.String()
	assert.Contains(t, output, "Configuration initialized at")
}

func TestConfigInitCmdAlreadyExists(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)
	
	// Create config file first
	err := config.InitConfig()
	require.NoError(t, err)
	
	cmd := cli.NewConfigInitCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	
	err = cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestConfigSetCmd(t *testing.T) {
	tempDir := t.TempDir()
	setupTestConfig(t, tempDir)
	
	tests := []struct {
		name  string
		args  []string
		flags []string
		want  string
	}{
		{
			name: "set output format",
			args: []string{"output.default_format", "json"},
			want: "Configuration output.default_format set to json",
		},
		{
			name: "set precision",
			args: []string{"output.precision", "4"},
			want: "Configuration output.precision set to 4",
		},
		{
			name: "set plugin region",
			args: []string{"plugins.aws.region", "us-west-2"},
			want: "Configuration plugins.aws.region set to us-west-2",
		},
		{
			name:  "set credential",
			args:  []string{"plugins.aws.access_key", "AKIATEST"},
			flags: []string{"--credential"},
			want:  "Credential access_key set for plugin aws (encrypted)",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := cli.NewConfigSetCmd()
			var out bytes.Buffer
			cmd.SetOut(&out)
			
			args := append(tt.args, tt.flags...)
			cmd.SetArgs(args)
			
			err := cmd.Execute()
			require.NoError(t, err)
			
			output := out.String()
			assert.Contains(t, output, tt.want)
		})
	}
}

func TestConfigSetCmdInvalid(t *testing.T) {
	tempDir := t.TempDir()
	setupTestConfig(t, tempDir)
	
	tests := []struct {
		name string
		args []string
	}{
		{"invalid key", []string{"invalid.key", "value"}},
		{"invalid output format", []string{"output.default_format", "invalid"}},
		{"invalid precision", []string{"output.precision", "invalid"}},
		{"precision out of range", []string{"output.precision", "11"}},
		{"invalid log level", []string{"logging.level", "invalid"}},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := cli.NewConfigSetCmd()
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			
			cmd.SetArgs(tt.args)
			
			err := cmd.Execute()
			assert.Error(t, err)
		})
	}
}

func TestConfigGetCmd(t *testing.T) {
	tempDir := t.TempDir()
	setupTestConfig(t, tempDir)
	
	// Set up some test data
	cfg, err := config.Load()
	require.NoError(t, err)
	
	err = cfg.Set("output.default_format", "json")
	require.NoError(t, err)
	
	err = cfg.Set("plugins.aws.region", "us-east-1")
	require.NoError(t, err)
	
	err = cfg.SetCredential("aws", "access_key", "AKIATEST")
	require.NoError(t, err)
	
	err = cfg.Save()
	require.NoError(t, err)
	
	tests := []struct {
		name  string
		args  []string
		flags []string
		want  string
	}{
		{
			name: "get output format",
			args: []string{"output.default_format"},
			want: "json",
		},
		{
			name: "get plugin region",
			args: []string{"plugins.aws.region"},
			want: "us-east-1",
		},
		{
			name:  "get credential",
			args:  []string{"plugins.aws.access_key"},
			flags: []string{"--credential"},
			want:  "AKIATEST",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := cli.NewConfigGetCmd()
			var out bytes.Buffer
			cmd.SetOut(&out)
			
			args := append(tt.args, tt.flags...)
			cmd.SetArgs(args)
			
			err := cmd.Execute()
			require.NoError(t, err)
			
			output := strings.TrimSpace(out.String())
			assert.Equal(t, tt.want, output)
		})
	}
}

func TestConfigGetCmdNonexistent(t *testing.T) {
	tempDir := t.TempDir()
	setupTestConfig(t, tempDir)
	
	cmd := cli.NewConfigGetCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	
	cmd.SetArgs([]string{"nonexistent.key"})
	
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown config section")
}

func TestConfigListCmd(t *testing.T) {
	tempDir := t.TempDir()
	setupTestConfig(t, tempDir)
	
	// Set up some test data
	cfg, err := config.Load()
	require.NoError(t, err)
	
	err = cfg.Set("output.default_format", "json")
	require.NoError(t, err)
	
	err = cfg.Set("plugins.aws.region", "us-east-1")
	require.NoError(t, err)
	
	err = cfg.Save()
	require.NoError(t, err)
	
	cmd := cli.NewConfigListCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	
	err = cmd.Execute()
	require.NoError(t, err)
	
	output := out.String()
	assert.Contains(t, output, "output.default_format")
	assert.Contains(t, output, "json")
	assert.Contains(t, output, "plugins.aws.region")
	assert.Contains(t, output, "us-east-1")
}

func TestConfigListCmdEmpty(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)
	
	// No config file exists
	cmd := cli.NewConfigListCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	
	err := cmd.Execute()
	require.NoError(t, err)
	
	output := out.String()
	assert.Contains(t, output, "No configuration found")
}

func TestConfigValidateCmd(t *testing.T) {
	tempDir := t.TempDir()
	setupTestConfig(t, tempDir)
	
	cmd := cli.NewConfigValidateCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	
	err := cmd.Execute()
	require.NoError(t, err)
	
	output := out.String()
	assert.Contains(t, output, "✅ Configuration is valid")
}

func TestConfigValidateCmdInvalid(t *testing.T) {
	tempDir := t.TempDir()
	setupTestConfig(t, tempDir)
	
	// Create invalid config
	cfg, err := config.Load()
	require.NoError(t, err)
	
	cfg.Output.DefaultFormat = "invalid"
	err = cfg.Save()
	require.NoError(t, err)
	
	cmd := cli.NewConfigValidateCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	
	err = cmd.Execute()
	assert.Error(t, err)
	
	output := out.String()
	assert.Contains(t, output, "❌ Configuration validation failed")
}

// setupTestConfig creates a test configuration directory
func setupTestConfig(t *testing.T, tempDir string) {
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() { os.Setenv("HOME", originalHome) })
	
	// Initialize config
	err := config.InitConfig()
	require.NoError(t, err)
}