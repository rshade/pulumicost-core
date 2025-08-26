package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the complete configuration structure
type Config struct {
	// Legacy fields for backward compatibility
	PluginDir string `yaml:"-" json:"-"`
	SpecDir   string `yaml:"-" json:"-"`
	
	// New comprehensive configuration
	Output  OutputConfig          `yaml:"output" json:"output"`
	Plugins map[string]PluginConfig `yaml:"plugins" json:"plugins"`
	Logging LoggingConfig         `yaml:"logging" json:"logging"`
	
	// Internal fields
	configPath string
	encKey     []byte
}

// OutputConfig defines output formatting preferences
type OutputConfig struct {
	DefaultFormat string `yaml:"default_format" json:"default_format"`
	Precision     int    `yaml:"precision" json:"precision"`
}

// PluginConfig defines plugin-specific configuration
type PluginConfig struct {
	Config map[string]interface{} `yaml:",inline" json:",inline"`
}

// LoggingConfig defines logging preferences
type LoggingConfig struct {
	Level string `yaml:"level" json:"level"`
	File  string `yaml:"file" json:"file"`
}

// EncryptedValue represents an encrypted configuration value
type EncryptedValue struct {
	Data string `yaml:"data" json:"data"`
}

// New creates a new configuration with defaults
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
			Precision:     2,
		},
		Plugins: make(map[string]PluginConfig),
		Logging: LoggingConfig{
			Level: "info",
			File:  filepath.Join(pulumicostDir, "logs", "pulumicost.log"),
		},
		
		configPath: filepath.Join(pulumicostDir, "config.yaml"),
		encKey:     deriveKey(),
	}
	
	// Load from file if exists
	if err := cfg.Load(); err != nil && !os.IsNotExist(err) {
		// Log error but continue with defaults
		fmt.Fprintf(os.Stderr, "Warning: Failed to load config: %v\n", err)
	}
	
	// Apply environment variable overrides
	cfg.applyEnvOverrides()
	
	return cfg
}

// Load loads configuration from the config file
func (c *Config) Load() error {
	data, err := os.ReadFile(c.configPath)
	if err != nil {
		return err
	}
	
	return yaml.Unmarshal(data, c)
}

// Save saves the current configuration to the config file
func (c *Config) Save() error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(c.configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	return os.WriteFile(c.configPath, data, 0600)
}

// Set sets a configuration value using dot notation
func (c *Config) Set(key, value string) error {
	parts := strings.Split(key, ".")
	if len(parts) < 1 {
		return fmt.Errorf("invalid key format")
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

// Get gets a configuration value using dot notation
func (c *Config) Get(key string) (interface{}, error) {
	parts := strings.Split(key, ".")
	if len(parts) < 1 {
		return nil, fmt.Errorf("invalid key format")
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

// List returns all configuration as a map
func (c *Config) List() map[string]interface{} {
	return map[string]interface{}{
		"output":  c.Output,
		"plugins": c.Plugins,
		"logging": c.Logging,
	}
}

// Validate validates the configuration
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
	
	// Validate logging level
	validLevels := []string{"debug", "info", "warn", "error"}
	valid = false
	for _, level := range validLevels {
		if c.Logging.Level == level {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid log level: %s (must be one of: %v)", c.Logging.Level, validLevels)
	}
	
	return nil
}

// EncryptValue encrypts a sensitive value
func (c *Config) EncryptValue(value string) (string, error) {
	block, err := aes.NewCipher(c.encKey)
	if err != nil {
		return "", err
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	
	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return "", err
	}
	
	ciphertext := gcm.Seal(nonce, nonce, []byte(value), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptValue decrypts a sensitive value
func (c *Config) DecryptValue(encryptedValue string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedValue)
	if err != nil {
		return "", err
	}
	
	block, err := aes.NewCipher(c.encKey)
	if err != nil {
		return "", err
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	
	return string(plaintext), nil
}

// PluginPath returns the path for a specific plugin version (backward compatibility)
func (c *Config) PluginPath(name, version string) string {
	return filepath.Join(c.PluginDir, name, version)
}

// GetPluginConfig returns configuration for a specific plugin
func (c *Config) GetPluginConfig(pluginName string) (map[string]interface{}, error) {
	if plugin, exists := c.Plugins[pluginName]; exists {
		return plugin.Config, nil
	}
	return make(map[string]interface{}), nil
}

// SetPluginConfig sets configuration for a specific plugin
func (c *Config) SetPluginConfig(pluginName string, config map[string]interface{}) {
	if c.Plugins == nil {
		c.Plugins = make(map[string]PluginConfig)
	}
	c.Plugins[pluginName] = PluginConfig{Config: config}
}

// applyEnvOverrides applies environment variable overrides
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
	if logFile := os.Getenv("PULUMICOST_LOG_FILE"); logFile != "" {
		c.Logging.File = logFile
	}
	
	// Plugin overrides (PULUMICOST_PLUGIN_<NAME>_<KEY>=value)
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "PULUMICOST_PLUGIN_") {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) != 2 {
				continue
			}
			
			key := parts[0]
			value := parts[1]
			
			// Extract plugin name and config key
			keyParts := strings.Split(strings.TrimPrefix(key, "PULUMICOST_PLUGIN_"), "_")
			if len(keyParts) < 2 {
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
}

// Helper methods for setting values
func (c *Config) setOutputValue(parts []string, value string) error {
	if len(parts) != 1 {
		return fmt.Errorf("invalid output key")
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
	if len(parts) < 2 {
		return fmt.Errorf("plugin key must be in format plugins.<name>.<key>")
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
		return fmt.Errorf("invalid logging key")
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

// Helper methods for getting values
func (c *Config) getOutputValue(parts []string) (interface{}, error) {
	if len(parts) != 1 {
		return nil, fmt.Errorf("invalid output key")
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
		return nil, fmt.Errorf("invalid logging key")
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

// deriveKey derives an encryption key from machine-specific data
func deriveKey() []byte {
	// Use hostname + user as seed for machine-specific key
	hostname, _ := os.Hostname()
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME") // Windows
	}
	
	seed := fmt.Sprintf("%s-%s", hostname, user)
	hash := sha256.Sum256([]byte(seed))
	return hash[:]
}
