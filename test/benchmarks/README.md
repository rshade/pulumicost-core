# Benchmarks

This directory contains performance benchmarks for FinFocus Core.

## Running Benchmarks

```bash
go test -bench=. -benchmem ./test/benchmarks/...
```

## Files

- `engine_bench_test.go`: Benchmarks for engine core logic.
- `parse_bench_test.go`: Benchmarks for JSON parsing.
- `plugin_bench_test.go`: Benchmarks for plugin communication.
- `scale_test.go`: Benchmarks for large-scale scenarios.

## Baseline

A baseline is stored in `baseline.txt`. Use `scripts/compare-benchmarks.sh` to compare current performance against baseline.
