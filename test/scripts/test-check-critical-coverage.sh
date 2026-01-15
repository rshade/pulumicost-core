#!/usr/bin/env bash

# Test script for scripts/check-critical-coverage.sh
# Tests critical path coverage validation for specific packages

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TEST_DIR="$(mktemp -d)"

cleanup() {
    rm -rf "$TEST_DIR"
}
trap cleanup EXIT

# Test 1: All critical paths above threshold should pass
test_all_critical_paths_pass() {
    echo "Test 1: All critical paths above threshold (should pass)"

    # Create mock coverage file with high coverage for critical packages
    cat > "$TEST_DIR/coverage.out" <<EOF
mode: atomic
github.com/rshade/finfocus/internal/cli/root.go:10.1,12.2 1 1
github.com/rshade/finfocus/internal/cli/root.go:14.1,16.2 1 1
github.com/rshade/finfocus/internal/cli/root.go:18.1,20.2 1 1
github.com/rshade/finfocus/internal/cli/root.go:22.1,24.2 1 1
github.com/rshade/finfocus/internal/cli/root.go:26.1,28.2 1 1
github.com/rshade/finfocus/internal/cli/root.go:30.1,32.2 1 1
github.com/rshade/finfocus/internal/cli/root.go:34.1,36.2 1 1
github.com/rshade/finfocus/internal/cli/root.go:38.1,40.2 1 1
github.com/rshade/finfocus/internal/cli/root.go:42.1,44.2 1 1
github.com/rshade/finfocus/internal/cli/root.go:46.1,48.2 1 1
github.com/rshade/finfocus/internal/engine/engine.go:10.1,12.2 1 1
github.com/rshade/finfocus/internal/engine/engine.go:14.1,16.2 1 1
github.com/rshade/finfocus/internal/engine/engine.go:18.1,20.2 1 1
github.com/rshade/finfocus/internal/engine/engine.go:22.1,24.2 1 1
github.com/rshade/finfocus/internal/engine/engine.go:26.1,28.2 1 1
github.com/rshade/finfocus/internal/engine/engine.go:30.1,32.2 1 1
github.com/rshade/finfocus/internal/engine/engine.go:34.1,36.2 1 1
github.com/rshade/finfocus/internal/engine/engine.go:38.1,40.2 1 1
github.com/rshade/finfocus/internal/engine/engine.go:42.1,44.2 1 1
github.com/rshade/finfocus/internal/engine/engine.go:46.1,48.2 1 1
github.com/rshade/finfocus/internal/pluginhost/host.go:10.1,12.2 1 1
github.com/rshade/finfocus/internal/pluginhost/host.go:14.1,16.2 1 1
github.com/rshade/finfocus/internal/pluginhost/host.go:18.1,20.2 1 1
github.com/rshade/finfocus/internal/pluginhost/host.go:22.1,24.2 1 1
github.com/rshade/finfocus/internal/pluginhost/host.go:26.1,28.2 1 1
github.com/rshade/finfocus/internal/pluginhost/host.go:30.1,32.2 1 1
github.com/rshade/finfocus/internal/pluginhost/host.go:34.1,36.2 1 1
github.com/rshade/finfocus/internal/pluginhost/host.go:38.1,40.2 1 1
github.com/rshade/finfocus/internal/pluginhost/host.go:42.1,44.2 1 1
github.com/rshade/finfocus/internal/pluginhost/host.go:46.1,48.2 1 1
EOF

    # Run check-critical-coverage.sh with 50% threshold (all packages have 100% coverage)
    if [ -f "$PROJECT_ROOT/scripts/check-critical-coverage.sh" ]; then
        if bash "$PROJECT_ROOT/scripts/check-critical-coverage.sh" "$TEST_DIR/coverage.out" 50; then
            echo "✓ Test 1 passed: Critical path coverage check succeeded"
            return 0
        else
            echo "✗ Test 1 failed: Critical path coverage check should have passed"
            return 1
        fi
    else
        echo "⚠ Test 1 skipped: scripts/check-critical-coverage.sh not yet implemented"
        return 0
    fi
}

