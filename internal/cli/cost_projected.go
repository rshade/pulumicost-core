package cli

import (
	"context"
	"fmt"

	"github.com/rshade/pulumicost-core/internal/config"
	"github.com/rshade/pulumicost-core/internal/engine"
	"github.com/rshade/pulumicost-core/internal/ingest"
	"github.com/rshade/pulumicost-core/internal/registry"
	"github.com/rshade/pulumicost-core/internal/spec"
	"github.com/spf13/cobra"
)

func NewCostProjectedCmd() *cobra.Command {
	var planPath, specDir, adapter, output, filter string

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
		RunE: func(_ *cobra.Command, _ []string) error {
			ctx := context.Background()

			plan, err := ingest.LoadPulumiPlan(planPath)
			if err != nil {
				return fmt.Errorf("loading Pulumi plan: %w", err)
			}

			pulumiResources := plan.GetResources()
			resources, err := ingest.MapResources(pulumiResources)
			if err != nil {
				return fmt.Errorf("mapping resources: %w", err)
			}

			// Apply resource filter if specified
			if filter != "" {
				resources = engine.FilterResources(resources, filter)
			}

			if specDir == "" {
				cfg := config.New()
				specDir = cfg.SpecDir
			}
			loader := spec.NewLoader(specDir)

			reg := registry.NewDefault()
			clients, cleanup, err := reg.Open(ctx, adapter)
			if err != nil {
				return fmt.Errorf("opening plugins: %w", err)
			}
			defer cleanup()

			eng := engine.New(clients, loader)
			results, err := eng.GetProjectedCost(ctx, resources)
			if err != nil {
				return fmt.Errorf("calculating projected costs: %w", err)
			}

			// Use configuration-aware output format selection
			finalOutput := config.GetOutputFormat(output)
			outputFormat := engine.OutputFormat(finalOutput)
			return engine.RenderResults(outputFormat, results)
		},
	}

	cmd.Flags().StringVar(&planPath, "pulumi-json", "", "Path to Pulumi preview JSON output (required)")
	cmd.Flags().StringVar(&specDir, "spec-dir", "", "Directory containing pricing spec files")
	cmd.Flags().StringVar(&adapter, "adapter", "", "Use only the specified adapter plugin")

	// Use configuration default if no output format specified
	defaultFormat := config.GetDefaultOutputFormat()
	cmd.Flags().StringVar(&output, "output", defaultFormat, "Output format: table, json, or ndjson")
	cmd.Flags().StringVar(&filter, "filter", "", "Resource filter expressions (e.g., 'type=aws:ec2/instance')")

	_ = cmd.MarkFlagRequired("pulumi-json")

	return cmd
}
