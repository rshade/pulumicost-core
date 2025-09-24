package cli

import (
	"errors"
	"fmt"
	"sort"

	"github.com/rshade/pulumicost-core/internal/config"
	"github.com/spf13/cobra"
)

func NewConfigGetCmd() *cobra.Command {
	var decrypt bool
	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Long:  "Gets a configuration value using dot notation from ~/.pulumicost/config.yaml.",
		Example: `  # Get output format
  pulumicost config get output.default_format
  
  # Get output precision
  pulumicost config get output.precision
  
  # Get plugin configuration
  pulumicost config get plugins.aws.region
  pulumicost config get plugins.aws
  
  # Get all plugins
  pulumicost config get plugins
  
  # Get logging level
  pulumicost config get logging.level
  
  # Decrypt encrypted value
  pulumicost config get plugins.aws.secret_key --decrypt`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			// config.New() already loads from disk and applies env overrides
			cfg := config.New()

			// Get the value
			value, err := cfg.Get(key)
			if err != nil {
				return fmt.Errorf("failed to get config value: %w", err)
			}

			// Decrypt value if requested and it's a string
			if decrypt {
				if strValue, ok := value.(string); ok {
					decryptedValue, decryptErr := cfg.DecryptValue(strValue)
					if decryptErr != nil {
						return fmt.Errorf("failed to decrypt value: %w", decryptErr)
					}
					value = decryptedValue
				} else {
					return errors.New("can only decrypt string values")
				}
			}

			// Format and output the value
			formatAndPrintValue(cmd, key, value)

			return nil
		},
	}

	cmd.Flags().BoolVar(&decrypt, "decrypt", false, "decrypt the value if it's encrypted")

	return cmd
}

// formatAndPrintValue formats and prints configuration values based on their type.
func formatAndPrintValue(cmd *cobra.Command, key string, value interface{}) {
	switch v := value.(type) {
	case string:
		cmd.Printf("%s\n", v)
	case int:
		cmd.Printf("%d\n", v)
	case map[string]interface{}:
		cmd.Printf("%s:\n", key)
		// Sort keys for deterministic output
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, subKey := range keys {
			cmd.Printf("  %s: %v\n", subKey, v[subKey])
		}
	case map[string]config.PluginConfig:
		cmd.Printf("%s:\n", key)
		// Sort plugin names for deterministic output
		pluginNames := make([]string, 0, len(v))
		for name := range v {
			pluginNames = append(pluginNames, name)
		}
		sort.Strings(pluginNames)
		for _, pluginName := range pluginNames {
			pluginConfig := v[pluginName]
			cmd.Printf("  %s:\n", pluginName)
			// Sort config keys for deterministic output
			configKeys := make([]string, 0, len(pluginConfig.Config))
			for k := range pluginConfig.Config {
				configKeys = append(configKeys, k)
			}
			sort.Strings(configKeys)
			for _, configKey := range configKeys {
				cmd.Printf("    %s: %v\n", configKey, pluginConfig.Config[configKey])
			}
		}
	default:
		cmd.Printf("%v\n", v)
	}
}
