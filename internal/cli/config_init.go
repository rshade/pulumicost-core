package cli

import (
	"github.com/spf13/cobra"
	"github.com/rshade/pulumicost-core/internal/config"
)

// NewConfigInitCmd creates the 'config init' command
func NewConfigInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration file with defaults",
		Long:  "Create a default configuration file at ~/.pulumicost/config.yaml if it doesn't already exist.",
		Example: `  # Initialize configuration with defaults
  pulumicost config init`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.InitConfig(); err != nil {
				return err
			}
			
			cfg := config.DefaultConfig()
			cmd.Printf("Configuration initialized at %s\n", cfg.ConfigFile)
			return nil
		},
	}
	
	return cmd
}