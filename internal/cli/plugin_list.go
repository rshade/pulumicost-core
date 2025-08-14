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
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed plugins",
		Long:  "List all installed plugins with their versions and paths",
		RunE: func(cmd *cobra.Command, args []string) error {
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
			fmt.Fprintln(w, "Name\tVersion\tPath")
			fmt.Fprintln(w, "----\t-------\t----")

			for _, plugin := range plugins {
				fmt.Fprintf(w, "%s\t%s\t%s\n", plugin.Name, plugin.Version, plugin.Path)
			}

			return w.Flush()
		},
	}

	return cmd
}
