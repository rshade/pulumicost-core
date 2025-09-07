package cli

import (
	"encoding/json"
	"fmt"

	"github.com/rshade/pulumicost-core/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewConfigListCmd() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all configuration values",
		Long:  "Lists all configuration values from ~/.pulumicost/config.yaml in the specified format.",
		Example: `  # List all configuration (default YAML format)
  pulumicost config list
  
  # List configuration in JSON format
  pulumicost config list --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {

			cfg := config.New()

			// Load existing config
			if err := cfg.Load(); err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Get all configuration
			allConfig := cfg.List()

			// Format and output based on requested format
			switch format {
			case "json":
				jsonData, err := json.MarshalIndent(allConfig, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal config to JSON: %w", err)
				}
				cmd.Printf("%s\n", jsonData)

			case "yaml", "yml":
				yamlData, err := yaml.Marshal(allConfig)
				if err != nil {
					return fmt.Errorf("failed to marshal config to YAML: %w", err)
				}
				cmd.Printf("%s", yamlData)

			default:
				return fmt.Errorf("unsupported format: %s (supported: json, yaml, yml)", format)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "yaml", "output format (yaml, json)")

	return cmd
}
