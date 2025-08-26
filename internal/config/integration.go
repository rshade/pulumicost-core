package config

import (
	"os"
	"path/filepath"
)

// GlobalConfig holds the global configuration instance
var GlobalConfig *Config

// InitGlobalConfig initializes the global configuration
func InitGlobalConfig() {
	GlobalConfig = New()
}

// GetGlobalConfig returns the global configuration, initializing it if needed
func GetGlobalConfig() *Config {
	if GlobalConfig == nil {
		InitGlobalConfig()
	}
	return GlobalConfig
}

// GetDefaultOutputFormat returns the configured default output format
func GetDefaultOutputFormat() string {
	cfg := GetGlobalConfig()
	return cfg.Output.DefaultFormat
}

// GetOutputPrecision returns the configured output precision
func GetOutputPrecision() int {
	cfg := GetGlobalConfig()
	return cfg.Output.Precision
}

// GetLogLevel returns the configured log level
func GetLogLevel() string {
	cfg := GetGlobalConfig()
	return cfg.Logging.Level
}

// GetLogFile returns the configured log file path
func GetLogFile() string {
	cfg := GetGlobalConfig()
	return cfg.Logging.File
}

// GetPluginConfiguration returns configuration for a specific plugin
func GetPluginConfiguration(pluginName string) (map[string]interface{}, error) {
	cfg := GetGlobalConfig()
	return cfg.GetPluginConfig(pluginName)
}

// EnsureConfigDir ensures the pulumicost configuration directory exists
func EnsureConfigDir() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	
	configDir := filepath.Join(homeDir, ".pulumicost")
	return os.MkdirAll(configDir, 0755)
}

// EnsureLogDir ensures the pulumicost log directory exists
func EnsureLogDir() error {
	cfg := GetGlobalConfig()
	logDir := filepath.Dir(cfg.Logging.File)
	return os.MkdirAll(logDir, 0755)
}