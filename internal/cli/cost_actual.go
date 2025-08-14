package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/rshade/pulumicost-core/internal/engine"
	"github.com/rshade/pulumicost-core/internal/ingest"
	"github.com/rshade/pulumicost-core/internal/registry"
)

func newCostActualCmd() *cobra.Command {
	var planPath, adapter, output, fromStr, toStr string
	
	cmd := &cobra.Command{
		Use:   "actual",
		Short: "Fetch actual historical costs",
		Long:  "Fetch actual historical costs for resources from cloud provider billing APIs",
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
			
			from, to, err := parseTimeRange(fromStr, toStr)
			if err != nil {
				return fmt.Errorf("parsing time range: %w", err)
			}
			
			reg := registry.NewDefault()
			clients, cleanup, err := reg.Open(ctx, adapter)
			if err != nil {
				return fmt.Errorf("opening plugins: %w", err)
			}
			defer cleanup()
			
			eng := engine.New(clients, nil)
			results, err := eng.GetActualCost(ctx, resources, from, to)
			if err != nil {
				return fmt.Errorf("fetching actual costs: %w", err)
			}
			
			outputFormat := engine.OutputFormat(output)
			return engine.RenderResults(outputFormat, results)
		},
	}
	
	cmd.Flags().StringVar(&planPath, "pulumi-json", "", "Path to Pulumi preview JSON output (required)")
	cmd.Flags().StringVar(&fromStr, "from", "", "Start date (YYYY-MM-DD or RFC3339) (required)")
	cmd.Flags().StringVar(&toStr, "to", "", "End date (YYYY-MM-DD or RFC3339) (required)")
	cmd.Flags().StringVar(&adapter, "adapter", "", "Use only the specified adapter plugin")
	cmd.Flags().StringVar(&output, "output", "table", "Output format: table, json, or ndjson")
	
	cmd.MarkFlagRequired("pulumi-json")
	cmd.MarkFlagRequired("from")
	cmd.MarkFlagRequired("to")
	
	return cmd
}

func parseTimeRange(fromStr, toStr string) (time.Time, time.Time, error) {
	from, err := parseTime(fromStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("parsing 'from' date: %w", err)
	}
	
	to, err := parseTime(toStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("parsing 'to' date: %w", err)
	}
	
	if to.Before(from) {
		return time.Time{}, time.Time{}, fmt.Errorf("'to' date must be after 'from' date")
	}
	
	return from, to, nil
}

func parseTime(str string) (time.Time, error) {
	layouts := []string{
		"2006-01-02",
		time.RFC3339,
		"2006-01-02T15:04:05Z",
	}
	
	for _, layout := range layouts {
		t, err := time.Parse(layout, str)
		if err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse date: %s (use YYYY-MM-DD or RFC3339)", str)
}
