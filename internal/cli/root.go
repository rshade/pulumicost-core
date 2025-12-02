package cli

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rshade/pulumicost-core/internal/config"
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
			// Get logging configuration from config file (includes env var overrides)
			loggingCfg := config.GetLoggingConfig()

			// Check --debug flag (highest priority override)
			debug, _ := cmd.Flags().GetBool("debug")
			if debug {
				loggingCfg.Level = "debug"
				loggingCfg.Format = "console"
				loggingCfg.File = "" // Clear file to force stderr output for immediate visibility
			}

			// Check environment variable overrides (between config file and --debug)
			// Note: PULUMICOST_LOG_FORMAT can override even in debug mode for CI/scripting needs
			if envLevel := os.Getenv("PULUMICOST_LOG_LEVEL"); envLevel != "" && !debug {
				loggingCfg.Level = envLevel
			}
			if envFormat := os.Getenv("PULUMICOST_LOG_FORMAT"); envFormat != "" {
				loggingCfg.Format = envFormat
			}

			// Convert config to logging package format and create logger
			logCfg := loggingCfg.ToLoggingConfig()
			logResult := logging.NewLoggerWithPath(logCfg)
			logger = logging.ComponentLogger(logResult.Logger, "cli")

			// Display log path to operator (T020, T021, T022)
			// Print to stderr to avoid polluting command output (JSON, tables, etc.)
			if logResult.UsingFile {
				logging.PrintLogPathMessage(cmd.ErrOrStderr(), logResult.FilePath)
			} else if logResult.FallbackUsed {
				logging.PrintFallbackWarning(cmd.ErrOrStderr(), logResult.FallbackReason)
			}
			// When using stderr/stdout directly, no message is printed (T022)

			// Generate trace ID and store in context
			ctx := cmd.Context()
			traceID := logging.GetOrGenerateTraceID(ctx)
			ctx = logging.ContextWithTraceID(ctx, traceID)
			ctx = logger.WithContext(ctx)

			// Initialize audit logger (T032)
			auditCfg := logging.AuditLoggerConfig{
				Enabled: loggingCfg.Audit.Enabled,
				File:    loggingCfg.Audit.File,
			}
			auditLogger := logging.NewAuditLogger(auditCfg)
			ctx = logging.ContextWithAuditLogger(ctx, auditLogger)

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
