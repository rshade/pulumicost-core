package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rshade/pulumicost-core/internal/engine"
)

// RecommendationSortField represents the field to sort recommendations by.
type RecommendationSortField int

const (
	// SortBySavings sorts by estimated savings (descending).
	SortBySavings RecommendationSortField = iota
	// SortByResourceID sorts by resource ID (ascending).
	SortByResourceID
	// SortByActionType sorts by action type (ascending).
	SortByActionType
)

const (
	// numRecommendationSortFields is the number of available sort fields.
	numRecommendationSortFields = 3

	// topRecommendationsLimit is the maximum number of recommendations to show in summary.
	topRecommendationsLimit = 5

	// recSummaryHeight is the height reserved for the summary section.
	recSummaryHeight = 10

	// Table column widths for recommendations.
	recColWidthResource    = 35
	recColWidthAction      = 15
	recColWidthSavings     = 12
	recColWidthDescription = 30

	// recDescTruncateLen is the maximum description length in table rows.
	recDescTruncateLen = 27

	// defaultCurrency is used when no currency is specified.
	defaultCurrency = "USD"
)

// RecommendationsSummary contains aggregated statistics for recommendations display.
type RecommendationsSummary struct {
	// TotalCount is the total number of recommendations.
	TotalCount int

	// TotalSavings is the sum of all estimated savings.
	TotalSavings float64

	// Currency is the currency for savings (typically "USD").
	Currency string

	// CountByAction maps action type to count of recommendations.
	CountByAction map[string]int

	// SavingsByAction maps action type to total savings.
	SavingsByAction map[string]float64

	// TopRecommendations contains the top 5 recommendations sorted by savings.
	TopRecommendations []engine.Recommendation
}

// NewRecommendationsSummary creates a summary from a list of recommendations.
// It calculates aggregate statistics and extracts the top 5 by savings.
func NewRecommendationsSummary(recs []engine.Recommendation) *RecommendationsSummary {
	summary := &RecommendationsSummary{
		TotalCount:      len(recs),
		CountByAction:   make(map[string]int),
		SavingsByAction: make(map[string]float64),
	}

	if len(recs) == 0 {
		return summary
	}

	// Calculate aggregates.
	for _, rec := range recs {
		summary.TotalSavings += rec.EstimatedSavings
		summary.CountByAction[rec.Type]++
		summary.SavingsByAction[rec.Type] += rec.EstimatedSavings

		// Set currency from first recommendation with a currency.
		if summary.Currency == "" && rec.Currency != "" {
			summary.Currency = rec.Currency
		}
	}

	// Sort by savings descending and take top 5.
	sorted := make([]engine.Recommendation, len(recs))
	copy(sorted, recs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].EstimatedSavings > sorted[j].EstimatedSavings
	})

	if len(sorted) > topRecommendationsLimit {
		summary.TopRecommendations = sorted[:topRecommendationsLimit]
	} else {
		summary.TopRecommendations = sorted
	}

	return summary
}

// Messages for RecommendationsViewModel.
type recommendationsLoadingMsg struct {
	recommendations []engine.Recommendation
	err             error
}

// RecommendationsViewModel is the Bubble Tea model for interactive recommendations display.
type RecommendationsViewModel struct {
	// View state
	state              ViewState
	allRecommendations []engine.Recommendation // Source of truth
	recommendations    []engine.Recommendation // Filtered/sorted for display

	// Interactive components
	table     table.Model
	textInput textinput.Model
	selected  int

	// Display configuration
	width      int
	height     int
	sortBy     RecommendationSortField
	showFilter bool
	verbose    bool

	// Loading state
	loading  *LoadingState
	fetchCmd tea.Cmd

	// Aggregated data
	summary *RecommendationsSummary

	// Error handling
	err error
}

// NewRecommendationsViewModel creates a new model with the given recommendations.
func NewRecommendationsViewModel(recs []engine.Recommendation) *RecommendationsViewModel {
	m := &RecommendationsViewModel{
		state:              ViewStateList,
		allRecommendations: recs,
		recommendations:    recs,
		textInput:          newRecTextInput(),
		summary:            NewRecommendationsSummary(recs),
		width:              defaultWidth,
		height:             defaultHeight,
	}
	m.applySort()
	m.rebuildTable()
	return m
}

// RecommendationFetcher is a context-aware function that fetches recommendations.
// The fetcher should check ctx.Done() to support cancellation.
type RecommendationFetcher func(ctx context.Context) ([]engine.Recommendation, error)

