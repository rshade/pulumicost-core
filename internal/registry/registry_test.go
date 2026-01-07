package registry // needs access to unexported methods

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/rshade/pulumicost-core/internal/pluginhost"
	"google.golang.org/grpc"
)

func TestListLatestPlugins(t *testing.T) {
	tests := []struct {
		name              string
		setupDir          func(t *testing.T) string
		wantCount         int
		wantPlugins       []PluginInfo
		wantWarningsCount int
		wantErr           bool
	}{
		{
			name: "empty directory",
			setupDir: func(t *testing.T) string {
				return t.TempDir()
			},
			wantCount: 0,
		},
		{
			name: "single plugin single version",
			setupDir: func(t *testing.T) string {
				return createSinglePluginDir(t, "testplugin", "v1.0.0")
			},
			wantCount: 1,
			wantPlugins: []PluginInfo{
				{Name: "testplugin", Version: "v1.0.0"},
			},
		},
		{
			name:      "multiple plugins single version",
			setupDir:  createMultiplePluginsDir,
			wantCount: 2,
			wantPlugins: []PluginInfo{
				{Name: "aws-plugin", Version: "v1.0.0"},
				{Name: "kubecost", Version: "v2.1.0"},
			},
		},
		{
			name:      "same plugin multiple versions selects latest",
			setupDir:  createMultiVersionPluginDir,
			wantCount: 1,
			wantPlugins: []PluginInfo{
				{Name: "testplugin", Version: "v2.0.0"},
			},
		},
		{
			name: "three versions selects highest",
			setupDir: func(t *testing.T) string {
				dir := t.TempDir()
				versions := []string{"v1.0.0", "v1.1.0", "v2.0.0"}
				for _, v := range versions {
					if err := createPluginVersion(dir, "tri-ver", v); err != nil {
						t.Fatal(err)
					}
				}
				return dir
			},
			wantCount: 1,
			wantPlugins: []PluginInfo{
				{Name: "tri-ver", Version: "v2.0.0"},
			},
		},
		{
			name:              "edge cases: pre-release, invalid, corrupt",
			setupDir:          createEdgeCasePluginDir,
			wantCount:         2,
			wantWarningsCount: 1, // Invalid version triggers warning
			wantPlugins: []PluginInfo{
				{Name: "prerelease-test", Version: "v1.0.0"},  // Stable over pre-release
				{Name: "invalid-ver-test", Version: "v1.0.0"}, // Valid one picked
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootDir := tt.setupDir(t)
			reg := &Registry{
				root:     rootDir,
				launcher: pluginhost.NewProcessLauncher(),
			}

			plugins, warnings, err := reg.ListLatestPlugins()
			verifyListLatestPluginsResult(
				t,
				plugins,
				warnings,
				err,
				tt.wantErr,
				tt.wantCount,
				tt.wantPlugins,
				tt.wantWarningsCount,
			)
		})
	}
}

func TestListLatestPlugins_FSErrors(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission tests on Windows")
	}

	dir := t.TempDir()
	// Create a directory with no permissions
	noPermDir := filepath.Join(dir, "no-perm-plugin")
	if err := os.MkdirAll(noPermDir, 0000); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chmod(noPermDir, 0755) // Cleanup
	}()

	reg := &Registry{
		root:     dir,
		launcher: pluginhost.NewProcessLauncher(),
	}

	plugins, _, err := reg.ListLatestPlugins()
	if err != nil {
		// It might fail or just return empty depending on implementation
		// Ideally it should not crash
		t.Logf(
			"ListLatestPlugins returned error (expected behavior for root read failure): %v",
			err,
		)
	}

	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins from unreadable dir, got %d", len(plugins))
	}
}

func TestListLatestPlugins_BinaryValidation(t *testing.T) {
	dir := t.TempDir()

	// 1. Missing Binary
	if err := os.MkdirAll(filepath.Join(dir, "missing-bin", "v1.0.0"), 0755); err != nil {
		t.Fatal(err)
	}

	// 2. Non-executable Binary (Linux/Mac only)
	if runtime.GOOS != "windows" {
		nonExecDir := filepath.Join(dir, "non-exec", "v1.0.0")
		if err := os.MkdirAll(nonExecDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(nonExecDir, "non-exec"), []byte("data"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	reg := &Registry{
		root:     dir,
		launcher: pluginhost.NewProcessLauncher(),
	}

	plugins, _, err := reg.ListLatestPlugins()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins (invalid binaries), got %d: %v", len(plugins), plugins)
	}
}

func TestListLatestPlugins_Concurrency(t *testing.T) {
	dir := createMultiVersionPluginDir(t)
	reg := &Registry{
		root:     dir,
		launcher: pluginhost.NewProcessLauncher(),
	}

	concurrency := 10
	errCh := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			_, _, err := reg.ListLatestPlugins()
			errCh <- err
		}()
	}

	for i := 0; i < concurrency; i++ {
		if err := <-errCh; err != nil {
			t.Errorf("concurrent ListLatestPlugins failed: %v", err)
		}
	}
}

