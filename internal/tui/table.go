package tui

import (
	"github.com/charmbracelet/bubbles/table"
)

// DefaultTableStyles returns a configured table.Styles struct using the package's
// DefaultTableStyles returns a table.Styles value with standardized header and selection styles applied.
// It starts from table.DefaultStyles() and sets Header to TableHeaderStyle and Selected to TableSelectedStyle.
func DefaultTableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = TableHeaderStyle
	s.Selected = TableSelectedStyle
	return s
}

// NewTable creates a new table with the default styles applied.
// NewTable creates a table.Model configured with the provided columns, rows, and height.
// The created table is focused and has DefaultTableStyles applied.
// columns is the slice of table.Column defining the table schema, rows is the slice of table.Row providing the initial data, and height is the table's visible height in rows.
// The configured table.Model is returned.
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