# Test 2: CLI package below threshold should fail
test_cli_below_threshold() {
    echo "Test 2: CLI package below threshold (should fail)"

    # Create mock coverage file with low CLI coverage
    cat > "$TEST_DIR/coverage.out" <<EOF
mode: atomic
github.com/rshade/finfocus/internal/cli/root.go:10.1,12.2 1 1
github.com/rshade/finfocus/internal/cli/root.go:14.1,16.2 1 1
github.com/rshade/finfocus/internal/cli/root.go:18.1,20.2 1 0
github.com/rshade/finfocus/internal/cli/root.go:22.1,24.2 1 0
github.com/rshade/finfocus/internal/cli/root.go:26.1,28.2 1 0
github.com/rshade/finfocus/internal/cli/root.go:30.1,32.2 1 0
github.com/rshade/finfocus/internal/cli/root.go:34.1,36.2 1 0
github.com/rshade/finfocus/internal/cli/root.go:38.1,40.2 1 0
github.com/rshade/finfocus/internal/cli/root.go:42.1,44.2 1 0
github.com/rshade/finfocus/internal/cli/root.go:46.1,48.2 1 0
github.com/rshade/finfocus/internal/engine/engine.go:10.1,12.2 1 1
github.com/rshade/finfocus/internal/engine/engine.go:14.1,16.2 1 1
github.com/rshade/finfocus/internal/pluginhost/host.go:10.1,12.2 1 1
github.com/rshade/finfocus/internal/pluginhost/host.go:14.1,16.2 1 1
EOF

    # Run check-critical-coverage.sh with 50% threshold (CLI has 20% coverage)
    if [ -f "$PROJECT_ROOT/scripts/check-critical-coverage.sh" ]; then
        if bash "$PROJECT_ROOT/scripts/check-critical-coverage.sh" "$TEST_DIR/coverage.out" 50; then
            echo "✗ Test 2 failed: Should fail when CLI coverage is low"
            return 1
        else
            echo "✓ Test 2 passed: Correctly failed when CLI coverage below threshold"
            return 0
        fi
    else
        echo "⚠ Test 2 skipped: scripts/check-critical-coverage.sh not yet implemented"
        return 0
    fi
}

# Test 3: Engine package below threshold should fail
test_engine_below_threshold() {
    echo "Test 3: Engine package below threshold (should fail)"

    # Create mock coverage file with low engine coverage
    cat > "$TEST_DIR/coverage.out" <<EOF
mode: atomic
github.com/rshade/finfocus/internal/cli/root.go:10.1,12.2 1 1
github.com/rshade/finfocus/internal/cli/root.go:14.1,16.2 1 1
github.com/rshade/finfocus/internal/engine/engine.go:10.1,12.2 1 1
github.com/rshade/finfocus/internal/engine/engine.go:14.1,16.2 1 0
github.com/rshade/finfocus/internal/engine/engine.go:18.1,20.2 1 0
github.com/rshade/finfocus/internal/engine/engine.go:22.1,24.2 1 0
github.com/rshade/finfocus/internal/engine/engine.go:26.1,28.2 1 0
github.com/rshade/finfocus/internal/pluginhost/host.go:10.1,12.2 1 1
github.com/rshade/finfocus/internal/pluginhost/host.go:14.1,16.2 1 1
EOF

    # Run check-critical-coverage.sh with 50% threshold (engine has 20% coverage)
    if [ -f "$PROJECT_ROOT/scripts/check-critical-coverage.sh" ]; then
        if bash "$PROJECT_ROOT/scripts/check-critical-coverage.sh" "$TEST_DIR/coverage.out" 50; then
            echo "✗ Test 3 failed: Should fail when engine coverage is low"
            return 1
        else
            echo "✓ Test 3 passed: Correctly failed when engine coverage below threshold"
            return 0
        fi
    else
        echo "⚠ Test 3 skipped: scripts/check-critical-coverage.sh not yet implemented"
        return 0
    fi
}

