package e2e

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// ComparisonReport holds details about a cost comparison.
type ComparisonReport struct {
	Expected    float64
	Actual      float64
	Diff        float64
	PercentDiff float64
	WithinLimit bool
	Message     string
}

func (r ComparisonReport) String() string {
	status := "PASS"
	if !r.WithinLimit {
		status = "FAIL"
	}
	return fmt.Sprintf("[%s] Expected: $%.4f, Actual: $%.4f, Diff: $%.4f (%.2f%%) - %s", 
		status, r.Expected, r.Actual, r.Diff, r.PercentDiff, r.Message)
}

// CostValidator defines the interface for validating cost calculations.
type CostValidator interface {
	ValidateProjected(actual float64, expected float64) error
	ValidateActual(calculated float64, runtime time.Duration, expectedHourly float64) error
	Compare(actual float64, expected float64) ComparisonReport
}

// DefaultCostValidator is a concrete implementation of CostValidator.
type DefaultCostValidator struct {
	TolerancePercent float64
}

// NewDefaultCostValidator creates a new DefaultCostValidator with the given tolerance.
func NewDefaultCostValidator(tolerance float64) *DefaultCostValidator {
	return &DefaultCostValidator{
		TolerancePercent: tolerance,
	}
}

// Compare generates a structured report comparing two cost values.
func (v *DefaultCostValidator) Compare(actual float64, expected float64) ComparisonReport {
	diff := math.Abs(actual - expected)
	var percentDiff float64
	if expected != 0 {
		percentDiff = (diff / expected) * 100
	} else if actual != 0 {
		percentDiff = 100.0 // Infinite difference if expected 0 but actual > 0
	}

	withinLimit := percentDiff <= v.TolerancePercent
	
	msg := "Within tolerance"
	if !withinLimit {
		msg = fmt.Sprintf("Exceeds tolerance of %.2f%%", v.TolerancePercent)
	}

	return ComparisonReport{
		Expected:    expected,
		Actual:      actual,
		Diff:        diff,
		PercentDiff: percentDiff,
		WithinLimit: withinLimit,
		Message:     msg,
	}
}

// ValidateProjected checks if the actual projected cost is within tolerance of the expected cost.
func (v *DefaultCostValidator) ValidateProjected(actual float64, expected float64) error {
	report := v.Compare(actual, expected)
	if !report.WithinLimit {
		return fmt.Errorf("projected cost mismatch: %s", report.String())
	}
	return nil
}

// ValidateActual checks if the calculated actual cost is proportional to runtime.
// Fallback formula: projected_cost * runtime_hours / 730
func (v *DefaultCostValidator) ValidateActual(calculated float64, runtime time.Duration, expectedHourly float64) error {
	// Minimum billing of 1 hour logic (simplified for now, assuming runtime is sufficient)
	// For this validator, we'll compare against the expected hourly rate * runtime
	
	runtimeHours := runtime.Hours()
	expectedTotal := expectedHourly * runtimeHours

	// Use a slightly looser tolerance for actual costs due to timing variations
	// or billing granularity if needed. For now, using the same tolerance.
	return v.ValidateProjected(calculated, expectedTotal)
}
