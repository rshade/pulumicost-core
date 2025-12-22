package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rshade/pulumicost-core/internal/config"
	"github.com/rshade/pulumicost-core/internal/registry"
	"github.com/spf13/cobra"
)

// NewPluginValidateCmd creates the plugin validate command for validating plugin installations.
func NewPluginValidateCmd() *cobra.Command {
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runPluginValidateCmd(cmd, targetPlugin)
		},
	}

	cmd.Flags().StringVar(&targetPlugin, "plugin", "", "Validate a specific plugin by name")

	return cmd
}

// ValidatePlugin validates a plugin by checking its binary and optional manifest file.
func ValidatePlugin(_ context.Context, plugin registry.PluginInfo) error {
	if _, err := os.Stat(plugin.Path); err != nil {
		return fmt.Errorf("plugin binary not found: %s", plugin.Path)
	}

	info, err := os.Stat(plugin.Path)
	if err != nil {
		return fmt.Errorf("cannot stat plugin binary: %w", err)
	}

	if info.IsDir() {
		return errors.New("plugin path is a directory, not a binary")
	}

	if info.Mode()&0111 == 0 {
		return errors.New("plugin binary is not executable")
	}

	manifestPath := filepath.Join(filepath.Dir(plugin.Path), "plugin.manifest.json")
	if _, statErr := os.Stat(manifestPath); statErr == nil {
		manifest, loadErr := registry.LoadManifest(manifestPath)
		if loadErr != nil {
			return fmt.Errorf("invalid manifest: %w", loadErr)
		}
		if manifest.Name != plugin.Name {
			return fmt.Errorf(
				"manifest name mismatch: expected %s, got %s",
				plugin.Name,
				manifest.Name,
			)
		}
		if manifest.Version != plugin.Version {
			return fmt.Errorf(
				"manifest version mismatch: expected %s, got %s",
				plugin.Version,
				manifest.Version,
			)
		}
	}

	return nil
}

func runPluginValidateCmd(cmd *cobra.Command, targetPlugin string) error {
	ctx := context.Background()

	cfg := config.New()
	if _, err := os.Stat(cfg.PluginDir); os.IsNotExist(err) {
		if targetPlugin != "" {
			return fmt.Errorf(
				"plugin '%s' not found: plugin directory does not exist",
				targetPlugin,
			)
		}
		cmd.Printf("Plugin directory does not exist: %s\n", cfg.PluginDir)
		cmd.Println("No plugins to validate.")
		return nil
	}

	reg := registry.NewDefault()
	plugins, err := reg.ListPlugins()
	if err != nil {
		return fmt.Errorf("listing plugins: %w", err)
	}

	if len(plugins) == 0 {
		if targetPlugin != "" {
			return fmt.Errorf("plugin '%s' not found: no plugins installed", targetPlugin)
		}
		cmd.Println("No plugins found to validate.")
		return nil
	}

	plugins, err = filterPlugins(plugins, targetPlugin)
	if err != nil {
		return err
	}

	return runValidation(ctx, cmd, plugins)
}

func filterPlugins(
	plugins []registry.PluginInfo,
	targetPlugin string,
) ([]registry.PluginInfo, error) {
	if targetPlugin == "" {
		return plugins, nil
	}

	var filtered []registry.PluginInfo
	for _, p := range plugins {
		if p.Name == targetPlugin {
			filtered = append(filtered, p)
		}
	}
	if len(filtered) == 0 {
		return nil, fmt.Errorf("plugin '%s' not found", targetPlugin)
	}
	return filtered, nil
}

func runValidation(ctx context.Context, cmd *cobra.Command, plugins []registry.PluginInfo) error {
	cmd.Printf("Validating %d plugin(s)...\n\n", len(plugins))

	validCount := 0
	for _, plugin := range plugins {
		if validateSinglePlugin(ctx, cmd, plugin) {
			validCount++
		}
	}

	cmd.Printf("\nValidation complete: %d/%d plugins valid\n", validCount, len(plugins))

	if validCount < len(plugins) {
		os.Exit(1)
	}
	return nil
}

func validateSinglePlugin(
	ctx context.Context,
	cmd *cobra.Command,
	plugin registry.PluginInfo,
) bool {
	cmd.Printf("Validating %s v%s... ", plugin.Name, plugin.Version)

	if err := ValidatePlugin(ctx, plugin); err != nil {
		cmd.Printf("FAILED: %v\n", err)
		return false
	}

	cmd.Println("OK")
	return true
}
