package cli

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rshade/pulumicost-core/internal/logging"
	"github.com/spf13/cobra"
)

// logger is the package-level logger for CLI operations.
var logger zerolog.Logger //nolint:gochecknoglobals // Required for zerolog context integration

// NewRootCmd creates the root Cobra command for the pulumicost CLI and configures its subcommands.
// The returned command has its Version set from ver, a persistent "debug" flag, usage examples, and
// - config: init, set, get, list, and validate configuration commands.
//
//nolint:funlen // Comprehensive logging setup and subcommand registration requires additional lines
func NewRootCmd(ver string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pulumicost",
		Short:   "PulumiCost CLI and plugin host",
		Long:    "PulumiCost: Calculate projected and actual cloud costs via plugins",
		Version: ver,
		Example: `  # Calculate projected costs from a Pulumi plan
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
	  pulumicost config set output.default_format json`,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// Initialize logging based on --debug flag
			debug, _ := cmd.Flags().GetBool("debug")

			level := "info"
			format := "json"
			if debug {
				level = "debug"
				format = "console"
			}

			// Check environment variable overrides
			if envLevel := os.Getenv("PULUMICOST_LOG_LEVEL"); envLevel != "" {
				level = envLevel
			}
			if envFormat := os.Getenv("PULUMICOST_LOG_FORMAT"); envFormat != "" {
				format = envFormat
			}

			cfg := logging.LoggingConfig{
				Level:  level,
				Format: format,
				Output: "stderr",
			}

			logger = logging.NewLogger(cfg)
			logger = logging.ComponentLogger(logger, "cli")

			// Generate trace ID and store in context
			ctx := cmd.Context()
			traceID := logging.GetOrGenerateTraceID(ctx)
			ctx = logging.ContextWithTraceID(ctx, traceID)
			ctx = logger.WithContext(ctx)
			cmd.SetContext(ctx)

			// Log command start
			logger.Info().
				Ctx(ctx).
				Str("command", cmd.Name()).
				Str("version", ver).
				Msg("command started")

			return nil
		},
	}

	cmd.PersistentFlags().Bool("debug", false, "enable debug logging")

	costCmd := &cobra.Command{
		Use:   "cost",
		Short: "Cost calculation commands",
	}
	costCmd.AddCommand(
		NewCostProjectedCmd(),
		NewCostActualCmd(),
	)

	pluginCmd := &cobra.Command{
		Use:   "plugin",
		Short: "Plugin management commands",
	}
	pluginCmd.AddCommand(
		NewPluginValidateCmd(),
		NewPluginListCmd(),
		NewPluginInitCmd(),
		NewPluginInstallCmd(),
		NewPluginUpdateCmd(),
		NewPluginRemoveCmd(),
	)

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management commands",
	}
	configCmd.AddCommand(
		NewConfigInitCmd(),
		NewConfigSetCmd(),
		NewConfigGetCmd(),
		NewConfigListCmd(),
		NewConfigValidateCmd(),
	)

	cmd.AddCommand(costCmd, pluginCmd, configCmd)

	return cmd
}
