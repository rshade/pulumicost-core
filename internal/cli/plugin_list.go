package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/rshade/pulumicost-core/internal/config"
	"github.com/rshade/pulumicost-core/internal/registry"
	"github.com/spf13/cobra"
)

func NewPluginListCmd() *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed plugins",
		Long:  "List all installed plugins with their versions and paths",
		Example: `  # List all installed plugins
  pulumicost plugin list

  # List plugins with detailed information
  pulumicost plugin list --verbose`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runPluginListCmd(cmd, verbose)
		},
	}

	cmd.Flags().BoolVar(&verbose, "verbose", false, "Show detailed plugin information")

	return cmd
}

func runPluginListCmd(cmd *cobra.Command, verbose bool) error {
	cfg := config.New()
	if _, err := os.Stat(cfg.PluginDir); os.IsNotExist(err) {
		cmd.Printf("Plugin directory does not exist: %s\n", cfg.PluginDir)
		cmd.Println("No plugins installed.")
		return nil
	}

	reg := registry.NewDefault()
	plugins, err := reg.ListPlugins()
	if err != nil {
		return fmt.Errorf("listing plugins: %w", err)
	}

	if len(plugins) == 0 {
		cmd.Println("No plugins found.")
		return nil
	}

	return displayPlugins(plugins, verbose)
}

func displayPlugins(plugins []registry.PluginInfo, verbose bool) error {
	const tabPadding = 2
	w := tabwriter.NewWriter(os.Stdout, 0, 0, tabPadding, ' ', 0)

	if verbose {
		return displayVerbosePlugins(w, plugins)
	}
	return displaySimplePlugins(w, plugins)
}

func displayVerbosePlugins(w *tabwriter.Writer, plugins []registry.PluginInfo) error {
	fmt.Fprintln(w, "Name\tVersion\tPath\tExecutable")
	fmt.Fprintln(w, "----\t-------\t----\t----------")

	for _, plugin := range plugins {
		execStatus := getExecutableStatus(plugin.Path)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", plugin.Name, plugin.Version, plugin.Path, execStatus)
	}
	return w.Flush()
}

func displaySimplePlugins(w *tabwriter.Writer, plugins []registry.PluginInfo) error {
	fmt.Fprintln(w, "Name\tVersion\tPath")
	fmt.Fprintln(w, "----\t-------\t----")

	for _, plugin := range plugins {
		fmt.Fprintf(w, "%s\t%s\t%s\n", plugin.Name, plugin.Version, plugin.Path)
	}
	return w.Flush()
}

func getExecutableStatus(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return "No"
	}
	if info.Mode()&0111 != 0 {
		return "Yes"
	}
	return "No"
}
