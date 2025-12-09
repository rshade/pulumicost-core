# Data Model & Entity Definitions

**Feature**: Create shared TUI package with Bubble Tea/Lip Gloss components
**Date**: Tue Dec 09 2025

## Overview

The TUI package provides reusable components and utilities for CLI interfaces. No persistent storage or external data models are required - all entities are in-memory representations for UI rendering.

## Core Entities

### OutputMode

Represents the rendering mode for CLI output based on terminal capabilities and environment.

```go
type OutputMode int

const (
    OutputModePlain OutputMode = iota  // No styling, plain text
    OutputModeStyled                   // Lip Gloss styling, no interactivity
    OutputModeInteractive              // Full Bubble Tea TUI
)
```

**Attributes**:

- Plain: Basic text output
- Styled: ANSI colors and formatting
- Interactive: Full TUI with user interaction

**Validation Rules**:

- Must be one of the defined constants
- Determined by TTY detection logic

### ProgressBar

Configurable component for displaying progress with visual indicators.

```go
type ProgressBar struct {
    Width    int     // Bar width in characters (default: 30)
    Filled   string  // Character for filled portion (default: "█")
    Empty    string  // Character for empty portion (default: "░")
    ShowPct  bool    // Whether to show percentage (default: true)
}
```

**Attributes**:

- Width: Visual width of the progress bar
- Filled: Character representing completed progress
- Empty: Character representing remaining progress
- ShowPct: Toggle for percentage display

**Validation Rules**:

- Width must be > 0
- Filled and Empty must be single characters
- Progress percentage clamped to 0-100%

**State Transitions**:

- Render method accepts percentage (0-100)
- Color changes based on percentage thresholds (<80%: OK, >=80%: Warning, >=100%: Critical)

### ColorScheme

Constants defining the color palette for consistent UI styling.

**Status Colors**:

- ColorOK: Green (#5fd700) - Success states
- ColorWarning: Orange (#ff8700) - Warning states
- ColorCritical: Red (#ff0000) - Error/critical states
- ColorInfo: Blue (#0087ff) - Information states

**UI Element Colors**:

- ColorHeader: Purple (#875fff) - Headings and titles
- ColorLabel: Gray (#8a8a8a) - Field labels
- ColorValue: White (#eeeeee) - Data values
- ColorBorder: Dark Gray (#444444) - Borders and separators
- ColorHighlight: Yellow (#ffffaf) - Selected/highlighted items
- ColorMuted: Dim Gray (#585858) - Secondary information

**Priority Colors**:

- ColorPriorityCritical: Red - Critical priority
- ColorPriorityHigh: Orange - High priority
- ColorPriorityMedium: Yellow - Medium priority
- ColorPriorityLow: Green - Low priority

**Validation Rules**:

- All colors use valid ANSI 256-color codes
- Colors chosen for accessibility and readability

### StyleDefinitions

Predefined Lip Gloss styles for consistent application.

**Text Styles**:

- HeaderStyle: Bold, Header color
- LabelStyle: Label color
- ValueStyle: Value color

**Status Styles**:

- OKStyle: OK color, bold
- WarningStyle: Warning color, bold
- CriticalStyle: Critical color, bold

**Container Styles**:

- BoxStyle: Rounded border, border color, padding

**Table Styles**:

- TableHeaderStyle: Bold, header color, bottom border
- TableSelectedStyle: Background and highlight colors

**Validation Rules**:

- All styles use defined color constants
- Styles are immutable once defined

## Relationships

- OutputMode affects all rendering decisions
- ProgressBar uses StyleDefinitions for color coding
- StyleDefinitions depend on ColorScheme constants
- All entities are independent with no inter-entity relationships

## Business Rules

1. **Color Consistency**: All UI elements must use colors from the defined scheme
2. **Fallback Behavior**: Plain mode strips all styling, Styled mode applies colors only
3. **Accessibility**: Color choices consider contrast and readability
4. **Performance**: All rendering operations must complete in <50ms

## Data Flow

1. TTY detection determines OutputMode
2. Components check OutputMode for rendering decisions
3. StyleDefinitions applied based on component state
4. ColorScheme provides consistent palette across all components

## Testing Considerations

- OutputMode testing requires mocking terminal detection
- ProgressBar testing covers all percentage ranges and color transitions
- ColorScheme validation ensures ANSI code correctness
- StyleDefinitions testing verifies Lip Gloss rendering
