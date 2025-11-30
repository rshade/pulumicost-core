// Package cli implements the Cobra-based command-line interface for PulumiCost.
//
// The CLI provides the primary user interface with subcommands for:
//   - cost projected: Calculate projected costs from Pulumi preview JSON
//   - cost actual: Fetch actual historical costs with time ranges
//   - plugin list: List installed plugins
//   - plugin validate: Validate plugin installations
//
// # Usage Patterns
//
// Commands use RunE for proper error handling and cmd.Printf() for output.
// Date inputs support multiple formats including "2006-01-02" and RFC3339.
//
// # Configuration
//
// CLI flags take precedence over environment variables and config file settings.
// Debug output can be enabled with --debug flag or PULUMICOST_LOG_LEVEL=debug.
package cli
