# TUI Package Quickstart

**Package**: `internal/tui`
**Audience**: Go developers implementing CLI commands
**Date**: Tue Dec 09 2025

## Installation

The package is part of the internal codebase. Import it in your CLI commands:

```go
import "github.com/pulumi/pulumicost/internal/tui"
```

## Basic Usage

### Output Mode Detection

```go
// Detect appropriate output mode based on environment
mode := tui.DetectOutputMode(forceColor, noColor, plain)

switch mode {
case tui.OutputModePlain:
    // Basic text output, no ANSI colors
case tui.OutputModeStyled:
    // ANSI colors and formatting
case tui.OutputModeInteractive:
    // Full TUI with interactivity (future use)
}
```

### TTY Detection

```go
// Check if running in terminal
if tui.IsTTY() {
    // Safe to use styled output
}

// Get terminal width for responsive layouts
width := tui.TerminalWidth() // Returns 80 as fallback
```

### Status Messages

```go
// Render status indicators
successMsg := tui.RenderStatus("ok")        // "✓ OK" (green)
warningMsg := tui.RenderStatus("warning")   // "⚠ WARNING" (orange)
errorMsg := tui.RenderStatus("critical")    // "🚨 CRITICAL" (red)
```

### Cost Deltas

```go
// Show cost changes
increase := tui.RenderDelta(25.50)   // "+$25.50 ↑" (orange)
decrease := tui.RenderDelta(-10.00)  // "-$10.00 ↓" (green)
noChange := tui.RenderDelta(0)       // "$0.00 →" (gray)
```

### Priorities

```go
// Display priority levels
critical := tui.RenderPriority("CRITICAL")  // "🚨 CRITICAL" (red)
high := tui.RenderPriority("HIGH")          // "⚠ HIGH" (orange)
medium := tui.RenderPriority("MEDIUM")      // "◉ MEDIUM" (yellow)
low := tui.RenderPriority("LOW")            // "✓ LOW" (green)
```

### Progress Bars

```go
// Create default progress bar
pb := tui.DefaultProgressBar()

// Render at 75% completion
progress := pb.Render(75.0)  // "████████████████████████░░░ 75.0%"

// Custom progress bar
customPb := tui.ProgressBar{
    Width:   20,
    Filled:  "■",
    Empty:   "□",
    ShowPct: false,
}
customProgress := customPb.Render(50.0)  // "■■■■■■■■■■□□□□□□□□"
```

### Money Formatting

```go
// Format currency values
price := tui.FormatMoney(1234.56, "USD")     // "$1,234.56 USD"
cost := tui.FormatMoneyShort(99.99)          // "$99.99"

// Format percentages
pct := tui.FormatPercent(85.7)               // "85.7%"
```

### Using Styles

```go
// Apply predefined styles
header := tui.HeaderStyle.Render("Section Title")
label := tui.LabelStyle.Render("Field:")
value := tui.ValueStyle.Render("Data Value")

// Create bordered content
content := tui.BoxStyle.Render("Boxed content with padding")
```

## Integration Patterns

### CLI Command Structure

```go
func runCommand(cmd *cobra.Command, args []string) error {
    // Detect output mode early
    mode := tui.DetectOutputMode(
        cmd.Flags().Changed("force-color"),
        cmd.Flags().Changed("no-color"),
        cmd.Flags().Changed("plain"),
    )

    // Use appropriate rendering based on mode
    switch mode {
    case tui.OutputModePlain:
        // Plain text output
        fmt.Println("Plain text result")
    case tui.OutputModeStyled:
        // Styled output
        styled := tui.OKStyle.Render("✓ Success")
        fmt.Println(styled)
    case tui.OutputModeInteractive:
        // Interactive TUI (future use)
        // tea.NewProgram(model).Run()
    }

    return nil
}
```

### Error Handling

```go
func displayResult(result *CostResult, mode tui.OutputMode) {
    if mode == tui.OutputModePlain {
        fmt.Printf("Total: $%.2f\n", result.Total)
        return
    }

    // Styled output with status
    total := tui.ValueStyle.Render(tui.FormatMoneyShort(result.Total))
    status := tui.RenderStatus("ok")
    fmt.Printf("%s %s\n", status, total)
}
```

## Best Practices

1. **Detect Mode Early**: Call `DetectOutputMode` at command start
2. **Respect User Preferences**: Honor NO_COLOR and explicit flags
3. **Graceful Degradation**: Always provide plain text fallbacks
4. **Consistent Styling**: Use predefined styles and colors
5. **Test Multiple Modes**: Verify output in plain, styled, and interactive modes

## Common Patterns

### Budget Status Display

```go
func renderBudgetStatus(budget, spent float64) string {
    pct := (spent / budget) * 100
    progress := tui.DefaultProgressBar().Render(pct)

    remaining := budget - spent
    delta := tui.RenderDelta(remaining)

    return fmt.Sprintf("Budget: %s\nRemaining: %s",
        progress, delta)
}
```

### Cost Table Row

```go
func renderCostRow(resource, cost string) string {
    resourceStyled := tui.ValueStyle.Render(resource)
    costStyled := tui.OKStyle.Render(tui.FormatMoneyShort(parseFloat(cost)))

    return fmt.Sprintf("%-30s %s", resourceStyled, costStyled)
}
```

## Troubleshooting

- **No colors in output**: Check TTY detection and NO_COLOR environment
- **Wide characters**: Ensure terminal supports Unicode
- **Performance issues**: Profile rendering in tight loops
- **Testing failures**: Mock TTY detection for consistent test output