type mockLauncher struct {
	startCalled map[string]int
}

func (m *mockLauncher) Start(
	ctx context.Context,
	path string,
	args ...string,
) (*grpc.ClientConn, func() error, error) {
	if m.startCalled == nil {
		m.startCalled = make(map[string]int)
	}
	m.startCalled[filepath.Base(path)]++
	return nil, func() error { return nil }, errors.New("mock launch failed")
}

func TestRegistry_Open_WithWarnings(t *testing.T) {
	// Create directory with valid and invalid plugins
	dir := createEdgeCasePluginDir(t)
	mock := &mockLauncher{}
	reg := &Registry{
		root:     dir,
		launcher: mock,
	}

	ctx := context.Background()
	// Open should proceed despite warnings about invalid versions
	// It will try to launch valid plugins. Mock launcher returns error,
	// so Open will log warnings and return (clients=empty, err=nil).
	clients, cleanup, err := reg.Open(ctx, "")
	if cleanup != nil {
		defer cleanup()
	}

	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	if len(clients) != 0 {
		t.Errorf("Expected 0 clients (mock fails), got %d", len(clients))
	}

	// Verify it attempted to launch the valid plugins
	// Expected valid plugins: "prerelease-test", "invalid-ver-test" (names from createEdgeCasePluginDir)
	// The binary names are created by createPluginVersion as the plugin Name.
	expectedAttempts := []string{"prerelease-test", "invalid-ver-test"}
	if len(mock.startCalled) != len(expectedAttempts) {
		t.Errorf(
			"Expected %d launch attempts, got %d",
			len(expectedAttempts),
			len(mock.startCalled),
		)
	}

	for _, name := range expectedAttempts {
		// On windows it adds .exe
		key := name
		if runtime.GOOS == "windows" {
			key += ".exe"
		}
		if mock.startCalled[key] != 1 {
			t.Errorf("Expected 1 launch attempt for %s, got %d", key, mock.startCalled[key])
		}
	}
}

func TestListPlugins(t *testing.T) {
	tests := []struct {
		name        string
		setupDir    func(t *testing.T) string
		wantCount   int
		wantPlugins []PluginInfo
		wantErr     bool
	}{
		{
			name: "empty directory",
			setupDir: func(t *testing.T) string {
				return t.TempDir()
			},
			wantCount: 0,
		},
		{
			name: "no plugin directory",
			setupDir: func(t *testing.T) string {
				dir := t.TempDir()
				return filepath.Join(dir, "nonexistent")
			},
			wantCount: 0,
		},
		{
			name: "single plugin with version",
			setupDir: func(t *testing.T) string {
				return createSinglePluginDir(t, "testplugin", "v1.0.0")
			},
			wantCount: 1,
			wantPlugins: []PluginInfo{
				{Name: "testplugin", Version: "v1.0.0"},
			},
		},
		{
			name:      "multiple plugins with versions",
			setupDir:  createMultiplePluginsDir,
			wantCount: 2,
			wantPlugins: []PluginInfo{
				{Name: "aws-plugin", Version: "v1.0.0"},
				{Name: "kubecost", Version: "v2.1.0"},
			},
		},
		{
			name:      "multiple versions same plugin",
			setupDir:  createMultiVersionPluginDir,
			wantCount: 2,
			wantPlugins: []PluginInfo{
				{Name: "testplugin", Version: "v1.0.0"},
				{Name: "testplugin", Version: "v2.0.0"},
			},
		},
		{
			name:      "non-executable file ignored",
			setupDir:  createNonExecutablePluginDir,
			wantCount: 0,
		},
		{
			name: "files in root ignored",
			setupDir: func(t *testing.T) string {
				dir := t.TempDir()
				if err := os.WriteFile(filepath.Join(dir, "somefile"), []byte("test"), 0755); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootDir := tt.setupDir(t)
			reg := &Registry{
				root:     rootDir,
				launcher: pluginhost.NewProcessLauncher(),
			}

			plugins, err := reg.ListPlugins()
			verifyListPluginsResult(t, plugins, err, tt.wantErr, tt.wantCount, tt.wantPlugins)
		})
	}
}

