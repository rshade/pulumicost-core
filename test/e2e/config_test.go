package e2e

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2ETestConfig_LoadFromEnv(t *testing.T) {
	// Set environment variables using t.Setenv for automatic cleanup
	t.Setenv("PULUMICOST_E2E_AWS_REGION", "us-west-2")
	t.Setenv("PULUMICOST_E2E_TOLERANCE", "0.05")
	t.Setenv("PULUMICOST_E2E_TIMEOUT", "5m")

	config := LoadConfig()
	require.NotNil(t, config)

	assert.Equal(t, "us-west-2", config.AWSRegion)
	assert.Equal(t, 0.05, config.Tolerance)
	assert.Equal(t, 5*time.Minute, config.Timeout)
}

func TestE2ETestConfig_Defaults(t *testing.T) {
	// Save and restore current env to avoid polluting other tests
	savedRegion := os.Getenv("PULUMICOST_E2E_AWS_REGION")
	savedTol := os.Getenv("PULUMICOST_E2E_TOLERANCE")
	savedTimeout := os.Getenv("PULUMICOST_E2E_TIMEOUT")
	t.Cleanup(func() {
		os.Setenv("PULUMICOST_E2E_AWS_REGION", savedRegion)
		os.Setenv("PULUMICOST_E2E_TOLERANCE", savedTol)
		os.Setenv("PULUMICOST_E2E_TIMEOUT", savedTimeout)
	})

	// Unset the specific env vars we care about
	os.Unsetenv("PULUMICOST_E2E_AWS_REGION")
	os.Unsetenv("PULUMICOST_E2E_TOLERANCE")
	os.Unsetenv("PULUMICOST_E2E_TIMEOUT")

	config := LoadConfig()
	require.NotNil(t, config)

	// Check defaults
	assert.Equal(t, "us-east-1", config.AWSRegion)
	assert.Equal(t, 0.05, config.Tolerance) // Default 5%
	assert.Equal(t, 10*time.Minute, config.Timeout)
}
