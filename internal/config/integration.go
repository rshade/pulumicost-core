package config

import (
	"fmt"
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
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".pulumicost")
	return os.MkdirAll(configDir, 0700)
}

// EnsureLogDir ensures the directory for the configured PulumiCost log file exists.
// It reads the global configuration and, if a log file is configured, creates its
// parent directory with permission 0700. If no log file is configured, it does nothing.
// It returns any error encountered while creating the directory.
func EnsureLogDir() error {
	cfg := GetGlobalConfig()
	if cfg.Logging.File == "" {
		return nil
	}
	logDir := filepath.Dir(cfg.Logging.File)
	if err := os.MkdirAll(logDir, 0700); err != nil {
		return fmt.Errorf("failed to create log directory %q: %w", logDir, err)
	}
	return nil
}

// GetConfigDir returns the path to the pulumicost configuration directory.
// It yields "<home>/.pulumicost" or an error if the user's home directory cannot be determined.
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".pulumicost"), nil
}

// GetPluginDir returns the path to the plugins subdirectory under the user's configuration directory (for example, ~/.pulumicost/plugins).
// It returns an error if the base configuration directory cannot be determined.
func GetPluginDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "plugins"), nil
}

// GetSpecDir returns the path to the specs directory under the user's config directory
// (typically ~/.pulumicost/specs). It returns an error if the base config directory
// cannot be determined.
func GetSpecDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "specs"), nil
}

// EnsureSubDirs creates the standard configuration subdirectories under the user's
// config directory and ensures the log directory exists.
//
// It ensures the base config directory exists, creates the "plugins" and "specs"
// subdirectories with permission 0700, and then ensures the configured log
// directory exists. It returns an error if the user's home directory cannot be
// determined or if any directory creation operation fails.
func EnsureSubDirs() error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	// Create plugins directory
	pluginDir, err := GetPluginDir()
	if err != nil {
		return fmt.Errorf("failed to get plugin directory: %w", err)
	}
	if mkdirErr := os.MkdirAll(pluginDir, 0700); mkdirErr != nil {
		return fmt.Errorf("failed to create plugin directory %q: %w", pluginDir, mkdirErr)
	}

	// Create specs directory
	specDir, err := GetSpecDir()
	if err != nil {
		return fmt.Errorf("failed to get spec directory: %w", err)
	}
	if mkdirErr := os.MkdirAll(specDir, 0700); mkdirErr != nil {
		return fmt.Errorf("failed to create spec directory %q: %w", specDir, mkdirErr)
	}

	// Create logs directory
	return EnsureLogDir()
}
