package cli

import (
	"github.com/spf13/cobra"
)

// NewAnalyzerCmd creates the analyzer command group for Pulumi Analyzer plugin functionality.
//
// The analyzer command provides subcommands for running PulumiCost as a Pulumi Analyzer plugin.
// NewAnalyzerCmd creates the "analyzer" Cobra command group for the Pulumi Analyzer plugin.
// The command provides help text and examples for running PulumiCost as an analyzer (for
// example, the "serve" subcommand) so the analyzer can perform cost estimation during
// `pulumi preview` operations. It returns a configured *cobra.Command containing those subcommands.
func NewAnalyzerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyzer",
		Short: "Pulumi Analyzer plugin commands",
		Long: `Commands for running PulumiCost as a Pulumi Analyzer plugin.

The analyzer plugin provides cost estimation during pulumi preview operations.
It communicates with the Pulumi engine via gRPC and returns cost diagnostics
that appear in the CLI output.`,
		Example: `  # Start the analyzer server (used by Pulumi engine)
  pulumicost analyzer serve

  # Start with debug logging
  pulumicost analyzer serve --debug`,
	}

	cmd.AddCommand(NewAnalyzerServeCmd())

	return cmd
}