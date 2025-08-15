package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rshade/pulumicost-core/internal/config"
	"github.com/rshade/pulumicost-core/internal/registry"
	"github.com/spf13/cobra"
)

func newPluginValidateCmd() *cobra.Command {
	var targetPlugin string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate installed plugins",
		Long:  "Validate that installed plugins can be loaded and respond to basic API calls",
		Example: `  # Validate all installed plugins
  pulumicost plugin validate

  # Validate a specific plugin
  pulumicost plugin validate --plugin aws-plugin

  # Validate kubecost plugin specifically
  pulumicost plugin validate --plugin kubecost`,
		RunE: func(_ *cobra.Command, _ []string) error {
			ctx := context.Background()

			cfg := config.New()
			if _, err := os.Stat(cfg.PluginDir); os.IsNotExist(err) {
				fmt.Printf("Plugin directory does not exist: %s\n", cfg.PluginDir)
				fmt.Println("No plugins to validate.")
				return nil
			}

			reg := registry.NewDefault()
			plugins, err := reg.ListPlugins()
			if err != nil {
				return fmt.Errorf("listing plugins: %w", err)
			}

			if len(plugins) == 0 {
				fmt.Println("No plugins found to validate.")
				return nil
			}

			// Filter plugins if --plugin flag is specified
			if targetPlugin != "" {
				var filtered []registry.PluginInfo
				for _, p := range plugins {
					if p.Name == targetPlugin {
						filtered = append(filtered, p)
					}
				}
				if len(filtered) == 0 {
					return fmt.Errorf("plugin '%s' not found", targetPlugin)
				}
				plugins = filtered
			}

			fmt.Printf("Validating %d plugin(s)...\n\n", len(plugins))

			validCount := 0
			for _, plugin := range plugins {
				fmt.Printf("Validating %s v%s... ", plugin.Name, plugin.Version)

				if err := validatePlugin(ctx, plugin); err != nil {
					fmt.Printf("FAILED: %v\n", err)
				} else {
					fmt.Println("OK")
					validCount++
				}
			}

			fmt.Printf("\nValidation complete: %d/%d plugins valid\n", validCount, len(plugins))

			if validCount < len(plugins) {
				os.Exit(1)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&targetPlugin, "plugin", "", "Validate a specific plugin by name")

	return cmd
}

func validatePlugin(_ context.Context, plugin registry.PluginInfo) error {
	if _, err := os.Stat(plugin.Path); err != nil {
		return fmt.Errorf("plugin binary not found: %s", plugin.Path)
	}

	info, err := os.Stat(plugin.Path)
	if err != nil {
		return fmt.Errorf("cannot stat plugin binary: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("plugin path is a directory, not a binary")
	}

	if info.Mode()&0111 == 0 {
		return fmt.Errorf("plugin binary is not executable")
	}

	manifestPath := filepath.Join(filepath.Dir(plugin.Path), "plugin.manifest.json")
	if _, err := os.Stat(manifestPath); err == nil {
		manifest, err := registry.LoadManifest(manifestPath)
		if err != nil {
			return fmt.Errorf("invalid manifest: %w", err)
		}
		if manifest.Name != plugin.Name {
			return fmt.Errorf("manifest name mismatch: expected %s, got %s", plugin.Name, manifest.Name)
		}
		if manifest.Version != plugin.Version {
			return fmt.Errorf("manifest version mismatch: expected %s, got %s", plugin.Version, manifest.Version)
		}
	}

	return nil
}
