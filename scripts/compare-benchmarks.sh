#!/bin/bash
# Compare benchmark results against baseline

BASELINE_FILE="test/benchmarks/baseline.txt"
CURRENT_FILE="test/benchmarks/current.txt"

if [ ! -f "$BASELINE_FILE" ]; then
    echo "No baseline found. Creating..."
    go test -bench=. -benchmem ./test/benchmarks/... > "$BASELINE_FILE"
    echo "Baseline created."
    exit 0
fi

echo "Running benchmarks..."
go test -bench=. -benchmem ./test/benchmarks/... > "$CURRENT_FILE"

echo "Comparing results..."
# Simple comparison using benchstat if available, or diff
if command -v benchstat >/dev/null; then
    benchstat "$BASELINE_FILE" "$CURRENT_FILE"
else
    echo "benchstat not found. Showing raw diff:"
    diff "$BASELINE_FILE" "$CURRENT_FILE"
fi
