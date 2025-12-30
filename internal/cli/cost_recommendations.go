package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/rshade/pulumicost-core/internal/config"
	"github.com/rshade/pulumicost-core/internal/engine"
	"github.com/rshade/pulumicost-core/internal/logging"
	"github.com/rshade/pulumicost-core/internal/proto"
	"github.com/spf13/cobra"
)

// Note: The engine.Recommendation struct has these fields:
// - ResourceID string
// - Type string (maps from proto ActionType)
// - Description string
// - EstimatedSavings float64
// - Currency string

// costRecommendationsParams holds the parameters for the recommendations command execution.
type costRecommendationsParams struct {
	planPath string
	adapter  string
	output   string
	filter   []string
}

// NewCostRecommendationsCmd creates the "recommendations" subcommand that fetches cost optimization
// recommendations for resources described in a Pulumi preview JSON.
//
// The command is configured with flags:
//   - --pulumi-json (required): path to Pulumi preview JSON output
//   - --adapter: restrict to a specific adapter plugin
//   - --output: output format (table, json, ndjson; defaults from configuration)
//   - --filter: filter expressions for recommendations (e.g., 'action=MIGRATE')
//
// The returned *cobra.Command is ready to be added to the CLI command tree.
func NewCostRecommendationsCmd() *cobra.Command {
	var params costRecommendationsParams

	cmd := &cobra.Command{
		Use:   "recommendations",
		Short: "Get cost optimization recommendations",
		Long: `Fetch cost optimization recommendations for resources from cloud provider APIs and plugins.

Valid action types for filtering:
  RIGHTSIZE, TERMINATE, PURCHASE_COMMITMENT, ADJUST_REQUESTS, MODIFY,
  DELETE_UNUSED, MIGRATE, CONSOLIDATE, SCHEDULE, REFACTOR, OTHER`,
		Example: `  # Get all cost optimization recommendations
  pulumicost cost recommendations --pulumi-json plan.json

  # Output recommendations as JSON
  pulumicost cost recommendations --pulumi-json plan.json --output json

  # Filter recommendations by action type
  pulumicost cost recommendations --pulumi-json plan.json --filter "action=MIGRATE"

  # Filter by multiple action types (comma-separated)
  pulumicost cost recommendations --pulumi-json plan.json --filter "action=RIGHTSIZE,TERMINATE"

  # Use a specific adapter plugin
  pulumicost cost recommendations --pulumi-json plan.json --adapter kubecost`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return executeCostRecommendations(cmd, params)
		},
	}

	cmd.Flags().
		StringVar(&params.planPath, "pulumi-json", "", "Path to Pulumi preview JSON output (required)")
	cmd.Flags().StringVar(&params.adapter, "adapter", "", "Use only the specified adapter plugin")

	// Use configuration default if no output format specified
	defaultFormat := config.GetDefaultOutputFormat()
	cmd.Flags().StringVar(&params.output, "output", defaultFormat, "Output format: table, json, or ndjson")
	cmd.Flags().StringArrayVar(&params.filter, "filter", []string{},
		"Filter expressions (e.g., 'action=MIGRATE,RIGHTSIZE')")

	_ = cmd.MarkFlagRequired("pulumi-json")

	return cmd
}

