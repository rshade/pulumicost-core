package e2e

import "fmt"

// ShouldSkip checks if tests for a provider should be skipped due to missing credentials.
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
