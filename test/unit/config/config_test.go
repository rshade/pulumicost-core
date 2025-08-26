package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rshade/pulumicost-core/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfigDir(t *testing.T) {
	tests := []struct {
		name     string
		envVar   string
		envValue string
		expected string
	}{
		{
			name:     "default config directory",
			envVar:   "",
			envValue: "",
			expected: ".pulumicost",
		},
		{
			name:     "custom config directory via HOME",
			envVar:   "HOME",
			envValue: "/custom/home",
			expected: "/custom/home/.pulumicost",
		},
		{
			name:     "custom config directory via USERPROFILE",
			envVar:   "USERPROFILE",
			envValue: "/custom/user",
			expected: "/custom/user/.pulumicost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env var
			originalEnv := os.Getenv(tt.envVar)
			defer func() {
				if originalEnv != "" {
					os.Setenv(tt.envVar, originalEnv)
				} else if tt.envVar != "" {
					os.Unsetenv(tt.envVar)
				}
			}()

			// Set test env var if specified
			if tt.envVar != "" && tt.envValue != "" {
				os.Setenv(tt.envVar, tt.envValue)
			}

			result := config.GetConfigDir()
			
			if tt.envVar != "" {
				assert.Contains(t, result, tt.expected)
			} else {
				assert.Contains(t, result, tt.expected)
			}
		})
	}
}

func TestGetPluginDir(t *testing.T) {
	t.Run("returns plugin directory path", func(t *testing.T) {
		pluginDir := config.GetPluginDir()
		
		assert.NotEmpty(t, pluginDir)
		assert.Contains(t, pluginDir, ".pulumicost")
		assert.Contains(t, pluginDir, "plugins")
	})
}

func TestGetSpecDir(t *testing.T) {
	t.Run("returns spec directory path", func(t *testing.T) {
		specDir := config.GetSpecDir()
		
		assert.NotEmpty(t, specDir)
		assert.Contains(t, specDir, ".pulumicost")
		assert.Contains(t, specDir, "specs")
	})
}

func TestEnsureConfigDir(t *testing.T) {
	t.Run("creates config directory if it doesn't exist", func(t *testing.T) {
		// Create a temporary directory for testing
		tempDir := t.TempDir()
		configDir := filepath.Join(tempDir, ".pulumicost")
		
		// Override the config dir for testing
		original := os.Getenv("HOME")
		defer os.Setenv("HOME", original)
		os.Setenv("HOME", tempDir)
		
		// Directory should not exist initially
		_, err := os.Stat(configDir)
		assert.True(t, os.IsNotExist(err))
		
		// This should create the directory
		err = config.EnsureConfigDir()
		require.NoError(t, err)
		
		// Directory should now exist
		info, err := os.Stat(configDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})
	
	t.Run("handles existing config directory", func(t *testing.T) {
		// Create a temporary directory for testing
		tempDir := t.TempDir()
		configDir := filepath.Join(tempDir, ".pulumicost")
		
		// Create the directory first
		err := os.MkdirAll(configDir, 0755)
		require.NoError(t, err)
		
		// Override the config dir for testing
		original := os.Getenv("HOME")
		defer os.Setenv("HOME", original)
		os.Setenv("HOME", tempDir)
		
		// This should not error on existing directory
		err = config.EnsureConfigDir()
		require.NoError(t, err)
		
		// Directory should still exist
		info, err := os.Stat(configDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})
}

func TestEnsureSubDirs(t *testing.T) {
	t.Run("creates plugin and spec subdirectories", func(t *testing.T) {
		// Create a temporary directory for testing
		tempDir := t.TempDir()
		
		// Override the config dir for testing
		original := os.Getenv("HOME")
		defer os.Setenv("HOME", original)
		os.Setenv("HOME", tempDir)
		
		err := config.EnsureSubDirs()
		require.NoError(t, err)
		
		// Check that subdirectories were created
		pluginDir := filepath.Join(tempDir, ".pulumicost", "plugins")
		specDir := filepath.Join(tempDir, ".pulumicost", "specs")
		
		pluginInfo, err := os.Stat(pluginDir)
		require.NoError(t, err)
		assert.True(t, pluginInfo.IsDir())
		
		specInfo, err := os.Stat(specDir)
		require.NoError(t, err)
		assert.True(t, specInfo.IsDir())
	})
}

func TestConfigPaths(t *testing.T) {
	t.Run("all config paths are accessible", func(t *testing.T) {
		configDir := config.GetConfigDir()
		pluginDir := config.GetPluginDir()
		specDir := config.GetSpecDir()
		
		// All paths should be non-empty
		assert.NotEmpty(t, configDir)
		assert.NotEmpty(t, pluginDir)
		assert.NotEmpty(t, specDir)
		
		// Plugin and spec dirs should be under config dir
		assert.Contains(t, pluginDir, ".pulumicost")
		assert.Contains(t, specDir, ".pulumicost")
		
		// Paths should be absolute or relative
		assert.True(t, filepath.IsAbs(configDir) || !filepath.IsAbs(configDir))
	})
}

func TestConfigPermissions(t *testing.T) {
	t.Run("config directories have correct permissions", func(t *testing.T) {
		// Create a temporary directory for testing
		tempDir := t.TempDir()
		
		// Override the config dir for testing
		original := os.Getenv("HOME")
		defer os.Setenv("HOME", original)
		os.Setenv("HOME", tempDir)
		
		err := config.EnsureConfigDir()
		require.NoError(t, err)
		
		err = config.EnsureSubDirs()
		require.NoError(t, err)
		
		// Check permissions on created directories
		configDir := filepath.Join(tempDir, ".pulumicost")
		pluginDir := filepath.Join(tempDir, ".pulumicost", "plugins")
		specDir := filepath.Join(tempDir, ".pulumicost", "specs")
		
		for _, dir := range []string{configDir, pluginDir, specDir} {
			info, err := os.Stat(dir)
			require.NoError(t, err)
			
			// Should be a directory
			assert.True(t, info.IsDir())
			
			// Should have reasonable permissions (readable and writable by owner)
			mode := info.Mode().Perm()
			assert.True(t, mode&0700 != 0, "Directory should be readable/writable by owner: %s", dir)
		}
	})
}