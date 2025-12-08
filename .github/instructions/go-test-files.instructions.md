---
applyTo: '**/*_test.go'
---

# Go Test Files Instructions

Test files follow specific patterns and requirements for this Go project.

## Test Structure:

- Use `_test.go` suffix for test files
- Place test files in the same package as the code being tested
- Use descriptive test function names: `TestXxx`
- Use table-driven tests for multiple test cases
- Group related tests appropriately

## Test Coverage Requirements:

- Achieve 80% minimum code coverage
- Test both success and error paths
- Include edge cases and boundary conditions
- Test error conditions and failure scenarios
- Verify proper error handling and propagation
- Test concurrent behavior with race detection

## Testing Patterns:

- Use `github.com/stretchr/testify` for assertions
- Use `require` for fatal assertions, `assert` for non-fatal
- Mock external dependencies appropriately
- Use test helpers and setup/teardown functions
- Isolate tests to prevent interference

## Test Organization:

- Use descriptive test names that explain what is being tested
- Group related test cases in table-driven tests
- Use subtests for complex scenarios
- Include benchmark tests where performance is critical
- Document test intent and expectations

## CI/CD Integration:

- Tests run automatically in CI pipeline
- Coverage reports are generated and uploaded
- Race detection is enabled for concurrent code
- Test failures block merges

## Mock and Stub Usage:

- Use interfaces for dependency injection
- Create mock implementations for external services
- Use test-specific fixtures and data
- Avoid testing implementation details

## Performance Testing:

- Include benchmarks for performance-critical code
- Use `testing.B` for benchmark functions
- Profile memory usage and allocations
- Compare performance across changes
