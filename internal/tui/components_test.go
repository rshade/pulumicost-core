package tui

import (
	"strings"
	"testing"
)

func TestRenderStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{"OK status", "ok", "✓ OK"},
		{"OK uppercase", "OK", "✓ OK"},
		{"Success status", "success", "✓ OK"},
		{"Warning status", "warning", "⚠ WARNING"},
		{"WARNING uppercase", "WARNING", "⚠ WARNING"},
		{"Critical status", "critical", "🚨 CRITICAL"},
		{"CRITICAL uppercase", "CRITICAL", "🚨 CRITICAL"},
		{"Exceeded status", "exceeded", "🚨 CRITICAL"},
		{"EXCEEDED uppercase", "EXCEEDED", "🚨 CRITICAL"},
		{"Unknown status", "unknown", "○ unknown"},
		{"Empty status", "", "○ "},
		{"Custom status", "processing", "○ processing"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderStatus(tt.status)
			if !strings.Contains(result, tt.expected) {
				t.Errorf(
					"RenderStatus(%q) = %q, expected to contain %q",
					tt.status,
					result,
					tt.expected,
				)
			}
		})
	}
}

func TestRenderDelta(t *testing.T) {
	tests := []struct {
		name     string
		delta    float64
		expected string
	}{
		{"Positive delta", 25.50, "+$25.50 ↑"},
		{"Positive delta with decimals", 10.99, "+$10.99 ↑"},
		{"Negative delta", -15.75, "-$15.75 ↓"},
		{"Zero delta", 0.0, "$0.00 →"},
		{"Zero delta negative", -0.0, "$0.00 →"},
		{"Small positive", 0.01, "+$0.01 ↑"},
		{"Small negative", -0.01, "-$0.01 ↓"},
		{"Large positive", 1234.56, "+$1,234.56 ↑"},
		{"Large negative", -9999.99, "-$9,999.99 ↓"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderDelta(tt.delta)
			if !strings.Contains(result, tt.expected) {
				t.Errorf(
					"RenderDelta(%.2f) = %q, expected to contain %q",
					tt.delta,
					result,
					tt.expected,
				)
			}
		})
	}
}

func TestRenderPriority(t *testing.T) {
	tests := []struct {
		name     string
		priority string
		expected string
	}{
		{"Critical priority", "CRITICAL", "🚨 CRITICAL"},
		{"critical lowercase", "critical", "🚨 CRITICAL"},
		{"High priority", "HIGH", "⚠ HIGH"},
		{"high lowercase", "high", "⚠ HIGH"},
		{"Medium priority", "MEDIUM", "◉ MEDIUM"},
		{"medium lowercase", "medium", "◉ MEDIUM"},
		{"Low priority", "LOW", "✓ LOW"},
		{"low lowercase", "low", "✓ LOW"},
		{"Unknown priority", "urgent", "○ urgent"},
		{"Empty priority", "", "○ "},
		{"Custom priority", "normal", "○ normal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderPriority(tt.priority)
			if !strings.Contains(result, tt.expected) {
				t.Errorf(
					"RenderPriority(%q) = %q, expected to contain %q",
					tt.priority,
					result,
					tt.expected,
				)
			}
		})
	}
}

// T027: Test FormatActionType for TUI action type label rendering.
func TestFormatActionType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Valid action types
		{"RIGHTSIZE", "RIGHTSIZE", "Rightsize"},
		{"TERMINATE", "TERMINATE", "Terminate"},
		{"PURCHASE_COMMITMENT", "PURCHASE_COMMITMENT", "Purchase Commitment"},
		{"ADJUST_REQUESTS", "ADJUST_REQUESTS", "Adjust Requests"},
		{"MODIFY", "MODIFY", "Modify"},
		{"DELETE_UNUSED", "DELETE_UNUSED", "Delete Unused"},
		{"MIGRATE", "MIGRATE", "Migrate"},
		{"CONSOLIDATE", "CONSOLIDATE", "Consolidate"},
		{"SCHEDULE", "SCHEDULE", "Schedule"},
		{"REFACTOR", "REFACTOR", "Refactor"},
		{"OTHER", "OTHER", "Other"},
		// Case insensitivity
		{"lowercase migrate", "migrate", "Migrate"},
		{"mixed case Consolidate", "Consolidate", "Consolidate"},
		{"lowercase other", "other", "Other"},
		// Unknown types (returned as-is for forward compatibility)
		{"unknown type", "UNKNOWN", "UNKNOWN"},
		{"custom type", "CUSTOM_TYPE", "CUSTOM_TYPE"},
		// Edge cases
		{"empty string", "", ""},
		{"with spaces", "  MIGRATE  ", "Migrate"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatActionType(tt.input)
			if result != tt.expected {
				t.Errorf("FormatActionType(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// T030: Test FormatActionType for unknown action types.
func TestFormatActionType_UnknownTypes(t *testing.T) {
	// Unknown types should be returned as-is for forward compatibility
	unknownTypes := []string{
		"UNKNOWN",
		"FUTURE_TYPE",
		"NOT_A_TYPE",
		"NEW_ACTION",
	}

	for _, unknown := range unknownTypes {
		t.Run(unknown, func(t *testing.T) {
			result := FormatActionType(unknown)
			// Unknown types are returned as-is (not transformed)
			if result != unknown {
				t.Errorf("FormatActionType(%q) = %q, expected %q (unchanged)", unknown, result, unknown)
			}
		})
	}
}

func TestRenderFunctions_BasicOutput(t *testing.T) {
	// Test that the functions produce expected output (styling may be disabled in test env)
	tests := []struct {
		name     string
		function func() string
		contains string
	}{
		{"OK status output", func() string { return RenderStatus("ok") }, "✓ OK"},
		{"Warning status output", func() string { return RenderStatus("warning") }, "⚠ WARNING"},
		{"Critical status output", func() string { return RenderStatus("critical") }, "🚨 CRITICAL"},
		{"Positive delta output", func() string { return RenderDelta(10.0) }, "+$10.00 ↑"},
		{"Negative delta output", func() string { return RenderDelta(-10.0) }, "-$10.00 ↓"},
		{"Zero delta output", func() string { return RenderDelta(0) }, "$0.00 →"},
		{
			"Critical priority output",
			func() string { return RenderPriority("CRITICAL") },
			"🚨 CRITICAL",
		},
		{"High priority output", func() string { return RenderPriority("HIGH") }, "⚠ HIGH"},
		{"Medium priority output", func() string { return RenderPriority("MEDIUM") }, "◉ MEDIUM"},
		{"Low priority output", func() string { return RenderPriority("LOW") }, "✓ LOW"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function()

			// Should contain expected text content
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected output to contain %q, got: %q", tt.contains, result)
			}

			// Should not be empty
			if result == "" {
				t.Error("Expected non-empty result")
			}
		})
	}
}
