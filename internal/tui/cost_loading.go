package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LoadingState tracks the progress of plugin queries.
// It provides visual feedback while plugins are being queried asynchronously.
type LoadingState struct {
	spinner spinner.Model
	message string
}

// NewLoadingState creates a new loading state with spinner.
func NewLoadingState() *LoadingState {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorSpinner)
	return &LoadingState{
		spinner: s,
		message: "Querying cost data from plugins...",
	}
}

// Init initializes the loading state (starts spinner).
func (l *LoadingState) Init() tea.Cmd {
	return l.spinner.Tick
}

// Update updates the loading state (spinner).
func (l *LoadingState) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	l.spinner, cmd = l.spinner.Update(msg)
	return cmd
}
