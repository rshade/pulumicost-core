// Package main provides the pulumicost CLI tool for calculating cloud infrastructure costs.
// It supports both projected costs from Pulumi infrastructure definitions and actual historical
// costs from cloud provider APIs via a plugin-based architecture.
package main

import (
	"fmt"
	"os"

	"github.com/rshade/pulumicost-core/internal/cli"
	"github.com/rshade/pulumicost-core/internal/logging"
	"github.com/rshade/pulumicost-core/pkg/version"
)

// and exits the process with status code 1.
func main() {
	// Initialize a minimal startup logger for early error reporting
	// Full logger initialization happens in PersistentPreRunE with debug/config options
	startupCfg := logging.LoggingConfig{
		Level:  "error",
		Format: "json",
		Output: "stderr",
	}
	startupLogger := logging.NewLogger(startupCfg)
	startupLogger = logging.ComponentLogger(startupLogger, "main")

	root := cli.NewRootCmd(version.GetVersion())
	if err := root.Execute(); err != nil {
		// Print user-friendly error to stderr for immediate visibility
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		// Also log for debugging purposes
		startupLogger.Error().Err(err).Msg("command execution failed")
		os.Exit(1)
	}
}