package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/rshade/pulumicost-core/internal/config"
	"github.com/rshade/pulumicost-core/internal/registry"
	"github.com/spf13/cobra"
)

func newPluginListCmd() *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed plugins",
		Long:  "List all installed plugins with their versions and paths",
		Example: `  # List all installed plugins
  pulumicost plugin list

  # List plugins with detailed information
  pulumicost plugin list --verbose`,
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg := config.New()
			if _, err := os.Stat(cfg.PluginDir); os.IsNotExist(err) {
				fmt.Printf("Plugin directory does not exist: %s\n", cfg.PluginDir)
				fmt.Println("No plugins installed.")
				return nil
			}

			reg := registry.NewDefault()
			plugins, err := reg.ListPlugins()
			if err != nil {
				return fmt.Errorf("listing plugins: %w", err)
			}

			if len(plugins) == 0 {
				fmt.Println("No plugins found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			if verbose {
				fmt.Fprintln(w, "Name\tVersion\tPath\tExecutable")
				fmt.Fprintln(w, "----\t-------\t----\t----------")

				for _, plugin := range plugins {
					// Check if plugin is executable
					execStatus := "No"
					if info, err := os.Stat(plugin.Path); err == nil {
						if info.Mode()&0111 != 0 {
							execStatus = "Yes"
						}
					}
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", plugin.Name, plugin.Version, plugin.Path, execStatus)
				}
			} else {
				fmt.Fprintln(w, "Name\tVersion\tPath")
				fmt.Fprintln(w, "----\t-------\t----")

				for _, plugin := range plugins {
					fmt.Fprintf(w, "%s\t%s\t%s\n", plugin.Name, plugin.Version, plugin.Path)
				}
			}

			return w.Flush()
		},
	}

	cmd.Flags().BoolVar(&verbose, "verbose", false, "Show detailed plugin information")

	return cmd
}
