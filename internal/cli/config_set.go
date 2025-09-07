package cli

import (
	"fmt"

	"github.com/rshade/pulumicost-core/internal/config"
	"github.com/spf13/cobra"
)

func NewConfigSetCmd() *cobra.Command {
	var encrypt bool
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long:  "Sets a configuration value using dot notation. The configuration will be saved to ~/.pulumicost/config.yaml.",
		Example: `  # Set output format
  pulumicost config set output.default_format json
  
  # Set output precision
  pulumicost config set output.precision 4
  
  # Set plugin configuration
  pulumicost config set plugins.aws.region us-west-2
  pulumicost config set plugins.aws.profile production
  
  # Set logging level
  pulumicost config set logging.level debug
  
  # Set encrypted credential (sensitive values)
  pulumicost config set plugins.aws.secret_key "mysecret" --encrypt`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]
			
			cfg := config.New()
			
			var displayValue string
			
			// Encrypt value if requested
			if encrypt {
				encryptedValue, err := cfg.EncryptValue(value)
				if err != nil {
					return fmt.Errorf("failed to encrypt value: %w", err)
				}
				value = encryptedValue
				displayValue = "[encrypted]"
				cmd.Printf("Value encrypted before storage\n")
			} else {
				displayValue = value
			}
			
			// Set the value
			if err := cfg.Set(key, value); err != nil {
				return fmt.Errorf("failed to set config value: %w", err)
			}
			
			// Validate the configuration
			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("configuration validation failed: %w", err)
			}
			
			// Save the configuration
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			
			cmd.Printf("Configuration updated: %s = %s\n", key, displayValue)
			
			return nil
		},
	}
	
	cmd.Flags().BoolVar(&encrypt, "encrypt", false, "encrypt the value before storing (for sensitive data)")
	
	return cmd
}