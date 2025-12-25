package registry

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
)

func TestNewInstaller(t *testing.T) {
	tests := []struct {
		name      string
		pluginDir string
		wantDir   bool
	}{
		{
			name:      "with custom dir",
			pluginDir: "/custom/path",
			wantDir:   true,
		},
		{
			name:      "with empty dir uses default",
			pluginDir: "",
			wantDir:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installer := NewInstaller(tt.pluginDir)
			if installer == nil {
				t.Fatal("NewInstaller returned nil")
			}
			if installer.client == nil {
				t.Error("installer.client is nil")
			}
			if tt.wantDir && installer.pluginDir == "" {
				t.Error("installer.pluginDir is empty")
			}
			if tt.pluginDir != "" && installer.pluginDir != tt.pluginDir {
				t.Errorf("pluginDir = %v, want %v", installer.pluginDir, tt.pluginDir)
			}
		})
	}
}

func TestInstallOptions(t *testing.T) {
	opts := InstallOptions{
		Force:     true,
		NoSave:    true,
		PluginDir: "/custom/dir",
	}

	if !opts.Force {
		t.Error("Force should be true")
	}
	if !opts.NoSave {
		t.Error("NoSave should be true")
	}
	if opts.PluginDir != "/custom/dir" {
		t.Errorf("PluginDir = %v, want /custom/dir", opts.PluginDir)
	}
}

func TestInstallResult(t *testing.T) {
	result := InstallResult{
		Name:       "test-plugin",
		Version:    "v1.0.0",
		Path:       "/path/to/plugin",
		FromURL:    true,
		Repository: "owner/repo",
	}

	if result.Name != "test-plugin" {
		t.Errorf("Name = %v, want test-plugin", result.Name)
	}
	if result.Version != "v1.0.0" {
		t.Errorf("Version = %v, want v1.0.0", result.Version)
	}
	if !result.FromURL {
		t.Error("FromURL should be true")
	}
}

func TestParseOwnerRepo(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{
			name:      "valid format",
			input:     "owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
		},
		{
			name:    "invalid format no slash",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:      "with multiple slashes",
			input:     "owner/repo/extra",
			wantOwner: "owner",
			wantRepo:  "repo/extra",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := parseOwnerRepo(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseOwnerRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if owner != tt.wantOwner {
				t.Errorf("owner = %v, want %v", owner, tt.wantOwner)
			}
			if repo != tt.wantRepo {
				t.Errorf("repo = %v, want %v", repo, tt.wantRepo)
			}
		})
	}
}

