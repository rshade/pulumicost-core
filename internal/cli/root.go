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
