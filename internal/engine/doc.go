// Package engine provides the core cost calculation logic for FinFocus.
//
// The engine orchestrates between plugins and local pricing specifications to
// calculate both projected and actual infrastructure costs.
//
// # Projected Cost Calculation
//
// For projected costs, the engine:
//  1. Validates resource descriptors
//  2. Queries registered plugins for cost estimates
//  3. Falls back to local YAML specs when plugins don't provide pricing
//  4. Aggregates results into a unified cost breakdown
//
// # Actual Cost Pipeline
//
// For actual costs, the engine supports:
//   - Time range queries with start/end dates
//   - Tag-based filtering using "tag:key=value" syntax
//   - Grouping by resource, type, provider, or date
//   - Cross-provider cost aggregation with currency validation
//
// # Output Formats
//
// Results can be rendered in three formats:
//   - table: Human-readable tabular output
//   - json: Structured JSON for programmatic use
//   - ndjson: Newline-delimited JSON for streaming
//
// # Timeouts
//
// Operations are protected by context timeouts:
//   - 60s overall query timeout
//   - 30s per-plugin call timeout
//   - 5s per-resource calculation timeout
package engine
