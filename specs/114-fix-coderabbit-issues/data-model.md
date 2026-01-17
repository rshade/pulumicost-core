# Data Model Changes

**Feature**: CodeRabbit Issue Resolution

## New Constants

### Package: `internal/constants`

| Constant | Type | Value | Description |
|----------|------|-------|-------------|
| `EnvAnalyzerMode` | `string` | `"FINFOCUS_ANALYZER_MODE"` | Environment variable to enable Analyzer mode. |

## Struct Updates

### Package: `internal/proto`

**Struct**: `FieldMapping`

Updated tags for serialization control:

```go
type FieldMapping struct {
    // ... other fields
    Condition    string `json:"condition,omitempty" yaml:"condition,omitempty"`       // Added omitempty
    ExpectedType string `json:"expectedType,omitempty" yaml:"expectedType,omitempty"` // Added omitempty
}
```

## Enum Stringers

### Package: `internal/pluginhost`

**Type**: `CompatibilityResult`

New `String()` method output:

| Value | String Representation |
|-------|-----------------------|
| `Compatible` | `"Compatible"` |
| `MajorMismatch` | `"MajorMismatch"` |
| `Invalid` | `"Invalid"` |
| *(unknown)* | `"CompatibilityResult(<n>)"` |