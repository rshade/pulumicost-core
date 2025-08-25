package cli

import (
	"fmt"
	"sort"
	"text/tabwriter"
	
	"github.com/spf13/cobra"
	"github.com/rshade/pulumicost-core/internal/config"
)

// NewConfigListCmd creates the 'config list' command
func NewConfigListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all configuration values",
		Long: `List all configuration values in a formatted table.

Credentials are shown as <encrypted> for security.`,
		Example: `  # List all configuration
  pulumicost config list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}
			
			configMap := cfg.ListAll()
			
			if len(configMap) == 0 {
				cmd.Println("No configuration found. Use 'pulumicost config init' to create a default configuration.")
				return nil
			}
			
			// Create a tabwriter for formatted output
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			defer w.Flush()
			
			fmt.Fprintf(w, "KEY\tVALUE\n")
			fmt.Fprintf(w, "---\t-----\n")
			
			// Sort keys for consistent output
			keys := make([]string, 0, len(configMap))
			for key := range configMap {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			
			for _, key := range keys {
				value := configMap[key]
				fmt.Fprintf(w, "%s\t%v\n", key, value)
			}
			
			return nil
		},
	}
	
	return cmd
}