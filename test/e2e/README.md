# End-to-End Tests

This directory contains E2E tests for Pulumicost Core.

## Running Tests

Prerequisites:
- `pulumicost` binary built and available in `bin/` or `PATH`.
- Fixtures available in `test/fixtures/plans/`.

Run all E2E tests:
```bash
go test -v -tags e2e ./test/e2e/...
```

## Test Structure

- `projected_cost_test.go`: Validates projected cost workflow.
- `output_*.go`: Validates different output formats.
- `errors_test.go`: Validates error handling.
- `*_test.go`: Provider-specific tests.
- `actual_cost_test.go`: Placeholder for actual cost workflow.

## Helpers

The tests rely on `findPulumicostBinary()` to locate the executable.
Tests use `exec.Command` to run the binary against fixture plans.
