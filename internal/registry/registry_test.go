package registry // needs access to unexported methods

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/rshade/pulumicost-core/internal/pluginhost"
)

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
