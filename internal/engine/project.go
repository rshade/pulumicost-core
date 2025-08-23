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
	switch format {
	case OutputTable:
		return renderTable(results)
	case OutputJSON:
		return renderJSON(results)
	case OutputNDJSON:
		return renderNDJSON(results)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func RenderActualCostResults(format OutputFormat, results []CostResult) error {
	switch format {
	case OutputTable:
		return renderActualCostTable(results)
	case OutputJSON:
		return renderJSON(results)
	case OutputNDJSON:
		return renderNDJSON(results)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func renderTable(results []CostResult) error {
	const tabPadding = 2
	w := tabwriter.NewWriter(os.Stdout, 0, 0, tabPadding, ' ', 0)
	fmt.Fprintln(w, "Resource\tAdapter\tProjected Monthly\tCurrency\tNotes")
	fmt.Fprintln(w, "--------\t-------\t-----------------\t--------\t-----")

	for _, result := range results {
		resource := fmt.Sprintf("%s/%s", result.ResourceType, result.ResourceID)
		const maxResourceLen = 50
		if len(resource) > maxResourceLen {
			resource = resource[:maxResourceLen-3] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%.2f\t%s\t%s\n",
			resource,
			result.Adapter,
			result.Monthly,
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

func renderJSON(results []CostResult) error {
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
