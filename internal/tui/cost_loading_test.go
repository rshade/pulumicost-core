package tui

import (
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadingState_Update(t *testing.T) {
	loading := NewLoadingState()
	assert.NotNil(t, loading.spinner)
	assert.Equal(t, "Querying cost data from plugins...", loading.message)

	// Test Update with nil msg (spinner update usually requires specific msg, but returns cmd)
	// We just want to ensure it doesn't panic. cmd might be nil.
	_ = loading.Update(nil)

	// Check Init returns a command
	initCmd := loading.Init()
	assert.NotNil(t, initCmd)
}

func TestLoadingState_SpinnerTick(t *testing.T) {
	loading := NewLoadingState()

	// Test that spinner tick messages are handled without panic.
	initCmd := loading.Init()
	require.NotNil(t, initCmd)

	// Simulate a spinner tick message.
	tickMsg := spinner.TickMsg{Time: time.Now(), ID: 0}
	cmd := loading.Update(tickMsg)
	// The spinner should return another tick command.
	assert.NotNil(t, cmd)
}

func TestLoadingState_MessageUpdate(t *testing.T) {
	loading := NewLoadingState()
	assert.Equal(t, "Querying cost data from plugins...", loading.message)

	// Test updating the message.
	loading.message = "Processing results..."
	assert.Equal(t, "Processing results...", loading.message)

	output := RenderLoading(loading)
	assert.Contains(t, output, "Processing results...")
}

func TestLoadingState_StartTime(t *testing.T) {
	before := time.Now()
	loading := NewLoadingState()
	after := time.Now()

	// Verify startTime is set to approximately now.
	assert.False(t, loading.startTime.IsZero(), "startTime should be set")
	assert.True(t, loading.startTime.After(before) || loading.startTime.Equal(before),
		"startTime should be >= before")
	assert.True(t, loading.startTime.Before(after) || loading.startTime.Equal(after),
		"startTime should be <= after")
}

func TestRenderLoading(t *testing.T) {
	loading := NewLoadingState()
	output := RenderLoading(loading)
	assert.Contains(t, output, "Querying cost data from plugins...")
}

func TestRenderLoading_NilLoading(t *testing.T) {
	output := RenderLoading(nil)
	assert.Equal(t, "Loading...", output)
}
