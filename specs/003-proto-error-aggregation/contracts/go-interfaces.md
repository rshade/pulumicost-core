# Go Interface Contracts: Error Aggregation

**Date**: 2025-11-24
**Feature**: 003-proto-error-aggregation

## New Types Contract

### ErrorDetail

```go
// ErrorDetail captures information about a failed resource cost calculation
type ErrorDetail struct {
    ResourceType string    // e.g., "aws:ec2:Instance"
    ResourceID   string    // e.g., "i-1234567890abcdef0"
    PluginName   string    // e.g., "kubecost", "vantage"
    Error        error     // underlying error
    Timestamp    time.Time // when error occurred
}
```

### CostResultWithErrors

```go
// CostResultWithErrors wraps results and any errors encountered
type CostResultWithErrors struct {
    Results []engine.CostResult
    Errors  []ErrorDetail
}

// HasErrors returns true if any errors were encountered
func (c *CostResultWithErrors) HasErrors() bool

// ErrorSummary returns a human-readable summary of errors
// Truncates after 5 errors with count of remaining
func (c *CostResultWithErrors) ErrorSummary() string
```

---

## Modified Method Signatures

### Adapter (internal/proto/adapter.go)

```go
// GetProjectedCost calculates projected costs for resources
// Returns results for all resources (placeholders for failures) and error details
func (a *Adapter) GetProjectedCost(
    ctx context.Context,
    client *pluginhost.Client,
    resources []engine.ResourceDescriptor,
) *CostResultWithErrors

// GetActualCost retrieves actual historical costs for resources
// Returns results for all resources (placeholders for failures) and error details
func (a *Adapter) GetActualCost(
    ctx context.Context,
    client *pluginhost.Client,
    resources []engine.ResourceDescriptor,
    from, to time.Time,
) *CostResultWithErrors
```

### Engine (internal/engine/engine.go)

```go
// GetProjectedCost calculates projected costs, aggregating across all plugins
// Returns combined results and errors from all plugin calls
func (e *Engine) GetProjectedCost(
    ctx context.Context,
    resources []ResourceDescriptor,
) (*CostResultWithErrors, error)

// GetActualCost retrieves actual costs, aggregating across all plugins
// Returns combined results and errors from all plugin calls
func (e *Engine) GetActualCost(
    ctx context.Context,
    resources []ResourceDescriptor,
    from, to time.Time,
) (*CostResultWithErrors, error)
```

---

## CLI Display Contract

### Table Output

For each resource in Results:
- Display all columns as normal
- Notes column shows "ERROR: {message}" for failed resources

### Error Summary (after table)

```text
⚠️ {N} resource(s) failed:
  - {ResourceType} ({ResourceID}): {error message}
  - {ResourceType} ({ResourceID}): {error message}
  ... and {remaining} more errors
```

---

## Logging Contract

### zerolog Fields

Each error logged with:
```go
log.Error().
    Str("resource_type", err.ResourceType).
    Str("resource_id", err.ResourceID).
    Str("plugin", err.PluginName).
    Time("timestamp", err.Timestamp).
    Err(err.Error).
    Msg("Plugin cost calculation failed")
```

---

## Error Handling Contract

1. **Never fail fast**: Continue processing remaining resources on error
2. **Always return results**: Include placeholder for failed resources
3. **Preserve error context**: Wrap plugin errors with "plugin call failed: %w"
4. **Aggregate across plugins**: Combine errors from multiple plugin calls
5. **Truncate summaries**: Limit user-facing summary to 5 errors
