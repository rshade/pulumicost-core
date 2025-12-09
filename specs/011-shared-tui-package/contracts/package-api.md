# Package API Contracts

**Package**: `internal/tui`
**Date**: Tue Dec 09 2025

## Overview

The `internal/tui` package provides a stable API for CLI TUI components. All exported functions and types maintain backward compatibility within the same major version.

## Public API

### Types

#### OutputMode

```go
type OutputMode int
```

Enumeration for rendering modes.

**Constants**:

- `OutputModePlain`: Plain text output
- `OutputModeStyled`: ANSI styled output
- `OutputModeInteractive`: Interactive TUI

#### ProgressBar

```go
type ProgressBar struct {
    Width    int
    Filled   string
    Empty    string
    ShowPct  bool
}
```

Configurable progress bar component.

### Functions

#### DetectOutputMode

```go
func DetectOutputMode(forceColor, noColor, plain bool) OutputMode
```

Determines appropriate output mode based on flags and environment.

**Parameters**:

- `forceColor`: Force color output
- `noColor`: Force plain output
- `plain`: Force plain output (alias for noColor)

**Returns**: Appropriate OutputMode

**Contract**:

- Respects explicit flags over environment detection
- NO_COLOR environment variable takes precedence
- Falls back to Interactive mode when TTY detected and no restrictions

#### IsTTY

```go
func IsTTY() bool
```

Returns true if stdout is connected to a terminal.

**Returns**: bool indicating TTY status

#### TerminalWidth

```go
func TerminalWidth() int
```

Returns terminal width or default if unavailable.

**Returns**: Terminal width in characters (minimum 80)

#### DefaultProgressBar

```go
func DefaultProgressBar() ProgressBar
```

Returns progress bar with default settings.

**Returns**: ProgressBar with Width=30, Filled="█", Empty="░", ShowPct=true

#### RenderStatus

```go
func RenderStatus(status string) string
```

Renders a styled status indicator.

**Parameters**:

- `status`: Status string (case-insensitive)

**Returns**: ANSI-styled status string with icon

**Supported Status Values**:

- "ok", "OK", "success" → Green ✓ OK
- "warning", "WARNING" → Orange ⚠ WARNING
- "critical", "CRITICAL", "exceeded", "EXCEEDED" → Red 🚨 CRITICAL
- Other values → Gray ○ {status}

#### RenderDelta

```go
func RenderDelta(delta float64) string
```

Renders a styled cost delta indicator.

**Parameters**:

- `delta`: Numeric delta value

**Returns**: ANSI-styled delta string with icon and sign

**Formatting**:

- Positive: Orange ↑ +$X.XX
- Negative: Green ↓ -$X.XX
- Zero: Gray → $0.00

#### RenderPriority

```go
func RenderPriority(priority string) string
```

Renders a styled priority indicator.

**Parameters**:

- `priority`: Priority string (case-insensitive)

**Returns**: ANSI-styled priority string with icon

**Supported Priorities**:

- "CRITICAL" → Red 🚨 CRITICAL
- "HIGH" → Orange ⚠ HIGH
- "MEDIUM" → Yellow ◉ MEDIUM
- "LOW" → Green ✓ LOW
- Other values → Gray ○ {priority}

#### FormatMoney

```go
func FormatMoney(amount float64, currency string) string
```

Formats monetary value with currency.

**Parameters**:

- `amount`: Numeric amount
- `currency`: Currency code (defaults to "USD")

**Returns**: Formatted string like "$1,234.56 USD"

#### FormatMoneyShort

```go
func FormatMoneyShort(amount float64) string
```

Formats monetary value without currency.

**Parameters**:

- `amount`: Numeric amount

**Returns**: Formatted string like "$1,234.56"

#### FormatPercent

```go
func FormatPercent(value float64) string
```

Formats percentage value.

**Parameters**:

- `value`: Numeric percentage (e.g., 85.5)

**Returns**: Formatted string like "85.0%"

### Methods

#### ProgressBar.Render

```go
func (p ProgressBar) Render(percent float64) string
```

Renders progress bar at specified percentage.

**Parameters**:

- `percent`: Completion percentage (0-100, clamped)

**Returns**: ANSI-styled progress bar string

**Color Coding**:

- < 80%: Green
- 80-99%: Orange
- ≥ 100%: Red

## Constants

### Colors

All color constants use `lipgloss.Color` type with ANSI 256-color codes.

### Icons

String constants for Unicode icons used in rendering.

### Styles

Predefined `lipgloss.Style` variables for consistent styling.

## Error Handling

All functions are designed to be panic-free:

- Invalid inputs are clamped or defaulted
- Terminal detection failures return safe defaults
- Color rendering gracefully degrades

## Thread Safety

All exported functions and methods are safe for concurrent use:

- No global mutable state
- Pure functions where possible
- Immutable style definitions

## Version Compatibility

- **v1.0.0**: Initial release with all specified APIs
- Breaking changes will increment major version
- New features added as minor versions
- Bug fixes as patch versions
