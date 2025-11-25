# Research: Error Aggregation in Proto Adapter

**Date**: 2025-11-24
**Feature**: 003-proto-error-aggregation

## Research Summary

All technical decisions were resolved during the clarification phase. No additional research required.

---

## Decision 1: Structured Logging Library

**Decision**: Use `zerolog` for structured logging

**Rationale**:
- High performance with zero-allocation JSON logging
- Structured fields for error details (resource_type, resource_id, plugin, timestamp)
- Well-maintained and widely adopted in Go ecosystem
- Supports leveled logging for different verbosity needs

**Alternatives Considered**:
- `log.Printf` (standard library): Simpler but lacks structured field support
- `logrus`: More mature but higher allocation overhead than zerolog

---

## Decision 2: PR Size Constraints

**Decision**: Allow ~150-200 lines (expanded from original ~100 lines estimate)

**Rationale**:
- Adding zerolog dependency increases scope
- Error aggregation types and methods require adequate space
- Still qualifies as "small" PR by conventional standards
- Single responsibility maintained (error handling only)

**Alternatives Considered**:
- Strict ~100 lines: Would require cutting features or splitting PR
- Split into 2 PRs: Adds coordination overhead for simple feature

---

## Decision 3: Engine Return Type

**Decision**: Change `Engine.GetProjectedCost` and `Engine.GetActualCost` to return `*CostResultWithErrors`

**Rationale**:
- Propagates error information to CLI layer
- Single implementer means no backward compatibility concerns
- Consistent interface across both methods
- CLI can display both inline errors and summary

**Alternatives Considered**:
- Keep signature, log only: Errors not visible to CLI layer
- Separate `GetErrors()` method: Splits related data, complicates usage

---

## Decision 4: GetActualCost Standardization

**Decision**: Update both engine methods with same return type

**Rationale**:
- Consistency in error handling across projected and actual costs
- Same downstream impact either way
- Users expect same behavior for both cost types

**Alternatives Considered**:
- Only GetProjectedCost: Creates inconsistent API surface

---

## Decision 5: CLI Error Display

**Decision**: Display errors both inline (Notes column) AND as summary after table

**Rationale**:
- Inline errors prevent confusion about $0 costs (shows ERROR: in Notes)
- Summary provides actionable debugging information
- Users see errors in context and in aggregate
- Matches example output in original issue

**Alternatives Considered**:
- Summary only: Users might miss which rows failed
- Inline only: Error details truncated in table column

---

## Dependencies to Add

```go
import "github.com/rs/zerolog"
```

**Installation**: `go get github.com/rs/zerolog`

---

## Best Practices Applied

### zerolog Integration Pattern

```go
import (
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

// In error handling:
log.Error().
    Str("resource_type", err.ResourceType).
    Str("resource_id", err.ResourceID).
    Str("plugin", err.PluginName).
    Time("timestamp", err.Timestamp).
    Err(err.Error).
    Msg("Plugin cost calculation failed")
```

### Error Aggregation Pattern

- Collect errors during iteration, don't fail fast
- Return both results and errors together
- Provide summary method for user-friendly output
- Truncate long error lists (>5 errors)

---

## No Outstanding Research Items

All technical decisions resolved. Ready for Phase 1: Design & Contracts.
