package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/rshade/pulumicost-core/internal/config"
	"github.com/rshade/pulumicost-core/internal/engine"
	"github.com/rshade/pulumicost-core/internal/ingest"
	"github.com/rshade/pulumicost-core/internal/logging"
	"github.com/rshade/pulumicost-core/internal/registry"
	"github.com/rshade/pulumicost-core/internal/spec"
	"github.com/spf13/cobra"
)

// displayErrorSummary prints an error summary to the command output.
// It only displays for table format since JSON/NDJSON formats include errors in their structure.
func displayErrorSummary(
	cmd *cobra.Command,
	resultWithErrors *engine.CostResultWithErrors,
	outputFormat engine.OutputFormat,
) {
	if resultWithErrors.HasErrors() && outputFormat == engine.OutputTable {
		cmd.Println() // Add blank line before error summary
		cmd.Println("ERRORS")
		cmd.Println("======")
		cmd.Print(resultWithErrors.ErrorSummary())
	}
}

// NewCostProjectedCmd creates the "projected" subcommand that calculates estimated costs from a Pulumi plan.
//
// NewCostProjectedCmd creates the "projected" subcommand for calculating estimated costs from a Pulumi preview JSON.
// The command loads a Pulumi plan, maps and optionally filters resources, loads pricing specs and adapter plugins,
// computes projected monthly costs (including per-resource errors), renders results in the configured output format,
// and emits structured audit entries for failures and for successful completion.
// Flags: --pulumi-json (required), --spec-dir, --adapter, --output, and --filter.
//
//nolint:funlen // CLI command setup requires many configuration steps
func NewCostProjectedCmd() *cobra.Command {
	var planPath, specDir, adapter, output, filter string
	var utilization float64

	cmd := &cobra.Command{
		Use:   "projected",
		Short: "Calculate projected costs from a Pulumi plan",
		Long:  "Calculate projected costs by analyzing a Pulumi preview JSON output",
		Example: `  # Basic usage
  pulumicost cost projected --pulumi-json plan.json

  # Filter resources by type
  pulumicost cost projected --pulumi-json plan.json --filter "type=aws:ec2/instance"

  # Output as JSON
  pulumicost cost projected --pulumi-json plan.json --output json

  # Use a specific adapter plugin
  pulumicost cost projected --pulumi-json plan.json --adapter aws-plugin

  # Use custom spec directory
  pulumicost cost projected --pulumi-json plan.json --spec-dir ./custom-specs`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			start := time.Now()
			ctx := cmd.Context()

			// Inject utilization into context
			ctx = context.WithValue(ctx, engine.ContextKeyUtilization, utilization)

			// Get logger from context (set by PersistentPreRunE)
			log := logging.FromContext(ctx)
			log.Debug().
				Ctx(ctx).
				Str("operation", "cost_projected").
				Str("plan_path", planPath).
				Msg("starting projected cost calculation")

			// Get audit logger and trace ID for audit entry
			auditLogger := logging.AuditLoggerFromContext(ctx)
			traceID := logging.TraceIDFromContext(ctx)
			auditParams := map[string]string{
				"pulumi_json": planPath,
				"output":      output,
			}
			if filter != "" {
				auditParams["filter"] = filter
			}

			plan, err := ingest.LoadPulumiPlanWithContext(ctx, planPath)
			if err != nil {
				log.Error().
					Ctx(ctx).
					Err(err).
					Str("plan_path", planPath).
					Msg("failed to load Pulumi plan")
				// Log audit entry for failure
				auditEntry := logging.NewAuditEntry("cost projected", traceID).
					WithParameters(auditParams).
					WithError(err.Error()).
					WithDuration(start)
				auditLogger.Log(ctx, *auditEntry)
				return fmt.Errorf("loading Pulumi plan: %w", err)
			}

			pulumiResources := plan.GetResourcesWithContext(ctx)
			resources, err := ingest.MapResources(pulumiResources)
			if err != nil {
				log.Error().Ctx(ctx).Err(err).Msg("failed to map resources")
				// Log audit entry for failure
				auditEntry := logging.NewAuditEntry("cost projected", traceID).
					WithParameters(auditParams).
					WithError(err.Error()).
					WithDuration(start)
				auditLogger.Log(ctx, *auditEntry)
				return fmt.Errorf("mapping resources: %w", err)
			}

			log.Debug().
				Ctx(ctx).
				Int("resource_count", len(resources)).
				Msg("resources loaded from plan")

			// Apply resource filter if specified
			if filter != "" {
				resources = engine.FilterResources(resources, filter)
				log.Debug().
					Ctx(ctx).
					Str("filter", filter).
					Int("filtered_count", len(resources)).
					Msg("applied resource filter")
			}

			if specDir == "" {
				cfg := config.New()
				specDir = cfg.SpecDir
			}
			loader := spec.NewLoader(specDir)

			reg := registry.NewDefault()
			clients, cleanup, err := reg.Open(ctx, adapter)
			if err != nil {
				log.Error().Ctx(ctx).Err(err).Str("adapter", adapter).Msg("failed to open plugins")
				// Log audit entry for failure
				auditEntry := logging.NewAuditEntry("cost projected", traceID).
					WithParameters(auditParams).
					WithError(err.Error()).
					WithDuration(start)
				auditLogger.Log(ctx, *auditEntry)
				return fmt.Errorf("opening plugins: %w", err)
			}
			defer cleanup()

			log.Debug().
				Ctx(ctx).
				Int("plugin_count", len(clients)).
				Msg("plugins opened")

			eng := engine.New(clients, loader)
			resultWithErrors, err := eng.GetProjectedCostWithErrors(ctx, resources)
			if err != nil {
				log.Error().Ctx(ctx).Err(err).Msg("failed to calculate projected costs")
				// Log audit entry for failure
				auditEntry := logging.NewAuditEntry("cost projected", traceID).
					WithParameters(auditParams).
					WithError(err.Error()).
					WithDuration(start)
				auditLogger.Log(ctx, *auditEntry)
				return fmt.Errorf("calculating projected costs: %w", err)
			}

			// Use configuration-aware output format selection
			finalOutput := config.GetOutputFormat(output)
			outputFormat := engine.OutputFormat(finalOutput)
			if renderErr := engine.RenderResults(cmd.OutOrStdout(), outputFormat, resultWithErrors.Results); renderErr != nil {
				return renderErr
			}

			displayErrorSummary(cmd, resultWithErrors, outputFormat)

			// Log completion with duration
			log.Info().
				Ctx(ctx).
				Str("operation", "cost_projected").
				Int("result_count", len(resultWithErrors.Results)).
				Dur("duration_ms", time.Since(start)).
				Msg("projected cost calculation complete")

			// Log audit entry for success
			totalCost := 0.0
			for _, r := range resultWithErrors.Results {
				totalCost += r.Monthly
			}
			auditEntry := logging.NewAuditEntry("cost projected", traceID).
				WithParameters(auditParams).
				WithSuccess(len(resultWithErrors.Results), totalCost).
				WithDuration(start)
			auditLogger.Log(ctx, *auditEntry)

			return nil
		},
	}

	cmd.Flags().
		StringVar(&planPath, "pulumi-json", "", "Path to Pulumi preview JSON output (required)")
	cmd.Flags().StringVar(&specDir, "spec-dir", "", "Directory containing pricing spec files")
	cmd.Flags().StringVar(&adapter, "adapter", "", "Use only the specified adapter plugin")

	// Use configuration default if no output format specified
	defaultFormat := config.GetDefaultOutputFormat()
	cmd.Flags().StringVar(&output, "output", defaultFormat, "Output format: table, json, or ndjson")
	cmd.Flags().
		StringVar(&filter, "filter", "", "Resource filter expressions (e.g., 'type=aws:ec2/instance')")
	cmd.Flags().
		Float64Var(&utilization, "utilization", 1.0, "Utilization rate for sustainability calculations (0.0 to 1.0)")

	_ = cmd.MarkFlagRequired("pulumi-json")

	return cmd
}
