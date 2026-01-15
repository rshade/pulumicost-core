#!/usr/bin/env bash

# Test script for scripts/check-coverage.sh
# Tests coverage threshold validation logic

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TEST_DIR="$(mktemp -d)"

cleanup() {
    rm -rf "$TEST_DIR"
}
trap cleanup EXIT

# Test 1: Coverage above threshold should pass
test_coverage_above_threshold() {
    echo "Test 1: Coverage above threshold (should pass)"

    # Create mock coverage file with 70% coverage
    cat > "$TEST_DIR/coverage.out" <<EOF
mode: atomic
github.com/rshade/finfocus/internal/cli/root.go:10.1,12.2 1 1
github.com/rshade/finfocus/internal/cli/root.go:14.1,16.2 1 1
github.com/rshade/finfocus/internal/cli/root.go:18.1,20.2 1 1
github.com/rshade/finfocus/internal/cli/root.go:22.1,24.2 1 1
github.com/rshade/finfocus/internal/cli/root.go:26.1,28.2 1 1
github.com/rshade/finfocus/internal/cli/root.go:30.1,32.2 1 1
github.com/rshade/finfocus/internal/cli/root.go:34.1,36.2 1 1
github.com/rshade/finfocus/internal/cli/root.go:38.1,40.2 1 0
github.com/rshade/finfocus/internal/cli/root.go:42.1,44.2 1 0
github.com/rshade/finfocus/internal/cli/root.go:46.1,48.2 1 0
EOF

    # Run check-coverage.sh with 50% threshold (should pass with 70%)
    if [ -f "$PROJECT_ROOT/scripts/check-coverage.sh" ]; then
        if bash "$PROJECT_ROOT/scripts/check-coverage.sh" "$TEST_DIR/coverage.out" 50; then
            echo "✓ Test 1 passed: Coverage check succeeded"
            return 0
        else
            echo "✗ Test 1 failed: Coverage check should have passed"
            return 1
        fi
    else
        echo "⚠ Test 1 skipped: scripts/check-coverage.sh not yet implemented"
        return 0
    fi
}

# Test 2: Coverage below threshold should fail
test_coverage_below_threshold() {
    echo "Test 2: Coverage below threshold (should fail)"

    # Create mock coverage file with 30% coverage
    cat > "$TEST_DIR/coverage.out" <<EOF
mode: atomic
github.com/rshade/finfocus/internal/cli/root.go:10.1,12.2 1 1
github.com/rshade/finfocus/internal/cli/root.go:14.1,16.2 1 1
github.com/rshade/finfocus/internal/cli/root.go:18.1,20.2 1 1
github.com/rshade/finfocus/internal/cli/root.go:22.1,24.2 1 0
github.com/rshade/finfocus/internal/cli/root.go:26.1,28.2 1 0
github.com/rshade/finfocus/internal/cli/root.go:30.1,32.2 1 0
github.com/rshade/finfocus/internal/cli/root.go:34.1,36.2 1 0
github.com/rshade/finfocus/internal/cli/root.go:38.1,40.2 1 0
github.com/rshade/finfocus/internal/cli/root.go:42.1,44.2 1 0
github.com/rshade/finfocus/internal/cli/root.go:46.1,48.2 1 0
EOF

    # Run check-coverage.sh with 50% threshold (should fail with 30%)
    if [ -f "$PROJECT_ROOT/scripts/check-coverage.sh" ]; then
        if bash "$PROJECT_ROOT/scripts/check-coverage.sh" "$TEST_DIR/coverage.out" 50; then
            echo "✗ Test 2 failed: Coverage check should have failed"
            return 1
        else
            echo "✓ Test 2 passed: Coverage check correctly failed"
            return 0
        fi
    else
        echo "⚠ Test 2 skipped: scripts/check-coverage.sh not yet implemented"
        return 0
    fi
}

# Test 3: Missing coverage file should fail gracefully
test_missing_coverage_file() {
    echo "Test 3: Missing coverage file (should fail gracefully)"

    # Run check-coverage.sh with non-existent file
    if [ -f "$PROJECT_ROOT/scripts/check-coverage.sh" ]; then
        if bash "$PROJECT_ROOT/scripts/check-coverage.sh" "$TEST_DIR/nonexistent.out" 50 2>/dev/null; then
            echo "✗ Test 3 failed: Should fail with missing file"
            return 1
        else
            echo "✓ Test 3 passed: Correctly failed with missing file"
            return 0
        fi
    else
        echo "⚠ Test 3 skipped: scripts/check-coverage.sh not yet implemented"
        return 0
    fi
}

# Test 4: Empty coverage file should fail
test_empty_coverage_file() {
    echo "Test 4: Empty coverage file (should fail)"

    # Create empty coverage file
    touch "$TEST_DIR/empty.out"

    # Run check-coverage.sh with empty file
    if [ -f "$PROJECT_ROOT/scripts/check-coverage.sh" ]; then
        if bash "$PROJECT_ROOT/scripts/check-coverage.sh" "$TEST_DIR/empty.out" 50 2>/dev/null; then
            echo "✗ Test 4 failed: Should fail with empty file"
            return 1
        else
            echo "✓ Test 4 passed: Correctly failed with empty file"
            return 0
        fi
    else
        echo "⚠ Test 4 skipped: scripts/check-coverage.sh not yet implemented"
        return 0
    fi
}

# Test 5: Invalid threshold should fail
test_invalid_threshold() {
    echo "Test 5: Invalid threshold (should fail)"

    # Create mock coverage file
    cat > "$TEST_DIR/coverage.out" <<EOF
mode: atomic
github.com/rshade/finfocus/internal/cli/root.go:10.1,12.2 1 1
EOF

    # Run check-coverage.sh with invalid threshold
    if [ -f "$PROJECT_ROOT/scripts/check-coverage.sh" ]; then
        if bash "$PROJECT_ROOT/scripts/check-coverage.sh" "$TEST_DIR/coverage.out" abc 2>/dev/null; then
            echo "✗ Test 5 failed: Should fail with invalid threshold"
            return 1
        else
            echo "✓ Test 5 passed: Correctly failed with invalid threshold"
            return 0
        fi
    else
        echo "⚠ Test 5 skipped: scripts/check-coverage.sh not yet implemented"
        return 0
    fi
}

# Run all tests
echo "================================"
echo "Testing scripts/check-coverage.sh"
echo "================================"
echo ""

FAILED=0

test_coverage_above_threshold || FAILED=$((FAILED + 1))
echo ""
test_coverage_below_threshold || FAILED=$((FAILED + 1))
echo ""
test_missing_coverage_file || FAILED=$((FAILED + 1))
echo ""
test_empty_coverage_file || FAILED=$((FAILED + 1))
echo ""
test_invalid_threshold || FAILED=$((FAILED + 1))
echo ""

echo "================================"
if [ $FAILED -eq 0 ]; then
    echo "✓ All tests passed"
    exit 0
else
    echo "✗ $FAILED test(s) failed"
    exit 1
fi
