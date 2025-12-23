package e2e

import (
	"os"
	"strconv"
	"time"
)

// Config holds configuration for E2E tests.
type Config struct {
	AWSRegion string
	Tolerance float64
	Timeout   time.Duration
}

// LoadConfig returns a Config populated from environment variables with sensible defaults.
// 
// It initializes defaults (AWSRegion "us-east-1", Tolerance 0.05, Timeout 10 minutes) and
// overrides them when the following environment variables are set and valid:
// PULUMICOST_E2E_AWS_REGION (string), PULUMICOST_E2E_TOLERANCE (float64), and
// PULUMICOST_E2E_TIMEOUT (duration parseable by time.ParseDuration).
// Invalid numeric or duration values are ignored and the defaults are retained.
// The function always returns a non-nil *Config and a nil error.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		AWSRegion: "us-east-1",
		Tolerance: 0.05,
		Timeout:   10 * time.Minute,
	}

	if region := os.Getenv("PULUMICOST_E2E_AWS_REGION"); region != "" {
		cfg.AWSRegion = region
	}

	if tolStr := os.Getenv("PULUMICOST_E2E_TOLERANCE"); tolStr != "" {
		if tol, err := strconv.ParseFloat(tolStr, 64); err == nil {
			cfg.Tolerance = tol
		}
	}

	if timeoutStr := os.Getenv("PULUMICOST_E2E_TIMEOUT"); timeoutStr != "" {
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			cfg.Timeout = timeout
		}
	}

	return cfg, nil
}