# Test 4: PluginHost package below threshold should fail
test_pluginhost_below_threshold() {
    echo "Test 4: PluginHost package below threshold (should fail)"

    # Create mock coverage file with low pluginhost coverage
    cat > "$TEST_DIR/coverage.out" <<EOF
mode: atomic
github.com/rshade/finfocus/internal/cli/root.go:10.1,12.2 1 1
github.com/rshade/finfocus/internal/cli/root.go:14.1,16.2 1 1
github.com/rshade/finfocus/internal/engine/engine.go:10.1,12.2 1 1
github.com/rshade/finfocus/internal/engine/engine.go:14.1,16.2 1 1
github.com/rshade/finfocus/internal/pluginhost/host.go:10.1,12.2 1 1
github.com/rshade/finfocus/internal/pluginhost/host.go:14.1,16.2 1 0
github.com/rshade/finfocus/internal/pluginhost/host.go:18.1,20.2 1 0
github.com/rshade/finfocus/internal/pluginhost/host.go:22.1,24.2 1 0
github.com/rshade/finfocus/internal/pluginhost/host.go:26.1,28.2 1 0
EOF

    # Run check-critical-coverage.sh with 50% threshold (pluginhost has 20% coverage)
    if [ -f "$PROJECT_ROOT/scripts/check-critical-coverage.sh" ]; then
        if bash "$PROJECT_ROOT/scripts/check-critical-coverage.sh" "$TEST_DIR/coverage.out" 50; then
            echo "✗ Test 4 failed: Should fail when pluginhost coverage is low"
            return 1
        else
            echo "✓ Test 4 passed: Correctly failed when pluginhost coverage below threshold"
            return 0
        fi
    else
        echo "⚠ Test 4 skipped: scripts/check-critical-coverage.sh not yet implemented"
        return 0
    fi
}

# Test 5: Missing critical package should be reported
test_missing_critical_package() {
    echo "Test 5: Missing critical package (should report warning)"

    # Create mock coverage file without any critical packages
    cat > "$TEST_DIR/coverage.out" <<EOF
mode: atomic
github.com/rshade/finfocus/internal/config/config.go:10.1,12.2 1 1
github.com/rshade/finfocus/internal/spec/loader.go:10.1,12.2 1 1
EOF

    # Run check-critical-coverage.sh - should warn about missing packages
    if [ -f "$PROJECT_ROOT/scripts/check-critical-coverage.sh" ]; then
        if bash "$PROJECT_ROOT/scripts/check-critical-coverage.sh" "$TEST_DIR/coverage.out" 50 2>&1 | grep -q "WARNING\|missing\|not found"; then
            echo "✓ Test 5 passed: Correctly reported missing critical packages"
            return 0
        else
            echo "✗ Test 5 failed: Should warn about missing critical packages"
            return 1
        fi
    else
        echo "⚠ Test 5 skipped: scripts/check-critical-coverage.sh not yet implemented"
        return 0
    fi
}

# Run all tests
echo "========================================"
echo "Testing scripts/check-critical-coverage.sh"
echo "========================================"
echo ""

FAILED=0

test_all_critical_paths_pass || FAILED=$((FAILED + 1))
echo ""
test_cli_below_threshold || FAILED=$((FAILED + 1))
echo ""
test_engine_below_threshold || FAILED=$((FAILED + 1))
echo ""
test_pluginhost_below_threshold || FAILED=$((FAILED + 1))
echo ""
test_missing_critical_package || FAILED=$((FAILED + 1))
echo ""

echo "========================================"
if [ $FAILED -eq 0 ]; then
    echo "✓ All tests passed"
    exit 0
else
    echo "✗ $FAILED test(s) failed"
    exit 1
fi
