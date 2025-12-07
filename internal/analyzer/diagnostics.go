package analyzer

import (
	"fmt"

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
// CostToDiagnostic converts a CostResult into a Pulumi AnalyzeDiagnostic for a single resource.
// 
// The diagnostic uses the "pulumicost" policy pack and the "cost-estimate" policy name, is emitted
// at the ADVISORY enforcement level, and is attributed to the provided resource URN. By default the
// diagnostic severity is POLICY_SEVERITY_LOW; if the cost has no monthly estimate but contains notes,
// severity is elevated to POLICY_SEVERITY_MEDIUM to indicate partial or fallback estimation.
// 
// Parameters:
//   - cost: the engine.CostResult to represent in the diagnostic.
//   - urn: the resource URN to attribute the diagnostic to.
//   - version: the policy pack version to include in the diagnostic metadata.
// 
// Returns:
//   - A pointer to a pulumirpc.AnalyzeDiagnostic describing the estimated resource cost.
func CostToDiagnostic(cost engine.CostResult, urn string, version string) *pulumirpc.AnalyzeDiagnostic {
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
// StackSummaryDiagnostic returns a stack-level AnalyzeDiagnostic that summarizes the total estimated
// monthly cost across the provided CostResult slice.
// 
// The `costs` parameter is a slice of engine.CostResult used to compute the aggregated monthly total,
// the currency (the last non-empty currency encountered), and the number of resources with a monthly
// estimate greater than zero. The `version` parameter is recorded as the PolicyPackVersion in the
// resulting diagnostic.
// 
// The returned *pulumirpc.AnalyzeDiagnostic represents a stack-level summary (no URN), uses an
// ADVISORY enforcement level, and is assigned low policy severity.
func StackSummaryDiagnostic(costs []engine.CostResult, version string) *pulumirpc.AnalyzeDiagnostic {
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
// formatCostMessage returns a human-readable message describing the estimated monthly cost for the given CostResult.
// If `Monthly` is greater than zero the message includes the dollar amount, currency, and adapter source; if `Monthly` is zero and `Notes` is non-empty it returns `Notes`; otherwise it returns "Unable to estimate cost".
func formatCostMessage(cost engine.CostResult) string {
	if cost.Monthly > 0 {
		return fmt.Sprintf("Estimated Monthly Cost: $%.2f %s (source: %s)",
			cost.Monthly, cost.Currency, cost.Adapter)
	}
	if cost.Notes != "" {
		return cost.Notes
	}
	return "Unable to estimate cost"
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
// WarningDiagnostic creates an AnalyzeDiagnostic representing a non-critical cost estimation warning.
// WarningDiagnostic sets the policy name to "cost-estimate", the policy pack name to "pulumicost",
// the enforcement level to ADVISORY, and the severity to MEDIUM.
// message is the user-facing diagnostic message, urn is the resource URN to attribute the diagnostic to,
// and version is the policy pack version used in the diagnostic.
// The returned value is a configured *pulumirpc.AnalyzeDiagnostic ready to be reported by the Pulumi engine.
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