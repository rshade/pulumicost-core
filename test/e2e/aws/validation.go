package aws

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ValidateCost checks if actual cost is within tolerance of expected cost.
func ValidateCost(t *testing.T, resourceID string, actual, expected, tolerance float64) {
	delta := math.Abs(actual - expected)
	allowedDelta := expected * tolerance
	
	assert.True(t, delta <= allowedDelta, 
		"Cost for %s mismatch: actual=%f, expected=%f, tolerance=%f, delta=%f, allowed=%f",
		resourceID, actual, expected, tolerance, delta, allowedDelta)
}
