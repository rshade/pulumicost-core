package e2e

import "os"

// HasAWSCredentials reports whether AWS credentials appear in the environment.
// It returns true if both the AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables are non-empty, false otherwise.
func HasAWSCredentials() bool {
	return os.Getenv("AWS_ACCESS_KEY_ID") != "" && os.Getenv("AWS_SECRET_ACCESS_KEY") != ""
}