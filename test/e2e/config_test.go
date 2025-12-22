package e2e

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2ETestConfig_LoadFromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("PULUMICOST_E2E_AWS_REGION", "us-west-2")
	os.Setenv("PULUMICOST_E2E_TOLERANCE", "0.05")
	os.Setenv("PULUMICOST_E2E_TIMEOUT", "5m")
	defer func() {
		os.Unsetenv("PULUMICOST_E2E_AWS_REGION")
		os.Unsetenv("PULUMICOST_E2E_TOLERANCE")
		os.Unsetenv("PULUMICOST_E2E_TIMEOUT")
	}()

	config, err := LoadConfig()
	require.NoError(t, err)
	
	assert.Equal(t, "us-west-2", config.AWSRegion)
	assert.Equal(t, 0.05, config.Tolerance)
	assert.Equal(t, 5*time.Minute, config.Timeout)
}

func TestE2ETestConfig_Defaults(t *testing.T) {
	os.Clearenv() // Ensure no env vars interfere
	
	config, err := LoadConfig()
	require.NoError(t, err)
	
	// Check defaults
	assert.Equal(t, "us-east-1", config.AWSRegion)
	assert.Equal(t, 0.05, config.Tolerance) // Default 5%
	assert.Equal(t, 10*time.Minute, config.Timeout)
}