func TestFindBinary(t *testing.T) {
	tests := []struct {
		name       string
		setupDir   func(t *testing.T) string
		wantExists bool
	}{
		{"executable file found", createExecutableBinaryDir, true},
		{"non-executable file ignored", createNonExecutableBinaryDir, false},
		{"directory ignored", createDirectoryInsteadOfBinary, false},
		{"empty directory", func(t *testing.T) string { return t.TempDir() }, false},
		{"multiple files, finds executable", createMixedFilesDir, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setupDir(t)
			reg := &Registry{root: "", launcher: pluginhost.NewProcessLauncher()}
			binPath := reg.findBinary(dir)
			verifyBinaryResult(t, binPath, tt.wantExists)
		})
	}
}

func TestRegistry_Open(t *testing.T) {
	tests := []struct {
		name        string
		setupDir    func(t *testing.T) string
		onlyName    string
		wantClients int
		wantErr     bool
	}{
		{"no plugins", func(t *testing.T) string { return t.TempDir() }, "", 0, false},
		{"plugin directory doesn't exist", createNonexistentPluginDir, "", 0, false},
		{"filter by specific plugin name", createMultiplePluginsDir, "aws-plugin", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootDir := tt.setupDir(t)
			reg := &Registry{root: rootDir, launcher: pluginhost.NewProcessLauncher()}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			clients, cleanup, err := reg.Open(ctx, tt.onlyName)
			verifyRegistryOpenResult(t, clients, cleanup, err, tt.wantErr, tt.wantClients)
		})
	}
}

// Helper functions to reduce cognitive complexity