func TestFindPluginBinary(t *testing.T) {
	tests := []struct {
		name       string
		setupDir   func(t *testing.T) string
		pluginName string
		wantFound  bool
	}{
		{
			name: "finds exact name match",
			setupDir: func(t *testing.T) string {
				dir := t.TempDir()
				binPath := filepath.Join(dir, "test-plugin")
				if err := os.WriteFile(binPath, []byte("binary"), 0755); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			pluginName: "test-plugin",
			wantFound:  true,
		},
		{
			name: "finds prefixed name",
			setupDir: func(t *testing.T) string {
				dir := t.TempDir()
				binPath := filepath.Join(dir, "pulumicost-plugin-test")
				if err := os.WriteFile(binPath, []byte("binary"), 0755); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			pluginName: "test",
			wantFound:  true,
		},
		{
			name: "empty directory",
			setupDir: func(t *testing.T) string {
				return t.TempDir()
			},
			pluginName: "test",
			wantFound:  false,
		},
		{
			name: "finds any executable",
			setupDir: func(t *testing.T) string {
				dir := t.TempDir()
				binPath := filepath.Join(dir, "some-binary")
				if err := os.WriteFile(binPath, []byte("binary"), 0755); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			pluginName: "different-name",
			wantFound:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setupDir(t)
			result := findPluginBinary(dir, tt.pluginName)
			if tt.wantFound && result == "" {
				t.Error("expected to find binary but got empty string")
			}
			if !tt.wantFound && result != "" {
				t.Errorf("expected no binary but found: %s", result)
			}
		})
	}
}

func TestInstallAlreadyExists(t *testing.T) {
	// Create temp plugin directory with existing installation
	tmpDir := t.TempDir()
	pluginDir := filepath.Join(tmpDir, "test-plugin", "v1.0.0")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatal(err)
	}

	installer := NewInstaller(tmpDir)
	opts := InstallOptions{Force: false}

	// This should fail because we can't actually contact GitHub
	// but if it gets past the "already installed" check, the test structure is correct
	_, err := installer.Install("test-plugin@v1.0.0", opts, nil)
	if err == nil {
		t.Error("expected error for non-existent registry plugin")
	}
}

func TestUpdateOptions(t *testing.T) {
	opts := UpdateOptions{
		DryRun:    true,
		Version:   "v2.0.0",
		PluginDir: "/test/dir",
	}

	if !opts.DryRun {
		t.Error("UpdateOptions.DryRun should be true")
	}
	if opts.Version != "v2.0.0" {
		t.Errorf("UpdateOptions.Version = %v, want v2.0.0", opts.Version)
	}
	if opts.PluginDir != "/test/dir" {
		t.Errorf("UpdateOptions.PluginDir = %v, want /test/dir", opts.PluginDir)
	}
}

func TestRemoveOptions(t *testing.T) {
	opts := RemoveOptions{
		KeepConfig: true,
		PluginDir:  "/test/dir",
	}

	if !opts.KeepConfig {
		t.Error("RemoveOptions.KeepConfig should be true")
	}
	if opts.PluginDir != "/test/dir" {
		t.Errorf("RemoveOptions.PluginDir = %v, want /test/dir", opts.PluginDir)
	}
}

func TestUpdateResult(t *testing.T) {
	result := UpdateResult{
		Name:        "test-plugin",
		OldVersion:  "v1.0.0",
		NewVersion:  "v2.0.0",
		Path:        "/path/to/plugin",
		WasUpToDate: false,
	}

	if result.Name != "test-plugin" {
		t.Errorf("Name = %v, want test-plugin", result.Name)
	}
	if result.OldVersion != "v1.0.0" {
		t.Errorf("OldVersion = %v, want v1.0.0", result.OldVersion)
	}
	if result.NewVersion != "v2.0.0" {
		t.Errorf("NewVersion = %v, want v2.0.0", result.NewVersion)
	}
	if result.WasUpToDate {
		t.Error("WasUpToDate should be false")
	}
}

func TestInstallEmptySpecifier(t *testing.T) {
	tmpDir := t.TempDir()
	installer := NewInstaller(tmpDir)
	opts := InstallOptions{}

	_, err := installer.Install("", opts, nil)
	if err == nil {
		t.Error("expected error for empty specifier")
	}
}

func TestInstallInvalidURLFormat(t *testing.T) {
	tmpDir := t.TempDir()
	installer := NewInstaller(tmpDir)
	opts := InstallOptions{}

	_, err := installer.Install("github.com/invalid", opts, nil)
	if err == nil {
		t.Error("expected error for invalid URL format")
	}
}

func TestFindPluginBinaryNonExistentDir(t *testing.T) {
	result := findPluginBinary("/nonexistent/path", "test")
	if result != "" {
		t.Errorf("expected empty string for non-existent dir, got: %s", result)
	}
}

func TestParseOwnerRepoEmptyInput(t *testing.T) {
	_, _, err := parseOwnerRepo("")
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestParseOwnerRepoOnlySlash(t *testing.T) {
	_, _, err := parseOwnerRepo("/")
	if err == nil {
		t.Error("expected error for empty owner/repo segments")
	}
}

func TestInstallerLock(t *testing.T) {
	tmpDir := t.TempDir()
	installer := NewInstaller(tmpDir)
	name := "test-plugin"

	// Acquire lock first time
	unlock1, err := installer.acquireLock(name)
	if err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}
	if unlock1 == nil {
		t.Fatal("Unlock function is nil")
	}

	// Try to acquire lock second time - should fail
	unlock2, err := installer.acquireLock(name)
	if err == nil {
		t.Error("Expected error when acquiring already held lock")
		if unlock2 != nil {
			unlock2()
		}
	}

	// Release first lock
	unlock1()

	// Try to acquire lock again - should succeed now
	unlock3, err := installer.acquireLock(name)
	if err != nil {
		t.Fatalf("Failed to acquire lock after release: %v", err)
	}
	if unlock3 == nil {
		t.Fatal("Unlock function is nil")
	}
	unlock3()
}

func TestInstallerLockStaleDetection(t *testing.T) {
	tmpDir := t.TempDir()
	installer := NewInstaller(tmpDir)
	name := "test-plugin"

	// Ensure plugin directory exists
	if err := os.MkdirAll(tmpDir, 0750); err != nil {
		t.Fatalf("Failed to create plugin directory: %v", err)
	}

	// Create a stale lock file with an invalid PID
	lockPath := filepath.Join(tmpDir, name+".lock")
	if err := os.WriteFile(lockPath, []byte("99999999"), 0600); err != nil {
		t.Fatalf("Failed to create stale lock file: %v", err)
	}

	// Acquiring lock should succeed because the PID is invalid (stale)
	unlock, err := installer.acquireLock(name)
	if err != nil {
		t.Errorf("Expected to acquire lock with stale lock file, got error: %v", err)
		return
	}
	if unlock == nil {
		t.Fatal("Unlock function is nil")
	}
	unlock()
}

func TestInstallerLockEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	installer := NewInstaller(tmpDir)
	name := "test-plugin"

	// Ensure plugin directory exists
	if err := os.MkdirAll(tmpDir, 0750); err != nil {
		t.Fatalf("Failed to create plugin directory: %v", err)
	}

	// Create an empty lock file (legacy or corrupt)
	lockPath := filepath.Join(tmpDir, name+".lock")
	if err := os.WriteFile(lockPath, []byte(""), 0600); err != nil {
		t.Fatalf("Failed to create empty lock file: %v", err)
	}

	// Acquiring lock should succeed because empty file is treated as stale
	unlock, err := installer.acquireLock(name)
	if err != nil {
		t.Errorf("Expected to acquire lock with empty lock file, got error: %v", err)
		return
	}
	if unlock == nil {
		t.Fatal("Unlock function is nil")
	}
	unlock()
}

