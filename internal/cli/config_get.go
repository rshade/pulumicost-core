package cli

import (
	"fmt"
	"strings"
	
	"github.com/spf13/cobra"
	"github.com/rshade/pulumicost-core/internal/config"
)

// NewConfigGetCmd creates the 'config get' command
func NewConfigGetCmd() *cobra.Command {
	var isCredential bool
	
	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Long: `Get a configuration value using dot notation.

Configuration keys:
  output.default_format  - Output format
  output.precision       - Decimal precision
  logging.level          - Log level
  logging.file           - Log file path
  plugins.<name>.region  - Plugin region setting
  plugins.<name>.profile - Plugin profile setting
  
Use --credential flag to retrieve encrypted credentials.`,
		Example: `  # Get output format
  pulumicost config get output.default_format
  
  # Get AWS plugin region
  pulumicost config get plugins.aws.region
  
  # Get encrypted credential
  pulumicost config get plugins.aws.access_key --credential`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}
			
			// Handle credential retrieval
			if isCredential {
				parts := strings.Split(key, ".")
				if len(parts) < 3 || parts[0] != "plugins" {
					return fmt.Errorf("credential key must be in format: plugins.<name>.<key>")
				}
				
				pluginName := parts[1]
				credentialKey := strings.Join(parts[2:], ".")
				
				value, err := cfg.GetCredential(pluginName, credentialKey)
				if err != nil {
					return fmt.Errorf("getting credential: %w", err)
				}
				
				cmd.Println(value)
			} else {
				// Handle regular configuration retrieval
				value, err := cfg.Get(key)
				if err != nil {
					return fmt.Errorf("getting config value: %w", err)
				}
				
				cmd.Printf("%v\n", value)
			}
			
			return nil
		},
	}
	
	cmd.Flags().BoolVar(&isCredential, "credential", false, "Retrieve value as a decrypted credential")
	
	return cmd
}