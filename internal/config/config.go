package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the complete PulumiCost configuration
type Config struct {
	// Legacy fields for backward compatibility
	PluginDir string `yaml:"-"`
	SpecDir   string `yaml:"-"`
	
	// Configuration file fields
	Output  OutputConfig            `yaml:"output"`
	Plugins map[string]PluginConfig `yaml:"plugins"`
	Logging LoggingConfig           `yaml:"logging"`
	
	// Internal fields
	ConfigDir  string `yaml:"-"`
	ConfigFile string `yaml:"-"`
}

// OutputConfig contains output formatting preferences
type OutputConfig struct {
	DefaultFormat string `yaml:"default_format"`
	Precision     int    `yaml:"precision"`
}

// PluginConfig contains plugin-specific configuration
type PluginConfig struct {
	// Common cloud provider fields
	Region         string            `yaml:"region,omitempty"`
	Profile        string            `yaml:"profile,omitempty"`
	SubscriptionID string            `yaml:"subscription_id,omitempty"`
	TenantID       string            `yaml:"tenant_id,omitempty"`
	
	// Encrypted credential fields
	Credentials map[string]string `yaml:"credentials,omitempty"`
	
	// Custom plugin settings
	Settings map[string]interface{} `yaml:"settings,omitempty"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".pulumicost")
	
	return &Config{
		// Legacy fields
		PluginDir: filepath.Join(configDir, "plugins"),
		SpecDir:   filepath.Join(configDir, "specs"),
		
		// Configuration fields
		Output: OutputConfig{
			DefaultFormat: "table",
			Precision:     2,
		},
		Plugins: make(map[string]PluginConfig),
		Logging: LoggingConfig{
			Level: "info",
			File:  filepath.Join(configDir, "logs", "pulumicost.log"),
		},
		
		// Internal fields
		ConfigDir:  configDir,
		ConfigFile: filepath.Join(configDir, "config.yaml"),
	}
}

// New creates a Config with defaults - maintains backward compatibility
func New() *Config {
	return DefaultConfig()
}

// PluginPath returns the path for a specific plugin version
func (c *Config) PluginPath(name, version string) string {
	return filepath.Join(c.PluginDir, name, version)
}

// Load loads configuration from the config file with environment variable overrides
func Load() (*Config, error) {
	config := DefaultConfig()
	
	// Load from config file if it exists
	if _, err := os.Stat(config.ConfigFile); err == nil {
		if err := config.LoadFromFile(config.ConfigFile); err != nil {
			return nil, fmt.Errorf("loading config file: %w", err)
		}
	}
	
	// Apply environment variable overrides
	config.applyEnvOverrides()
	
	return config, nil
}

// LoadFromFile loads configuration from a YAML file
func (c *Config) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}
	
	if err := yaml.Unmarshal(data, c); err != nil {
		return fmt.Errorf("parsing config YAML: %w", err)
	}
	
	return nil
}

// Save saves the configuration to the config file
func (c *Config) Save() error {
	// Ensure config directory exists
	if err := os.MkdirAll(c.ConfigDir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	
	// Ensure logs directory exists
	logsDir := filepath.Dir(c.Logging.File)
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("creating logs directory: %w", err)
	}
	
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config to YAML: %w", err)
	}
	
	if err := os.WriteFile(c.ConfigFile, data, 0600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}
	
	return nil
}

// applyEnvOverrides applies environment variable overrides
func (c *Config) applyEnvOverrides() {
	// Output format override
	if format := os.Getenv("PULUMICOST_OUTPUT_FORMAT"); format != "" {
		c.Output.DefaultFormat = format
	}
	
	// Output precision override
	if precisionStr := os.Getenv("PULUMICOST_OUTPUT_PRECISION"); precisionStr != "" {
		if precision, err := strconv.Atoi(precisionStr); err == nil {
			c.Output.Precision = precision
		}
	}
	
	// Logging level override
	if level := os.Getenv("PULUMICOST_LOG_LEVEL"); level != "" {
		c.Logging.Level = level
	}
	
	// Logging file override
	if file := os.Getenv("PULUMICOST_LOG_FILE"); file != "" {
		c.Logging.File = file
	}
}

// Set sets a configuration value using dot notation (e.g., "output.default_format")
func (c *Config) Set(key, value string) error {
	parts := strings.Split(key, ".")
	if len(parts) == 0 {
		return fmt.Errorf("invalid key format: %s", key)
	}
	
	switch parts[0] {
	case "output":
		return c.setOutputValue(parts[1:], value)
	case "logging":
		return c.setLoggingValue(parts[1:], value)
	case "plugins":
		return c.setPluginValue(parts[1:], value)
	default:
		return fmt.Errorf("unknown config section: %s", parts[0])
	}
}

func (c *Config) setOutputValue(parts []string, value string) error {
	if len(parts) != 1 {
		return fmt.Errorf("invalid output key format")
	}
	
	switch parts[0] {
	case "default_format":
		if !isValidOutputFormat(value) {
			return fmt.Errorf("invalid output format: %s (valid: table, json, ndjson)", value)
		}
		c.Output.DefaultFormat = value
	case "precision":
		precision, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid precision value: %s", value)
		}
		if precision < 0 || precision > 10 {
			return fmt.Errorf("precision must be between 0 and 10")
		}
		c.Output.Precision = precision
	default:
		return fmt.Errorf("unknown output setting: %s", parts[0])
	}
	
	return nil
}

func (c *Config) setLoggingValue(parts []string, value string) error {
	if len(parts) != 1 {
		return fmt.Errorf("invalid logging key format")
	}
	
	switch parts[0] {
	case "level":
		if !isValidLogLevel(value) {
			return fmt.Errorf("invalid log level: %s (valid: debug, info, warn, error)", value)
		}
		c.Logging.Level = value
	case "file":
		c.Logging.File = value
	default:
		return fmt.Errorf("unknown logging setting: %s", parts[0])
	}
	
	return nil
}

func (c *Config) setPluginValue(parts []string, value string) error {
	if len(parts) < 2 {
		return fmt.Errorf("invalid plugin key format (expected plugins.name.setting)")
	}
	
	pluginName := parts[0]
	setting := parts[1]
	
	if c.Plugins == nil {
		c.Plugins = make(map[string]PluginConfig)
	}
	
	pluginConfig, exists := c.Plugins[pluginName]
	if !exists {
		pluginConfig = PluginConfig{
			Credentials: make(map[string]string),
			Settings:    make(map[string]interface{}),
		}
	}
	
	switch setting {
	case "region":
		pluginConfig.Region = value
	case "profile":
		pluginConfig.Profile = value
	case "subscription_id":
		pluginConfig.SubscriptionID = value
	case "tenant_id":
		pluginConfig.TenantID = value
	default:
		// Store in custom settings
		if pluginConfig.Settings == nil {
			pluginConfig.Settings = make(map[string]interface{})
		}
		pluginConfig.Settings[setting] = value
	}
	
	c.Plugins[pluginName] = pluginConfig
	return nil
}

// Get retrieves a configuration value using dot notation
func (c *Config) Get(key string) (interface{}, error) {
	parts := strings.Split(key, ".")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid key format: %s", key)
	}
	
	switch parts[0] {
	case "output":
		return c.getOutputValue(parts[1:])
	case "logging":
		return c.getLoggingValue(parts[1:])
	case "plugins":
		return c.getPluginValue(parts[1:])
	default:
		return nil, fmt.Errorf("unknown config section: %s", parts[0])
	}
}

func (c *Config) getOutputValue(parts []string) (interface{}, error) {
	if len(parts) != 1 {
		return nil, fmt.Errorf("invalid output key format")
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

func (c *Config) getLoggingValue(parts []string) (interface{}, error) {
	if len(parts) != 1 {
		return nil, fmt.Errorf("invalid logging key format")
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

func (c *Config) getPluginValue(parts []string) (interface{}, error) {
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid plugin key format (expected plugins.name.setting)")
	}
	
	pluginName := parts[0]
	setting := parts[1]
	
	pluginConfig, exists := c.Plugins[pluginName]
	if !exists {
		return nil, fmt.Errorf("plugin %s not configured", pluginName)
	}
	
	switch setting {
	case "region":
		return pluginConfig.Region, nil
	case "profile":
		return pluginConfig.Profile, nil
	case "subscription_id":
		return pluginConfig.SubscriptionID, nil
	case "tenant_id":
		return pluginConfig.TenantID, nil
	default:
		if pluginConfig.Settings != nil {
			if value, exists := pluginConfig.Settings[setting]; exists {
				return value, nil
			}
		}
		return nil, fmt.Errorf("setting %s not found for plugin %s", setting, pluginName)
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if !isValidOutputFormat(c.Output.DefaultFormat) {
		return fmt.Errorf("invalid output format: %s", c.Output.DefaultFormat)
	}
	
	if c.Output.Precision < 0 || c.Output.Precision > 10 {
		return fmt.Errorf("precision must be between 0 and 10")
	}
	
	if !isValidLogLevel(c.Logging.Level) {
		return fmt.Errorf("invalid log level: %s", c.Logging.Level)
	}
	
	return nil
}

func isValidOutputFormat(format string) bool {
	return format == "table" || format == "json" || format == "ndjson"
}

func isValidLogLevel(level string) bool {
	return level == "debug" || level == "info" || level == "warn" || level == "error"
}

// Credential management with encryption

// SetCredential sets an encrypted credential for a plugin
func (c *Config) SetCredential(pluginName, key, value string) error {
	if c.Plugins == nil {
		c.Plugins = make(map[string]PluginConfig)
	}
	
	pluginConfig, exists := c.Plugins[pluginName]
	if !exists {
		pluginConfig = PluginConfig{
			Credentials: make(map[string]string),
			Settings:    make(map[string]interface{}),
		}
	}
	
	if pluginConfig.Credentials == nil {
		pluginConfig.Credentials = make(map[string]string)
	}
	
	// Encrypt the credential value
	encrypted, err := c.encryptValue(value)
	if err != nil {
		return fmt.Errorf("encrypting credential: %w", err)
	}
	
	pluginConfig.Credentials[key] = encrypted
	c.Plugins[pluginName] = pluginConfig
	
	return nil
}

// GetCredential retrieves and decrypts a credential for a plugin
func (c *Config) GetCredential(pluginName, key string) (string, error) {
	pluginConfig, exists := c.Plugins[pluginName]
	if !exists {
		return "", fmt.Errorf("plugin %s not configured", pluginName)
	}
	
	encrypted, exists := pluginConfig.Credentials[key]
	if !exists {
		return "", fmt.Errorf("credential %s not found for plugin %s", key, pluginName)
	}
	
	// Decrypt the credential value
	decrypted, err := c.decryptValue(encrypted)
	if err != nil {
		return "", fmt.Errorf("decrypting credential: %w", err)
	}
	
	return decrypted, nil
}

// encryptValue encrypts a value using AES-GCM
func (c *Config) encryptValue(value string) (string, error) {
	key := c.getEncryptionKey()
	
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("creating cipher: %w", err)
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("creating GCM: %w", err)
	}
	
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generating nonce: %w", err)
	}
	
	ciphertext := gcm.Seal(nonce, nonce, []byte(value), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptValue decrypts a value using AES-GCM
func (c *Config) decryptValue(encrypted string) (string, error) {
	key := c.getEncryptionKey()
	
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("decoding base64: %w", err)
	}
	
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("creating cipher: %w", err)
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("creating GCM: %w", err)
	}
	
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypting: %w", err)
	}
	
	return string(plaintext), nil
}

// getEncryptionKey generates or retrieves the encryption key
func (c *Config) getEncryptionKey() []byte {
	// Use a deterministic key based on config directory path and machine ID
	// In production, this should use a more secure key management system
	source := c.ConfigDir + getMachineID()
	hash := sha256.Sum256([]byte(source))
	return hash[:]
}

// getMachineID returns a machine-specific identifier
func getMachineID() string {
	// Try to get a machine-specific ID from various sources
	if hostname, err := os.Hostname(); err == nil {
		return hostname
	}
	
	// Fallback to a constant (not secure, but works for demo)
	return "pulumicost-default-key"
}

// ListAll returns all configuration as a map for display
func (c *Config) ListAll() map[string]interface{} {
	result := make(map[string]interface{})
	
	// Output settings
	result["output.default_format"] = c.Output.DefaultFormat
	result["output.precision"] = c.Output.Precision
	
	// Logging settings
	result["logging.level"] = c.Logging.Level
	result["logging.file"] = c.Logging.File
	
	// Plugin settings (excluding encrypted credentials)
	for pluginName, pluginConfig := range c.Plugins {
		if pluginConfig.Region != "" {
			result[fmt.Sprintf("plugins.%s.region", pluginName)] = pluginConfig.Region
		}
		if pluginConfig.Profile != "" {
			result[fmt.Sprintf("plugins.%s.profile", pluginName)] = pluginConfig.Profile
		}
		if pluginConfig.SubscriptionID != "" {
			result[fmt.Sprintf("plugins.%s.subscription_id", pluginName)] = pluginConfig.SubscriptionID
		}
		if pluginConfig.TenantID != "" {
			result[fmt.Sprintf("plugins.%s.tenant_id", pluginName)] = pluginConfig.TenantID
		}
		
		// Custom settings
		for key, value := range pluginConfig.Settings {
			result[fmt.Sprintf("plugins.%s.%s", pluginName, key)] = value
		}
		
		// Show credential keys (but not values)
		for key := range pluginConfig.Credentials {
			result[fmt.Sprintf("plugins.%s.credentials.%s", pluginName, key)] = "<encrypted>"
		}
	}
	
	return result
}

// InitConfig creates a default configuration file
func InitConfig() error {
	config := DefaultConfig()
	
	// Check if config file already exists
	if _, err := os.Stat(config.ConfigFile); err == nil {
		return fmt.Errorf("configuration file already exists: %s", config.ConfigFile)
	}
	
	return config.Save()
}
