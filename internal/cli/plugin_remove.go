package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rshade/pulumicost-core/internal/registry"
)

// NewPluginRemoveCmd creates the plugin remove command.
func NewPluginRemoveCmd() *cobra.Command {
	var (
		keepConfig bool
		pluginDir  string
	)

	cmd := &cobra.Command{
		Use:     "remove <plugin>",
		Aliases: []string{"uninstall", "rm"},
		Short:   "Remove an installed plugin",
		Long: `Remove an installed plugin from the system.

This will delete the plugin files and remove it from the configuration.`,
		Example: `  # Remove a plugin
  pulumicost plugin remove kubecost

  # Remove but keep in config (for reinstalling later)
  pulumicost plugin remove kubecost --keep-config

  # Using alias
  pulumicost plugin uninstall kubecost`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Create installer
			installer := registry.NewInstaller(pluginDir)

			opts := registry.RemoveOptions{
				KeepConfig: keepConfig,
				PluginDir:  pluginDir,
			}

			// Progress callback
			progress := func(msg string) {
				cmd.Printf("%s\n", msg)
			}

			// Remove
			if err := installer.Remove(name, opts, progress); err != nil {
				return fmt.Errorf("removing plugin %q: %w", name, err)
			}

			cmd.Printf("\n✓ Plugin %s removed successfully\n", name)

			return nil
		},
	}

	cmd.Flags().BoolVar(&keepConfig, "keep-config", false, "Keep plugin entry in config file")
	cmd.Flags().StringVar(&pluginDir, "plugin-dir", "", "Custom plugin directory")

	return cmd
}
