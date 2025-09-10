package cli

import (
	"fmt"

	"github.com/rshade/pulumicost-core/internal/config"
	"github.com/spf13/cobra"
)

func NewConfigInitCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration file with default values",
		Long:  "Creates a new configuration file at ~/.pulumicost/config.yaml with default values.",
		Example: `  # Create default configuration
  pulumicost config init
  
  # Create default configuration, overwriting existing
  pulumicost config init --force`,
		RunE: func(cmd *cobra.Command, args []string) error {

			cfg := config.New()

			// Check if config already exists and force isn't set
			if !force {
				if err := cfg.Load(); err == nil {
					return fmt.Errorf("configuration file already exists, use --force to overwrite")
				}
			}

			// Save the default configuration
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}

			cmd.Printf("Configuration initialized successfully\n")
			cmd.Printf("Configuration file: ~/.pulumicost/config.yaml\n")

			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing configuration file")

	return cmd
}
