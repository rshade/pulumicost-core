package config

import (
	"fmt"
	"os"
	"strings"
)

// GetOutputFormat returns the configured output format, falling back to CLI flag or default
func GetOutputFormat(flagValue string) string {
	// CLI flag takes precedence
	if flagValue != "" {
		return flagValue
	}
	
	// Environment variable override
	if envFormat := os.Getenv("PULUMICOST_OUTPUT_FORMAT"); envFormat != "" {
		return envFormat
	}
	
	// Try to load from config file
	cfg, err := Load()
	if err == nil && cfg.Output.DefaultFormat != "" {
		return cfg.Output.DefaultFormat
	}
	
	// Final fallback
	return "table"
}

// GetOutputPrecision returns the configured output precision, falling back to default
func GetOutputPrecision() int {
	// Environment variable override
	if envPrecision := os.Getenv("PULUMICOST_OUTPUT_PRECISION"); envPrecision != "" {
		// This is handled in applyEnvOverrides, but check here too for safety
		if precision, err := parseIntSafe(envPrecision); err == nil {
			if precision >= 0 && precision <= 10 {
				return precision
			}
		}
	}
	
	// Try to load from config file
	cfg, err := Load()
	if err == nil {
		return cfg.Output.Precision
	}
	
	// Default precision
	return 2
}

// GetPluginConfig returns configuration for a specific plugin
func GetPluginConfig(pluginName string) (PluginConfig, error) {
	cfg, err := Load()
	if err != nil {
		return PluginConfig{}, err
	}
	
	if pluginConfig, exists := cfg.Plugins[pluginName]; exists {
		return pluginConfig, nil
	}
	
	// Return empty config if plugin not configured
	return PluginConfig{
		Credentials: make(map[string]string),
		Settings:    make(map[string]interface{}),
	}, nil
}

// IsDebugMode returns true if debug logging is enabled
func IsDebugMode() bool {
	// Environment variable check
	if level := os.Getenv("PULUMICOST_LOG_LEVEL"); level != "" {
		return strings.ToLower(level) == "debug"
	}
	
	// Config file check
	cfg, err := Load()
	if err == nil {
		return strings.ToLower(cfg.Logging.Level) == "debug"
	}
	
	return false
}

// GetLogFile returns the configured log file path
func GetLogFile() string {
	// Environment variable override
	if file := os.Getenv("PULUMICOST_LOG_FILE"); file != "" {
		return file
	}
	
	// Try to load from config file
	cfg, err := Load()
	if err == nil && cfg.Logging.File != "" {
		return cfg.Logging.File
	}
	
	// Default log file
	return DefaultConfig().Logging.File
}

func parseIntSafe(s string) (int, error) {
	// Handle empty string case
	if s == "" {
		return 0, nil
	}
	
	// Check for negative numbers (not supported)
	if strings.HasPrefix(s, "-") {
		return 0, fmt.Errorf("negative integers not supported: %s", s)
	}
	
	// Implement simple int parsing without importing strconv again
	result := 0
	for _, char := range s {
		if char < '0' || char > '9' {
			return 0, fmt.Errorf("invalid integer: %s", s)
		}
		result = result*10 + int(char-'0')
	}
	return result, nil
}