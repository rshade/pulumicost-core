package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LoadingState tracks the progress of plugin queries.
type LoadingState struct {
	spinner spinner.Model
	// TODO: plugins will track individual plugin query status for detailed progress display.
	plugins map[string]*PluginStatus
	// TODO: startTime will be used to show elapsed time during long-running queries.
	startTime time.Time
	message   string
}

// PluginStatus represents the status of a single plugin's query.
type PluginStatus struct {
	Name      string
	Done      bool
	Count     int
	Error     error
	StartTime time.Time
}

// NewLoadingState creates a new loading state with spinner.
func NewLoadingState() *LoadingState {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorSpinner)
	return &LoadingState{
		spinner:   s,
		plugins:   make(map[string]*PluginStatus),
		startTime: time.Now(),
		message:   "Querying cost data from plugins...",
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
