// Package config provides configuration management for PulumiCost.
//
// This package handles:
// - Loading and saving configuration from YAML files
// - Output formatting preferences (format, precision)
// - Plugin-specific configuration
// - Logging configuration with multiple output destinations
// - Configuration validation with detailed error reporting
//
// The configuration is stored in ~/.pulumicost/config.yaml by default.
// Sensitive values (API keys, credentials) should be stored as environment variables.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	defaultPrecision     = 2
	envKeySeparatorCount = 2
	minPluginKeyParts    = 2
)

// Config represents the complete configuration structure.
type Config struct {
	// Legacy fields for backward compatibility
	PluginDir string `yaml:"-" json:"-"`
	SpecDir   string `yaml:"-" json:"-"`

	// New comprehensive configuration
	Output  OutputConfig            `yaml:"output"  json:"output"`
	Plugins map[string]PluginConfig `yaml:"plugins" json:"plugins"`
	Logging LoggingConfig           `yaml:"logging" json:"logging"`

	// Internal fields
	configPath string
}

// OutputConfig defines output formatting preferences.
type OutputConfig struct {
	DefaultFormat string `yaml:"default_format" json:"default_format"`
	Precision     int    `yaml:"precision"      json:"precision"`
}

// PluginConfig defines plugin-specific configuration.
type PluginConfig struct {
	Config map[string]interface{} `yaml:",inline" json:",inline"`
}

// LoggingConfig defines logging preferences.
type LoggingConfig struct {
	Level   string      `yaml:"level"   json:"level"`
	Format  string      `yaml:"format"  json:"format"`  // "json" or "text"
	Outputs []LogOutput `yaml:"outputs" json:"outputs"` // Multiple output destinations
	File    string      `yaml:"file"    json:"file"`    // Legacy: single file output
}

// LogOutput defines a logging output destination.
type LogOutput struct {
	Type      string `yaml:"type"                  json:"type"`                  // "console", "file", "syslog"
	Level     string `yaml:"level,omitempty"       json:"level,omitempty"`       // Optional: override global level
	Path      string `yaml:"path,omitempty"        json:"path,omitempty"`        // For file type
	Format    string `yaml:"format,omitempty"      json:"format,omitempty"`      // Optional: override global format
	MaxSizeMB int    `yaml:"max_size_mb,omitempty" json:"max_size_mb,omitempty"` // File rotation
	MaxFiles  int    `yaml:"max_files,omitempty"   json:"max_files,omitempty"`   // File rotation
}

// New creates a new configuration with defaults.
func New() *Config {
	homeDir, _ := os.UserHomeDir()
	pulumicostDir := filepath.Join(homeDir, ".pulumicost")

	cfg := &Config{
		// Legacy fields
		PluginDir: filepath.Join(pulumicostDir, "plugins"),
		SpecDir:   filepath.Join(pulumicostDir, "specs"),

		// New configuration
		Output: OutputConfig{
			DefaultFormat: "table",
			Precision:     defaultPrecision,
		},
		Plugins: make(map[string]PluginConfig),
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
			File:   filepath.Join(pulumicostDir, "logs", "pulumicost.log"),
			Outputs: []LogOutput{
				{
					Type:   "console",
					Level:  "info",
					Format: "text",
				},
			},
		},

		configPath: filepath.Join(pulumicostDir, "config.yaml"),
	}

	// Load from file if exists
	if err := cfg.Load(); err != nil {
		switch {
		case os.IsNotExist(err):
			// Config file doesn't exist - this is fine, use defaults
		case os.IsPermission(err):
			// Permission error - warn but continue
			fmt.Fprintf(os.Stderr, "Warning: Permission denied reading config file: %v\n", err)
		default:
			// Likely a corrupted config file - warn about potential data corruption
			fmt.Fprintf(os.Stderr, "Warning: Config file may be corrupted, using defaults: %v\n", err)
		}
	}

	// Apply environment variable overrides
	cfg.applyEnvOverrides()

	return cfg
}

// Load loads configuration from the config file.
func (c *Config) Load() error {
	data, err := os.ReadFile(c.configPath)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, c)
}

// Save saves the current configuration to the config file.
func (c *Config) Save() error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(c.configPath), 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(c.configPath, data, 0600)
}

