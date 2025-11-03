package cli

import (
	"fmt"

	"github.com/rshade/pulumicost-core/internal/config"
	"github.com/spf13/cobra"
)

// NewConfigSetCmd creates the config set command for setting configuration values.
func NewConfigSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long: `Sets a configuration value using dot notation. The configuration will be saved to ~/.pulumicost/config.yaml.

For sensitive values like API keys or credentials, use environment variables instead:
  export PULUMICOST_PLUGIN_AWS_SECRET_KEY="mysecret"
  export PULUMICOST_PLUGIN_AZURE_CLIENT_SECRET="secret"`,
		Example: `  # Set output format
  pulumicost config set output.default_format json

  # Set output precision
  pulumicost config set output.precision 4

  # Set plugin configuration
  pulumicost config set plugins.aws.region us-west-2
  pulumicost config set plugins.aws.profile production

  # Set logging level
  pulumicost config set logging.level debug

  # For sensitive values, use environment variables instead
  export PULUMICOST_PLUGIN_AWS_SECRET_KEY="mysecret"`,
		Args: cobra.ExactArgs(2), //nolint:mnd // Exactly 2 args: key and value
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			cfg := config.New()

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

			cmd.Printf("Configuration updated: %s = %s\n", key, value)

			return nil
		},
	}

	return cmd
}
