package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
)

type OutputFormat string

const (
	OutputTable  OutputFormat = "table"
	OutputJSON   OutputFormat = "json"
	OutputNDJSON OutputFormat = "ndjson"
)

func RenderResults(format OutputFormat, results []CostResult) error {
	// Aggregate results for enhanced reporting
	aggregated := AggregateResults(results)

	switch format {
	case OutputTable:
		return renderTable(aggregated)
	case OutputJSON:
		return renderJSON(aggregated)
	case OutputNDJSON:
		return renderNDJSON(results) // NDJSON doesn't need aggregation
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func RenderActualCostResults(format OutputFormat, results []CostResult) error {
	switch format {
	case OutputTable:
		return renderActualCostTable(results)
	case OutputJSON:
		return renderJSONCostResults(results)
	case OutputNDJSON:
		return renderNDJSON(results)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func RenderCrossProviderAggregation(
	format OutputFormat,
	aggregations []CrossProviderAggregation,
	groupBy GroupBy,
) error {
	switch format {
	case OutputTable:
		return renderCrossProviderTable(aggregations, groupBy)
	case OutputJSON:
		return renderJSONCrossProvider(aggregations)
	case OutputNDJSON:
		return renderNDJSONCrossProvider(aggregations)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func renderTable(aggregated *AggregatedResults) error {
	const tabPadding = 2
	w := tabwriter.NewWriter(os.Stdout, 0, 0, tabPadding, ' ', 0)

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
		const maxResourceLen = 40
		if len(resource) > maxResourceLen {
			resource = resource[:maxResourceLen-3] + "..."
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

func renderActualCostTable(results []CostResult) error {
	const tabPadding = 2
	w := tabwriter.NewWriter(os.Stdout, 0, 0, tabPadding, ' ', 0)

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
		const maxResourceLen = 40
		if len(resource) > maxResourceLen {
			resource = resource[:maxResourceLen-3] + "..."
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

func renderJSON(aggregated *AggregatedResults) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(aggregated)
}

func renderJSONCostResults(results []CostResult) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

func renderNDJSON(results []CostResult) error {
	encoder := json.NewEncoder(os.Stdout)
	for _, result := range results {
		if err := encoder.Encode(result); err != nil {
			return err
		}
	}
	return nil
}

func renderCrossProviderTable(aggregations []CrossProviderAggregation, groupBy GroupBy) error {
	if len(aggregations) == 0 {
		_, err := fmt.Fprintln(os.Stdout, "No cost data available for cross-provider aggregation")
		return err
	}

	const tabPadding = 2
	w := tabwriter.NewWriter(os.Stdout, 0, 0, tabPadding, ' ', 0)

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

func renderJSONCrossProvider(aggregations []CrossProviderAggregation) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(aggregations)
}

func renderNDJSONCrossProvider(aggregations []CrossProviderAggregation) error {
	encoder := json.NewEncoder(os.Stdout)
	for _, agg := range aggregations {
		if err := encoder.Encode(agg); err != nil {
			return err
		}
	}
	return nil
}

// getCurrencySymbol returns the appropriate currency symbol for a given currency code.
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
