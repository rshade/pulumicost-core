package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

// CleanupManager handles the destruction of Pulumi stacks with timeouts and retries.
type CleanupManager struct {
	T       *testing.T
	Timeout time.Duration
}

// NewCleanupManager creates a new CleanupManager.
func NewCleanupManager(t *testing.T, timeout time.Duration) *CleanupManager {
	return &CleanupManager{
		T:       t,
		Timeout: timeout,
	}
}

// PerformCleanup destroys the stack and removes it.
func (cm *CleanupManager) PerformCleanup(ctx context.Context, stack auto.Stack) error {
	cm.T.Logf("Starting cleanup for stack %s (timeout: %v)", stack.Name(), cm.Timeout)

	ctx, cancel := context.WithTimeout(ctx, cm.Timeout)
	defer cancel()

	// Retry logic for destroy
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		_, err := stack.Destroy(ctx)
		if err == nil {
			break
		}
		if i == maxRetries-1 {
			return fmt.Errorf("failed to destroy stack after %d attempts: %w", maxRetries, err)
		}
		cm.T.Logf("Destroy failed, retrying in 10s... (%d/%d)", i+1, maxRetries)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
			// continue
		}
	}

	// Remove the stack
	err := stack.Workspace().RemoveStack(ctx, stack.Name())
	if err != nil {
		return fmt.Errorf("failed to remove stack: %w", err)
	}

	cm.T.Log("Cleanup successful")
	return nil
}
