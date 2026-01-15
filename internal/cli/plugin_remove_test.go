package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rshade/finfocus/internal/cli"
)

func TestPluginRemoveCmd_Help(t *testing.T) {
	// Set log level to error to avoid cluttering test output with debug logs
	t.Setenv("FINFOCUS_LOG_LEVEL", "error")
	rootCmd := cli.NewRootCmd("test")

	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"plugin", "remove", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()

	// Check for expected content
	expectedStrings := []string{
		"Remove",
		"--keep-config",
		"--plugin-dir",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("help output missing expected string: %q", expected)
		}
	}
}

func TestPluginRemoveCmd_NoArgs(t *testing.T) {
	// Set log level to error to avoid cluttering test output with debug logs
	t.Setenv("FINFOCUS_LOG_LEVEL", "error")
	rootCmd := cli.NewRootCmd("test")

	var stderr bytes.Buffer
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"plugin", "remove"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error when no plugin specified")
	}

	errOutput := stderr.String()
	if !strings.Contains(errOutput, "accepts 1 arg") {
		t.Errorf("expected 'accepts 1 arg' error, got: %s", errOutput)
	}
}

func TestPluginRemoveCmd_NotInstalled(t *testing.T) {
	// Set log level to error to avoid cluttering test output with debug logs
	t.Setenv("FINFOCUS_LOG_LEVEL", "error")
	rootCmd := cli.NewRootCmd("test")

	// Set HOME to temp directory
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	var stderr bytes.Buffer
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"plugin", "remove", "nonexistent-plugin"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for non-installed plugin")
	}
}

func TestPluginRemoveCmd_Flags(t *testing.T) {
	// Set log level to error to avoid cluttering test output with debug logs
	t.Setenv("FINFOCUS_LOG_LEVEL", "error")
	rootCmd := cli.NewRootCmd("test")

	// Get the remove command to check flags
	pluginCmd, _, err := rootCmd.Find([]string{"plugin", "remove"})
	if err != nil {
		t.Fatalf("failed to find remove command: %v", err)
	}

	// Check that expected flags exist
	expectedFlags := []string{"keep-config", "plugin-dir"}
	for _, flag := range expectedFlags {
		if pluginCmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag --%s not found", flag)
		}
	}
}

func TestPluginRemoveCmd_Aliases(t *testing.T) {
	// Set log level to error to avoid cluttering test output with debug logs
	t.Setenv("FINFOCUS_LOG_LEVEL", "error")
	rootCmd := cli.NewRootCmd("test")

	// Test that "uninstall" alias works
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"plugin", "uninstall", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error with uninstall alias: %v", err)
	}
}
