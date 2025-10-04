package config

import (
	"os"
	"path/filepath"
	"sync"
)

// GlobalConfig holds the global configuration instance.
var GlobalConfig *Config        //nolint:gochecknoglobals // Singleton pattern for configuration
var globalConfigMu sync.RWMutex //nolint:gochecknoglobals // Protects globalConfigInit flag
var globalConfigInit bool       //nolint:gochecknoglobals // Tracks if global config has been initialized

// InitGlobalConfig initializes the global configuration.
func InitGlobalConfig() {
	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()

	if globalConfigInit {
		return
	}

	GlobalConfig = New()
	globalConfigInit = true
}

// ResetGlobalConfigForTest resets the global config for testing purposes.
func ResetGlobalConfigForTest() {
	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()

	GlobalConfig = nil
	globalConfigInit = false
}

// GetGlobalConfig returns the global configuration, initializing it if needed.
func GetGlobalConfig() *Config {
	InitGlobalConfig()
	return GlobalConfig
}

// GetDefaultOutputFormat returns the configured default output format.
func GetDefaultOutputFormat() string {
	cfg := GetGlobalConfig()
	return cfg.Output.DefaultFormat
}

// GetOutputPrecision returns the configured output precision.
func GetOutputPrecision() int {
	cfg := GetGlobalConfig()
	return cfg.Output.Precision
}

// GetLogLevel returns the configured log level.
func GetLogLevel() string {
	cfg := GetGlobalConfig()
	return cfg.Logging.Level
}

// GetLogFile returns the configured log file path.
func GetLogFile() string {
	cfg := GetGlobalConfig()
	return cfg.Logging.File
}

// GetPluginConfiguration returns configuration for a specific plugin.
func GetPluginConfiguration(pluginName string) (map[string]interface{}, error) {
	cfg := GetGlobalConfig()
	return cfg.GetPluginConfig(pluginName)
}

// EnsureConfigDir ensures the pulumicost configuration directory exists.
func EnsureConfigDir() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".pulumicost")
	return os.MkdirAll(configDir, 0700)
}

// EnsureLogDir ensures the pulumicost log directory exists.
func EnsureLogDir() error {
	cfg := GetGlobalConfig()
	if cfg.Logging.File == "" {
		return nil
	}
	logDir := filepath.Dir(cfg.Logging.File)
	return os.MkdirAll(logDir, 0700)
}

// GetConfigDir returns the path to the pulumicost configuration directory.
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".pulumicost"), nil
}

// GetPluginDir returns the path to the plugins directory.
func GetPluginDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "plugins"), nil
}

// GetSpecDir returns the path to the specs directory.
func GetSpecDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "specs"), nil
}

// EnsureSubDirs ensures all standard subdirectories exist.
func EnsureSubDirs() error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	// Create plugins directory
	pluginDir, err := GetPluginDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(pluginDir, 0700); err != nil {
		return err
	}

	// Create specs directory
	specDir, err := GetSpecDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(specDir, 0700); err != nil {
		return err
	}

	// Create logs directory
	return EnsureLogDir()
}
