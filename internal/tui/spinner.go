package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

// DefaultSpinner returns a spinner model configured with the standard style.
// DefaultSpinner returns a spinner.Model configured with the Dot spinner and styled using ColorInfo.
func DefaultSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorInfo)
	return s
}