package conformance

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatus_Constants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   Status
		expected string
	}{
		{"pass status", StatusPass, "pass"},
		{"fail status", StatusFail, "fail"},
		{"skip status", StatusSkip, "skip"},
		{"error status", StatusError, "error"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, string(tc.status))
		})
	}
}

func TestCategory_Constants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		category Category
		expected string
	}{
		{"protocol category", CategoryProtocol, "protocol"},
		{"performance category", CategoryPerformance, "performance"},
		{"error category", CategoryError, "error"},
		{"context category", CategoryContext, "context"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, string(tc.category))
		})
	}
}

func TestAllCategories(t *testing.T) {
	t.Parallel()

	categories := AllCategories()

	require.Len(t, categories, 4)
	assert.Contains(t, categories, CategoryProtocol)
	assert.Contains(t, categories, CategoryPerformance)
	assert.Contains(t, categories, CategoryError)
	assert.Contains(t, categories, CategoryContext)
}

func TestIsValidCategory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		category string
		expected bool
	}{
		{"valid protocol", "protocol", true},
		{"valid performance", "performance", true},
		{"valid error", "error", true},
		{"valid context", "context", true},
		{"invalid empty", "", false},
		{"invalid unknown", "unknown", false},
		{"invalid case", "Protocol", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, IsValidCategory(tc.category))
		})
	}
}

func TestVerbosity_Constants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		verbosity Verbosity
		expected  string
	}{
		{"quiet verbosity", VerbosityQuiet, "quiet"},
		{"normal verbosity", VerbosityNormal, "normal"},
		{"verbose verbosity", VerbosityVerbose, "verbose"},
		{"debug verbosity", VerbosityDebug, "debug"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, string(tc.verbosity))
		})
	}
}

func TestIsValidVerbosity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		verbosity string
		expected  bool
	}{
		{"valid quiet", "quiet", true},
		{"valid normal", "normal", true},
		{"valid verbose", "verbose", true},
		{"valid debug", "debug", true},
		{"invalid empty", "", false},
		{"invalid unknown", "unknown", false},
		{"invalid case", "Quiet", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, IsValidVerbosity(tc.verbosity))
		})
	}
}

func TestCommMode_Constants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		commMode CommMode
		expected string
	}{
		{"tcp mode", CommModeTCP, "tcp"},
		{"stdio mode", CommModeStdio, "stdio"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, string(tc.commMode))
		})
	}
}

func TestIsValidCommMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		mode     string
		expected bool
	}{
		{"valid tcp", "tcp", true},
		{"valid stdio", "stdio", true},
		{"invalid empty", "", false},
		{"invalid unknown", "unknown", false},
		{"invalid case", "TCP", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, IsValidCommMode(tc.mode))
		})
	}
}

func TestTestCase(t *testing.T) {
	t.Parallel()

	tc := TestCase{
		Name:            "Name_ReturnsPluginIdentifier",
		Category:        CategoryProtocol,
		Description:     "Verifies plugin returns its identifier via Name RPC",
		Timeout:         5 * time.Second,
		RequiredMethods: []string{"Name"},
	}

	assert.Equal(t, "Name_ReturnsPluginIdentifier", tc.Name)
	assert.Equal(t, CategoryProtocol, tc.Category)
	assert.Equal(t, "Verifies plugin returns its identifier via Name RPC", tc.Description)
	assert.Equal(t, 5*time.Second, tc.Timeout)
	assert.Equal(t, []string{"Name"}, tc.RequiredMethods)
}

func TestTestResult(t *testing.T) {
	t.Parallel()

	now := time.Now()
	result := TestResult{
		TestName:  "Name_ReturnsPluginIdentifier",
		Category:  CategoryProtocol,
		Status:    StatusPass,
		Duration:  50 * time.Millisecond,
		Timestamp: now,
	}

	assert.Equal(t, "Name_ReturnsPluginIdentifier", result.TestName)
	assert.Equal(t, CategoryProtocol, result.Category)
	assert.Equal(t, StatusPass, result.Status)
	assert.Equal(t, 50*time.Millisecond, result.Duration)
	assert.Equal(t, now, result.Timestamp)
	assert.Empty(t, result.Error)
	assert.Empty(t, result.Details)
}

func TestTestResult_WithError(t *testing.T) {
	t.Parallel()

	result := TestResult{
		TestName: "GetProjectedCost_InvalidResource",
		Category: CategoryError,
		Status:   StatusFail,
		Duration: 110 * time.Millisecond,
		Error:    "expected NotFound, got InvalidArgument",
		Details:  "Request: {...}, Response: {...}",
	}

	assert.Equal(t, StatusFail, result.Status)
	assert.Equal(t, "expected NotFound, got InvalidArgument", result.Error)
	assert.Equal(t, "Request: {...}, Response: {...}", result.Details)
}

