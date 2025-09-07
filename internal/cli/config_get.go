package cli

import (
	"fmt"

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

			cfg := config.New()

			// Load existing config
			if err := cfg.Load(); err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Get the value
			value, err := cfg.Get(key)
			if err != nil {
				return fmt.Errorf("failed to get config value: %w", err)
			}

			// Decrypt value if requested and it's a string
			if decrypt {
				if strValue, ok := value.(string); ok {
					decryptedValue, err := cfg.DecryptValue(strValue)
					if err != nil {
						return fmt.Errorf("failed to decrypt value: %w", err)
					}
					value = decryptedValue
				} else {
					return fmt.Errorf("can only decrypt string values")
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

// formatAndPrintValue formats and prints configuration values based on their type
func formatAndPrintValue(cmd *cobra.Command, key string, value interface{}) {
	switch v := value.(type) {
	case string:
		cmd.Printf("%s\n", v)
	case int:
		cmd.Printf("%d\n", v)
	case map[string]interface{}:
		cmd.Printf("%s:\n", key)
		for subKey, subValue := range v {
			cmd.Printf("  %s: %v\n", subKey, subValue)
		}
	case map[string]config.PluginConfig:
		cmd.Printf("%s:\n", key)
		for pluginName, pluginConfig := range v {
			cmd.Printf("  %s:\n", pluginName)
			for configKey, configValue := range pluginConfig.Config {
				cmd.Printf("    %s: %v\n", configKey, configValue)
			}
		}
	default:
		cmd.Printf("%v\n", v)
	}
}