// NewRecommendationsViewModelWithLoading creates a model that starts in loading state.
// The fetcher receives a context that can be used for cancellation.
func NewRecommendationsViewModelWithLoading(
	ctx context.Context,
	fetcher RecommendationFetcher,
) *RecommendationsViewModel {
	m := &RecommendationsViewModel{
		state:     ViewStateLoading,
		loading:   NewLoadingState(),
		textInput: newRecTextInput(),
		summary:   &RecommendationsSummary{Currency: defaultCurrency}, // Initialize with empty summary
		width:     defaultWidth,
		height:    defaultHeight,
		fetchCmd: func() tea.Msg {
			recs, err := fetcher(ctx)
			return recommendationsLoadingMsg{recommendations: recs, err: err}
		},
	}
	return m
}

// newRecTextInput creates a new text input for filtering recommendations.
func newRecTextInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Filter recommendations..."
	ti.CharLimit = filterInputCharLimit
	ti.Width = filterInputWidth
	return ti
}

// SetVerbose sets whether to show all recommendations (verbose mode).
func (m *RecommendationsViewModel) SetVerbose(verbose bool) {
	m.verbose = verbose
}

// Init initializes the model.
func (m *RecommendationsViewModel) Init() tea.Cmd {
	var cmds []tea.Cmd
	if m.state == ViewStateLoading {
		cmds = append(cmds, m.loading.Init(), m.fetchCmd)
	} else if m.showFilter {
		cmds = append(cmds, textinput.Blink)
	}
	return tea.Batch(cmds...)
}

// Update handles messages and updates the model state.
func (m *RecommendationsViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle window resizing
	if winMsg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = winMsg.Width
		m.height = winMsg.Height
		m.rebuildTable()
	}

	// Handle loading complete
	if loadMsg, ok := msg.(recommendationsLoadingMsg); ok {
		return m.handleLoadingComplete(loadMsg)
	}

	// Handle filter input
	if m.showFilter {
		return m.handleFilterInput(msg)
	}

	// Handle state-specific updates
	switch m.state {
	case ViewStateLoading:
		return m.handleLoadingUpdate(msg)
	case ViewStateList:
		return m.handleListUpdate(msg)
	case ViewStateDetail:
		return m.handleDetailUpdate(msg)
	case ViewStateQuitting, ViewStateError:
		return m.handleQuitUpdate(msg)
	default:
		return m, nil
	}
}

func (m *RecommendationsViewModel) handleLoadingComplete(
	msg recommendationsLoadingMsg,
) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		m.state = ViewStateError
		return m, tea.Quit
	}
	m.allRecommendations = msg.recommendations
	m.recommendations = msg.recommendations
	m.summary = NewRecommendationsSummary(msg.recommendations)
	m.state = ViewStateList
	m.applySort()
	m.rebuildTable()
	return m, nil
}

func (m *RecommendationsViewModel) handleFilterInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case keyEnter, keyEsc:
			m.showFilter = false
			m.textInput.Blur()
			m.applyFilter()
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *RecommendationsViewModel) handleLoadingUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, m.loading.Update(msg)
}

func (m *RecommendationsViewModel) handleListUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case keyQuit, keyCtrlC:
			m.state = ViewStateQuitting
			return m, tea.Quit
		case keyEnter:
			if len(m.recommendations) > 0 {
				m.selected = m.table.Cursor()
				m.state = ViewStateDetail
			}
			return m, nil
		case keySlash:
			m.showFilter = true
			m.textInput.Focus()
			return m, nil
		case keyS:
			m.cycleSort()
			return m, nil
		case keyEsc:
			if m.textInput.Value() != "" {
				m.textInput.SetValue("")
				m.applyFilter()
			}
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *RecommendationsViewModel) handleDetailUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case keyQuit, keyCtrlC:
			m.state = ViewStateQuitting
			return m, tea.Quit
		case keyEsc:
			m.state = ViewStateList
			m.table.Focus()
			return m, nil
		}
	}
	return m, nil
}

func (m *RecommendationsViewModel) handleQuitUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case keyQuit, keyCtrlC:
			m.state = ViewStateQuitting
			return m, tea.Quit
		}
	}
	return m, nil
}

// applyFilter filters recommendations based on the text input value.
func (m *RecommendationsViewModel) applyFilter() {
	val := m.textInput.Value()
	if val == "" {
		m.recommendations = m.allRecommendations
	} else {
		var filtered []engine.Recommendation
		query := strings.ToLower(val)
		for _, r := range m.allRecommendations {
			if strings.Contains(strings.ToLower(r.ResourceID), query) ||
				strings.Contains(strings.ToLower(r.Type), query) ||
				strings.Contains(strings.ToLower(r.Description), query) {
				filtered = append(filtered, r)
			}
		}
		m.recommendations = filtered
	}
	m.summary = NewRecommendationsSummary(m.recommendations)
	m.applySort()
	m.rebuildTable()
}

// cycleSort cycles through the available sort fields.
func (m *RecommendationsViewModel) cycleSort() {
	m.sortBy = (m.sortBy + 1) % numRecommendationSortFields
	m.applySort()
	m.rebuildTable()
}