// Set sets a configuration value using dot notation.
func (c *Config) Set(key, value string) error {
	parts := strings.Split(key, ".")
	if len(parts) < 1 {
		return errors.New("invalid key format")
	}

	switch parts[0] {
	case "output":
		return c.setOutputValue(parts[1:], value)
	case "plugins":
		return c.setPluginValue(parts[1:], value)
	case "logging":
		return c.setLoggingValue(parts[1:], value)
	default:
		return fmt.Errorf("unknown configuration section: %s", parts[0])
	}
}

// Get gets a configuration value using dot notation.
func (c *Config) Get(key string) (interface{}, error) {
	parts := strings.Split(key, ".")
	if len(parts) < 1 {
		return nil, errors.New("invalid key format")
	}

	switch parts[0] {
	case "output":
		return c.getOutputValue(parts[1:])
	case "plugins":
		return c.getPluginValue(parts[1:])
	case "logging":
		return c.getLoggingValue(parts[1:])
	default:
		return nil, fmt.Errorf("unknown configuration section: %s", parts[0])
	}
}

// List returns all configuration as a map.
func (c *Config) List() map[string]interface{} {
	return map[string]interface{}{
		"output":  c.Output,
		"plugins": c.Plugins,
		"logging": c.Logging,
	}
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	// Validate output format
	validFormats := []string{"table", "json", "ndjson"}
	valid := false
	for _, format := range validFormats {
		if c.Output.DefaultFormat == format {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid output format: %s (must be one of: %v)", c.Output.DefaultFormat, validFormats)
	}

	// Validate precision
	if c.Output.Precision < 0 || c.Output.Precision > 10 {
		return fmt.Errorf("invalid precision: %d (must be between 0 and 10)", c.Output.Precision)
	}

	// Validate logging configuration
	if err := c.validateLogging(); err != nil {
		return fmt.Errorf("logging configuration validation failed: %w", err)
	}

	// Validate plugin configurations
	if err := c.validatePluginConfigurations(); err != nil {
		return fmt.Errorf("plugin configuration validation failed: %w", err)
	}

	return nil
}

// validateLogging validates logging configuration.
func (c *Config) validateLogging() error {
	// Validate logging level
	if c.Logging.Level != "" {
		if err := isValidLevel(c.Logging.Level); err != nil {
			return err
		}
	}

	// Validate logging format
	if c.Logging.Format != "" {
		if err := isValidFormat(c.Logging.Format); err != nil {
			return err
		}
	}

	// Validate legacy file path (if specified)
	if c.Logging.File != "" {
		if err := validateFilePath(c.Logging.File); err != nil {
			return err
		}
	}

	// Validate outputs
	for i, output := range c.Logging.Outputs {
		if err := validateLogOutput(output); err != nil {
			return fmt.Errorf("output %d: %w", i, err)
		}
	}

	return nil
}

func isValidLevel(level string) error {
	validLevels := []string{"debug", "info", "warn", "error"}
	for _, validLevel := range validLevels {
		if level == validLevel {
			return nil
		}
	}
	return fmt.Errorf("invalid log level: %s (must be one of: %v)", level, validLevels)
}

func isValidFormat(format string) error {
	validFormats := []string{"json", "text"}
	for _, validFormat := range validFormats {
		if format == validFormat {
			return nil
		}
	}
	return fmt.Errorf("invalid log format: %s (must be one of: %v)", format, validFormats)
}

func validateFilePath(path string) error {
	if !filepath.IsAbs(path) {
		return fmt.Errorf("log file path must be absolute: %s", path)
	}

	// Check if parent directory is accessible
	logDir := filepath.Dir(path)
	if _, statErr := os.Stat(logDir); statErr != nil && os.IsNotExist(statErr) {
		// Try to create the directory
		if mkdirErr := os.MkdirAll(logDir, 0700); mkdirErr != nil {
			return fmt.Errorf("cannot create log directory %s: %w", logDir, mkdirErr)
		}
	}

	return nil
}

const (
	outputTypeFile = "file"
)

// validateLogOutput validates a single log output configuration.
func validateLogOutput(output LogOutput) error {
	// Validate output type
	if err := validateOutputType(output.Type); err != nil {
		return err
	}

	// Validate level if specified
	if output.Level != "" {
		if err := isValidLevel(output.Level); err != nil {
			return err
		}
	}

	// Validate format if specified
	if output.Format != "" {
		if err := isValidFormat(output.Format); err != nil {
			return err
		}
	}

	// File-specific validations
	if output.Type == outputTypeFile {
		return validateFileOutput(output)
	}

	return nil
}

func validateOutputType(outputType string) error {
	validTypes := []string{"console", outputTypeFile, "syslog"}
	for _, t := range validTypes {
		if outputType == t {
			return nil
		}
	}
	return fmt.Errorf("invalid output type: %s (must be one of: %v)", outputType, validTypes)
}

func validateFileOutput(output LogOutput) error {
	if output.Path == "" {
		return errors.New("file output requires 'path' field")
	}
	if !filepath.IsAbs(output.Path) {
		return fmt.Errorf("file path must be absolute: %s", output.Path)
	}

	// Check if parent directory is accessible
	logDir := filepath.Dir(output.Path)
	if _, statErr := os.Stat(logDir); statErr != nil && os.IsNotExist(statErr) {
		// Try to create the directory
		if mkdirErr := os.MkdirAll(logDir, 0700); mkdirErr != nil {
			return fmt.Errorf("cannot create log directory %s: %w", logDir, mkdirErr)
		}
	}

	// Validate rotation settings
	if output.MaxSizeMB < 0 {
		return fmt.Errorf("max_size_mb must be non-negative (0 means unlimited), got: %d", output.MaxSizeMB)
	}
	if output.MaxFiles < 0 {
		return fmt.Errorf("max_files must be non-negative (0 means unlimited), got: %d", output.MaxFiles)
	}

	return nil
}

// validatePluginConfigurations validates plugin-specific configurations.
func (c *Config) validatePluginConfigurations() error {
	for pluginName, pluginConfig := range c.Plugins {
		// Validate plugin name (must be valid identifier)
		if pluginName == "" {
			return errors.New("plugin name cannot be empty")
		}

		// Check for suspicious characters that might cause issues
		for _, char := range pluginName {
			if (char < 'a' || char > 'z') &&
				(char < 'A' || char > 'Z') &&
				(char < '0' || char > '9') &&
				char != '-' && char != '_' {
				return fmt.Errorf("plugin name contains invalid character: %s", pluginName)
			}
		}

		// Validate plugin configuration keys
		for configKey := range pluginConfig.Config {
			if configKey == "" {
				return fmt.Errorf("plugin %s has empty configuration key", pluginName)
			}
		}
	}

	return nil
}

// PluginPath returns the path for a specific plugin version (backward compatibility).
func (c *Config) PluginPath(name, version string) string {
	return filepath.Join(c.PluginDir, name, version)
}

// GetPluginConfig returns configuration for a specific plugin.
func (c *Config) GetPluginConfig(pluginName string) (map[string]interface{}, error) {
	if plugin, exists := c.Plugins[pluginName]; exists {
		return plugin.Config, nil
	}
	return make(map[string]interface{}), nil
}

// SetPluginConfig sets configuration for a specific plugin.
func (c *Config) SetPluginConfig(pluginName string, config map[string]interface{}) {
	if c.Plugins == nil {
		c.Plugins = make(map[string]PluginConfig)
	}
	c.Plugins[pluginName] = PluginConfig{Config: config}
}

// applyEnvOverrides applies environment variable overrides.
func (c *Config) applyEnvOverrides() {
	// Output overrides
	if format := os.Getenv("PULUMICOST_OUTPUT_FORMAT"); format != "" {
		c.Output.DefaultFormat = format
	}
	if precision := os.Getenv("PULUMICOST_OUTPUT_PRECISION"); precision != "" {
		if p, err := strconv.Atoi(precision); err == nil {
			c.Output.Precision = p
		}
	}

	// Logging overrides
	if level := os.Getenv("PULUMICOST_LOG_LEVEL"); level != "" {
		c.Logging.Level = level
	}
	if format := os.Getenv("PULUMICOST_LOG_FORMAT"); format != "" {
		c.Logging.Format = format
	}
	if logFile := os.Getenv("PULUMICOST_LOG_FILE"); logFile != "" {
		c.Logging.File = logFile
	}

	// Plugin overrides (PULUMICOST_PLUGIN_<NAME>_<KEY>=value)
	// More efficient: only scan environment variables that start with our prefix
	c.scanPluginEnvironmentVars()
}

// scanPluginEnvironmentVars efficiently scans for plugin-specific environment variables.
func (c *Config) scanPluginEnvironmentVars() {
	const pluginPrefix = "PULUMICOST_PLUGIN_"

	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, pluginPrefix) {
			continue
		}

		parts := strings.SplitN(env, "=", envKeySeparatorCount)
		if len(parts) != envKeySeparatorCount {
			continue
		}

		key := parts[0]
		value := parts[1]

		// Extract plugin name and config key
		keyParts := strings.Split(strings.TrimPrefix(key, pluginPrefix), "_")
		if len(keyParts) < minPluginKeyParts {
			continue
		}

		pluginName := strings.ToLower(keyParts[0])
		configKey := strings.ToLower(strings.Join(keyParts[1:], "_"))

		if c.Plugins == nil {
			c.Plugins = make(map[string]PluginConfig)
		}

		if _, exists := c.Plugins[pluginName]; !exists {
			c.Plugins[pluginName] = PluginConfig{Config: make(map[string]interface{})}
		}

		c.Plugins[pluginName].Config[configKey] = value
	}
}

