package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

// DefaultSpinner returns a spinner model configured with the standard style.
// The default spinner uses the "Dot" type and is colored with the Info color.
func DefaultSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorInfo)
	return s
}
