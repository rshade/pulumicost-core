package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/rshade/pulumicost-core/internal/config"
	"github.com/rshade/pulumicost-core/internal/registry"
	"github.com/spf13/cobra"
)

// NewPluginListCmd creates a Cobra "list" command for displaying plugins.
// The command lists installed plugins by default and supports an `--verbose`
// flag for detailed output and an `--available` flag to list plugins from the registry.
// It returns the configured *cobra.Command.
func NewPluginListCmd() *cobra.Command {
	var (
		verbose   bool
		available bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed plugins",
		Long:  "List all installed plugins with their versions and paths",
		Example: `  # List all installed plugins
  pulumicost plugin list

  # List plugins with detailed information
  pulumicost plugin list --verbose

  # List available plugins from registry
  pulumicost plugin list --available`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if available {
				return runPluginListAvailable(cmd)
			}
			return runPluginListCmd(cmd, verbose)
		},
	}

	cmd.Flags().BoolVar(&verbose, "verbose", false, "Show detailed plugin information")
	cmd.Flags().BoolVar(&available, "available", false, "List available plugins from registry")

	return cmd
}

// runPluginListAvailable lists plugins available in the registry and writes a tabulated
// table (Name, Description, Repository, Security) to the command's output.
//
// If the registry cannot be loaded the function returns an error wrapping the underlying
// cause. If no entries exist the function prints "No plugins available in registry."
// to the command output and returns nil. For entries with an empty security level the
// security column defaults to "community". Any error produced while flushing the
// output is returned.
func runPluginListAvailable(cmd *cobra.Command) error {
	entries, err := registry.GetAllPluginEntries()
	if err != nil {
		return fmt.Errorf("loading registry: %w", err)
	}

	if len(entries) == 0 {
		cmd.Println("No plugins available in registry.")
		return nil
	}

	const tabPadding = 2
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, tabPadding, ' ', 0)

	fmt.Fprintln(w, "Name\tDescription\tRepository\tSecurity")
	fmt.Fprintln(w, "----\t-----------\t----------\t--------")

	for _, entry := range entries {
		security := entry.SecurityLevel
		if security == "" {
			security = "community"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", entry.Name, entry.Description, entry.Repository, security)
	}
	return w.Flush()
}

// runPluginListCmd lists installed plugins and writes a tabulated listing to the provided Cobra command output.
// It checks whether the configured plugin directory exists and prints a message and returns nil if it does not.
// If no plugins are installed it prints 'No plugins found.' and returns nil.
// cmd is the Cobra command used for printing. verbose controls whether plugin details are shown.
// Returns an error if querying the registry for installed plugins fails; otherwise nil.
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

	return displayPlugins(cmd, plugins, verbose)
}

func displayPlugins(cmd *cobra.Command, plugins []registry.PluginInfo, verbose bool) error {
	const tabPadding = 2
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, tabPadding, ' ', 0)

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