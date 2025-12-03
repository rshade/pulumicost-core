# Research: Engine Test Coverage Completion

**Branch**: `001-engine-test-coverage`
**Date**: 2025-12-02
**Purpose**: Document findings for functions requiring test coverage

## Coverage Gap Analysis

Current engine coverage: 78.5% → Target: 80%+

### Functions at 0% Coverage (Priority Analysis)

| Function | Lines | Worth Testing? | Rationale |
|----------|-------|----------------|-----------|
| `tryStoragePricing` | 12 | YES | Size multiplier calculation with dependency chain |
| `getDefaultMonthlyByType` | 12 | YES | String matching logic for resource type classification |
| `parseFloatValue` | 15 | YES | Type conversion with multiple input types (float/int/string) |
| `distributeDailyCosts` | 14 | YES | Date iteration algorithm with map mutation |
| `GetActualCost` | 3 | NO | Simple delegation to GetActualCostWithOptions |
| `getStorageSize` | 17 | MAYBE | Property lookup with multiple key fallbacks |
| `tryFallbackNumericValue` | 10 | NO | Last-resort iteration with minimal logic |

**Decision**: Focus on the YES candidates. Testing simple delegations adds coverage
without value (anti-slop principle).

### Functions at Partial Coverage (<70%)

| Function | Current | Gap Analysis |
|----------|---------|--------------|
| `GetProjectedCostWithErrors` | 41.7% | Error aggregation paths not covered |
| `getActualCostForResource` | 35.0% | Plugin fallback paths |
| `calculateCostsFromSpec` | 50.0% | Spec loading fallback branches |
| `tryExtractCostsFromPricing` | 57.1% | Various pricing field combinations |
| `AggregateResultsInternal` | 60.0% | Edge cases in aggregation |
| `GroupResults` | 70.8% | Grouping edge cases |
| `validateCurrencyConsistency` | 75.0% | Currency mismatch paths |

## Testing Strategy Decisions

### Decision 1: Table-Driven Tests for Type Conversion

**Function**: `parseFloatValue`

**Approach**: Table-driven test with cases:

- float64 input → returns as-is
- int input → converts to float64
- string numeric input → parses successfully
- string non-numeric input → returns false
- nil input → returns false
- other types (bool, struct) → returns false

**Rationale**: Pure function with no side effects; table-driven is idiomatic Go
and matches anti-slop requirement (variations on same behavior).

### Decision 2: Integration-Style Tests for tryStoragePricing

**Function**: `tryStoragePricing`

**Approach**: Test through exposed API where possible, or create focused unit tests:

- Resource with size property and pricePerGBMonth → calculates correctly
- Resource without size property → returns false
- Resource with size but no pricing → returns false
- Size values: 0, 1, 100, 1000 (boundary testing)

**Rationale**: Function has dependencies (getStorageSize, getFloatFromPricing).
Testing at appropriate integration level avoids mocking internals.

### Decision 3: Type Classification Tests

**Function**: `getDefaultMonthlyByType`

**Approach**: Table-driven with resource type strings:

- "aws:rds:Instance" → database cost
- "aws:s3:Bucket" → storage cost
- "aws:ec2:Instance" → compute cost (default)
- Case insensitivity: "AWS:RDS:Instance" → database cost
- Unknown types → compute cost (default)

**Rationale**: String matching logic; table-driven ensures comprehensive coverage
of classification branches.

### Decision 4: Date Distribution Algorithm

**Function**: `distributeDailyCosts`

**Approach**: Focused tests for algorithm correctness:

- 3 daily costs starting Jan 1 → entries for Jan 1, 2, 3
- Monthly grouping → entries grouped by month
- Multi-day span crossing month boundary → correct distribution
- Empty DailyCosts array → no entries added

**Rationale**: Algorithm with date manipulation; correctness is critical for
actual cost reporting.

## Benchmark Extension Strategy

### Current Benchmark Coverage

Existing benchmarks cover:

- Single resource (baseline)
- 10 resources (small batch)
- 100 resources (medium batch)

### Required Extension

| Scale | Purpose | Test Type |
|-------|---------|-----------|
| 1,000 | Enterprise small | Benchmark |
| 10,000 | Enterprise medium | Benchmark |
| 100,000 | Enterprise large | Benchmark |

**Memory Tracking**: Use `-benchmem` flag to track allocations/op.

**Baseline Establishment**: First run establishes baseline; subsequent runs
compare for regressions.

## Alternatives Considered

### Alternative 1: Mock-Heavy Unit Tests

**Rejected Because**: Would require mocking internal functions like getStorageSize,
creating test infrastructure without proportional value. Violates anti-slop
principle of "helpers that don't reduce complexity."

### Alternative 2: Skip Low-Coverage Simple Functions

**Accepted With Limits**: Functions like `GetActualCost` (simple delegation) and
`tryFallbackNumericValue` (last-resort fallback) have minimal logic. Adding tests
would be coverage-chasing slop. These remain at 0% coverage by design.

### Alternative 3: Property-Based Testing

**Deferred**: Could use rapid (Go property testing library) for parseFloatValue
and date calculations. Deferred to future iteration; table-driven tests provide
sufficient coverage for current scope.

## Test File Organization

Following existing conventions:

```text
test/unit/engine/
├── aggregation_test.go    # CrossProviderAggregation tests (exists)
├── engine_test.go         # General engine tests (exists)
├── errors_test.go         # Error type tests (exists)
├── render_test.go         # Rendering tests (exists)
├── pricing_test.go        # NEW: tryStoragePricing, getDefaultMonthlyByType
├── conversion_test.go     # NEW: parseFloatValue, type conversions
└── distribution_test.go   # NEW: distributeDailyCosts, date algorithms

test/benchmarks/
└── engine_bench_test.go   # Extended with 1K, 10K, 100K benchmarks
```

## Quality Gates

Before merging:

1. `go test ./internal/engine/... -coverprofile=coverage.out -race`
2. `go tool cover -func=coverage.out | grep total` shows ≥80%
3. `make lint` passes
4. `make test` passes
5. Each new test has comment explaining its distinct purpose