func createSinglePluginDir(t *testing.T, name, version string) string {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, name, version)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatal(err)
	}
	binPath := filepath.Join(pluginDir, name)
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}
	if err := os.WriteFile(binPath, []byte("#!/bin/bash\necho test"), 0755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func createMultiplePluginsDir(t *testing.T) string {
	dir := t.TempDir()
	plugins := []struct{ name, version string }{
		{"aws-plugin", "v1.0.0"},
		{"kubecost", "v2.1.0"},
	}
	for _, p := range plugins {
		pluginDir := filepath.Join(dir, p.name, p.version)
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			t.Fatal(err)
		}
		binPath := filepath.Join(pluginDir, p.name)
		if runtime.GOOS == "windows" {
			binPath += ".exe"
		}
		if err := os.WriteFile(binPath, []byte("#!/bin/bash\necho "+p.name+"\nexit 1"), 0755); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func createMultiVersionPluginDir(t *testing.T) string {
	dir := t.TempDir()
	versions := []string{"v1.0.0", "v2.0.0"}
	for _, version := range versions {
		pluginDir := filepath.Join(dir, "testplugin", version)
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			t.Fatal(err)
		}
		binPath := filepath.Join(pluginDir, "testplugin")
		if runtime.GOOS == "windows" {
			binPath += ".exe"
		}
		if err := os.WriteFile(binPath, []byte("#!/bin/bash\necho "+version), 0755); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func createNonExecutablePluginDir(t *testing.T) string {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "testplugin", "v1.0.0")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatal(err)
	}
	binPath := filepath.Join(pluginDir, "testplugin")
	if err := os.WriteFile(binPath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func verifyListPluginsResult(
	t *testing.T,
	plugins []PluginInfo,
	err error,
	wantErr bool,
	wantCount int,
	wantPlugins []PluginInfo,
) {
	if wantErr && err == nil {
		t.Error("expected error but got none")
		return
	}
	if !wantErr && err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if len(plugins) != wantCount {
		t.Errorf("expected %d plugins, got %d", wantCount, len(plugins))
		return
	}
	if wantPlugins != nil {
		verifyExpectedPlugins(t, plugins, wantPlugins)
	}
}

func verifyExpectedPlugins(t *testing.T, plugins []PluginInfo, wantPlugins []PluginInfo) {
	for _, wantPlugin := range wantPlugins {
		found := false
		for _, plugin := range plugins {
			if plugin.Name == wantPlugin.Name && plugin.Version == wantPlugin.Version {
				found = true
				if plugin.Path == "" {
					t.Errorf("plugin %s v%s has empty path", plugin.Name, plugin.Version)
				}
				break
			}
		}
		if !found {
			t.Errorf("expected plugin %s v%s not found", wantPlugin.Name, wantPlugin.Version)
		}
	}
}

func TestNewDefault(t *testing.T) {
	reg := NewDefault()

	if reg == nil {
		t.Fatal("NewDefault returned nil")
	}

	if reg.root == "" {
		t.Error("registry root is empty")
	}

	if reg.launcher == nil {
		t.Error("registry launcher is nil")
	}

	// Verify the root path structure
	expectedSuffix := filepath.Join(".pulumicost", "plugins")
	if !filepath.IsAbs(reg.root) {
		t.Error("registry root is not absolute path")
	}

	if !strings.HasSuffix(reg.root, expectedSuffix) {
		t.Errorf("registry root doesn't end with expected path: %s", expectedSuffix)
	}
}

// Additional helper functions for TestFindBinary and TestRegistry_Open

func createExecutableBinaryDir(t *testing.T) string {
	dir := t.TempDir()
	binPath := filepath.Join(dir, "plugin")
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}
	if err := os.WriteFile(binPath, []byte("#!/bin/bash\necho test"), 0755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func createNonExecutableBinaryDir(t *testing.T) string {
	dir := t.TempDir()
	binPath := filepath.Join(dir, "plugin")
	if err := os.WriteFile(binPath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func createDirectoryInsteadOfBinary(t *testing.T) string {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "plugin")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func createMixedFilesDir(t *testing.T) string {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("readme"), 0644); err != nil {
		t.Fatal(err)
	}
	binPath := filepath.Join(dir, "plugin")
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}
	if err := os.WriteFile(binPath, []byte("#!/bin/bash\necho test"), 0755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func createNonexistentPluginDir(t *testing.T) string {
	dir := t.TempDir()
	return filepath.Join(dir, "nonexistent")
}

func verifyBinaryResult(t *testing.T, binPath string, wantExists bool) {
	if wantExists && binPath == "" {
		t.Error("expected to find binary but got empty path")
	}
	if !wantExists && binPath != "" {
		t.Errorf("expected no binary but got path: %s", binPath)
	}
	if binPath != "" {
		verifyBinaryExecutable(t, binPath)
	}
}

func verifyBinaryExecutable(t *testing.T, binPath string) {
	info, err := os.Stat(binPath)
	if err != nil {
		t.Errorf("binary path invalid: %v", err)
		return
	}
	if runtime.GOOS != "windows" {
		if info.Mode()&0111 == 0 {
			t.Error("found binary is not executable")
		}
	} else {
		if filepath.Ext(binPath) != ".exe" {
			t.Error("found binary on Windows doesn't have .exe extension")
		}
	}
}

func createEdgeCasePluginDir(t *testing.T) string {
	dir := t.TempDir()

	// 1. Pre-release vs Stable: v1.0.0-alpha vs v1.0.0
	// Should select v1.0.0
	if err := createPluginVersion(dir, "prerelease-test", "v1.0.0-alpha"); err != nil {
		t.Fatal(err)
	}
	if err := createPluginVersion(dir, "prerelease-test", "v1.0.0"); err != nil {
		t.Fatal(err)
	}

	// 2. Invalid Version: invalid-version
	// Should be skipped/warned
	if err := createPluginVersion(dir, "invalid-ver-test", "invalid-version"); err != nil {
		t.Fatal(err)
	}
	// Add a valid one to ensure it's picked up
	if err := createPluginVersion(dir, "invalid-ver-test", "v1.0.0"); err != nil {
		t.Fatal(err)
	}

	// 3. Corrupted Directory (file instead of dir)
	// Should be skipped/warned
	corruptPath := filepath.Join(dir, "corrupt-plugin")
	if err := os.WriteFile(corruptPath, []byte("file-not-dir"), 0644); err != nil {
		t.Fatal(err)
	}

	return dir
}

func createPluginVersion(rootDir, name, version string) error {
	pluginDir := filepath.Join(rootDir, name, version)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return err
	}
	binPath := filepath.Join(pluginDir, name)
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}
	return os.WriteFile(binPath, []byte("#!/bin/bash\necho "+version), 0755)
}

func verifyListLatestPluginsResult(
	t *testing.T,
	plugins []PluginInfo,
	warnings []string,
	err error,
	wantErr bool,
	wantCount int,
	wantPlugins []PluginInfo,
	wantWarningsCount int,
) {
	if wantErr && err == nil {
		t.Error("expected error but got none")
		return
	}
	if !wantErr && err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if len(plugins) != wantCount {
		t.Errorf("expected %d plugins, got %d", wantCount, len(plugins))
	}
	if len(warnings) != wantWarningsCount {
		t.Errorf(
			"expected %d warnings, got %d. Warnings: %v",
			wantWarningsCount,
			len(warnings),
			warnings,
		)
	}
	if wantPlugins != nil {
		verifyExpectedPlugins(t, plugins, wantPlugins)
	}
}

func verifyRegistryOpenResult(
	t *testing.T,
	clients []*pluginhost.Client,
	cleanup func(),
	err error,
	wantErr bool,
	wantClients int,
) {
	if cleanup != nil {
		defer cleanup()
	}
	if wantErr && err == nil {
		t.Error("expected error but got none")
		return
	}
	if !wantErr && err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if len(clients) != wantClients {
		t.Errorf("expected %d clients, got %d", wantClients, len(clients))
	}
	if cleanup != nil {
		cleanup() // Should not panic
	}
}
