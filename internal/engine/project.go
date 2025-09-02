package engine

import (
	"encoding/json"
	"fmt"
	"os"
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

func renderJSON(aggregated *AggregatedResults) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(aggregated)
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
