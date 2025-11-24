package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// InstalledPlugin represents an installed plugin entry in config.yaml.
type InstalledPlugin struct {
	Name    string `yaml:"name"    json:"name"`
	URL     string `yaml:"url"     json:"url"`
	Version string `yaml:"version" json:"version"`
}

// InstalledPluginsConfig holds the installed plugins list.
type InstalledPluginsConfig struct {
	InstalledPlugins []InstalledPlugin `yaml:"installed_plugins" json:"installed_plugins"`
}

// pluginsConfigPath returns the path to the config file.
func pluginsConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".pulumicost", "config.yaml"), nil
}

// LoadInstalledPlugins loads the list of installed plugins from config.
func LoadInstalledPlugins() ([]InstalledPlugin, error) {
	configPath, err := pluginsConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []InstalledPlugin{}, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg InstalledPluginsConfig
	if unmarshalErr := yaml.Unmarshal(data, &cfg); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to parse config: %w", unmarshalErr)
	}

	return cfg.InstalledPlugins, nil
}

// SaveInstalledPlugins saves the list of installed plugins to config.
func SaveInstalledPlugins(plugins []InstalledPlugin) error {
	configPath, err := pluginsConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Load existing config to preserve other settings
	var fullConfig map[string]interface{}
	if existingData, readErr := os.ReadFile(configPath); readErr == nil {
		if unmarshalErr := yaml.Unmarshal(existingData, &fullConfig); unmarshalErr != nil {
			fullConfig = make(map[string]interface{})
		}
	} else {
		fullConfig = make(map[string]interface{})
	}

	// Update installed_plugins
	fullConfig["installed_plugins"] = plugins

	// Marshal and save
	data, err := yaml.Marshal(fullConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to temp file first, then rename (atomic write)
	tmpPath := configPath + ".tmp"
	if writeErr := os.WriteFile(tmpPath, data, 0600); writeErr != nil {
		return fmt.Errorf("failed to write config: %w", writeErr)
	}

	if renameErr := os.Rename(tmpPath, configPath); renameErr != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to save config: %w", renameErr)
	}

	return nil
}

// AddInstalledPlugin adds a plugin to the installed plugins list.
func AddInstalledPlugin(plugin InstalledPlugin) error {
	plugins, err := LoadInstalledPlugins()
	if err != nil {
		return err
	}

	// Check if already exists and update
	found := false
	for i, p := range plugins {
		if p.Name == plugin.Name {
			plugins[i] = plugin
			found = true
			break
		}
	}

	if !found {
		plugins = append(plugins, plugin)
	}

	return SaveInstalledPlugins(plugins)
}

// RemoveInstalledPlugin removes a plugin from the installed plugins list.
func RemoveInstalledPlugin(name string) error {
	plugins, err := LoadInstalledPlugins()
	if err != nil {
		return err
	}

	// Find and remove
	var newPlugins []InstalledPlugin
	for _, p := range plugins {
		if p.Name != name {
			newPlugins = append(newPlugins, p)
		}
	}

	return SaveInstalledPlugins(newPlugins)
}

// GetInstalledPlugin returns a specific installed plugin by name.
func GetInstalledPlugin(name string) (*InstalledPlugin, error) {
	plugins, err := LoadInstalledPlugins()
	if err != nil {
		return nil, err
	}

	for _, p := range plugins {
		if p.Name == name {
			return &p, nil
		}
	}

	return nil, fmt.Errorf("plugin %q not found in config", name)
}

// UpdateInstalledPluginVersion updates the version of an installed plugin.
func UpdateInstalledPluginVersion(name, version string) error {
	plugins, err := LoadInstalledPlugins()
	if err != nil {
		return err
	}

	found := false
	for i, p := range plugins {
		if p.Name == name {
			plugins[i].Version = version
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("plugin %q not found in config", name)
	}

	return SaveInstalledPlugins(plugins)
}

// GetMissingPlugins returns plugins that are in config but not installed on disk.
func GetMissingPlugins() ([]InstalledPlugin, error) {
	plugins, err := LoadInstalledPlugins()
	if err != nil {
		return nil, err
	}

	homeDir, _ := os.UserHomeDir()
	pluginsDir := filepath.Join(homeDir, ".pulumicost", "plugins")

	var missing []InstalledPlugin
	for _, p := range plugins {
		pluginDir := filepath.Join(pluginsDir, p.Name, p.Version)
		if _, statErr := os.Stat(pluginDir); os.IsNotExist(statErr) {
			missing = append(missing, p)
		}
	}

	return missing, nil
}