// Helper methods for setting values.
func (c *Config) setOutputValue(parts []string, value string) error {
	if len(parts) != 1 {
		return errors.New("invalid output key")
	}

	switch parts[0] {
	case "default_format":
		c.Output.DefaultFormat = value
	case "precision":
		p, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("precision must be a number: %w", err)
		}
		c.Output.Precision = p
	default:
		return fmt.Errorf("unknown output setting: %s", parts[0])
	}

	return nil
}

func (c *Config) setPluginValue(parts []string, value string) error {
	if len(parts) < minPluginKeyParts {
		return errors.New("plugin key must be in format plugins.<name>.<key>")
	}

	pluginName := parts[0]
	configKey := strings.Join(parts[1:], ".")

	if c.Plugins == nil {
		c.Plugins = make(map[string]PluginConfig)
	}

	if _, exists := c.Plugins[pluginName]; !exists {
		c.Plugins[pluginName] = PluginConfig{Config: make(map[string]interface{})}
	}

	c.Plugins[pluginName].Config[configKey] = value
	return nil
}

func (c *Config) setLoggingValue(parts []string, value string) error {
	if len(parts) != 1 {
		return errors.New("invalid logging key")
	}

	switch parts[0] {
	case "level":
		c.Logging.Level = value
	case "file":
		c.Logging.File = value
	default:
		return fmt.Errorf("unknown logging setting: %s", parts[0])
	}

	return nil
}

