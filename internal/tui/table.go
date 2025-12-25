package tui

import (
	"github.com/charmbracelet/bubbles/table"
)

// DefaultTableStyles returns a configured table.Styles struct using the package's
// standardized styles for headers and selection.
func DefaultTableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = TableHeaderStyle
	s.Selected = TableSelectedStyle
	return s
}

// NewTable creates a new table with the default styles applied.
// This is a helper wrapper around table.New that ensures consistency.
func NewTable(columns []table.Column, rows []table.Row, height int) table.Model {
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height),
	)
	t.SetStyles(DefaultTableStyles())
	return t
}
