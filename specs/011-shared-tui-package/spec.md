# Feature Specification: Create shared TUI package with Bubble Tea and Lip Gloss components

**Feature Branch**: `011-shared-tui-package`  
**Created**: Tue Dec 09 2025  
**Status**: Draft  
**Input**: User description: "title: feat(tui): Create shared TUI package with Bubble Tea/Lip Gloss components
state: OPEN
author: rshade
labels:
comments: 0
assignees:
projects:
milestone:
number: 222
--

## Summary

Create a shared internal TUI package that provides common Bubble Tea and Lip Gloss components, styles, and utilities for use across all CLI commands. This is the foundational package that other TUI-related features will depend on.

## Priority

**P0 - Highest Priority** - This is a prerequisite for all other Bubble Tea and Lip Gloss features.

## Motivation

Multiple CLI commands will use Bubble Tea and Lip Gloss:

- Budget status display (#217)
- Recommendations browsing (#216)
- Cost actual/projected/estimate (#218)

A shared package ensures:

- Consistent visual styling across all commands
- DRY principle - no duplicated code
- Single source of truth for color schemes
- Reusable components (progress bars, tables, spinners)
- Centralized TTY detection and fallback logic

## Technical Specification

### Package Structure

```text
internal/tui/
├── styles.go        # Lip Gloss style definitions
├── colors.go        # Color scheme constants
├── components.go    # Reusable UI components
├── progress.go      # Progress bar rendering
├── table.go         # Table configuration helpers
├── spinner.go       # Loading spinner presets
├── detect.go        # TTY detection utilities
├── render.go        # Common rendering utilities
└── tui_test.go      # Unit tests
```

### Color Scheme (colors.go)

```go
package tui

import "github.com/charmbracelet/lipgloss"

// Status colors
const (
    ColorOK       = lipgloss.Color("82")   // #5fd700 - Green
    ColorWarning  = lipgloss.Color("208")  // #ff8700 - Orange
    ColorCritical = lipgloss.Color("196")  // #ff0000 - Red
    ColorInfo     = lipgloss.Color("33")   // #0087ff - Blue
)

// UI element colors
const (
    ColorHeader     = lipgloss.Color("99")   // #875fff - Purple
    ColorLabel      = lipgloss.Color("245")  // #8a8a8a - Gray
    ColorValue      = lipgloss.Color("255")  // #eeeeee - White
    ColorBorder     = lipgloss.Color("238")  // #444444 - Dark gray
    ColorHighlight  = lipgloss.Color("229")  // #ffffaf - Yellow
    ColorMuted      = lipgloss.Color("240")  // #585858 - Dim gray
)

// Priority colors
const (
    ColorPriorityCritical = ColorCritical
    ColorPriorityHigh     = ColorWarning
    ColorPriorityMedium   = lipgloss.Color("226")  // #ffff00 - Yellow
    ColorPriorityLow      = ColorOK
)
```

### Shared Styles (styles.go)

```go
package tui

import "github.com/charmbracelet/lipgloss"

var (
    // Text styles
    HeaderStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(ColorHeader)

    LabelStyle = lipgloss.NewStyle().
        Foreground(ColorLabel)

    ValueStyle = lipgloss.NewStyle().
        Foreground(ColorValue)

    // Status styles
    OKStyle = lipgloss.NewStyle().
        Foreground(ColorOK).
        Bold(true)

    WarningStyle = lipgloss.NewStyle().
        Foreground(ColorWarning).
        Bold(true)

    CriticalStyle = lipgloss.NewStyle().
        Foreground(ColorCritical).
        Bold(true)

    // Container styles
    BoxStyle = lipgloss.NewStyle().
        BorderStyle(lipgloss.RoundedBorder()).
        BorderForeground(ColorBorder).
        Padding(0, 1)

    // Table styles
    TableHeaderStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(ColorHeader).
        BorderStyle(lipgloss.NormalBorder()).
        BorderBottom(true)

    TableSelectedStyle = lipgloss.NewStyle().
        Background(lipgloss.Color("237")).
        Foreground(ColorHighlight)
)
```

### TTY Detection (detect.go)

```go
package tui

import (
    "os"
    "golang.org/x/term"
)

// OutputMode represents the rendering mode for CLI output
type OutputMode int

const (
    OutputModePlain OutputMode = iota  // No styling, plain text
    OutputModeStyled                   // Lip Gloss styling, no interactivity
    OutputModeInteractive              // Full Bubble Tea TUI
)

// DetectOutputMode determines the appropriate output mode based on
// terminal capabilities, environment variables, and flags.
func DetectOutputMode(forceColor, noColor, plain bool) OutputMode {
    // Explicit flags take precedence
    if plain || noColor {
        return OutputModePlain
    }

    // Check NO_COLOR environment variable (standard)
    if os.Getenv("NO_COLOR") != "" {
        return OutputModePlain
    }

    // Check if stdout is a TTY
    if !term.IsTerminal(int(os.Stdout.Fd())) {
        return OutputModePlain
    }

    // Check TERM environment
    if os.Getenv("TERM") == "dumb" {
        return OutputModePlain
    }

    // Check for CI environments
    if os.Getenv("CI") != "" {
        return OutputModeStyled  // Styled but not interactive
    }

    return OutputModeInteractive
}

// IsTTY returns true if stdout is connected to a terminal
func IsTTY() bool {
    return term.IsTerminal(int(os.Stdout.Fd()))
}

// TerminalWidth returns the terminal width or a default if unavailable
func TerminalWidth() int {
    width, _, err := term.GetSize(int(os.Stdout.Fd()))
    if err != nil || width <= 0 {
        return 80  // Default width
    }
    return width
}
```

### Progress Bar Component (progress.go)

```go
package tui

import (
    "fmt"
    "strings"
)

// ProgressBar renders a text-based progress bar
type ProgressBar struct {
    Width    int
    Filled   string
    Empty    string
    ShowPct  bool
}

// DefaultProgressBar returns a progress bar with default settings
func DefaultProgressBar() ProgressBar {
    return ProgressBar{
        Width:   30,
        Filled:  "█",
        Empty:   "░",
        ShowPct: true,
    }
}

// Render returns a styled progress bar string
func (p ProgressBar) Render(percent float64) string {
    if percent < 0 {
        percent = 0
    }
    if percent > 100 {
        percent = 100
    }

    filled := int(percent / 100 * float64(p.Width))
    bar := strings.Repeat(p.Filled, filled) +
           strings.Repeat(p.Empty, p.Width-filled)

    // Apply color based on percentage
    var style lipgloss.Style
    switch {
    case percent >= 100:
        style = CriticalStyle
    case percent >= 80:
        style = WarningStyle
    default:
        style = OKStyle
    }

    result := style.Render(bar)
    if p.ShowPct {
        result += fmt.Sprintf(" %.0f%%", percent)
    }
    return result
}
```

### Status Icons (components.go)

```go
package tui

// Status icons for different states
const (
    IconOK       = "✓"
    IconWarning  = "⚠"
    IconCritical = "🚨"
    IconPending  = "○"
    IconProgress = "◉"
    IconArrowUp  = "↑"
    IconArrowDown = "↓"
    IconArrowRight = "→"
)

// RenderStatus returns a styled status indicator
func RenderStatus(status string) string {
    switch status {
    case "ok", "OK", "success":
        return OKStyle.Render(IconOK + " OK")
    case "warning", "WARNING":
        return WarningStyle.Render(IconWarning + " WARNING")
    case "critical", "CRITICAL", "exceeded", "EXCEEDED":
        return CriticalStyle.Render(IconCritical + " CRITICAL")
    default:
        return LabelStyle.Render(IconPending + " " + status)
    }
}

// RenderDelta returns a styled cost delta indicator
func RenderDelta(delta float64) string {
    switch {
    case delta > 0:
        return WarningStyle.Render(fmt.Sprintf("+$%.2f %s", delta, IconArrowUp))
    case delta < 0:
        return OKStyle.Render(fmt.Sprintf("-$%.2f %s", -delta, IconArrowDown))
    default:
        return LabelStyle.Render(fmt.Sprintf("$0.00 %s", IconArrowRight))
    }
}

// RenderPriority returns a styled priority indicator
func RenderPriority(priority string) string {
    switch strings.ToUpper(priority) {
    case "CRITICAL":
        return CriticalStyle.Render(IconCritical + " CRITICAL")
    case "HIGH":
        return WarningStyle.Render(IconWarning + " HIGH")
    case "MEDIUM":
        return lipgloss.NewStyle().
            Foreground(ColorPriorityMedium).
            Render(IconProgress + " MEDIUM")
    case "LOW":
        return OKStyle.Render(IconOK + " LOW")
    default:
        return LabelStyle.Render(IconPending + " " + priority)
    }
}
```

### Money Formatting (render.go)

```go
package tui

import (
    "fmt"
    "golang.org/x/text/language"
    "golang.org/x/text/message"
)

// FormatMoney formats a monetary value with proper separators
func FormatMoney(amount float64, currency string) string {
    p := message.NewPrinter(language.English)
    if currency == "" {
        currency = "USD"
    }
    return p.Sprintf("$%.2f %s", amount, currency)
}

// FormatMoneyShort formats money without currency code
func FormatMoneyShort(amount float64) string {
    p := message.NewPrinter(language.English)
    return p.Sprintf("$%.2f", amount)
}

// FormatPercent formats a percentage value
func FormatPercent(value float64) string {
    return fmt.Sprintf("%.0f%%", value)
}
```

## Acceptance Criteria

- [ ] Package structure created at `internal/tui/`
- [ ] Color scheme constants defined
- [ ] Lip Gloss styles exported
- [ ] TTY detection with NO_COLOR support
- [ ] Progress bar component
- [ ] Status/priority rendering helpers
- [ ] Money formatting utilities
- [ ] Unit tests for all components
- [ ] Example usage documentation in code comments

## Dependencies

```go
require (
    github.com/charmbracelet/bubbletea v1.2.4
    github.com/charmbracelet/bubbles v0.20.0
    github.com/charmbracelet/lipgloss v1.0.0
    golang.org/x/term v0.27.0
    golang.org/x/text v0.21.0
)
```

## Testing Strategy

```go
func TestProgressBar(t *testing.T) {
    tests := []struct {
        percent float64
        wantLen int
    }{
        {0, 30},
        {50, 30},
        {100, 30},
        {150, 30},  // Should cap at 100
    }
    // ...
}

func TestDetectOutputMode(t *testing.T) {
    // Test NO_COLOR, CI, TTY scenarios
}

func TestRenderStatus(t *testing.T) {
    // Test all status types
}
```

## Blocked Issues

The following issues depend on this package:

- #216 - Recommendations CLI
- #217 - Budget alerts
- #218 - Cost commands upgrade
- #219 - Exit codes (uses styles for output)
- #220 - Notifications (uses styles)
- #221 - Flexible scoping (uses styles)

## Related Issues

- #216, #217, #218, #219, #220, #221 - All TUI consumers"

## User Scenarios & Testing _(mandatory)_

### User Story 1 - Consistent CLI Styling (Priority: P1)

As a CLI user running cost analysis commands, I want all commands to have consistent visual styling so that the interface feels cohesive and professional.

**Why this priority**: This is the foundational requirement that enables all other TUI features. Without consistent styling, users will experience a fragmented interface.

**Independent Test**: Can be fully tested by running multiple CLI commands and verifying visual consistency in styling, colors, and layout.

**Acceptance Scenarios**:

1. **Given** a user runs multiple CLI commands (budget status, recommendations, cost estimates), **When** they view the output, **Then** all commands use the same color scheme and styling patterns with 95% visual consistency (matching colors, fonts, and layout structures)
2. **Given** a user has NO_COLOR environment variable set, **When** they run any CLI command, **Then** output is plain text without styling
3. **Given** a user runs commands in a CI environment, **When** they view output, **Then** styling is applied but no interactive elements are present

---

### User Story 2 - Reusable UI Components (Priority: P2)

As a developer implementing new CLI features, I want access to reusable TUI components so that I can quickly build consistent interfaces without duplicating code.

**Why this priority**: Enables faster development of new features and ensures consistency across the codebase.

**Independent Test**: Can be tested by importing the TUI package and using components like progress bars and status indicators in new code.

**Acceptance Scenarios**:

1. **Given** a developer needs a progress bar, **When** they use the ProgressBar component, **Then** it renders correctly with configurable width and percentage display
2. **Given** a developer needs to show status, **When** they call RenderStatus function, **Then** it returns appropriately styled text with icons
3. **Given** a developer needs money formatting, **When** they use FormatMoney functions, **Then** values are formatted with proper currency symbols and separators

---

### User Story 3 - TTY Detection and Fallbacks (Priority: P3)

As a user running CLI commands in various environments (TTY, pipes, CI), I want the output to automatically adapt so that I always get appropriate formatting.

**Why this priority**: Ensures the tool works reliably across different usage contexts and environments.

**Independent Test**: Can be tested by running commands with different terminal configurations and environment variables.

**Acceptance Scenarios**:

1. **Given** a user pipes command output to another program, **When** the command runs, **Then** output is plain text without ANSI codes
2. **Given** a user runs command in CI environment, **When** they view output, **Then** styled output is shown but no interactive prompts
3. **Given** a user has TERM=dumb set, **When** they run command, **Then** plain text output is used

---

### Edge Cases

- What happens when terminal width is very narrow (< 40 characters)? System should truncate progress bars to fit available width and wrap text appropriately without breaking ANSI color codes
- How does system handle when Lip Gloss dependencies are not available?
- What happens when color codes are invalid or unsupported? System should validate ANSI color codes at initialization and fall back to safe defaults (grayscale) if unsupported

## Requirements _(mandatory)_

### Functional Requirements

- **FR-001**: System MUST provide a shared TUI package at `internal/tui/` with all specified files
- **FR-002**: System MUST define color scheme constants for status, UI elements, and priorities
- **FR-003**: System MUST export Lip Gloss style definitions for text, status, containers, and tables
- **FR-004**: System MUST implement TTY detection with support for NO_COLOR, CI, and TERM environment variables
- **FR-005**: System MUST provide a configurable progress bar component with color coding
- **FR-006**: System MUST provide status and priority rendering helpers with icons
- **FR-007**: System MUST provide money and percentage formatting utilities
- **FR-008**: System MUST include unit tests for all exported functions and types with 80% minimum coverage (95% for critical paths)
- **FR-009**: System MUST include example usage documentation in code comments

### Key Entities _(include if feature involves data)_

- **ColorScheme**: Defines color constants for different UI elements and states
- **ProgressBar**: Configurable component for displaying progress with styling
- **OutputMode**: Enumeration for different rendering modes (plain, styled, interactive)
- **StyleDefinitions**: Predefined Lip Gloss styles for consistent application

## Success Criteria _(mandatory)_

### Measurable Outcomes

- **SC-001**: All specified package files are created and compile successfully
- **SC-002**: Unit test coverage achieves minimum 80% overall (95% for critical paths) for all TUI components
- **SC-003**: TTY detection correctly identifies output modes in different environments
- **SC-004**: All acceptance criteria from user stories are met and testable
- **SC-005**: Package can be imported and used by other CLI commands without issues