// Helper methods for getting values.
func (c *Config) getOutputValue(parts []string) (interface{}, error) {
	if len(parts) != 1 {
		return nil, errors.New("invalid output key")
	}

	switch parts[0] {
	case "default_format":
		return c.Output.DefaultFormat, nil
	case "precision":
		return c.Output.Precision, nil
	default:
		return nil, fmt.Errorf("unknown output setting: %s", parts[0])
	}
}

func (c *Config) getPluginValue(parts []string) (interface{}, error) {
	if len(parts) < 1 {
		return c.Plugins, nil
	}

	pluginName := parts[0]
	plugin, exists := c.Plugins[pluginName]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginName)
	}

	if len(parts) == 1 {
		return plugin.Config, nil
	}

	configKey := strings.Join(parts[1:], ".")
	value, exists := plugin.Config[configKey]
	if !exists {
		return nil, fmt.Errorf("config key not found: %s", configKey)
	}

	return value, nil
}

func (c *Config) getLoggingValue(parts []string) (interface{}, error) {
	if len(parts) != 1 {
		return nil, errors.New("invalid logging key")
	}

	switch parts[0] {
	case "level":
		return c.Logging.Level, nil
	case "file":
		return c.Logging.File, nil
	default:
		return nil, fmt.Errorf("unknown logging setting: %s", parts[0])
	}
}

// GetOutputFormat returns the output format to use, preferring user choice over config default.
func GetOutputFormat(userChoice string) string {
	// If user provided a format, use it
	if userChoice != "" {
		return userChoice
	}

	// Try to get default from configuration (use singleton to avoid side effects)
	cfg := GetGlobalConfig()
	return cfg.Output.DefaultFormat
}
