package main

import (
	"testing"

	"github.com/rshade/finfocus/internal/cli"
	"github.com/rshade/finfocus/pkg/version"
)

func TestRun(t *testing.T) {
	// Test that run() can be called without panicking
	// Note: This is a basic smoke test. More comprehensive testing
	// would require mocking the CLI execution, which is complex
	// for a main package test.

	// We can't easily test the full execution without setting up
	// complex test harnesses, but we can test that the function
	// exists and can be called
	t.Run("run function exists", func(t *testing.T) {
		// This test mainly ensures the function can be called
		// In a real scenario, we'd mock dependencies
		_ = run
	})
}

func TestMainComponents(t *testing.T) {
	t.Run("version available", func(t *testing.T) {
		v := version.GetVersion()
		if v == "" {
			t.Error("expected version to be non-empty")
		}
	})

	t.Run("cli root command", func(t *testing.T) {
		root := cli.NewRootCmd(version.GetVersion())
		if root == nil {
			t.Error("expected root command to be non-nil")
		}
		if root.Use == "" {
			t.Error("expected root command to have a use string")
		}
	})
}
