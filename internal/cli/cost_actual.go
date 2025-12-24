package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/rshade/pulumicost-core/internal/config"
	"github.com/rshade/pulumicost-core/internal/engine"
	"github.com/rshade/pulumicost-core/internal/logging"
	"github.com/spf13/cobra"
)

const (
	filterKeyValueParts = 2   // For "key=value" pairs
	maxDateRangeDays    = 366 // Maximum date range (1 year + 1 day for leap years)
	maxPastYears        = 5   // Maximum years in the past allowed
	hoursPerDay         = 24  // Hours in a day for date calculations
)

// costActualParams holds the parameters for the actual cost command execution.
type costActualParams struct {
	planPath string
	adapter  string
	output   string
	fromStr  string
	toStr    string
	groupBy  string
}

// defaultToNow returns s if non-empty, otherwise returns the current time in RFC3339 format.
func defaultToNow(s string) string {
	if s == "" {
		return time.Now().Format(time.RFC3339)
	}
	return s
}

// NewCostActualCmd creates the "actual" subcommand which fetches historical cloud-provider billing
// costs for resources described in a Pulumi preview JSON.
//
// The command is configured with flags:
//   - --pulumi-json (required): path to Pulumi preview JSON output
//   - --from (required): start date (YYYY-MM-DD or RFC3339)
//   - --to: end date (YYYY-MM-DD or RFC3339; defaults to now)
//   - --adapter: restrict to a specific adapter plugin
//   - --output: output format (table, json, ndjson; defaults from configuration)
//   - --group-by: grouping or tag filter (resource, type, provider, date, daily, monthly, or tag:key=value)
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
			params := costActualParams{
				planPath: planPath,
				adapter:  adapter,
				output:   output,
				fromStr:  fromStr,
				toStr:    toStr,
				groupBy:  groupBy,
			}
			return executeCostActual(cmd, params)
		},
	}

	cmd.Flags().
		StringVar(&planPath, "pulumi-json", "", "Path to Pulumi preview JSON output (required)")
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

// executeCostActual orchestrates the "actual" cost workflow for a Pulumi plan.
// It loads and maps resources, parses the time range, opens adapter plugins, fetches actual costs,
// renders the results, and emits audit entries.
func executeCostActual(cmd *cobra.Command, params costActualParams) error {
	ctx := cmd.Context()
	log := logging.FromContext(ctx)

	log.Debug().Ctx(ctx).Str("operation", "cost_actual").Str("plan_path", params.planPath).
		Str("from", params.fromStr).Str("to", params.toStr).Str("group_by", params.groupBy).
		Msg("starting actual cost calculation")

	// Setup audit context for logging
	auditParams := map[string]string{
		"pulumi_json": params.planPath, "output": params.output,
		"from": params.fromStr, "to": params.toStr,
	}
	if params.groupBy != "" {
		auditParams["group_by"] = params.groupBy
	}
	audit := newAuditContext(ctx, "cost actual", auditParams)

	resources, err := loadAndMapResources(ctx, params.planPath, audit)
	if err != nil {
		return err
	}

	from, to, err := ParseTimeRange(params.fromStr, defaultToNow(params.toStr))
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("failed to parse time range")
		audit.logFailure(ctx, err)
		return fmt.Errorf("parsing time range: %w", err)
	}

	clients, cleanup, err := openPlugins(ctx, params.adapter, audit)
	if err != nil {
		return err
	}
	defer cleanup()

	tags, actualGroupBy := parseTagFilter(params.groupBy)
	request := engine.ActualCostRequest{
		Resources: resources, From: from, To: to,
		Adapter: params.adapter, GroupBy: actualGroupBy, Tags: tags,
	}

	resultWithErrors, err := engine.New(clients, nil).GetActualCostWithOptionsAndErrors(ctx, request)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("failed to fetch actual costs")
		audit.logFailure(ctx, err)
		return fmt.Errorf("fetching actual costs: %w", err)
	}

	outputFormat := engine.OutputFormat(config.GetOutputFormat(params.output))
	if renderErr := renderActualCostOutput(cmd.OutOrStdout(), outputFormat, resultWithErrors.Results, actualGroupBy); renderErr != nil {
		return renderErr
	}
	displayErrorSummary(cmd, resultWithErrors, outputFormat)

	log.Info().Ctx(ctx).Str("operation", "cost_actual").Int("result_count", len(resultWithErrors.Results)).
		Dur("duration_ms", time.Since(audit.start)).Msg("actual cost calculation complete")

	totalCost := 0.0
	for _, r := range resultWithErrors.Results {
		totalCost += r.TotalCost
	}
	audit.logSuccess(ctx, len(resultWithErrors.Results), totalCost)
	return nil
}

// ParseTimeRange parses the provided from and to date strings into time values and validates that the range is chronological.
//
// ParseTimeRange accepts two date strings, parses each into a time.Time, and ensures the 'to' time is after the 'from' time.
// It returns the parsed from and to times on success. If either date cannot be parsed or if the 'to' time is not after
// the 'from' time, an error is returned describing the failure.
// Additionally validates that the date range does not exceed maximum limits.
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

	// Validate date range is within acceptable limits
	if rangeErr := ValidateDateRange(from, to); rangeErr != nil {
		return time.Time{}, time.Time{}, rangeErr
	}

	return from, to, nil
}

// ParseTime parses str as a date in either "YYYY-MM-DD" or RFC3339 format.
// It validates that the parsed time is not in the future and is not more than maxPastYears years in the past.
func ParseTime(str string) (time.Time, error) {
	layouts := []string{
		"2006-01-02",
		time.RFC3339,
	}

	var parsedTime time.Time
	var parseErr error
	parsed := false

	for _, layout := range layouts {
		t, err := time.Parse(layout, str)
		if err == nil {
			parsedTime = t
			parsed = true
			break
		}
		parseErr = err
	}

	if !parsed {
		return time.Time{}, fmt.Errorf(
			"unable to parse date: %s (use YYYY-MM-DD or RFC3339): %w",
			str,
			parseErr,
		)
	}

	// Validate: date cannot be in the future
	now := time.Now()
	if parsedTime.After(now) {
		return time.Time{}, fmt.Errorf("date cannot be in the future: %s", str)
	}

	// Validate: date cannot be more than maxPastYears years in the past
	oldestAllowed := now.AddDate(-maxPastYears, 0, 0)
	if parsedTime.Before(oldestAllowed) {
		return time.Time{}, fmt.Errorf(
			"date too far in past: %s (max %d years ago)",
			str,
			maxPastYears,
		)
	}

	return parsedTime, nil
}

// ValidateDateRange validates that the date range is within acceptable limits.
// Returns an error if the range exceeds maxDateRangeDays (approximately 1 year).
func ValidateDateRange(from, to time.Time) error {
	days := int(to.Sub(from).Hours() / hoursPerDay)
	if days > maxDateRangeDays {
		return fmt.Errorf("date range too large: %d days (max %d days / ~1 year). "+
			"Tip: Use --group-by monthly to analyze longer periods efficiently", days, maxDateRangeDays)
	}
	return nil
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

// renderActualCostOutput renders actual cost results to the provided writer.
// If actualGroupBy denotes a time-based grouping, it creates cross-provider aggregations;
// otherwise it renders the raw results.
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
		return engine.RenderCrossProviderAggregation(
			writer,
			outputFormat,
			aggregations,
			groupByType,
		)
	}

	return engine.RenderActualCostResults(writer, outputFormat, results)
}
