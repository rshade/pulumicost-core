package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestPluginInitCommand(t *testing.T) {
	cmd := NewPluginInitCmd()

	if cmd.Use != "init <plugin-name>" {
		t.Errorf("Expected Use 'init <plugin-name>', got %s", cmd.Use)
	}

	if cmd.Short != "Initialize a new plugin development project" {
		t.Errorf("Expected short description about initializing plugin project, got %s", cmd.Short)
	}
}

func TestPluginInitValidation(t *testing.T) {
	testCases := []struct {
		name      string
		args      []string
		opts      PluginInitOptions
		expectErr bool
	}{
		{
			name: "valid plugin name",
			args: []string{"aws-plugin"},
			opts: PluginInitOptions{
				Author:    "Test Author",
				Providers: []string{"aws"},
			},
			expectErr: false,
		},
		{
			name: "invalid plugin name with uppercase",
			args: []string{"AWS-Plugin"},
			opts: PluginInitOptions{
				Author:    "Test Author",
				Providers: []string{"aws"},
			},
			expectErr: true,
		},
		{
			name: "invalid plugin name with underscore",
			args: []string{"aws_plugin"},
			opts: PluginInitOptions{
				Author:    "Test Author",
				Providers: []string{"aws"},
			},
			expectErr: true,
		},
		{
			name: "empty providers",
			args: []string{"aws-plugin"},
			opts: PluginInitOptions{
				Author:    "Test Author",
				Providers: []string{},
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tc.opts.OutputDir = tmpDir
			tc.opts.Name = tc.args[0]
			tc.opts.Force = true

			// Create a mock command for testing
			cmd := &cobra.Command{
				RunE: func(cmd *cobra.Command, args []string) error {
					return runPluginInit(cmd, &tc.opts)
				},
			}

			err := cmd.RunE(cmd, tc.args)

			if tc.expectErr && err == nil {
				t.Errorf("Expected error, got none")
			}
			if !tc.expectErr && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

func TestPluginInitProjectGeneration(t *testing.T) {
	tmpDir := t.TempDir()

	opts := &PluginInitOptions{
		Name:      "test-plugin",
		Author:    "Test Author",
		Providers: []string{"aws", "azure"},
		OutputDir: tmpDir,
		Force:     true,
	}

	cmd := &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPluginInit(cmd, opts)
		},
	}

	err := cmd.RunE(cmd, []string{"test-plugin"})
	if err != nil {
		t.Fatalf("Plugin init failed: %v", err)
	}

	// Check that project directory was created
	projectDir := filepath.Join(tmpDir, "test-plugin")
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		t.Errorf("Project directory was not created: %s", projectDir)
	}

	// Check key files exist
	expectedFiles := []string{
		"go.mod",
		"manifest.yaml",
		"cmd/plugin/main.go",
		"internal/pricing/calculator.go",
		"internal/pricing/data.go",
		"internal/client/client.go",
		"Makefile",
		"README.md",
		"internal/pricing/calculator_test.go",
	}

	for _, file := range expectedFiles {
		fullPath := filepath.Join(projectDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Expected file was not created: %s", file)
		}
	}

	// Check directories exist
	expectedDirs := []string{
		"cmd/plugin",
		"internal/pricing",
		"internal/client",
		"examples",
		"bin",
	}

	for _, dir := range expectedDirs {
		fullPath := filepath.Join(projectDir, dir)
		if info, err := os.Stat(fullPath); os.IsNotExist(err) || !info.IsDir() {
			t.Errorf("Expected directory was not created: %s", dir)
		}
	}
}

func TestPluginInitForceOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test-plugin")

	// Create existing directory
	err := os.MkdirAll(projectDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	opts := &PluginInitOptions{
		Name:      "test-plugin",
		Author:    "Test Author",
		Providers: []string{"aws"},
		OutputDir: tmpDir,
		Force:     false, // Don't force overwrite
	}

	cmd := &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPluginInit(cmd, opts)
		},
	}

	// Should fail without force
	err = cmd.RunE(cmd, []string{"test-plugin"})
	if err == nil {
		t.Errorf("Expected error when directory exists and force=false, got none")
	}

	// Should succeed with force
	opts.Force = true
	err = cmd.RunE(cmd, []string{"test-plugin"})
	if err != nil {
		t.Errorf("Expected success with force=true, got error: %v", err)
	}
}

func TestIsValidPluginName(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid simple name", "aws", true},
		{"valid hyphenated name", "aws-plugin", true},
		{"valid with numbers", "aws-plugin-v2", true},
		{"too short", "a", false},
		{"starts with hyphen", "-aws", false},
		{"ends with hyphen", "aws-", false},
		{"contains uppercase", "AWS", false},
		{"contains underscore", "aws_plugin", false},
		{"contains space", "aws plugin", false},
		{"contains dot", "aws.plugin", false},
		{"too long", "this-is-a-very-long-plugin-name-that-exceeds-the-maximum-allowed-length", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isValidPluginName(tc.input)
			if result != tc.expected {
				t.Errorf("isValidPluginName(%q) = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}
