package cli

import (
	"fmt"

	"github.com/rshade/pulumicost-core/internal/config"
	"github.com/spf13/cobra"
)

// NewConfigValidateCmd creates the 'config validate' command
func NewConfigValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration file",
		Long: `Validate the configuration file for syntax and semantic errors.

This command checks:
- YAML syntax validity
- Configuration schema compliance
- Value range validation
- Required field presence`,
		Example: `  # Validate current configuration
  pulumicost config validate`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}
			
			// Validate configuration
			if err := cfg.Validate(); err != nil {
				cmd.Printf("❌ Configuration validation failed: %v\n", err)
				return err
			}
			
			cmd.Println("✅ Configuration is valid")
			
			// Show configuration file location
			cmd.Printf("Configuration file: %s\n", cfg.ConfigFile)
			
			// Show summary of configuration
			configMap := cfg.ListAll()
			if len(configMap) > 0 {
				cmd.Printf("Found %d configuration settings\n", len(configMap))
			}
			
			return nil
		},
	}
	
	return cmd
}