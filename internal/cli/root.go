package cli

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rshade/pulumicost-core/internal/logging"
	"github.com/spf13/cobra"
)

// logger is the package-level logger for CLI operations.
var logger zerolog.Logger //nolint:gochecknoglobals // Required for zerolog context integration

// NewRootCmd creates the root Cobra command for the pulumicost CLI.
// It wires up logging, tracing, audit logging, and subcommands (cost, plugin, config, analyzer).
// The command dynamically adjusts its Use and Example strings based on whether it's running
// as a Pulumi tool plugin (detected via binary name or PULUMICOST_PLUGIN_MODE env var).
func NewRootCmd(ver string) *cobra.Command {
	return NewRootCmdWithArgs(ver, os.Args, os.LookupEnv)
}

// NewRootCmdWithArgs creates the root command with explicit args and env lookup for testability.
// This allows tests to inject custom args and environment variables.
func NewRootCmdWithArgs(ver string, args []string, lookupEnv func(string) (string, bool)) *cobra.Command {
	var logResult *logging.LogPathResult

	// Detect plugin mode from binary name or environment variable
	pluginMode := DetectPluginMode(args, lookupEnv)

	// Select the appropriate Use and Example strings based on mode
	useName := "pulumicost"
	example := rootCmdExample
	if pluginMode {
		useName = "pulumi plugin run tool cost"
		example = pluginCmdExample
	}

	cmd := &cobra.Command{
		Use:     useName,
		Short:   "PulumiCost CLI and plugin host",
		Long:    "PulumiCost: Calculate projected and actual cloud costs via plugins",
		Version: ver,
		Example: example,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			result := setupLogging(cmd)
			logResult = &result
			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, _ []string) error {
			return cleanupLogging(cmd, logResult)
		},
	}

	cmd.PersistentFlags().Bool("debug", false, "enable debug logging")
	cmd.PersistentFlags().Bool("skip-version-check", false, "skip plugin spec version compatibility check")
	cmd.AddCommand(newCostCmd(), newPluginCmd(), newConfigCmd(), NewAnalyzerCmd())

	return cmd
}

const rootCmdExample = `  # Calculate projected costs from a Pulumi plan
  pulumicost cost projected --pulumi-json plan.json

  # Get actual costs for the last 7 days
  pulumicost cost actual --pulumi-json plan.json --from 2025-01-07

  # Install a plugin from registry
  pulumicost plugin install kubecost

  # List installed plugins
  pulumicost plugin list

  # Initialize a new plugin project
  pulumicost plugin init aws-plugin --author "Your Name" --providers aws

  # Validate all plugins
  pulumicost plugin validate

  # Initialize configuration
  pulumicost config init

  # Set configuration values
  pulumicost config set output.default_format json`

// pluginCmdExample is the example text shown when running as a Pulumi tool plugin.
const pluginCmdExample = `  # Calculate projected costs from a Pulumi plan
  pulumi plugin run tool cost -- cost projected --pulumi-json plan.json

  # Get actual costs for the last 7 days
  pulumi plugin run tool cost -- cost actual --pulumi-json plan.json --from 2025-01-07

  # List installed plugins
  pulumi plugin run tool cost -- plugin list

  # Validate all plugins
  pulumi plugin run tool cost -- plugin validate

  # Initialize configuration
  pulumi plugin run tool cost -- config init

  # Set configuration values
  pulumi plugin run tool cost -- config set output.default_format json`

// newCostCmd creates the cost command group with projected, actual, and recommendations subcommands.
func newCostCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "cost", Short: "Cost calculation commands"}
	cmd.AddCommand(NewCostProjectedCmd(), NewCostActualCmd(), NewCostRecommendationsCmd())
	return cmd
}

// newPluginCmd creates the plugin command group with management subcommands.
func newPluginCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "plugin", Short: "Plugin management commands"}
	cmd.AddCommand(
		NewPluginValidateCmd(), NewPluginListCmd(), NewPluginInitCmd(),
		NewPluginInstallCmd(), NewPluginUpdateCmd(), NewPluginRemoveCmd(),
		NewPluginConformanceCmd(), NewPluginCertifyCmd(), NewPluginInspectCmd(),
	)
	return cmd
}

// newConfigCmd creates the config command group with configuration subcommands.
func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "config", Short: "Configuration management commands"}
	cmd.AddCommand(
		NewConfigInitCmd(), NewConfigSetCmd(), NewConfigGetCmd(),
		NewConfigListCmd(), NewConfigValidateCmd(),
	)
	return cmd
}