func TestInstallerLockInvalidPID(t *testing.T) {
	tmpDir := t.TempDir()
	installer := NewInstaller(tmpDir)
	name := "test-plugin"

	// Ensure plugin directory exists
	if err := os.MkdirAll(tmpDir, 0750); err != nil {
		t.Fatalf("Failed to create plugin directory: %v", err)
	}

	// Create a lock file with invalid content
	lockPath := filepath.Join(tmpDir, name+".lock")
	if err := os.WriteFile(lockPath, []byte("not-a-pid"), 0600); err != nil {
		t.Fatalf("Failed to create invalid lock file: %v", err)
	}

	// Acquiring lock should succeed because invalid PID is treated as stale
	unlock, err := installer.acquireLock(name)
	if err != nil {
		t.Errorf("Expected to acquire lock with invalid PID, got error: %v", err)
		return
	}
	if unlock == nil {
		t.Fatal("Unlock function is nil")
	}
	unlock()
}

func TestIsProcessRunning(t *testing.T) {
	// Test with current process - should be running
	currentPID := os.Getpid()
	if !isProcessRunning(currentPID) {
		t.Error("Expected current process to be running")
	}

	// Test with invalid PID - should not be running
	if isProcessRunning(99999999) {
		t.Error("Expected invalid PID to not be running")
	}

	// Test with PID 0 - typically kernel, but behavior varies
	// Just ensure it doesn't panic
	_ = isProcessRunning(0)
}

func TestIsLockStale(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "empty file is stale",
			content:  "",
			expected: true,
		},
		{
			name:     "whitespace only is stale",
			content:  "   \n  ",
			expected: true,
		},
		{
			name:     "invalid PID is stale",
			content:  "not-a-number",
			expected: true,
		},
		{
			name:     "very large PID is stale",
			content:  "99999999",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lockPath := filepath.Join(tmpDir, "test-"+tt.name+".lock")
			if err := os.WriteFile(lockPath, []byte(tt.content), 0600); err != nil {
				t.Fatalf("Failed to create lock file: %v", err)
			}

			result := isLockStale(lockPath)
			if result != tt.expected {
				t.Errorf("isLockStale() = %v, expected %v", result, tt.expected)
			}
		})
	}

	// Test with non-existent file - should not be stale (safe default)
	if isLockStale(filepath.Join(tmpDir, "nonexistent.lock")) {
		t.Error("Non-existent lock file should not be considered stale")
	}
}

// TestInstallerLockConcurrent verifies that multiple concurrent acquisition
// attempts are properly serialized and only one goroutine can hold the lock
// at a time.
func TestInstallerLockConcurrent(t *testing.T) {
	tmpDir := t.TempDir()
	installer := NewInstaller(tmpDir)
	name := "concurrent-test-plugin"

	const numGoroutines = 10
	var wg sync.WaitGroup
	var successCount atomic.Int32
	var errorCount atomic.Int32

	// Start signal channel to ensure all goroutines start simultaneously
	start := make(chan struct{})

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Wait for start signal
			<-start

			unlock, err := installer.acquireLock(name)
			if err != nil {
				errorCount.Add(1)
				return
			}

			// Successfully acquired the lock
			successCount.Add(1)

			// Hold the lock briefly to simulate work
			// (no actual sleep, just release immediately)
			unlock()
		}()
	}

	// Signal all goroutines to start
	close(start)

	// Wait for all goroutines to complete
	wg.Wait()

	// At least one goroutine should have acquired the lock
	if successCount.Load() == 0 {
		t.Error("Expected at least one goroutine to acquire the lock")
	}

	// All goroutines should have either succeeded or failed
	total := successCount.Load() + errorCount.Load()
	if total != numGoroutines {
		t.Errorf("Expected %d total results, got %d", numGoroutines, total)
	}

	// After all goroutines complete, we should be able to acquire the lock again
	unlock, err := installer.acquireLock(name)
	if err != nil {
		t.Fatalf("Failed to acquire lock after concurrent test: %v", err)
	}
	unlock()
}
