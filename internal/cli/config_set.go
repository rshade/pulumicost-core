package cli

import (
	"fmt"
	"strings"
	
	"github.com/spf13/cobra"
	"github.com/rshade/pulumicost-core/internal/config"
)

// NewConfigSetCmd creates the 'config set' command
func NewConfigSetCmd() *cobra.Command {
	var isCredential bool
	
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long: `Set a configuration value using dot notation.

Configuration keys:
  output.default_format  - Output format (table, json, ndjson)
  output.precision       - Decimal precision for numbers (0-10)
  logging.level          - Log level (debug, info, warn, error)  
  logging.file           - Log file path
  plugins.<name>.region  - Plugin region setting
  plugins.<name>.profile - Plugin profile setting
  
Credentials are encrypted and stored securely.`,
		Example: `  # Set output format to JSON
  pulumicost config set output.default_format json
  
  # Set AWS plugin region
  pulumicost config set plugins.aws.region us-west-2
  
  # Set encrypted credential
  pulumicost config set plugins.aws.access_key AKIA... --credential`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]
			
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}
			
			// Handle credential setting
			if isCredential {
				parts := strings.Split(key, ".")
				if len(parts) < 3 || parts[0] != "plugins" {
					return fmt.Errorf("credential key must be in format: plugins.<name>.<key>")
				}
				
				pluginName := parts[1]
				credentialKey := strings.Join(parts[2:], ".")
				
				if err := cfg.SetCredential(pluginName, credentialKey, value); err != nil {
					return fmt.Errorf("setting credential: %w", err)
				}
				
				cmd.Printf("Credential %s set for plugin %s (encrypted)\n", credentialKey, pluginName)
			} else {
				// Handle regular configuration setting
				if err := cfg.Set(key, value); err != nil {
					return fmt.Errorf("setting config value: %w", err)
				}
				
				cmd.Printf("Configuration %s set to %s\n", key, value)
			}
			
			// Save configuration
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}
			
			return nil
		},
	}
	
	cmd.Flags().BoolVar(&isCredential, "credential", false, "Encrypt and store value as a credential")
	
	return cmd
}