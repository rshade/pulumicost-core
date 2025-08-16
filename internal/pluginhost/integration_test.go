package pluginhost_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/rshade/pulumicost-core/internal/pluginhost"
)

func TestIntegration_ProcessLauncherWithClient(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Test integration between ProcessLauncher and NewClient
	launcher := pluginhost.NewProcessLauncher()

	// Create a mock plugin that will fail but test the integration
	mockPlugin := createFailingMockPlugin(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// This should fail but test the full integration path
	client, err := pluginhost.NewClient(ctx, launcher, mockPlugin)

	if err == nil {
		t.Log("Note: Mock plugin unexpectedly succeeded")
		if client != nil {
			client.Close()
		}
	} else {
		t.Logf("Expected integration error: %v", err)
	}

	// Should not panic or leave hanging processes
}

func TestIntegration_StdioLauncherWithClient(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Test integration between StdioLauncher and NewClient
	launcher := pluginhost.NewStdioLauncher()

	// Create a mock plugin that will fail but test the integration
	mockPlugin := createFailingMockPlugin(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// This should fail but test the full integration path
	client, err := pluginhost.NewClient(ctx, launcher, mockPlugin)

	if err == nil {
		t.Log("Note: Mock plugin unexpectedly succeeded")
		if client != nil {
			client.Close()
		}
	} else {
		t.Logf("Expected integration error: %v", err)
	}

	// Should not panic or leave hanging processes
}

func TestIntegration_LauncherSwitching(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Test that the same binary path works with different launchers
	mockPlugin := createFailingMockPlugin(t)

	launchers := []struct {
		name     string
		launcher pluginhost.Launcher
	}{
		{"ProcessLauncher", pluginhost.NewProcessLauncher()},
		{"StdioLauncher", pluginhost.NewStdioLauncher()},
	}

	for _, launcher := range launchers {
		t.Run(launcher.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			client, err := pluginhost.NewClient(ctx, launcher.launcher, mockPlugin)

			if client != nil {
				client.Close()
			}

			// We expect errors with our mock plugin
			if err != nil {
				t.Logf("%s integration error (expected): %v", launcher.name, err)
			}
		})
	}
}

func TestIntegration_ConcurrentClients(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Test creating multiple clients concurrently
	launcher := pluginhost.NewProcessLauncher()
	mockPlugin := createFailingMockPlugin(t)

	const numClients = 5
	results := make(chan error, numClients)

	// Start multiple client creation attempts concurrently
	for i := range numClients {
		go func(_ int) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			client, err := pluginhost.NewClient(ctx, launcher, mockPlugin)
			if client != nil {
				client.Close()
			}

			results <- err
		}(i)
	}

	// Collect results
	for i := range numClients {
		err := <-results
		if err != nil {
			t.Logf("Client %d error (expected): %v", i, err)
		}
	}

	// Should not panic or deadlock
}

func TestIntegration_RapidCreateDestroy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Test rapid creation and destruction of clients
	launcher := pluginhost.NewProcessLauncher()
	mockPlugin := createFailingMockPlugin(t)

	for i := range 10 {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)

		client, err := pluginhost.NewClient(ctx, launcher, mockPlugin)
		if client != nil {
			client.Close()
		}

		cancel()

		if err != nil {
			t.Logf("Iteration %d error (expected): %v", i, err)
		}

		// Brief pause between iterations
		time.Sleep(10 * time.Millisecond)
	}
}

func TestIntegration_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Test that context cancellation works properly throughout the integration
	launcher := pluginhost.NewProcessLauncher()
	mockPlugin := createWorkingMockPlugin(t)

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after a brief delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	client, err := pluginhost.NewClient(ctx, launcher, mockPlugin)
	if client != nil {
		client.Close()
	}

	// Should handle cancellation gracefully
	if err != nil {
		t.Logf("Cancellation error (expected): %v", err)
	}
}

func TestIntegration_PluginDirectoryStructure(t *testing.T) {
	// Test the plugin directory structure requirements
	tempDir := t.TempDir()

	// Create plugin directory structure: ~/.pulumicost/plugins/<name>/<version>/
	pluginDir := filepath.Join(tempDir, "test-plugin", "v1.0.0")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatalf("failed to create plugin directory: %v", err)
	}

	// Create plugin binary
	binPath := filepath.Join(pluginDir, "test-plugin")
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}

	if err := os.WriteFile(binPath, []byte("#!/bin/bash\necho test"), 0755); err != nil {
		t.Fatalf("failed to create plugin binary: %v", err)
	}

	// Test that plugin is detected as executable
	info, err := os.Stat(binPath)
	if err != nil {
		t.Fatalf("plugin binary not found: %v", err)
	}

	if runtime.GOOS != "windows" {
		if info.Mode()&0111 == 0 {
			t.Error("plugin binary is not executable")
		}
	} else {
		if filepath.Ext(binPath) != ".exe" {
			t.Error("Windows plugin doesn't have .exe extension")
		}
	}

	// Test with launcher (will fail but shouldn't panic)
	launcher := pluginhost.NewProcessLauncher()
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	client, err := pluginhost.NewClient(ctx, launcher, binPath)
	if client != nil {
		client.Close()
	}

	if err != nil {
		t.Logf("Expected error with test plugin: %v", err)
	}
}

func TestIntegration_ErrorRecovery(t *testing.T) {
	// Test that the system recovers gracefully from various error conditions
	launcher := pluginhost.NewProcessLauncher()

	testCases := []struct {
		name     string
		binPath  string
		expected string
	}{
		{
			name:     "nonexistent_binary",
			binPath:  "/nonexistent/plugin/binary",
			expected: "starting plugin",
		},
		{
			name:     "invalid_path",
			binPath:  "",
			expected: "starting plugin",
		},
		{
			name:     "directory_instead_of_binary",
			binPath:  t.TempDir(),
			expected: "starting plugin",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			client, err := pluginhost.NewClient(ctx, launcher, tc.binPath)
			if client != nil {
				client.Close()
			}

			if err == nil {
				t.Errorf("expected error for %s", tc.name)
			} else {
				t.Logf("Expected error for %s: %v", tc.name, err)
			}
		})
	}
}

// Helper functions for integration tests

func createFailingMockPlugin(t *testing.T) string {
	// Create a plugin that will start but fail to serve gRPC
	script := "#!/bin/bash\nexit 1"
	if runtime.GOOS == "windows" {
		script = "exit 1"
	}

	return createTestScript(t, script, ".sh")
}

func createWorkingMockPlugin(t *testing.T) string {
	// Create a plugin that will run but not serve gRPC
	script := "#!/bin/bash\nsleep 10"
	if runtime.GOOS == "windows" {
		script = "timeout 10"
	}

	return createTestScript(t, script, ".sh")
}

func createTestScript(t *testing.T, content, ext string) string {
	if runtime.GOOS == "windows" {
		ext = ".bat"
	}

	tmpfile, err := os.CreateTemp(t.TempDir(), "integration-test-*"+ext)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer tmpfile.Close()

	if _, writeErr := tmpfile.WriteString(content); writeErr != nil {
		t.Fatalf("failed to write script: %v", writeErr)
	}

	if chmodErr := tmpfile.Chmod(0755); chmodErr != nil {
		t.Fatalf("failed to make script executable: %v", chmodErr)
	}

	// Clean up after test
	t.Cleanup(func() {
		os.Remove(tmpfile.Name())
	})

	return tmpfile.Name()
}
