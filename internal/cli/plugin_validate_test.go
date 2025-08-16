package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rshade/pulumicost-core/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPluginValidateCmd(t *testing.T) {
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
			name:        "plugin flag",
			args:        []string{"--plugin", "test-plugin"},
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
			cmd := newPluginValidateCmd()
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

func TestPluginValidateCmdFlags(t *testing.T) {
	cmd := newPluginValidateCmd()

	// Check plugin flag
	pluginFlag := cmd.Flags().Lookup("plugin")
	assert.NotNil(t, pluginFlag)
	assert.Equal(t, "string", pluginFlag.Value.Type())
	assert.Equal(t, "", pluginFlag.DefValue)
	assert.Contains(t, pluginFlag.Usage, "Validate a specific plugin by name")
}

func TestPluginValidateCmdHelp(t *testing.T) {
	var buf bytes.Buffer
	cmd := newPluginValidateCmd()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Validate that installed plugins can be loaded")
	assert.Contains(t, output, "Validate that installed plugins can be loaded")
	assert.Contains(t, output, "--plugin")
	assert.Contains(t, output, "Validate a specific plugin by name")
}

func TestPluginValidateCmdExamples(t *testing.T) {
	cmd := newPluginValidateCmd()

	// Check that examples are present
	assert.NotEmpty(t, cmd.Example)
	assert.Contains(t, cmd.Example, "pulumicost plugin validate")
	assert.Contains(t, cmd.Example, "pulumicost plugin validate --plugin aws-plugin")
	assert.Contains(t, cmd.Example, "Validate a specific plugin")
	assert.Contains(t, cmd.Example, "kubecost")
}

func TestValidatePlugin(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "plugin-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a mock plugin binary
	pluginPath := filepath.Join(tmpDir, "test-plugin")
	err = os.WriteFile(pluginPath, []byte("#!/bin/sh\necho test"), 0755)
	require.NoError(t, err)

	// Create a valid manifest
	manifestPath := filepath.Join(tmpDir, "plugin.manifest.json")
	manifestContent := `{
		"name": "test-plugin",
		"version": "1.0.0",
		"description": "Test plugin"
	}`
	err = os.WriteFile(manifestPath, []byte(manifestContent), 0644)
	require.NoError(t, err)

	tests := []struct {
		name        string
		plugin      registry.PluginInfo
		setupFunc   func()
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid plugin",
			plugin: registry.PluginInfo{
				Name:    "test-plugin",
				Version: "1.0.0",
				Path:    pluginPath,
			},
			expectError: false,
		},
		{
			name: "plugin not found",
			plugin: registry.PluginInfo{
				Name:    "missing",
				Version: "1.0.0",
				Path:    filepath.Join(tmpDir, "missing"),
			},
			expectError: true,
			errorMsg:    "plugin binary not found",
		},
		{
			name: "plugin is directory",
			plugin: registry.PluginInfo{
				Name:    "dir-plugin",
				Version: "1.0.0",
				Path:    tmpDir,
			},
			expectError: true,
			errorMsg:    "plugin path is a directory",
		},
		{
			name: "plugin not executable",
			plugin: registry.PluginInfo{
				Name:    "non-exec",
				Version: "1.0.0",
				Path:    filepath.Join(tmpDir, "non-exec"),
			},
			setupFunc: func() {
				// Create non-executable file
				nonExecPath := filepath.Join(tmpDir, "non-exec")
				err := os.WriteFile(nonExecPath, []byte("test"), 0644)
				require.NoError(t, err)
			},
			expectError: true,
			errorMsg:    "plugin binary is not executable",
		},
		{
			name: "manifest name mismatch",
			plugin: registry.PluginInfo{
				Name:    "wrong-name",
				Version: "1.0.0",
				Path:    pluginPath,
			},
			expectError: true,
			errorMsg:    "manifest name mismatch",
		},
		{
			name: "manifest version mismatch",
			plugin: registry.PluginInfo{
				Name:    "test-plugin",
				Version: "2.0.0",
				Path:    pluginPath,
			},
			expectError: true,
			errorMsg:    "manifest version mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			ctx := context.Background()
			err := validatePlugin(ctx, tt.plugin)

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
