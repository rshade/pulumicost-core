//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
)

// GetAWSCredentialsEnv returns a slice of environment variables for AWS credentials
// using the AWS SDK v2. This avoids the need for external shell scripts to export credentials.
func GetAWSCredentialsEnv(ctx context.Context) ([]string, error) {
	// Load default config. It will pick up AWS_PROFILE, AWS_REGION, etc.
	// and also handle SSO session if configured.
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve AWS credentials: %w", err)
	}

	env := []string{
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", creds.AccessKeyID),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", creds.SecretAccessKey),
	}

	if creds.SessionToken != "" {
		env = append(env, fmt.Sprintf("AWS_SESSION_TOKEN=%s", creds.SessionToken))
	}

	// Also include the region if it was loaded from config and not already in env
	if cfg.Region != "" && os.Getenv("AWS_REGION") == "" {
		env = append(env, fmt.Sprintf("AWS_REGION=%s", cfg.Region))
	}

	return env, nil
}
