package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/crypto/pbkdf2"
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
	if err := cfg.Load(); err != nil {
		if os.IsNotExist(err) {
			// Config file doesn't exist - this is fine, use defaults
		} else if os.IsPermission(err) {
			// Permission error - warn but continue
			fmt.Fprintf(os.Stderr, "Warning: Permission denied reading config file: %v\n", err)
		} else {
			// Likely a corrupted config file - warn about potential data corruption
			fmt.Fprintf(os.Stderr, "Warning: Config file may be corrupted, using defaults: %v\n", err)
		}
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
	
	// Validate logging file path (if specified)
	if c.Logging.File != "" {
		if !filepath.IsAbs(c.Logging.File) {
			return fmt.Errorf("log file path must be absolute: %s", c.Logging.File)
		}
		
		// Check if parent directory is accessible
		logDir := filepath.Dir(c.Logging.File)
		if _, err := os.Stat(logDir); err != nil && os.IsNotExist(err) {
			// Try to create the directory
			if err := os.MkdirAll(logDir, 0755); err != nil {
				return fmt.Errorf("cannot create log directory %s: %w", logDir, err)
			}
		}
	}
	
	// Validate plugin configurations
	if err := c.validatePluginConfigurations(); err != nil {
		return fmt.Errorf("plugin configuration validation failed: %w", err)
	}
	
	return nil
}

// validatePluginConfigurations validates plugin-specific configurations
func (c *Config) validatePluginConfigurations() error {
	for pluginName, pluginConfig := range c.Plugins {
		// Validate plugin name (must be valid identifier)
		if pluginName == "" {
			return fmt.Errorf("plugin name cannot be empty")
		}
		
		// Check for suspicious characters that might cause issues
		for _, char := range pluginName {
			if !((char >= 'a' && char <= 'z') || 
				 (char >= 'A' && char <= 'Z') || 
				 (char >= '0' && char <= '9') || 
				 char == '-' || char == '_') {
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
	// More efficient: only scan environment variables that start with our prefix
	c.scanPluginEnvironmentVars()
}

// scanPluginEnvironmentVars efficiently scans for plugin-specific environment variables
func (c *Config) scanPluginEnvironmentVars() {
	const pluginPrefix = "PULUMICOST_PLUGIN_"
	
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, pluginPrefix) {
			continue
		}
		
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := parts[0]
		value := parts[1]
		
		// Extract plugin name and config key
		keyParts := strings.Split(strings.TrimPrefix(key, pluginPrefix), "_")
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

// deriveKey derives an encryption key from machine-specific data with proper entropy
func deriveKey() []byte {
	homeDir, _ := os.UserHomeDir()
	keyFilePath := filepath.Join(homeDir, ".pulumicost", ".keydata")
	
	// Try to load existing salt from file
	salt := loadOrCreateSalt(keyFilePath)
	
	// Gather machine-specific entropy
	var entropy strings.Builder
	
	// Basic identifiers (less predictable than previous version)
	if hostname, err := os.Hostname(); err == nil {
		entropy.WriteString(hostname)
	}
	
	if user := getUserIdentifier(); user != "" {
		entropy.WriteString(user)
	}
	
	// Add OS and architecture information
	entropy.WriteString(runtime.GOOS)
	entropy.WriteString(runtime.GOARCH)
	
	// Add home directory path as additional entropy
	entropy.WriteString(homeDir)
	
	// Use PBKDF2 with high iteration count for key derivation
	key := pbkdf2.Key([]byte(entropy.String()), salt, 100000, 32, sha256.New)
	return key
}

// loadOrCreateSalt loads an existing salt or creates a new one
func loadOrCreateSalt(keyFilePath string) []byte {
	// Try to read existing salt
	if data, err := os.ReadFile(keyFilePath); err == nil && len(data) >= 32 {
		return data[:32] // Use first 32 bytes as salt
	}
	
	// Create new random salt
	salt := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		// Fallback to deterministic salt if random generation fails
		// This maintains backward compatibility but is less secure
		hostname, _ := os.Hostname()
		user := getUserIdentifier()
		fallbackSeed := fmt.Sprintf("%s-%s-fallback", hostname, user)
		hash := sha256.Sum256([]byte(fallbackSeed))
		copy(salt, hash[:])
	} else {
		// Save the randomly generated salt for future use
		os.MkdirAll(filepath.Dir(keyFilePath), 0700)
		os.WriteFile(keyFilePath, salt, 0600)
	}
	
	return salt
}

// getUserIdentifier gets a user identifier across platforms
func getUserIdentifier() string {
	// Try different environment variables for cross-platform compatibility
	for _, envVar := range []string{"USER", "USERNAME", "LOGNAME"} {
		if user := os.Getenv(envVar); user != "" {
			return user
		}
	}
	
	// Fallback to user home directory path
	if homeDir, err := os.UserHomeDir(); err == nil {
		return filepath.Base(homeDir)
	}
	
	return "unknown"
}
