package cli

import (
	"fmt"

	"github.com/rshade/pulumicost-core/internal/config"
	"github.com/spf13/cobra"
)

func NewConfigValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration file",
		Long:  "Validates the configuration file at ~/.pulumicost/config.yaml for syntax and semantic correctness.",
		Example: `  # Validate current configuration
  pulumicost config validate
  
  # Validate and show detailed information
  pulumicost config validate --verbose`,
		RunE: func(cmd *cobra.Command, args []string) error {
			verbose, _ := cmd.Flags().GetBool("verbose")
			
			cfg := config.New()
			
			// Load existing config
			if err := cfg.Load(); err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			
			// Validate configuration
			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("configuration validation failed: %w", err)
			}
			
			cmd.Printf("✅ Configuration is valid\n")
			
			if verbose {
				cmd.Printf("\nConfiguration details:\n")
				cmd.Printf("- Output format: %s\n", cfg.Output.DefaultFormat)
				cmd.Printf("- Output precision: %d\n", cfg.Output.Precision)
				cmd.Printf("- Logging level: %s\n", cfg.Logging.Level)
				cmd.Printf("- Log file: %s\n", cfg.Logging.File)
				
				if len(cfg.Plugins) > 0 {
					cmd.Printf("- Configured plugins: %d\n", len(cfg.Plugins))
					for pluginName := range cfg.Plugins {
						cmd.Printf("  - %s\n", pluginName)
					}
				} else {
					cmd.Printf("- No plugins configured\n")
				}
			}
			
			return nil
		},
	}
	
	cmd.Flags().BoolP("verbose", "v", false, "show detailed validation information")
	
	return cmd
}