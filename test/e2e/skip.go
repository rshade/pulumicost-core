package e2e

import "fmt"

// ShouldSkip reports whether tests for the given provider should be skipped and returns a human-readable reason when skipping.
// The first return value is true if tests should be skipped; the second is the skip reason or an empty string when no skip is needed.
func ShouldSkip(provider string) (bool, string) {
	switch provider {
	case "aws":
		if !HasAWSCredentials() {
			return true, "missing credentials: AWS_ACCESS_KEY_ID or AWS_SECRET_ACCESS_KEY not set"
		}
	default:
		return false, ""
	}
	return false, ""
}