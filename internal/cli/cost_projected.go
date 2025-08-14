package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/rshade/pulumicost-core/internal/config"
	"github.com/rshade/pulumicost-core/internal/engine"
	"github.com/rshade/pulumicost-core/internal/ingest"
	"github.com/rshade/pulumicost-core/internal/registry"
	"github.com/rshade/pulumicost-core/internal/spec"
)

func newCostProjectedCmd() *cobra.Command {
	var planPath, specDir, adapter, output string
	
	cmd := &cobra.Command{
		Use:   "projected",
		Short: "Calculate projected costs from a Pulumi plan",
		Long:  "Calculate projected costs by analyzing a Pulumi preview JSON output",
		RunE: func(cmd *cobra.Command, args []string) error {
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
			
			outputFormat := engine.OutputFormat(output)
			return engine.RenderResults(outputFormat, results)
		},
	}
	
	cmd.Flags().StringVar(&planPath, "pulumi-json", "", "Path to Pulumi preview JSON output (required)")
	cmd.Flags().StringVar(&specDir, "spec-dir", "", "Directory containing pricing spec files")
	cmd.Flags().StringVar(&adapter, "adapter", "", "Use only the specified adapter plugin")
	cmd.Flags().StringVar(&output, "output", "table", "Output format: table, json, or ndjson")
	
	cmd.MarkFlagRequired("pulumi-json")
	
	return cmd
}
