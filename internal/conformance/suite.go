package conformance

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"time"

	"github.com/rs/zerolog"
	"github.com/rshade/pulumicost-core/internal/pluginhost"
	pbc "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
)

// Suite is the main conformance test suite orchestrator.
type Suite struct {
	config    SuiteConfig
	logger    zerolog.Logger
	testCases []TestCase
	filter    *regexp.Regexp
}

// NewSuite creates a new conformance test suite with the given configuration.
func NewSuite(cfg SuiteConfig) (*Suite, error) {
	// Validate required fields
	if cfg.PluginPath == "" {
		return nil, errors.New("plugin path is required")
	}

	// Apply defaults
	if cfg.CommMode == "" {
		cfg.CommMode = CommModeTCP
	} else if !IsValidCommMode(string(cfg.CommMode)) {
		return nil, fmt.Errorf("invalid comm mode: %s", cfg.CommMode)
	}

	if cfg.Verbosity == "" {
		cfg.Verbosity = VerbosityNormal
	} else if !IsValidVerbosity(string(cfg.Verbosity)) {
		return nil, fmt.Errorf("invalid verbosity: %s", cfg.Verbosity)
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = DefaultSuiteTimeout
	}

	// Compile test filter regex if provided
	var filter *regexp.Regexp
	if cfg.TestFilter != "" {
		var err error
		filter, err = regexp.Compile(cfg.TestFilter)
		if err != nil {
			return nil, fmt.Errorf("invalid test filter regex: %w", err)
		}
	}

	// Initialize logger - use zerolog.Nop() if disabled or not explicitly configured.
	// A zero-value zerolog.Logger has TraceLevel (0), while properly configured
	// loggers typically have InfoLevel or above. If the level is TraceLevel,
	// assume the logger wasn't explicitly configured and use Nop() instead.
	logger := cfg.Logger
	if logger.GetLevel() == zerolog.Disabled || logger.GetLevel() == zerolog.TraceLevel {
		logger = zerolog.Nop()
	}

	suite := &Suite{
		config: cfg,
		logger: logger,
		filter: filter,
	}

	// Register default test cases
	suite.registerDefaultTests()

	return suite, nil
}

// registerDefaultTests registers all built-in conformance test cases.
func (s *Suite) registerDefaultTests() {
	s.testCases = []TestCase{
		// Protocol tests
		{
			Name:            "Name_ReturnsPluginIdentifier",
			Category:        CategoryProtocol,
			Description:     "Verifies plugin returns its identifier via Name RPC",
			Timeout:         DefaultTimeout,
			RequiredMethods: []string{"Name"},
			TestFunc:        testNameReturnsIdentifier,
		},
		{
			Name:            "Name_ReturnsProtocolVersion",
			Category:        CategoryProtocol,
			Description:     "Verifies plugin returns protocol version via Name RPC",
			Timeout:         DefaultTimeout,
			RequiredMethods: []string{"Name"},
			TestFunc:        testNameReturnsProtocolVersion,
		},
		{
			Name:            "GetProjectedCost_ValidResource",
			Category:        CategoryProtocol,
			Description:     "Verifies GetProjectedCost returns cost for valid resource",
			Timeout:         DefaultTimeout,
			RequiredMethods: []string{"GetProjectedCost"},
			TestFunc:        testGetProjectedCostValid,
		},
		{
			Name:            "GetProjectedCost_InvalidResource",
			Category:        CategoryError,
			Description:     "Verifies GetProjectedCost returns NotFound for unsupported resource",
			Timeout:         DefaultTimeout,
			RequiredMethods: []string{"GetProjectedCost"},
			TestFunc:        testGetProjectedCostInvalid,
		},
		// Context tests
		{
			Name:            "Context_Cancellation",
			Category:        CategoryContext,
			Description:     "Verifies plugin respects context cancellation",
			Timeout:         DefaultTimeout,
			RequiredMethods: []string{"GetProjectedCost"},
			TestFunc:        testContextCancellation,
		},
		// Performance tests
		{
			Name:            "Timeout_Respected",
			Category:        CategoryPerformance,
			Description:     "Verifies plugin responds within timeout limits",
			Timeout:         DefaultTimeout,
			RequiredMethods: []string{"Name"},
			TestFunc:        testTimeoutRespected,
		},
		{
			Name:            "Batch_Handling",
			Category:        CategoryPerformance,
			Description:     "Verifies plugin handles multiple sequential requests",
			Timeout:         DefaultTimeout * batchTestTimeoutMultiplier,
			RequiredMethods: []string{"GetProjectedCost"},
			TestFunc:        testBatchHandling,
		},
	}
}

