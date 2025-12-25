package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/stretchr/testify/assert"
)

func TestDefaultSpinner(t *testing.T) {
	s := DefaultSpinner()
	assert.Equal(t, spinner.Dot, s.Spinner)
	// Verify the spinner uses ColorInfo (ANSI color 33 - blue)
	if s.Style.GetForeground() != ColorInfo {
		t.Error("DefaultSpinner should use ColorInfo")
	}
}
