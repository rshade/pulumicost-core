---
title: Testing Guide
description: Comprehensive guide to testing in FinFocus Core
layout: default
---

This guide covers the testing philosophy, strategy, and practical instructions for FinFocus Core.

## Testing Philosophy

1. **Test-Driven Development (TDD)**: Write tests before implementation.
2. **High Coverage**: Aim for 80% overall coverage, 95% for critical paths.
3. **Isolation**: Unit tests must not depend on external systems (use mocks).
4. **Integration**: Verify component interactions with dedicated integration tests.
5. **Performance**: Benchmarks must track critical path performance.

## Test Categories

| Category        | Path                | Description                                                 |
| --------------- | ------------------- | ----------------------------------------------------------- |
| **Unit**        | `test/unit/`        | Isolated tests for individual functions/methods.            |
| **Integration** | `test/integration/` | Tests for component interactions (CLI -> Engine -> Plugin). |
| **E2E**         | `test/e2e/`         | Full system tests running against fixtures.                 |
| **Benchmarks**  | `test/benchmarks/`  | Performance tests for regression detection.                 |

## Running Tests

### Unit Tests

```bash
make test-unit
# or
go test -v ./test/unit/...
```

### Integration Tests

```bash
make test-integration
# or
go test -v ./test/integration/...
```

### E2E Tests

```bash
make test-e2e
# or
go test -v -tags e2e ./test/e2e/...
```

### All Tests

```bash
make test
```

## Writing Tests

See [examples](./examples/) for patterns.
