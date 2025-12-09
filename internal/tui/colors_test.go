package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestColorConstants(t *testing.T) {
	tests := []struct {
		name     string
		color    lipgloss.Color
		expected string
	}{
		{"ColorOK", ColorOK, "82"},
		{"ColorWarning", ColorWarning, "208"},
		{"ColorCritical", ColorCritical, "196"},
		{"ColorInfo", ColorInfo, "33"},
		{"ColorHeader", ColorHeader, "99"},
		{"ColorLabel", ColorLabel, "245"},
		{"ColorValue", ColorValue, "255"},
		{"ColorBorder", ColorBorder, "238"},
		{"ColorHighlight", ColorHighlight, "229"},
		{"ColorMuted", ColorMuted, "240"},
		{"ColorPriorityCritical", ColorPriorityCritical, "196"},
		{"ColorPriorityHigh", ColorPriorityHigh, "208"},
		{"ColorPriorityMedium", ColorPriorityMedium, "226"},
		{"ColorPriorityLow", ColorPriorityLow, "82"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.color) != tt.expected {
				t.Errorf("Expected %s to be %s, got %s", tt.name, tt.expected, string(tt.color))
			}
		})
	}
}
