package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root Cobra command for the pulumicost CLI and configures its subcommands.
// The returned command has its Version set from ver, a persistent "debug" flag, usage examples, and
// - config: init, set, get, list, and validate configuration commands.
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
