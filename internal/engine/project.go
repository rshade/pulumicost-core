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

func renderTable(results []CostResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Resource\tAdapter\tProjected Monthly\tCurrency\tNotes")
	fmt.Fprintln(w, "--------\t-------\t-----------------\t--------\t-----")
	
	for _, result := range results {
		resource := fmt.Sprintf("%s/%s", result.ResourceType, result.ResourceID)
		if len(resource) > 50 {
			resource = resource[:47] + "..."
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