// executeCostRecommendations orchestrates the recommendations workflow for a Pulumi plan.
// It loads and maps resources, opens adapter plugins, fetches recommendations, applies filters,
// and renders the output.
//
// cmd is the Cobra command whose context and output writer are used.
// params supplies the plan path, adapter, output format, and filter expressions.
//
// Returns an error when resource loading fails, plugins cannot be opened, recommendation
// fetching fails, or output rendering fails.
func executeCostRecommendations(cmd *cobra.Command, params costRecommendationsParams) error {
	ctx := cmd.Context()
	log := logging.FromContext(ctx)

	log.Debug().Ctx(ctx).Str("operation", "cost_recommendations").Str("plan_path", params.planPath).
		Msg("starting recommendations fetch")

	// Setup audit context for logging
	auditParams := map[string]string{
		"pulumi_json": params.planPath,
		"output":      params.output,
	}
	if len(params.filter) > 0 {
		auditParams["filter"] = strings.Join(params.filter, ",")
	}
	audit := newAuditContext(ctx, "cost recommendations", auditParams)

	// Load and map resources from Pulumi plan
	resources, err := loadAndMapResources(ctx, params.planPath, audit)
	if err != nil {
		return err
	}

	// Open plugin connections
	clients, cleanup, err := openPlugins(ctx, params.adapter, audit)
	if err != nil {
		return err
	}
	defer cleanup()

	// Fetch recommendations from engine
	result, err := engine.New(clients, nil).GetRecommendationsForResources(ctx, resources)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("failed to fetch recommendations")
		audit.logFailure(ctx, err)
		return fmt.Errorf("fetching recommendations: %w", err)
	}

	// Apply action type filters if specified
	filteredRecommendations := result.Recommendations
	for _, f := range params.filter {
		if f != "" {
			filtered, filterErr := applyActionTypeFilter(ctx, filteredRecommendations, f)
			if filterErr != nil {
				return filterErr
			}
			filteredRecommendations = filtered
		}
	}

	// Create filtered result for rendering
	filteredResult := &engine.RecommendationsResult{
		Recommendations: filteredRecommendations,
		Errors:          result.Errors,
		TotalSavings:    calculateTotalSavings(filteredRecommendations),
		Currency:        result.Currency,
	}

	// Render output
	if renderErr := RenderRecommendationsOutput(ctx, cmd, params.output, filteredResult); renderErr != nil {
		return renderErr
	}

	log.Info().Ctx(ctx).Str("operation", "cost_recommendations").
		Int("recommendation_count", len(filteredRecommendations)).
		Dur("duration_ms", time.Since(audit.start)).
		Msg("recommendations fetch complete")

	audit.logSuccess(ctx, len(filteredRecommendations), filteredResult.TotalSavings)
	return nil
}

// applyActionTypeFilter filters recommendations by action type based on a filter expression.
// Filter format: "action=TYPE1,TYPE2,..."
// Matching is case-insensitive. Returns error for invalid action types.
func applyActionTypeFilter(
	ctx context.Context,
	recommendations []engine.Recommendation,
	filter string,
) ([]engine.Recommendation, error) {
	log := logging.FromContext(ctx)

	// Parse filter expression: "action=MIGRATE,RIGHTSIZE"
	parts := strings.SplitN(filter, "=", 2) //nolint:mnd // key=value format
	if len(parts) != 2 {                    //nolint:mnd // key=value has 2 parts
		return recommendations, nil // Not an action filter, return unchanged
	}

	key := strings.TrimSpace(strings.ToLower(parts[0]))
	if key != "action" {
		return recommendations, nil // Not an action filter, return unchanged
	}

	// Parse and validate action types using proto utilities
	actionTypesStr := strings.TrimSpace(parts[1])
	actionTypes, err := proto.ParseActionTypeFilter(actionTypesStr)
	if err != nil {
		return nil, fmt.Errorf("invalid action type filter: %w", err)
	}

	// Filter recommendations using proto.MatchesActionType
	var filtered []engine.Recommendation
	for _, rec := range recommendations {
		if proto.MatchesActionType(rec.Type, actionTypes) {
			filtered = append(filtered, rec)
		}
	}

	log.Debug().Ctx(ctx).
		Int("original_count", len(recommendations)).
		Int("filtered_count", len(filtered)).
		Str("filter", filter).
		Msg("applied action type filter")

	return filtered, nil
}

// calculateTotalSavings calculates the total estimated savings from recommendations.
func calculateTotalSavings(recommendations []engine.Recommendation) float64 {
	var total float64
	for _, rec := range recommendations {
		total += rec.EstimatedSavings
	}
	return total
}

// RenderRecommendationsOutput routes the recommendations results to the appropriate
// rendering function based on the output format.
func RenderRecommendationsOutput(
	_ context.Context,
	cmd *cobra.Command,
	outputFormat string,
	result *engine.RecommendationsResult,
) error {
	fmtType := engine.OutputFormat(config.GetOutputFormat(outputFormat))

	// Validate format is supported
	if !isValidOutputFormat(fmtType) {
		return fmt.Errorf("unsupported output format: %s", fmtType)
	}

	switch fmtType {
	case engine.OutputJSON:
		return renderRecommendationsJSON(cmd.OutOrStdout(), result)
	case engine.OutputNDJSON:
		return renderRecommendationsNDJSON(cmd.OutOrStdout(), result)
	case engine.OutputTable:
		return renderRecommendationsTable(cmd.OutOrStdout(), result)
	default:
		return renderRecommendationsTable(cmd.OutOrStdout(), result)
	}
}