func TestPluginUnderTest(t *testing.T) {
	t.Parallel()

	plugin := PluginUnderTest{
		Path:            "/path/to/plugin",
		Name:            "aws-cost",
		Version:         "1.2.0",
		ProtocolVersion: "1.0",
		CommMode:        CommModeTCP,
	}

	assert.Equal(t, "/path/to/plugin", plugin.Path)
	assert.Equal(t, "aws-cost", plugin.Name)
	assert.Equal(t, "1.2.0", plugin.Version)
	assert.Equal(t, "1.0", plugin.ProtocolVersion)
	assert.Equal(t, CommModeTCP, plugin.CommMode)
}

func TestSuiteConfig_Defaults(t *testing.T) {
	t.Parallel()

	cfg := SuiteConfig{
		PluginPath: "/path/to/plugin",
	}

	assert.Equal(t, "/path/to/plugin", cfg.PluginPath)
	assert.Empty(t, cfg.CommMode)   // Will be defaulted by NewSuite
	assert.Empty(t, cfg.Verbosity)  // Will be defaulted by NewSuite
	assert.Zero(t, cfg.Timeout)     // Will be defaulted by NewSuite
	assert.Nil(t, cfg.Categories)   // Empty means all categories
	assert.Empty(t, cfg.TestFilter) // Empty means no filter
}

func TestSuiteConfig_FullConfiguration(t *testing.T) {
	t.Parallel()

	cfg := SuiteConfig{
		PluginPath:   "/path/to/plugin",
		CommMode:     CommModeTCP,
		Verbosity:    VerbosityVerbose,
		OutputFormat: "json",
		OutputPath:   "/path/to/output.json",
		Timeout:      10 * time.Minute,
		Categories:   []Category{CategoryProtocol, CategoryError},
		TestFilter:   "Name_.*",
	}

	assert.Equal(t, "/path/to/plugin", cfg.PluginPath)
	assert.Equal(t, CommModeTCP, cfg.CommMode)
	assert.Equal(t, VerbosityVerbose, cfg.Verbosity)
	assert.Equal(t, "json", cfg.OutputFormat)
	assert.Equal(t, "/path/to/output.json", cfg.OutputPath)
	assert.Equal(t, 10*time.Minute, cfg.Timeout)
	assert.Len(t, cfg.Categories, 2)
	assert.Equal(t, "Name_.*", cfg.TestFilter)
}

func TestSummary(t *testing.T) {
	t.Parallel()

	summary := Summary{
		Total:   20,
		Passed:  18,
		Failed:  1,
		Skipped: 1,
		Errors:  0,
	}

	assert.Equal(t, 20, summary.Total)
	assert.Equal(t, 18, summary.Passed)
	assert.Equal(t, 1, summary.Failed)
	assert.Equal(t, 1, summary.Skipped)
	assert.Equal(t, 0, summary.Errors)

	// Verify counts add up
	assert.Equal(t, summary.Total, summary.Passed+summary.Failed+summary.Skipped+summary.Errors)
}

func TestSuiteReport(t *testing.T) {
	t.Parallel()

	startTime := time.Now()
	endTime := startTime.Add(4500 * time.Millisecond)

	report := SuiteReport{
		SuiteName: "conformance",
		Plugin: PluginUnderTest{
			Path:            "/path/to/plugin",
			Name:            "aws-cost",
			Version:         "1.2.0",
			ProtocolVersion: "1.0",
			CommMode:        CommModeTCP,
		},
		Results: []TestResult{
			{
				TestName: "Name_ReturnsPluginIdentifier",
				Category: CategoryProtocol,
				Status:   StatusPass,
				Duration: 50 * time.Millisecond,
			},
			{
				TestName: "GetProjectedCost_InvalidResource",
				Category: CategoryError,
				Status:   StatusFail,
				Duration: 110 * time.Millisecond,
				Error:    "expected NotFound, got InvalidArgument",
			},
		},
		Summary: Summary{
			Total:   20,
			Passed:  18,
			Failed:  1,
			Skipped: 1,
			Errors:  0,
		},
		StartTime: startTime,
		EndTime:   endTime,
		TotalTime: 4500 * time.Millisecond,
		Timestamp: endTime,
	}

	assert.Equal(t, "conformance", report.SuiteName)
	assert.Equal(t, "aws-cost", report.Plugin.Name)
	assert.Len(t, report.Results, 2)
	assert.Equal(t, 20, report.Summary.Total)
	assert.Equal(t, startTime, report.StartTime)
	assert.Equal(t, endTime, report.EndTime)
	assert.Equal(t, 4500*time.Millisecond, report.TotalTime)
}

func TestConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 10*time.Second, DefaultTimeout)
	assert.Equal(t, 5*time.Minute, DefaultSuiteTimeout)
	assert.Equal(t, 1000, MaxBatchSize)
	assert.Equal(t, "1.0", ProtocolVersion)
}
