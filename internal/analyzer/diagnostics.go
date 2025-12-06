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
// cost estimation never blocks deployments in MVP mode.
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
// The summary diagnostic has no URN (it's stack-level, not resource-specific).
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
//   - Zero cost no notes: "Unable to estimate cost"
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
