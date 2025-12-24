# Quickstart: Running the Hardened Test Suite

This guide explains how to use the new testing capabilities: fuzzing,
benchmarks, and cross-platform validation.

## 1. Fuzz Testing

Fuzz tests are located in `internal/ingest` and `internal/spec`.

### Running Fuzz Tests Locally

To run fuzz tests for a short duration (e.g., 30 seconds):

```bash
# Fuzz JSON parser
go test -fuzz=FuzzJSON$ -fuzztime=30s ./internal/ingest

# Fuzz YAML parser
go test -fuzz=FuzzYAML$ -fuzztime=30s ./internal/spec

# Fuzz full plan parsing
go test -fuzz=FuzzPulumiPlanParse$ -fuzztime=30s ./internal/ingest
```

### Analyzing Failures

If a crash is found, the input is saved to `testdata/fuzz`. Re-run the specific failure:

```bash
go test -run=FuzzJSON/cafebabe1234... ./internal/ingest
```

## 2. Performance Benchmarks

Benchmarks are located in `test/benchmarks`.

### Running Benchmarks

To run the full benchmark suite:

```bash
go test -bench=. -benchmem ./test/benchmarks/...
```

### Running Large Scale Tests

To run specifically the 100K resource stress test (takes ~3-5 seconds):

```bash
go test -bench=BenchmarkScale100K -benchtime=1x -benchmem ./test/benchmarks/...
```

### Available Benchmarks

| Benchmark               | Resources | Typical Time |
| ----------------------- | --------- | ------------ |
| `BenchmarkScale1K`      | 1,000     | ~13ms        |
| `BenchmarkScale10K`     | 10,000    | ~167ms       |
| `BenchmarkScale100K`    | 100,000   | ~2.3s        |
| `BenchmarkDeeplyNested` | 1,000     | ~10ms        |

## 3. Cross-Platform Testing

Cross-platform testing is automated in CI.

- **Pull Requests**: Runs on Ubuntu (Linux) only.
- **Nightly/Release**: Runs on Ubuntu, Windows, and macOS.

### Local Verification (if you have the OS)

On Windows (PowerShell):

```powershell
go test ./...
```

On macOS:

```bash
go test ./...
```
