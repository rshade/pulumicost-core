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
	assert.Equal(t, ColorInfo, s.Style.GetForeground(), "DefaultSpinner should use ColorInfo")
}
