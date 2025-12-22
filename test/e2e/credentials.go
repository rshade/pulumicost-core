package e2e

import "os"

// HasAWSCredentials checks if AWS credentials are present in the environment.
func HasAWSCredentials() bool {
	return os.Getenv("AWS_ACCESS_KEY_ID") != "" && os.Getenv("AWS_SECRET_ACCESS_KEY") != ""
}
