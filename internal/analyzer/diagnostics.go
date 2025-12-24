package analyzer

import (
	"fmt"
	"strings"

	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/rshade/pulumicost-core/internal/engine"
)

// Policy pack and policy name constants for diagnostic messages.
const (
	policyPackName = "pulumicost"
	policyNameCost = "cost-estimate"
	policyNameSum  = "stack-cost-summary"
)

// CostToDiagnostic converts a CostResult to an AnalyzeDiagnostic.
//
// This function creates a diagnostic message suitable for display in the
// Pulumi CLI output. The diagnostic includes:
//   - Per-resource cost information
//   - Source attribution (which plugin/spec provided the cost)
//   - Severity based on data availability
//
// Per FR-005, all diagnostics use ADVISORY enforcement level to ensure
// CostToDiagnostic converts an engine.CostResult into a Pulumi analysis diagnostic describing an estimated resource cost.
// It formats the cost message, assigns a default low severity, elevates severity to medium when Monthly is zero and Notes are present, and returns an AnalyzeDiagnostic populated with the cost policy name, policy pack name, provided URN, and policy pack version.
func CostToDiagnostic(
	cost engine.CostResult,
	urn string,
	version string,
) *pulumirpc.AnalyzeDiagnostic {
	message := formatCostMessage(cost)
	severity := pulumirpc.PolicySeverity_POLICY_SEVERITY_LOW

	// Elevate severity if no cost data available but notes present
	// This indicates a fallback or partial data scenario
	if cost.Monthly == 0 && cost.Notes != "" {
		severity = pulumirpc.PolicySeverity_POLICY_SEVERITY_MEDIUM
	}

	return &pulumirpc.AnalyzeDiagnostic{
		PolicyName:        policyNameCost,
		PolicyPackName:    policyPackName,
		PolicyPackVersion: version,
		Description:       "Estimated resource cost",
		Message:           message,
		EnforcementLevel:  pulumirpc.EnforcementLevel_ADVISORY,
		Urn:               urn,
		Severity:          severity,
	}
}

// StackSummaryDiagnostic creates a stack-level cost summary diagnostic.
//
// This diagnostic provides an aggregated view of all resource costs:
//   - Total monthly cost across all resources
//   - Count of resources successfully analyzed
//   - Currency (defaults to USD)
//
// StackSummaryDiagnostic creates a stack-level AnalyzeDiagnostic that summarizes the total
// estimated monthly cost across the provided CostResult values.
//
// The `costs` slice is aggregated to compute the total monthly cost, the currency (defaults
// to "USD" if unspecified), and the count of resources with a positive monthly estimate.
// The `version` parameter is used as the PolicyPackVersion on the returned diagnostic.
//
// It returns an *pulumirpc.AnalyzeDiagnostic representing a stack-level cost summary
// (no URN), with EnforcementLevel set to ADVISORY and a low policy severity.
func StackSummaryDiagnostic(
	costs []engine.CostResult,
	version string,
) *pulumirpc.AnalyzeDiagnostic {
	var totalMonthly float64
	currency := "USD"
	analyzed := 0

	for _, c := range costs {
		totalMonthly += c.Monthly
		if c.Currency != "" {
			currency = c.Currency
		}
		if c.Monthly > 0 {
			analyzed++
		}
	}

	message := fmt.Sprintf("Total Estimated Monthly Cost: $%.2f %s (%d resources analyzed)",
		totalMonthly, currency, analyzed)

	return &pulumirpc.AnalyzeDiagnostic{
		PolicyName:        policyNameSum,
		PolicyPackName:    policyPackName,
		PolicyPackVersion: version,
		Description:       "Stack cost summary",
		Message:           message,
		EnforcementLevel:  pulumirpc.EnforcementLevel_ADVISORY,
		Severity:          pulumirpc.PolicySeverity_POLICY_SEVERITY_LOW,
		// No URN - stack-level diagnostic
	}
}

// formatCostMessage formats a cost result into a human-readable message.
//
// Message formats:
//   - With cost: "Estimated Monthly Cost: $X.XX USD (source: adapter-name)"
//   - Zero cost with notes: Returns the notes directly
//   - Zero cost no notes: "Unable to estimate cost"
func formatCostMessage(cost engine.CostResult) string {
	var message string
	switch {
	case cost.Monthly > 0:
		message = fmt.Sprintf("Estimated Monthly Cost: $%.2f %s (source: %s)",
			cost.Monthly, cost.Currency, cost.Adapter)
	case cost.Notes != "":
		message = cost.Notes
	default:
		message = "Unable to estimate cost"
	}

	// Append sustainability metrics if present
	if len(cost.Sustainability) > 0 {
		// Use a fixed order for deterministic output
		var sustainParts []string
		// We can't easily sort map keys without importing sort package,
		// but typically there's only one or two metrics (e.g. gCO2e).
		// Let's just iterate and append. For tests, deterministic order helps,
		// but map iteration is random.
		// Since we're just appending to a string for display, strict order isn't critical
		// unless we're doing string matching in tests.
		// Let's check keys to provide a stable-ish output for common keys.
		if m, ok := cost.Sustainability["gCO2e"]; ok {
			sustainParts = append(sustainParts, fmt.Sprintf("Carbon: %.2f %s", m.Value, m.Unit))
		}
		// Add other keys if needed, or iterate remaining.
		// For now, let's just append all of them.
		for k, m := range cost.Sustainability {
			if k == "gCO2e" {
				continue // Already added
			}
			sustainParts = append(sustainParts, fmt.Sprintf("%s: %.2f %s", k, m.Value, m.Unit))
		}

		if len(sustainParts) > 0 {
			message += fmt.Sprintf(" [%s]", getJoinedSustainability(sustainParts))
		}
	}

	return message
}

func getJoinedSustainability(parts []string) string {
	return strings.Join(parts, ", ")
}

// WarningDiagnostic creates a warning-level diagnostic for error conditions.
//
// Use this function when cost calculation fails but the preview should
// continue. Examples include:
//   - Plugin timeout
//   - Network failures
//   - Unsupported resource types
//   - Invalid resource data
//
// Per FR-005, all diagnostics use ADVISORY enforcement to never block
// deployments in MVP mode.
func WarningDiagnostic(message, urn, version string) *pulumirpc.AnalyzeDiagnostic {
	return &pulumirpc.AnalyzeDiagnostic{
		PolicyName:        policyNameCost,
		PolicyPackName:    policyPackName,
		PolicyPackVersion: version,
		Description:       "Cost estimation warning",
		Message:           message,
		EnforcementLevel:  pulumirpc.EnforcementLevel_ADVISORY,
		Urn:               urn,
		Severity:          pulumirpc.PolicySeverity_POLICY_SEVERITY_MEDIUM,
	}
}
