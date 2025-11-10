package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/rshade/pulumicost-core/internal/config"
	"github.com/rshade/pulumicost-core/internal/engine"
	"github.com/rshade/pulumicost-core/internal/ingest"
	"github.com/rshade/pulumicost-core/internal/registry"
	"github.com/spf13/cobra"
)

const (
	filterKeyValueParts = 2 // For "key=value" pairs
)

// NewCostActualCmd returns a *cobra.Command that fetches actual historical costs from cloud provider billing APIs.
//
// The command accepts a Pulumi plan JSON and queries provider billing data for the specified time range.
// Flags:
//
//	--pulumi-json (required): path to Pulumi preview JSON output.
//	--from (required): start date in YYYY-MM-DD or RFC3339 format.
//	--to: end date in YYYY-MM-DD or RFC3339 format (defaults to now when omitted).
//	--adapter: restrict queries to a single adapter plugin.
//	--output: output format ("table", "json", or "ndjson"); defaults follow configuration.
//	--group-by: grouping mode: "resource", "type", "provider", "date", "daily", "monthly", or a tag filter of the form "tag:key=value".
//
// NewCostActualCmd creates the `actual` subcommand which fetches actual historical costs
// for resources declared in a Pulumi plan by querying cloud provider billing APIs.
// The command accepts a time range, supports grouping (resource, type, provider, date, daily, monthly)
// and tag-based filtering (`tag:key=value`), opens adapter plugins as needed, and renders results
// using the configured output format (table, json, or ndjson).
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

  # Daily cross-provider aggregation table
  pulumicost cost actual --pulumi-json plan.json --from 2025-01-01 --to 2025-01-07 --group-by daily

  # Monthly cross-provider aggregation table
  pulumicost cost actual --pulumi-json plan.json --from 2025-01-01 --to 2025-03-31 --group-by monthly

  # Output as JSON with grouping by provider
  pulumicost cost actual --pulumi-json plan.json --from 2025-01-01 --output json --group-by provider

  # Use RFC3339 timestamps
  pulumicost cost actual --pulumi-json plan.json --from 2025-01-01T00:00:00Z --to 2025-01-31T23:59:59Z`,
		RunE: func(cmd *cobra.Command, _ []string) error {
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
			tags, actualGroupBy := parseTagFilter(groupBy)

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

			// Use configuration-aware output format selection
			finalOutput := config.GetOutputFormat(output)
			outputFormat := engine.OutputFormat(finalOutput)

			return renderActualCostOutput(cmd.OutOrStdout(), outputFormat, results, actualGroupBy)
		},
	}

	cmd.Flags().StringVar(&planPath, "pulumi-json", "", "Path to Pulumi preview JSON output (required)")
	cmd.Flags().StringVar(&fromStr, "from", "", "Start date (YYYY-MM-DD or RFC3339) (required)")
	cmd.Flags().StringVar(&toStr, "to", "", "End date (YYYY-MM-DD or RFC3339) (defaults to now)")
	cmd.Flags().StringVar(&adapter, "adapter", "", "Use only the specified adapter plugin")

	// Use configuration default if no output format specified
	defaultFormat := config.GetDefaultOutputFormat()
	cmd.Flags().StringVar(&output, "output", defaultFormat, "Output format: table, json, or ndjson")
	cmd.Flags().
		StringVar(&groupBy, "group-by", "", "Group results by: resource, type, provider, date, daily, monthly, or filter by tag:key=value")

	_ = cmd.MarkFlagRequired("pulumi-json")
	_ = cmd.MarkFlagRequired("from")

	return cmd
}

// ParseTimeRange parses the provided from and to date strings into time values and validates that the range is chronological.
//
// ParseTimeRange accepts two date strings, parses each into a time.Time, and ensures the 'to' time is after the 'from' time.
// It returns the parsed from and to times on success. If either date cannot be parsed or if the 'to' time is not after
// the 'from' time, an error is returned describing the failure.
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

// ParseTime parses str interpreting it as either `YYYY-MM-DD` or an RFC3339 timestamp.
// ParseTime parses a date/time string in either YYYY-MM-DD or RFC3339 format.
// It returns the parsed time on success, or an error if the string does not match either supported format.
func ParseTime(str string) (time.Time, error) {
	layouts := []string{
		"2006-01-02",
		time.RFC3339,
	}

	for _, layout := range layouts {
		t, err := time.Parse(layout, str)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s (use YYYY-MM-DD or RFC3339)", str)
}

// parseTagFilter parses a group-by specifier for a tag filter and returns the parsed tags and the resulting groupBy.
// If groupBy is of the form "tag:key=value", it returns a map containing {key: value} and an empty actualGroupBy (indicating tag-based filtering).
// string (empty when filtering by tag).
func parseTagFilter(groupBy string) (map[string]string, string) {
	tags := make(map[string]string)
	actualGroupBy := groupBy

	if strings.HasPrefix(groupBy, "tag:") && strings.Contains(groupBy, "=") {
		tagPart := strings.TrimPrefix(groupBy, "tag:")
		if parts := strings.Split(tagPart, "="); len(parts) == filterKeyValueParts {
			tags[parts[0]] = parts[1]
			actualGroupBy = "" // Clear groupBy since we're filtering by tag
		}
	}

	return tags, actualGroupBy
}

// renderActualCostOutput renders cost output for the provided results using the given output format.
// If actualGroupBy specifies a time-based grouping (daily, monthly, etc.), it aggregates results across providers before rendering.
// Otherwise it renders the raw actual cost results as-is.
// Parameters:
//
//	outputFormat - the output format to use when rendering.
//	results - the list of cost results to render or aggregate.
//	actualGroupBy - the group-by specifier; may influence whether cross-provider time aggregation is performed.
//
// renderActualCostOutput renders actual cost results using the specified output format.
// If actualGroupBy denotes a time-based grouping, the function first creates
// cross-provider aggregations for the provided results and renders those;
// otherwise it renders the raw actual cost results.
// outputFormat is the desired rendering format, results are the cost entries
// to render or aggregate, and actualGroupBy controls grouping (time-based
// groupings trigger cross-provider aggregation).
// It returns an error if aggregation or rendering fails.
func renderActualCostOutput(
	writer io.Writer,
	outputFormat engine.OutputFormat,
	results []engine.CostResult,
	actualGroupBy string,
) error {
	// Check if we need cross-provider aggregation
	groupByType := engine.GroupBy(actualGroupBy)
	if groupByType.IsTimeBasedGrouping() {
		aggregations, err := engine.CreateCrossProviderAggregation(results, groupByType)
		if err != nil {
			return fmt.Errorf("creating cross-provider aggregation: %w", err)
		}
		return engine.RenderCrossProviderAggregation(writer, outputFormat, aggregations, groupByType)
	}

	return engine.RenderActualCostResults(writer, outputFormat, results)
}
