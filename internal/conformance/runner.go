package conformance

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
)

// Runner executes individual conformance test cases.
type Runner struct {
	logger    zerolog.Logger
	verbosity Verbosity
}

// NewRunner creates a new test runner with the given logger and verbosity.
func NewRunner(logger zerolog.Logger, verbosity Verbosity) *Runner {
	return &Runner{
		logger:    logger,
		verbosity: verbosity,
	}
}

// RunTest executes a single test case and returns the result.
// It handles timeout, panic recovery, and result recording.
func (r *Runner) RunTest(ctx context.Context, tc TestCase, client interface{}) *TestResult {
	result := &TestResult{
		TestName:  tc.Name,
		Category:  tc.Category,
		Timestamp: time.Now(),
	}

	// Check for nil test function
	if tc.TestFunc == nil {
		result.Status = StatusError
		result.Error = "test function is nil"
		return result
	}

	// Set up timeout
	timeout := tc.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	testCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create test context
	tctx := &TestContext{
		PluginClient: client,
		Logger:       r.logger,
		Verbosity:    r.verbosity,
		Timeout:      timeout,
	}

	// Record start time
	startTime := time.Now()

	// Run test with panic recovery
	resultChan := make(chan *TestResult, 1)
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				resultChan <- &TestResult{
					TestName:  tc.Name,
					Category:  tc.Category,
					Status:    StatusError,
					Error:     fmt.Sprintf("panic: %v", recovered),
					Timestamp: time.Now(),
				}
			}
		}()

		testResult := tc.TestFunc(tctx)
		if testResult != nil {
			testResult.TestName = tc.Name
			testResult.Category = tc.Category
		}
		resultChan <- testResult
	}()

	// Wait for result or timeout
	select {
	case testResult := <-resultChan:
		if testResult != nil {
			result = testResult
		} else {
			result.Status = StatusError
			result.Error = "test returned nil result"
		}
	case <-testCtx.Done():
		result.Status = StatusError
		result.Error = fmt.Sprintf("timeout after %v", timeout)
	}

	// Record duration
	result.Duration = time.Since(startTime)
	result.Timestamp = time.Now()

	// Log result based on verbosity
	r.logResult(tc, result)

	return result
}

// logResult logs the test result based on the configured verbosity level.
func (r *Runner) logResult(tc TestCase, result *TestResult) {
	event := r.logger.Info()

	switch result.Status {
	case StatusPass:
		event = r.logger.Debug()
	case StatusFail:
		event = r.logger.Warn()
	case StatusError:
		event = r.logger.Error()
	case StatusSkip:
		event = r.logger.Info()
	}

	event.
		Str("test", tc.Name).
		Str("category", string(tc.Category)).
		Str("status", string(result.Status)).
		Dur("duration", result.Duration)

	if result.Error != "" {
		event.Str("error", result.Error)
	}

	if r.verbosity == VerbosityVerbose && result.Details != "" {
		event.Str("details", result.Details)
	}

	event.Msg("test completed")
}

// RunTests executes multiple test cases and returns all results.
func (r *Runner) RunTests(ctx context.Context, tests []TestCase, client interface{}) []TestResult {
	results := make([]TestResult, 0, len(tests))

	for _, tc := range tests {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			// Mark remaining tests as skipped due to cancellation
			for i := len(results); i < len(tests); i++ {
				results = append(results, TestResult{
					TestName:  tests[i].Name,
					Category:  tests[i].Category,
					Status:    StatusSkip,
					Error:     "context cancelled",
					Timestamp: time.Now(),
				})
			}
			return results
		default:
		}

		result := r.RunTest(ctx, tc, client)
		results = append(results, *result)
	}

	return results
}
