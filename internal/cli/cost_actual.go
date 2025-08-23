package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rshade/pulumicost-core/internal/engine"
	"github.com/rshade/pulumicost-core/internal/ingest"
	"github.com/rshade/pulumicost-core/internal/registry"
	"github.com/spf13/cobra"
)

const (
	filterKeyValueParts = 2 // For "key=value" pairs
)

func NewCostActualCmd() *cobra.Command {
	var planPath, adapter, output, fromStr, toStr, groupBy string

	cmd := &cobra.Command{
		Use:   "actual",
		Short: "Fetch actual historical costs",
		Long:  "Fetch actual historical costs for resources from cloud provider billing APIs",
		Example: `  # Get costs for the last 7 days (to defaults to now)
  pulumicost cost actual --pulumi-json plan.json --from 2025-01-07

  # Get costs for a specific date range
  pulumicost cost actual --pulumi-json plan.json --from 2025-01-01 --to 2025-01-31

  # Group costs by resource type
  pulumicost cost actual --pulumi-json plan.json --from 2025-01-01 --group-by type

  # Output as JSON with grouping by provider
  pulumicost cost actual --pulumi-json plan.json --from 2025-01-01 --output json --group-by provider

  # Use RFC3339 timestamps
  pulumicost cost actual --pulumi-json plan.json --from 2025-01-01T00:00:00Z --to 2025-01-31T23:59:59Z`,
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

			// Default to now if --to is not provided
			if toStr == "" {
				toStr = time.Now().Format(time.RFC3339)
			}

			from, to, err := ParseTimeRange(fromStr, toStr)
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

			// Parse tags from groupBy if it's in tag:key=value format
			tags := make(map[string]string)
			actualGroupBy := groupBy
			if strings.HasPrefix(groupBy, "tag:") && strings.Contains(groupBy, "=") {
				tagPart := strings.TrimPrefix(groupBy, "tag:")
				if parts := strings.Split(tagPart, "="); len(parts) == filterKeyValueParts {
					tags[parts[0]] = parts[1]
					actualGroupBy = "" // Clear groupBy since we're filtering by tag
				}
			}

			request := engine.ActualCostRequest{
				Resources: resources,
				From:      from,
				To:        to,
				Adapter:   adapter,
				GroupBy:   actualGroupBy,
				Tags:      tags,
			}

			results, err := eng.GetActualCostWithOptions(ctx, request)
			if err != nil {
				return fmt.Errorf("fetching actual costs: %w", err)
			}

			outputFormat := engine.OutputFormat(output)
			return engine.RenderActualCostResults(outputFormat, results)
		},
	}

	cmd.Flags().StringVar(&planPath, "pulumi-json", "", "Path to Pulumi preview JSON output (required)")
	cmd.Flags().StringVar(&fromStr, "from", "", "Start date (YYYY-MM-DD or RFC3339) (required)")
	cmd.Flags().StringVar(&toStr, "to", "", "End date (YYYY-MM-DD or RFC3339) (defaults to now)")
	cmd.Flags().StringVar(&adapter, "adapter", "", "Use only the specified adapter plugin")
	cmd.Flags().StringVar(&output, "output", "table", "Output format: table, json, or ndjson")
	cmd.Flags().
		StringVar(&groupBy, "group-by", "", "Group results by: resource, type, provider, date, or filter by tag:key=value")

	_ = cmd.MarkFlagRequired("pulumi-json")
	_ = cmd.MarkFlagRequired("from")

	return cmd
}

func ParseTimeRange(fromStr, toStr string) (time.Time, time.Time, error) {
	from, err := ParseTime(fromStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("parsing 'from' date: %w", err)
	}

	to, err := ParseTime(toStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("parsing 'to' date: %w", err)
	}

	if !to.After(from) {
		return time.Time{}, time.Time{}, errors.New("'to' date must be after 'from' date")
	}

	return from, to, nil
}

func ParseTime(str string) (time.Time, error) {
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
