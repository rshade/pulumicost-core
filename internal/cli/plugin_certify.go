package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/rshade/pulumicost-core/internal/conformance"
	"github.com/rshade/pulumicost-core/internal/logging"
	"github.com/spf13/cobra"
)

// Certification timeout and output constants.
const (
	certificationTimeout     = 10 * time.Minute
	certificationOutputFile  = "certification.md"
	certificationExitFailure = 1
)

// NewPluginCertifyCmd creates the plugin certify command for running certification
// tests against a plugin binary and generating a certification report.
func NewPluginCertifyCmd() *cobra.Command {
	var (
		outputFile string
		mode       string
		timeout    string
	)

	cmd := &cobra.Command{
		Use:   "certify <plugin-path>",
		Short: "Run certification tests for a plugin",
		Long: `Runs the full conformance suite and generates a certification report.

Certification requires 100% of conformance tests to pass. The command outputs
a markdown certification report that can be distributed with your plugin.

A plugin is certified if:
- All protocol tests pass (Name, GetProjectedCost, GetActualCost)
- All error handling tests pass
- All context/timeout tests pass
- All performance tests pass`,
		Example: `  # Basic certification
  pulumicost plugin certify ./plugins/aws-cost

  # Save report to custom file
  pulumicost plugin certify --output certification-report.md ./plugins/aws-cost

  # Use stdio mode
  pulumicost plugin certify --mode stdio ./plugins/aws-cost`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPluginCertifyCmd(cmd, args[0], outputFile, mode, timeout)
		},
	}

	cmd.Flags().StringVarP(
		&outputFile, "output", "o", "",
		"Output file for certification report (default: stdout)",
	)
	cmd.Flags().StringVar(&mode, "mode", "tcp", "Communication mode: tcp, stdio")
	cmd.Flags().StringVar(&timeout, "timeout", "10m", "Global certification timeout")

	return cmd
}

func runPluginCertifyCmd(
	cmd *cobra.Command,
	pluginPath, outputFile, mode, timeout string,
) error {
	ctx := cmd.Context()

	// Validate plugin path exists
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin not found: %s", pluginPath)
	}

	// Parse timeout
	timeoutDuration, err := time.ParseDuration(timeout)
	if err != nil {
		return fmt.Errorf("invalid timeout: %w", err)
	}

	// Validate mode
	if !conformance.IsValidCommMode(mode) {
		return fmt.Errorf("invalid mode %q: must be tcp or stdio", mode)
	}

	cmd.Printf("🔍 Certifying plugin at %s...\n", pluginPath)

	// Get logger from context
	logger := logging.FromContext(ctx)

	// Create suite config - run ALL tests for certification
	cfg := conformance.SuiteConfig{
		PluginPath: pluginPath,
		CommMode:   conformance.CommMode(mode),
		Verbosity:  conformance.VerbosityNormal,
		Timeout:    timeoutDuration,
		Categories: nil, // Run all categories
		TestFilter: "",  // No filter - run all tests
		Logger:     *logger,
	}

	// Create and run suite
	suite, err := conformance.NewSuite(cfg)
	if err != nil {
		return fmt.Errorf("creating conformance suite: %w", err)
	}

	cmd.Println("Running conformance tests...")

	report, err := suite.Run(ctx)
	if err != nil {
		return fmt.Errorf("running conformance suite: %w", err)
	}

	// Generate certification report
	certReport := conformance.Certify(report, report.Plugin.Name, report.Plugin.Version)

	// Generate markdown
	markdown := certReport.GenerateMarkdown()

	// Write output
	if outputFile != "" {
		if writeErr := os.WriteFile(outputFile, []byte(markdown), 0o600); writeErr != nil {
			return fmt.Errorf("writing certification report: %w", writeErr)
		}
		cmd.Printf("📄 Certification report written to %s\n", outputFile)
	} else {
		cmd.Println()
		cmd.Println(markdown)
	}

	// Print summary
	if certReport.Certified {
		cmd.Println("✅ CERTIFIED - Plugin passed all conformance tests")
		return nil
	}

	cmd.Println("❌ NOT CERTIFIED - Plugin failed conformance tests")
	cmd.Printf("   Failed: %d, Errors: %d\n", report.Summary.Failed, report.Summary.Errors)

	return &exitError{
		code:    certificationExitFailure,
		message: "certification failed",
	}
}
