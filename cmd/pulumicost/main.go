// Package main provides the pulumicost CLI tool for calculating cloud infrastructure costs.
// It supports both projected costs from Pulumi infrastructure definitions and actual historical
// costs from cloud provider APIs via a plugin-based architecture.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rshade/pulumicost-core/internal/cli"
	"github.com/rshade/pulumicost-core/internal/logging"
	"github.com/rshade/pulumicost-core/pkg/version"
	"github.com/spf13/cobra"
)

// run is the main application logic, separated for testability.
func run() error {
	// Check if the binary is being run as a Pulumi Analyzer plugin
	// Supports both legacy policy pack mode (pulumi-analyzer-policy-pulumicost)
	// and direct analyzer mode (pulumi-analyzer-pulumicost)
	exeName := filepath.Base(os.Args[0])
	if strings.Contains(exeName, "pulumi-analyzer-policy-pulumicost") ||
		strings.Contains(exeName, "pulumi-analyzer-pulumicost") {
		// If run as an analyzer plugin, execute the analyzer serve logic directly.
		// Pulumi expects the plugin binary to start a gRPC server and output the port.
		// Setup a minimal stderr logger for the analyzer serve command, as it expects it.
		// This must use os.Stderr to avoid breaking the Pulumi handshake.

		// Determine log level from environment, default to info to reduce noise
		logLevel := zerolog.InfoLevel
		if envLevel := os.Getenv("PULUMICOST_LOG_LEVEL"); envLevel != "" {
			if parsed, err := zerolog.ParseLevel(envLevel); err == nil {
				logLevel = parsed
			}
		}

		stderrLogger := zerolog.New(os.Stderr).
			Level(logLevel).
			With().
			Str("component", "analyzer-plugin-autostart").
			Timestamp().
			Logger()

		dummyCmd := &cobra.Command{}
		dummyCmd.SetContext(stderrLogger.WithContext(context.Background())) // Initialize context with stderrLogger

		// Call the analyzer serve run function directly
		return cli.RunAnalyzerServe(dummyCmd)
	}

	// Original CLI execution for the main pulumicost CLI tool
	// Initialize a minimal startup logger for early error reporting
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
		return err
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}
