# Quickstart: Engine Test Coverage

**Branch**: `001-engine-test-coverage`

## Running Tests

### Full Test Suite

```bash
# Run all engine tests with coverage
go test ./internal/engine/... -coverprofile=coverage.out -race

# View coverage summary
go tool cover -func=coverage.out | grep total

# View detailed coverage report in browser
go tool cover -html=coverage.out
```

### Specific Test Categories

```bash
# Unit tests only (fast)
go test ./test/unit/engine/... -v

# Edge case tests
go test ./test/unit/engine/... -run TestEdgeCase -v

# Error type tests
go test ./test/unit/engine/... -run TestErr -v

# Cross-provider aggregation tests
go test ./test/unit/engine/... -run TestCrossProvider -v
```

### Benchmarks

```bash
# Run all engine benchmarks
go test ./test/benchmarks/... -bench=. -benchmem

# Run specific scale benchmarks
go test ./test/benchmarks/... -bench=1K -benchmem
go test ./test/benchmarks/... -bench=10K -benchmem
go test ./test/benchmarks/... -bench=100K -benchmem

# Compare benchmarks over time
go test ./test/benchmarks/... -bench=. -benchmem -count=5 > baseline.txt
# After changes:
go test ./test/benchmarks/... -bench=. -benchmem -count=5 > current.txt
benchstat baseline.txt current.txt
```

## Validation Checklist

Before claiming completion:

```bash
# 1. Run full test suite with race detection
go test ./internal/engine/... -race

# 2. Verify coverage threshold
go tool cover -func=coverage.out | grep total
# Must show ≥80%

# 3. Run linting
make lint

# 4. Run all project tests
make test
```

## Adding New Tests

### Test File Naming

- `*_test.go` suffix required by Go
- Place in `test/unit/engine/` for unit tests
- Place in `test/benchmarks/` for performance tests

### Test Quality Requirements

Each test MUST have:

```go
// TestParseFloatValue_FloatInput tests that float64 inputs return unchanged.
func TestParseFloatValue_FloatInput(t *testing.T) {
    // Test implementation
}
```

Avoid:

```go
// Bad: Generic name, no purpose comment
func TestParseFloatValue_Works(t *testing.T) {
    // ...
}

// Bad: Redundant test cases
func TestParseFloatValue_Works2(t *testing.T) {
    // Same thing as above...
}
```

### Table-Driven Test Pattern

```go
func TestParseFloatValue(t *testing.T) {
    tests := []struct {
        name     string
        input    interface{}
        expected float64
        wantOK   bool
    }{
        {"float64 input", 42.5, 42.5, true},
        {"int input", 42, 42.0, true},
        {"string numeric", "42.5", 42.5, true},
        {"string non-numeric", "hello", 0, false},
        {"nil input", nil, 0, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, ok := parseFloatValue(tt.input)
            assert.Equal(t, tt.wantOK, ok)
            if ok {
                assert.Equal(t, tt.expected, got)
            }
        })
    }
}
```

## Coverage Analysis

### View Per-Function Coverage

```bash
go tool cover -func=coverage.out | grep -v "100.0%"
```

This shows only functions needing more coverage.

### Focus Areas

From research.md, these functions are prioritized:

1. `tryStoragePricing` - Storage cost calculation
2. `getDefaultMonthlyByType` - Type classification
3. `parseFloatValue` - Type conversion
4. `distributeDailyCosts` - Date distribution algorithm
