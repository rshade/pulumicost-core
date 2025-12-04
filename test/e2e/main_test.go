// Teardown ensures resources are cleaned up.
func (tc *TestContext) Teardown(ctx context.Context) {
	// Capture interrupt signals to ensure cleanup happens
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	
	cleanupDone := make(chan struct{})

	go func() {
		select {
		case <-c:
			tc.T.Log("Interrupt received, cleaning up...")
			if err := tc.Cleanup.PerformCleanup(ctx, tc.Stack); err != nil {
				tc.T.Logf("Cleanup failed on interrupt: %v", err)
			}
			os.Exit(1)
		case <-cleanupDone:
			return
		}
	}()

	// Ensure cleanup happens using defer in case of panic within Teardown itself (unlikely but safe)
	defer func() {
		close(cleanupDone)
		signal.Stop(c)
	}()

	if err := tc.Cleanup.PerformCleanup(ctx, tc.Stack); err != nil {
		tc.T.Errorf("Teardown failed: %v", err)
	}
}