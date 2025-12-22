package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/rshade/pulumicost-core/internal/config"
	"github.com/rshade/pulumicost-core/internal/engine"
	"github.com/rshade/pulumicost-core/internal/ingest"
	"github.com/rshade/pulumicost-core/internal/logging"
	"github.com/rshade/pulumicost-core/internal/registry"
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
// NewCostActualCmd creates the "actual" subcommand that fetches historical cloud provider billing costs for resources.
// The command accepts flags for the Pulumi preview JSON path (--pulumi-json, required), start date (--from, required),
// optional end date (--to, defaults to now), optional adapter plugin (--adapter), output format (--output, defaults
// to the configured default), and grouping or tag filter (--group-by). It returns a configured *cobra.Command ready
// to be added to the CLI.
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

// executeCostActual executes the "actual" cost workflow using the provided command context and parameters.
// It loads and maps a Pulumi plan, resolves the query time range (defaulting the end to now if omitted),
// opens registry adapters, requests actual cost data from the engine, renders the results, and prints an
// error summary if any per-resource errors occurred.
//
// Parameters:
//   - cmd: the Cobra command used to access output streams and printing helpers.
//   - params: command parameters including the Pulumi plan path, adapter name, output format, time range, and grouping/tag filter.
//
// Returns an error when any step fails, including but not limited to:
//   - loading or mapping the Pulumi plan,
//   - parsing the from/to time range,
//   - opening adapter plugins,
//   - fetching actual cost data from the engine,
//   - rendering the output.
//
// executeCostActual orchestrates the "actual" cost workflow: it loads a Pulumi plan, maps resources,
// validates the time range, opens adapter plugins, calculates actual costs, renders output, and emits
// audit and operational logs.
//
// The cmd parameter provides command context and I/O; params configures input paths, adapters, date
// range, grouping, and output formatting.
//
// It returns an error when any step of the workflow fails (for example: loading the Pulumi plan,
// mapping resources, parsing the time range, opening adapter plugins, fetching costs, or rendering
// output). On success it returns nil.
//
//nolint:funlen // CLI command execution requires many sequential steps
func executeCostActual(cmd *cobra.Command, params costActualParams) error {
	start := time.Now()
	ctx := cmd.Context()

	// Get logger from context (set by PersistentPreRunE)
	log := logging.FromContext(ctx)
	log.Debug().
		Ctx(ctx).
		Str("operation", "cost_actual").
		Str("plan_path", params.planPath).
		Str("from", params.fromStr).
		Str("to", params.toStr).
		Str("group_by", params.groupBy).
		Msg("starting actual cost calculation")

	// Get audit logger and trace ID for audit entry
	auditLogger := logging.AuditLoggerFromContext(ctx)
	traceID := logging.TraceIDFromContext(ctx)
	auditParams := map[string]string{
		"pulumi_json": params.planPath,
		"output":      params.output,
		"from":        params.fromStr,
		"to":          params.toStr,
	}
	if params.groupBy != "" {
		auditParams["group_by"] = params.groupBy
	}

	plan, err := ingest.LoadPulumiPlanWithContext(ctx, params.planPath)
	if err != nil {
		log.Error().
			Ctx(ctx).
			Err(err).
			Str("plan_path", params.planPath).
			Msg("failed to load Pulumi plan")
		// Log audit entry for failure
		auditEntry := logging.NewAuditEntry("cost actual", traceID).
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
		auditEntry := logging.NewAuditEntry("cost actual", traceID).
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

	// Default to now if --to is not provided
	toStr := params.toStr
	if toStr == "" {
		toStr = time.Now().Format(time.RFC3339)
	}

	from, to, err := ParseTimeRange(params.fromStr, toStr)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("failed to parse time range")
		// Log audit entry for failure
		auditEntry := logging.NewAuditEntry("cost actual", traceID).
			WithParameters(auditParams).
			WithError(err.Error()).
			WithDuration(start)
		auditLogger.Log(ctx, *auditEntry)
		return fmt.Errorf("parsing time range: %w", err)
	}

	reg := registry.NewDefault()
	clients, cleanup, err := reg.Open(ctx, params.adapter)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Str("adapter", params.adapter).Msg("failed to open plugins")
		// Log audit entry for failure
		auditEntry := logging.NewAuditEntry("cost actual", traceID).
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

	eng := engine.New(clients, nil)

	// Parse tags from groupBy if it's in tag:key=value format
	tags, actualGroupBy := parseTagFilter(params.groupBy)

	request := engine.ActualCostRequest{
		Resources: resources,
		From:      from,
		To:        to,
		Adapter:   params.adapter,
		GroupBy:   actualGroupBy,
		Tags:      tags,
	}

	resultWithErrors, err := eng.GetActualCostWithOptionsAndErrors(ctx, request)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("failed to fetch actual costs")
		// Log audit entry for failure
		auditEntry := logging.NewAuditEntry("cost actual", traceID).
			WithParameters(auditParams).
			WithError(err.Error()).
			WithDuration(start)
		auditLogger.Log(ctx, *auditEntry)
		return fmt.Errorf("fetching actual costs: %w", err)
	}

	// Use configuration-aware output format selection
	finalOutput := config.GetOutputFormat(params.output)
	outputFormat := engine.OutputFormat(finalOutput)

	if renderErr := renderActualCostOutput(cmd.OutOrStdout(), outputFormat, resultWithErrors.Results, actualGroupBy); renderErr != nil {
		return renderErr
	}

	displayErrorSummary(cmd, resultWithErrors, outputFormat)

	// Log completion with duration
	log.Info().
		Ctx(ctx).
		Str("operation", "cost_actual").
		Int("result_count", len(resultWithErrors.Results)).
		Dur("duration_ms", time.Since(start)).
		Msg("actual cost calculation complete")

	// Log audit entry for success.
	// Use TotalCost for actual costs (not Monthly, which is for projected costs).
	totalCost := 0.0
	for _, r := range resultWithErrors.Results {
		totalCost += r.TotalCost
	}
	auditEntry := logging.NewAuditEntry("cost actual", traceID).
		WithParameters(auditParams).
		WithSuccess(len(resultWithErrors.Results), totalCost).
		WithDuration(start)
	auditLogger.Log(ctx, *auditEntry)

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

// ParseTime parses str interpreting it as either `YYYY-MM-DD` or an RFC3339 timestamp.
// ParseTime parses a date/time string in either YYYY-MM-DD or RFC3339 format.
// It returns the parsed time on success, or an error if the string does not match either supported format.
// Additionally validates that the date is not in the future and not more than 5 years in the past.
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
		return engine.RenderCrossProviderAggregation(
			writer,
			outputFormat,
			aggregations,
			groupByType,
		)
	}

	return engine.RenderActualCostResults(writer, outputFormat, results)
}