// tabPadding is the minimum column padding for tabwriter output.
const tabPadding = 2

// renderRecommendationsTable renders recommendations in table format.
func renderRecommendationsTable(w io.Writer, result *engine.RecommendationsResult) error {
	tw := tabwriter.NewWriter(w, 0, 0, tabPadding, ' ', 0)

	// Header
	fmt.Fprintln(tw, "RESOURCE\tACTION TYPE\tDESCRIPTION\tSAVINGS")
	fmt.Fprintln(tw, "--------\t-----------\t-----------\t-------")

	// Recommendations
	for _, rec := range result.Recommendations {
		savings := ""
		if rec.EstimatedSavings > 0 {
			savings = fmt.Sprintf("%.2f %s", rec.EstimatedSavings, rec.Currency)
		}

		// Truncate long descriptions
		description := rec.Description
		const maxDescLen = 50
		if len(description) > maxDescLen {
			description = description[:maxDescLen-3] + "..."
		}

		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			rec.ResourceID,
			formatActionTypeLabel(rec.Type),
			description,
			savings,
		)
	}

	if err := tw.Flush(); err != nil {
		return fmt.Errorf("flushing table writer: %w", err)
	}

	// Summary
	fmt.Fprintln(w)
	fmt.Fprintf(w, "Total Recommendations: %d\n", len(result.Recommendations))
	if result.TotalSavings > 0 {
		fmt.Fprintf(w, "Total Estimated Savings: %.2f %s\n", result.TotalSavings, result.Currency)
	}

	// Errors
	if result.HasErrors() {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "ERRORS")
		fmt.Fprintln(w, "======")
		fmt.Fprintln(w, result.ErrorSummary())
	}

	return nil
}

// renderRecommendationsJSON renders recommendations in JSON format.
func renderRecommendationsJSON(w io.Writer, result *engine.RecommendationsResult) error {
	output := recommendationsJSONOutput{
		Recommendations: make([]recommendationJSON, 0, len(result.Recommendations)),
		TotalSavings:    result.TotalSavings,
		Currency:        result.Currency,
		Errors:          result.Errors,
	}

	for _, rec := range result.Recommendations {
		jsonRec := recommendationJSON{
			ResourceID:       rec.ResourceID,
			ActionType:       rec.Type,
			Description:      rec.Description,
			EstimatedSavings: rec.EstimatedSavings,
			Currency:         rec.Currency,
		}
		output.Recommendations = append(output.Recommendations, jsonRec)
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}
	return nil
}

// renderRecommendationsNDJSON renders recommendations in NDJSON format.
func renderRecommendationsNDJSON(w io.Writer, result *engine.RecommendationsResult) error {
	encoder := json.NewEncoder(w)

	for _, rec := range result.Recommendations {
		jsonRec := recommendationJSON{
			ResourceID:       rec.ResourceID,
			ActionType:       rec.Type,
			Description:      rec.Description,
			EstimatedSavings: rec.EstimatedSavings,
			Currency:         rec.Currency,
		}
		if err := encoder.Encode(jsonRec); err != nil {
			return fmt.Errorf("encoding NDJSON: %w", err)
		}
	}
	return nil
}

// formatActionTypeLabel returns a human-readable label for an action type.
// Uses the proto utilities for consistent label formatting across the codebase.
func formatActionTypeLabel(actionType string) string {
	return proto.ActionTypeLabelFromString(actionType)
}

// JSON output structures for recommendations.
type recommendationsJSONOutput struct {
	Recommendations []recommendationJSON         `json:"recommendations"`
	TotalSavings    float64                      `json:"total_savings"`
	Currency        string                       `json:"currency"`
	Errors          []engine.RecommendationError `json:"errors,omitempty"`
}

type recommendationJSON struct {
	ResourceID       string  `json:"resource_id"`
	ActionType       string  `json:"action_type"`
	Description      string  `json:"description"`
	EstimatedSavings float64 `json:"estimated_savings,omitempty"`
	Currency         string  `json:"currency,omitempty"`
}