// applySort sorts recommendations based on the current sort field.
func (m *RecommendationsViewModel) applySort() {
	sort.Slice(m.recommendations, func(i, j int) bool {
		a, b := m.recommendations[i], m.recommendations[j]
		switch m.sortBy {
		case SortBySavings:
			return a.EstimatedSavings > b.EstimatedSavings
		case SortByResourceID:
			return a.ResourceID < b.ResourceID
		case SortByActionType:
			return a.Type < b.Type
		default:
			return false
		}
	})
}

// rebuildTable rebuilds the table model with current recommendations.
func (m *RecommendationsViewModel) rebuildTable() {
	availableHeight := m.height - recSummaryHeight - 1
	if availableHeight < minHeight {
		availableHeight = minHeight
	}
	m.table = NewRecommendationsTable(m.recommendations, availableHeight)
}

// View renders the current view.
func (m *RecommendationsViewModel) View() string {
	switch m.state {
	case ViewStateQuitting:
		return ""
	case ViewStateError:
		return fmt.Sprintf("Error: %v\n", m.err)
	case ViewStateLoading:
		return RenderLoading(m.loading)
	case ViewStateDetail:
		if m.selected >= 0 && m.selected < len(m.recommendations) {
			return RenderRecommendationDetail(m.recommendations[m.selected], m.width)
		}
		return "Error: selected index out of bounds"
	case ViewStateList:
		return m.renderListView()
	default:
		return ""
	}
}

func (m *RecommendationsViewModel) renderListView() string {
	summary := RenderRecommendationsSummaryTUI(m.summary, m.width)
	tableView := m.table.View()

	helpText := "\n[/] Filter  [s] Sort  [Enter] Details  [q] Quit"

	if m.showFilter {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			summary,
			tableView,
			"\nFilter: "+m.textInput.View(),
			helpText,
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left, summary, tableView, helpText)
}

// NewRecommendationsTable creates a table model for displaying recommendations.
func NewRecommendationsTable(recs []engine.Recommendation, height int) table.Model {
	columns := []table.Column{
		{Title: "Resource", Width: recColWidthResource},
		{Title: "Action", Width: recColWidthAction},
		{Title: "Savings", Width: recColWidthSavings},
		{Title: "Description", Width: recColWidthDescription},
	}

	rows := make([]table.Row, len(recs))
	for i, rec := range recs {
		savings := fmt.Sprintf("$%.2f", rec.EstimatedSavings)
		desc := rec.Description
		if len(desc) > recDescTruncateLen {
			desc = desc[:recDescTruncateLen] + "..."
		}
		rows[i] = table.Row{rec.ResourceID, rec.Type, savings, desc}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return t
}

// RenderRecommendationsSummaryTUI renders the recommendations summary for TUI display.
func RenderRecommendationsSummaryTUI(summary *RecommendationsSummary, _ int) string {
	if summary == nil {
		return "No recommendations available."
	}

	currency := summary.Currency
	if currency == "" {
		currency = defaultCurrency
	}

	var sb strings.Builder
	sb.WriteString("RECOMMENDATIONS SUMMARY\n")
	sb.WriteString("=======================\n")
	sb.WriteString(fmt.Sprintf("Total: %d recommendations\n", summary.TotalCount))
	sb.WriteString(fmt.Sprintf("Potential Savings: $%.2f %s\n", summary.TotalSavings, currency))

	if len(summary.CountByAction) > 0 {
		sb.WriteString("\nBy Action Type:\n")

		// Sort action types for deterministic output
		actionTypes := make([]string, 0, len(summary.CountByAction))
		for actionType := range summary.CountByAction {
			actionTypes = append(actionTypes, actionType)
		}
		sort.Strings(actionTypes)

		for _, actionType := range actionTypes {
			count := summary.CountByAction[actionType]
			savings := summary.SavingsByAction[actionType]
			sb.WriteString(fmt.Sprintf("  %s: %d ($%.2f)\n", actionType, count, savings))
		}
	}

	return sb.String()
}

// RenderRecommendationDetail renders a detailed view of a single recommendation.
func RenderRecommendationDetail(rec engine.Recommendation, width int) string {
	_ = width // Reserved for future width-aware rendering

	currency := rec.Currency
	if currency == "" {
		currency = defaultCurrency
	}

	var sb strings.Builder
	sb.WriteString("RECOMMENDATION DETAIL\n")
	sb.WriteString("=====================\n\n")
	sb.WriteString(fmt.Sprintf("Resource:    %s\n", rec.ResourceID))
	sb.WriteString(fmt.Sprintf("Action Type: %s\n", rec.Type))
	sb.WriteString(fmt.Sprintf("Savings:     $%.2f %s\n", rec.EstimatedSavings, currency))
	sb.WriteString(fmt.Sprintf("Description: %s\n", rec.Description))
	sb.WriteString("\n[Esc] Back to list  [q] Quit")

	return sb.String()
}
