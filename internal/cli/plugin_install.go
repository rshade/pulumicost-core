package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rshade/pulumicost-core/internal/registry"
)

// NewPluginInstallCmd creates the install command for installing plugins from registry or URL.
//
//	--plugin-dir    Custom plugin directory (default: ~/.pulumicost/plugins)
func NewPluginInstallCmd() *cobra.Command {
	var (
		force     bool
		noSave    bool
		pluginDir string
	)

	cmd := &cobra.Command{
		Use:   "install <plugin>",
		Short: "Install a plugin from registry or URL",
		Long: `Install a plugin from the registry or directly from a GitHub URL.

Plugins can be specified in several formats:
  - Registry name: kubecost
  - Registry name with version: kubecost@v1.0.0
  - GitHub URL: github.com/owner/repo
  - GitHub URL with version: github.com/owner/repo@v1.0.0`,
		Example: `  # Install latest version from registry
  pulumicost plugin install kubecost

  # Install specific version from registry
  pulumicost plugin install kubecost@v1.0.0

  # Install from GitHub URL
  pulumicost plugin install github.com/rshade/pulumicost-plugin-aws-public

  # Install specific version from URL
  pulumicost plugin install github.com/rshade/pulumicost-plugin-aws-public@v0.1.0

  # Force reinstall even if already installed
  pulumicost plugin install kubecost --force

  # Install without saving to config
  pulumicost plugin install kubecost --no-save`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			specifier := args[0]

			// Show security warning for URL installs
			spec, err := registry.ParsePluginSpecifier(specifier)
			if err != nil {
				return fmt.Errorf("parsing plugin specifier %q: %w", specifier, err)
			}

			if spec.IsURL {
				cmd.Printf("⚠️  Installing from URL: %s/%s\n", spec.Owner, spec.Repo)
				cmd.Printf("   URL-based plugins are not verified by the PulumiCost team.\n")
				cmd.Printf("   Only install from sources you trust.\n\n")
			} else {
				// Check security level for registry plugins
				entry, getErr := registry.GetPlugin(spec.Name)
				if getErr == nil && entry.SecurityLevel == "experimental" {
					cmd.Printf("⚠️  Plugin %q has security level: experimental\n", spec.Name)
					cmd.Printf("   This plugin is not yet fully reviewed.\n\n")
				}
			}

			// Create installer
			installer := registry.NewInstaller(pluginDir)

			opts := registry.InstallOptions{
				Force:     force,
				NoSave:    noSave,
				PluginDir: pluginDir,
			}

			// Progress callback
			progress := func(msg string) {
				cmd.Printf("%s\n", msg)
			}

			// Install
			result, err := installer.Install(specifier, opts, progress)
			if err != nil {
				return fmt.Errorf("installing plugin %q: %w", specifier, err)
			}

			cmd.Printf("\n✓ Plugin installed successfully\n")
			cmd.Printf("  Name:    %s\n", result.Name)
			cmd.Printf("  Version: %s\n", result.Version)
			cmd.Printf("  Path:    %s\n", result.Path)

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Reinstall even if version already exists")
	cmd.Flags().BoolVar(&noSave, "no-save", false, "Don't add plugin to config file")
	cmd.Flags().
		StringVar(&pluginDir, "plugin-dir", "", "Custom plugin directory (default: ~/.pulumicost/plugins)")

	return cmd
}
