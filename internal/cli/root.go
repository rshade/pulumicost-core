package cli

import (
	"github.com/spf13/cobra"
)

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

  # List installed plugins
  pulumicost plugin list

  # Validate all plugins
  pulumicost plugin validate`,
	}

	cmd.PersistentFlags().Bool("debug", false, "enable debug logging")

	costCmd := &cobra.Command{
		Use:   "cost",
		Short: "Cost calculation commands",
	}
	costCmd.AddCommand(
		newCostProjectedCmd(),
		newCostActualCmd(),
	)

	pluginCmd := &cobra.Command{
		Use:   "plugin",
		Short: "Plugin management commands",
	}
	pluginCmd.AddCommand(
		newPluginValidateCmd(),
		newPluginListCmd(),
	)

	cmd.AddCommand(costCmd, pluginCmd)

	return cmd
}
