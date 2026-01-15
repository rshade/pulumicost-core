// Package recorder implements a reference plugin that records all gRPC requests
// and optionally returns mock cost responses.
package recorder

import (
	"os"
	"strings"
)

// Config holds runtime configuration for the recorder plugin.
// Configuration is loaded from environment variables with sensible defaults.
type Config struct {
	// OutputDir is the directory where recorded JSON files are written.
	// Default: "./recorded_data"
	OutputDir string

	// MockResponse enables randomized mock cost responses.
	// When false (default), the plugin returns zero/empty costs.
	MockResponse bool
}

// Environment variable names for configuration.
const (
	EnvOutputDir    = "FINFOCUS_RECORDER_OUTPUT_DIR"
	EnvMockResponse = "FINFOCUS_RECORDER_MOCK_RESPONSE"
)

// Default configuration values.
const (
	DefaultOutputDir    = "./recorded_data"
	DefaultMockResponse = false
)

// LoadConfig creates a Config from environment variables.
// Missing variables use default values.
func LoadConfig() *Config {
	cfg := &Config{
		OutputDir:    DefaultOutputDir,
		MockResponse: DefaultMockResponse,
	}

	if dir := os.Getenv(EnvOutputDir); dir != "" {
		cfg.OutputDir = dir
	}

	if mock := os.Getenv(EnvMockResponse); mock != "" {
		cfg.MockResponse = parseBool(mock)
	}

	return cfg
}

// parseBool parses a boolean string case-insensitively.
// Returns true for "true", "1", "yes", "on" (case-insensitive).
// Returns false for all other values.
func parseBool(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	switch lower {
	case "true", "1", "yes", "on":
		return true
	default:
		return false
	}
}