// Placeholder test functions - return Skip until properly implemented.
// TODO: Implement actual test logic in error.go, etc.

// GetTestCases returns the filtered list of test cases based on configuration.
func (s *Suite) GetTestCases() []TestCase {
	filtered := make([]TestCase, 0, len(s.testCases))

	for _, tc := range s.testCases {
		// Filter by category
		if len(s.config.Categories) > 0 {
			if !slices.Contains(s.config.Categories, tc.Category) {
				continue
			}
		}

		// Filter by regex
		if s.filter != nil && !s.filter.MatchString(tc.Name) {
			continue
		}

		filtered = append(filtered, tc)
	}

	return filtered
}

// Run executes all conformance tests and returns the report.
func (s *Suite) Run(ctx context.Context) (*SuiteReport, error) {
	// Check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Create timeout context for entire suite
	suiteCtx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	startTime := time.Now()

	// Get filtered test cases
	testCases := s.GetTestCases()

	s.logger.Info().
		Str("plugin", s.config.PluginPath).
		Str("mode", string(s.config.CommMode)).
		Int("test_count", len(testCases)).
		Msg("starting conformance suite")

	// Start plugin process and establish connection
	launcher := pluginhost.NewProcessLauncher()
	conn, closeFn, err := launcher.Start(suiteCtx, s.config.PluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to start plugin: %w", err)
	}
	defer func() {
		if closeErr := closeFn(); closeErr != nil {
			s.logger.Warn().Err(closeErr).Msg("error closing plugin process")
		}
	}()

	client := pbc.NewCostSourceServiceClient(conn)

	// Fetch plugin info for report
	nameResp, err := client.Name(suiteCtx, &pbc.NameRequest{})
	pluginName := "unknown"
	pluginVersion := "unknown"
	if err == nil {
		pluginName = nameResp.GetName()
		// Version might be in metadata or we might add a Version() RPC
	}

	// Create runner
	runner := NewRunner(s.logger, s.config.Verbosity)

	// Run tests
	results := runner.RunTests(suiteCtx, testCases, client)

	endTime := time.Now()

	// Build report
	report := &SuiteReport{
		SuiteName: "conformance",
		Plugin: PluginUnderTest{
			Path:            s.config.PluginPath,
			Name:            pluginName,
			Version:         pluginVersion,
			ProtocolVersion: ProtocolVersion,
			CommMode:        s.config.CommMode,
		},
		Results:   results,
		Summary:   calculateSummary(results),
		StartTime: startTime,
		EndTime:   endTime,
		TotalTime: endTime.Sub(startTime),
		Timestamp: endTime,
	}

	s.logger.Info().
		Int("total", report.Summary.Total).
		Int("passed", report.Summary.Passed).
		Int("failed", report.Summary.Failed).
		Int("skipped", report.Summary.Skipped).
		Int("errors", report.Summary.Errors).
		Dur("duration", report.TotalTime).
		Msg("conformance suite completed")

	return report, nil
}

// calculateSummary computes aggregate counts from test results.
func calculateSummary(results []TestResult) Summary {
	summary := Summary{
		Total: len(results),
	}

	for _, r := range results {
		switch r.Status {
		case StatusPass:
			summary.Passed++
		case StatusFail:
			summary.Failed++
		case StatusSkip:
			summary.Skipped++
		case StatusError:
			summary.Errors++
		default:
			// Unknown status - count as error to surface the issue
			summary.Errors++
		}
	}

	return summary
}
