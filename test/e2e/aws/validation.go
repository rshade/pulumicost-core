package aws

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ValidateCost verifies that the actual cost for the specified resource is within the relative tolerance of the expected cost.
// If the absolute difference between actual and expected exceeds expected * tolerance, the test is failed on t with a detailed message.
// resourceID is the identifier of the resource being validated.
// actual and expected are the observed and expected cost values.
// tolerance is the allowed relative difference (for example, 0.1 represents 10%).
func ValidateCost(t *testing.T, resourceID string, actual, expected, tolerance float64) {
	delta := math.Abs(actual - expected)
	allowedDelta := expected * tolerance
	
	assert.True(t, delta <= allowedDelta, 
		"Cost for %s mismatch: actual=%f, expected=%f, tolerance=%f, delta=%f, allowed=%f",
		resourceID, actual, expected, tolerance, delta, allowedDelta)
}