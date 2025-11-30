package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"
)

// OutputFormat specifies the output format for cost results (table, JSON, NDJSON).
type OutputFormat string

const (
	// OutputTable renders results in a formatted table.
	OutputTable OutputFormat = "table"
	// OutputJSON renders results as pretty-printed JSON.
	OutputJSON OutputFormat = "json"
	// OutputNDJSON renders results as newline-delimited JSON for streaming.
	OutputNDJSON OutputFormat = "ndjson"
)

const (
	// defaultTabPadding is the number of spaces to use for padding in table output.
	defaultTabPadding = 2
	// maxResourceDisplayLen is the maximum length of the resource name in table output before truncation.
	maxResourceDisplayLen = 40
	// truncationEllipsis is the string to append when truncating resource names.
	truncationEllipsis = "..."
)

// RenderResults renders the given cost results using the specified output format.
// RenderResults aggregates the results for table and JSON summary outputs, emits NDJSON as individual records, and writes the output to the specified writer.
// The writer parameter specifies where the output should be written.
// The format parameter selects one of the supported OutputFormat values (OutputTable, OutputJSON, OutputNDJSON).
// The results parameter is the slice of CostResult to be rendered.
// It returns an error if rendering fails or if the provided format is unsupported.
func RenderResults(writer io.Writer, format OutputFormat, results []CostResult) error {
	// Aggregate results for enhanced reporting
	aggregated := AggregateResults(results)

	switch format {
	case OutputTable:
		return renderTable(writer, aggregated)
	case OutputJSON:
		return renderJSON(writer, aggregated)
	case OutputNDJSON:
		return renderNDJSON(writer, results) // NDJSON doesn't need aggregation
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// RenderActualCostResults renders actual cost results using the specified output format.
// It dispatches to the table, pretty JSON, or NDJSON renderer based on the provided format.
// Supported formats are OutputTable, OutputJSON, and OutputNDJSON; an error is returned for unsupported formats.
// Parameters:
//   - format: the desired OutputFormat (e.g., OutputTable, OutputJSON, OutputNDJSON).
//   - results: the slice of CostResult values to render.
//
// RenderActualCostResults dispatches actual cost results to the renderer that corresponds
// to the provided output format.
//
// The format parameter selects the output format and must be one of OutputTable,
// OutputJSON, or OutputNDJSON. The results parameter is the slice of actual cost
// results to be rendered.
//
// It returns an error if the selected renderer fails or if the format is unsupported.
func RenderActualCostResults(writer io.Writer, format OutputFormat, results []CostResult) error {
	switch format {
	case OutputTable:
		return renderActualCostTable(writer, results)
	case OutputJSON:
		return renderJSONCostResults(writer, results)
	case OutputNDJSON:
		return renderNDJSON(writer, results)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// RenderCrossProviderAggregation renders cross-provider aggregation data using the specified output format.
// It supports table, JSON, and NDJSON formats and dispatches to the corresponding renderer.
//
// Parameters:
//   - format: desired output format (e.g., OutputTable, OutputJSON, OutputNDJSON).
//   - aggregations: slice of cross-provider aggregation records to render.
//   - groupBy: grouping granularity used for table output (e.g., daily or monthly).
//
// RenderCrossProviderAggregation dispatches cross-provider aggregation data to the renderer
// corresponding to the given OutputFormat.
//
// The `format` parameter selects the output renderer: table, JSON, or NDJSON.
// `aggregations` is the slice of cross-provider aggregation records to render.
// `groupBy` controls the temporal grouping used by table renderers (for example, daily vs monthly).
//
// It returns an error if the specified format is not supported or if the selected renderer fails.
func RenderCrossProviderAggregation(
	writer io.Writer,
	format OutputFormat,
	aggregations []CrossProviderAggregation,
	groupBy GroupBy,
) error {
	switch format {
	case OutputTable:
		return renderCrossProviderTable(writer, aggregations, groupBy)
	case OutputJSON:
		return renderJSONCrossProvider(writer, aggregations)
	case OutputNDJSON:
		return renderNDJSONCrossProvider(writer, aggregations)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// renderTable writes a human-readable cost table for the given aggregated results to stdout.
// It prints a cost summary (total monthly/hourly and resource count), optional breakdowns by
// provider, service, and adapter, and a detailed per-resource table with columns for Resource,
// Adapter, Monthly, Hourly, Currency, and Notes.
// The aggregated parameter provides the summary totals, breakdown maps, and the slice of per-resource results to render.
// renderTable writes a tab-separated cost summary and detailed resource table for the
// provided aggregated results to standard output. It includes total monthly/hourly
// costs, counts, optional breakdowns by provider, service, and adapter, and a per-
// resource list with monthly/hourly costs and notes.
// aggregated is the precomputed aggregation to render.
// It returns an error if writing to or flushing the tabulated output fails.
func renderTable(writer io.Writer, aggregated *AggregatedResults) error {
	w := tabwriter.NewWriter(writer, 0, 0, defaultTabPadding, ' ', 0)

	// Print summary first
	fmt.Fprintf(w, "COST SUMMARY\n")
	fmt.Fprintf(w, "============\n")
	fmt.Fprintf(w, "Total Monthly Cost:\t%.2f %s\n", aggregated.Summary.TotalMonthly, aggregated.Summary.Currency)
	fmt.Fprintf(w, "Total Hourly Cost:\t%.2f %s\n", aggregated.Summary.TotalHourly, aggregated.Summary.Currency)
	fmt.Fprintf(w, "Total Resources:\t%d\n", len(aggregated.Resources))
	fmt.Fprintf(w, "\n")

	// Print breakdown by provider
	if len(aggregated.Summary.ByProvider) > 0 {
		fmt.Fprintf(w, "BY PROVIDER\n")
		fmt.Fprintf(w, "-----------\n")
		for provider, cost := range aggregated.Summary.ByProvider {
			fmt.Fprintf(w, "%s:\t%.2f %s\n", provider, cost, aggregated.Summary.Currency)
		}
		fmt.Fprintf(w, "\n")
	}

	// Print breakdown by service
	if len(aggregated.Summary.ByService) > 0 {
		fmt.Fprintf(w, "BY SERVICE\n")
		fmt.Fprintf(w, "----------\n")
		for service, cost := range aggregated.Summary.ByService {
			fmt.Fprintf(w, "%s:\t%.2f %s\n", service, cost, aggregated.Summary.Currency)
		}
		fmt.Fprintf(w, "\n")
	}

	// Print breakdown by adapter
	if len(aggregated.Summary.ByAdapter) > 0 {
		fmt.Fprintf(w, "BY ADAPTER\n")
		fmt.Fprintf(w, "----------\n")
		for adapter, cost := range aggregated.Summary.ByAdapter {
			fmt.Fprintf(w, "%s:\t%.2f %s\n", adapter, cost, aggregated.Summary.Currency)
		}
		fmt.Fprintf(w, "\n")
	}

	// Print detailed resource breakdown
	fmt.Fprintf(w, "RESOURCE DETAILS\n")
	fmt.Fprintf(w, "================\n")
	fmt.Fprintln(w, "Resource\tAdapter\tMonthly\tHourly\tCurrency\tNotes")
	fmt.Fprintln(w, "--------\t-------\t-------\t------\t--------\t-----")

	for _, result := range aggregated.Resources {
		resource := fmt.Sprintf("%s/%s", result.ResourceType, result.ResourceID)
		if len(resource) > maxResourceDisplayLen {
			resource = resource[:maxResourceDisplayLen-len(truncationEllipsis)] + truncationEllipsis
		}
		fmt.Fprintf(w, "%s\t%s\t%.2f\t%.4f\t%s\t%s\n",
			resource,
			result.Adapter,
			result.Monthly,
			result.Hourly,
			result.Currency,
			result.Notes,
		)
	}

	return w.Flush()
}

func renderActualCostTable(writer io.Writer, results []CostResult) error {
	w := tabwriter.NewWriter(writer, 0, 0, defaultTabPadding, ' ', 0)

	// Check if we have actual cost data to determine appropriate headers
	hasActualCosts := false
	for _, result := range results {
		if result.TotalCost > 0 || result.CostPeriod != "" {
			hasActualCosts = true
			break
		}
	}

	if hasActualCosts {
		fmt.Fprintln(w, "Resource\tAdapter\tTotal Cost\tPeriod\tCurrency\tNotes")
		fmt.Fprintln(w, "--------\t-------\t----------\t------\t--------\t-----")
	} else {
		fmt.Fprintln(w, "Resource\tAdapter\tProjected Monthly\tCurrency\tNotes")
		fmt.Fprintln(w, "--------\t-------\t-----------------\t--------\t-----")
	}

	for _, result := range results {
		resource := fmt.Sprintf("%s/%s", result.ResourceType, result.ResourceID)
		if len(resource) > maxResourceDisplayLen {
			resource = resource[:maxResourceDisplayLen-len(truncationEllipsis)] + truncationEllipsis
		}

		if hasActualCosts {
			costDisplay := fmt.Sprintf("%.2f", result.TotalCost)
			if result.TotalCost == 0 && result.Monthly > 0 {
				costDisplay = fmt.Sprintf("%.2f (est)", result.Monthly)
			}

			period := result.CostPeriod
			if period == "" {
				period = "monthly (est)"
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				resource,
				result.Adapter,
				costDisplay,
				period,
				result.Currency,
				result.Notes,
			)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%.2f\t%s\t%s\n",
				resource,
				result.Adapter,
				result.Monthly,
				result.Currency,
				result.Notes,
			)
		}
	}

	return w.Flush()
}

func renderJSON(writer io.Writer, aggregated *AggregatedResults) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(aggregated)
}

// renderJSONCostResults writes the provided cost results as pretty-printed JSON to the specified writer.
//
// writer is the destination for the JSON output.
// results is the slice of CostResult values to be encoded.
//
// It returns any error encountered while encoding/writing the JSON.
func renderJSONCostResults(writer io.Writer, results []CostResult) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

// renderNDJSON writes each CostResult in results as a separate JSON object on its own line to stdout,
// renderNDJSON encodes each CostResult in results as a single JSON object per line and writes them to stdout.
// It produces newline-delimited JSON (NDJSON).
//
// The results parameter is the slice of CostResult objects to encode.
// It returns any encoding error encountered while writing.
func renderNDJSON(writer io.Writer, results []CostResult) error {
	encoder := json.NewEncoder(writer)
	for _, result := range results {
		if err := encoder.Encode(result); err != nil {
			return err
		}
	}
	return nil
}

// renderCrossProviderTable writes a cross-provider cost table to stdout.
// It formats one row per aggregation period and one column per provider, with the first
// column labeled "Date" when groupBy is GroupByDaily or "Month" otherwise.
// The function sorts provider names alphabetically to produce a consistent column order,
// prefixes monetary values with the currency symbol from each aggregation, and formats
// amounts with two decimal places.
//
// Parameters:
//   - aggregations: slice of CrossProviderAggregation where each element represents a
//     period's total and per-provider costs along with the currency code.
//   - groupBy: controls whether the period column is presented as a date (GroupByDaily)
//     or as a month.
//
// Returns an error if writing to stdout or flushing the table writer fails. If
// aggregations is empty the function prints a "No cost data available..." message and
// renderCrossProviderTable prints a tabular cross-provider cost report to stdout.
//
// aggregations is a slice of CrossProviderAggregation describing per-period totals and per-provider breakdowns.
// groupBy controls the period label used in the table header (e.g., daily uses "Date", otherwise "Month").
//
// It returns an error if writing to stdout or flushing the table writer fails.
func renderCrossProviderTable(writer io.Writer, aggregations []CrossProviderAggregation, groupBy GroupBy) error {
	if len(aggregations) == 0 {
		_, err := fmt.Fprintln(writer, "No cost data available for cross-provider aggregation")
		return err
	}

	w := tabwriter.NewWriter(writer, 0, 0, defaultTabPadding, ' ', 0)

	// Collect all unique providers
	providerSet := make(map[string]bool)
	for _, agg := range aggregations {
		for provider := range agg.Providers {
			providerSet[provider] = true
		}
	}

	// Create sorted provider list
	var providers []string
	for provider := range providerSet {
		providers = append(providers, provider)
	}
	sort.Strings(providers) // Sort alphabetically for consistent ordering

	// Print header
	if groupBy == GroupByDaily {
		fmt.Fprintf(w, "Date\tTotal Cost")
	} else {
		fmt.Fprintf(w, "Month\tTotal Cost")
	}

	for _, provider := range providers {
		fmt.Fprintf(w, "\t%s", provider)
	}
	fmt.Fprintf(w, "\n")

	// Print separator
	if groupBy == GroupByDaily {
		fmt.Fprintf(w, "----\t----------")
	} else {
		fmt.Fprintf(w, "-----\t----------")
	}

	for range providers {
		fmt.Fprintf(w, "\t--------")
	}
	fmt.Fprintf(w, "\n")

	// Print data rows
	for _, agg := range aggregations {
		currencySymbol := getCurrencySymbol(agg.Currency)
		fmt.Fprintf(w, "%s\t%s%.2f", agg.Period, currencySymbol, agg.Total)
		for _, provider := range providers {
			cost := agg.Providers[provider]
			if cost > 0 {
				fmt.Fprintf(w, "\t%s%.2f", currencySymbol, cost)
			} else {
				fmt.Fprintf(w, "\t%s0.00", currencySymbol)
			}
		}
		fmt.Fprintf(w, "\n")
	}

	return w.Flush()
}

// renderJSONCrossProvider writes the provided cross-provider aggregations to stdout
// as pretty-printed (indented) JSON.
// renderJSONCrossProvider encodes the given cross-provider aggregations as indented JSON to stdout.
// The provided slice is written as a pretty-printed JSON array; any encoding error is returned.
func renderJSONCrossProvider(writer io.Writer, aggregations []CrossProviderAggregation) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(aggregations)
}

// renderNDJSONCrossProvider writes each CrossProviderAggregation in aggregations to stdout
// as newline-delimited JSON (NDJSON). It returns the first encoding error encountered, or
// renderNDJSONCrossProvider writes each CrossProviderAggregation as a separate NDJSON object to stdout.
// It returns an error if encoding any aggregation fails.
func renderNDJSONCrossProvider(writer io.Writer, aggregations []CrossProviderAggregation) error {
	encoder := json.NewEncoder(writer)
	for _, agg := range aggregations {
		if err := encoder.Encode(agg); err != nil {
			return err
		}
	}
	return nil
}

// getCurrencySymbol returns the currency symbol for the given ISO currency code.
// getCurrencySymbol returns the currency symbol for common ISO currency codes.
// It maps "USD" -> "$", "EUR" -> "€", "GBP" -> "£", "JPY" -> "¥", "CAD" -> "C$", and "AUD" -> "A$".
// For unknown or unmapped codes, it returns the original currency code.
func getCurrencySymbol(currency string) string {
	switch currency {
	case "USD":
		return "$"
	case "EUR":
		return "€"
	case "GBP":
		return "£"
	case "JPY":
		return "¥"
	case "CAD":
		return "C$"
	case "AUD":
		return "A$"
	default:
		return currency // Fall back to currency code if symbol is unknown
	}
}
