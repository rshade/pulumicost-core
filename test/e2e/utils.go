package e2e

import (
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
)

// LogComparisonReport logs the comparison report to the test logger.
func LogComparisonReport(t *testing.T, report ComparisonReport) {
	t.Logf("Cost Comparison Report: %s", report.String())
}

// ParseTimeRange parses start and end time strings in ISO 8601 or YYYY-MM-DD format.
func ParseTimeRange(startStr, endStr string) (time.Time, time.Time, error) {
	var start, end time.Time
	var err error

	formats := []string{time.RFC3339, "2006-01-02"}

	parse := func(s string) (time.Time, error) {
		for _, f := range formats {
			if t, err := time.Parse(f, s); err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
	}

	start, err = parse(startStr)
	if err != nil {
		return start, end, err
	}

	end, err = parse(endStr)
	if err != nil {
		return start, end, err
	}

	return start, end, nil
}

// GenerateStackName creates a unique stack name with a ULID suffix.
// Format: prefix-ULID
func GenerateStackName(prefix string) string {
	t := time.Now()
	entropy := ulid.Monotonic(rand.Reader, 0)
	id := ulid.MustNew(ulid.Timestamp(t), entropy)
	return fmt.Sprintf("%s-%s", prefix, id.String())
}

// findFinFocusBinary locates the finfocus binary.
func findFinFocusBinary() string {
	// Check common locations
	locations := []string{
		os.Getenv("FINFOCUS_BINARY"), // Environment override
		"../../bin/finfocus",         // From test/e2e relative to repo root
		"../../../bin/finfocus",      // Alternative
	}

	for _, loc := range locations {
		if loc == "" {
			continue
		}
		absPath, err := filepath.Abs(loc)
		if err != nil {
			continue
		}
		if _, err := os.Stat(absPath); err == nil {
			return absPath
		}
	}

	// Try PATH
	if path, err := exec.LookPath("finfocus"); err == nil {
		return path
	}

	return ""
}
