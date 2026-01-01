package engine

// Confidence represents the reliability level of a cost estimate.
//
// Confidence levels help users understand how reliable a cost estimate is:
//   - HIGH: Backed by real billing data from cloud provider APIs
//   - MEDIUM: Calculated from runtime using Pulumi state timestamps
//   - LOW: Estimated for imported resources where creation time is unknown
type Confidence string

// Confidence level constants.
const (
	// ConfidenceHigh indicates the cost is backed by real billing data.
	// This is the most reliable estimate, coming from actual cloud billing APIs.
	ConfidenceHigh Confidence = "high"

	// ConfidenceMedium indicates the cost is calculated from runtime.
	// The estimate uses hourly_rate × runtime from Pulumi state timestamps.
	// Reliable for resources created by Pulumi.
	ConfidenceMedium Confidence = "medium"

	// ConfidenceLow indicates reduced reliability for imported resources.
	// The "Created" timestamp reflects import time, not actual resource creation,
	// so runtime calculations may significantly underestimate actual usage.
	ConfidenceLow Confidence = "low"

	// ConfidenceUnknown indicates confidence could not be determined.
	// Used when no cost data is available or for error cases.
	ConfidenceUnknown Confidence = ""
)

// IsValid returns true if the Confidence value is valid.
func (c Confidence) IsValid() bool {
	switch c {
	case ConfidenceHigh, ConfidenceMedium, ConfidenceLow, ConfidenceUnknown:
		return true
	default:
		return false
	}
}

// String returns the string representation of the Confidence level.
func (c Confidence) String() string {
	return string(c)
}

// DisplayLabel returns an uppercase label for UI display.
// Returns "-" for unknown confidence to indicate missing data.
func (c Confidence) DisplayLabel() string {
	switch c {
	case ConfidenceHigh:
		return "HIGH"
	case ConfidenceMedium:
		return "MEDIUM"
	case ConfidenceLow:
		return "LOW"
	case ConfidenceUnknown:
		return "-"
	default:
		return "-"
	}
}

// DetermineConfidence calculates the appropriate confidence level based on data source.
//
// The logic is:
//   - HIGH: hasBillingData=true (real billing API data overrides all other factors)
//   - MEDIUM: hasBillingData=false, isExternal=false (runtime estimate, non-imported)
//   - LOW: hasBillingData=false, isExternal=true (runtime estimate, imported resource)
//
// Parameters:
//   - hasBillingData: true if cost came from actual billing APIs (e.g., AWS Cost Explorer)
//   - isExternal: true if the resource was imported (pulumi:external=true)
func DetermineConfidence(hasBillingData, isExternal bool) Confidence {
	if hasBillingData {
		return ConfidenceHigh
	}
	if isExternal {
		return ConfidenceLow
	}
	return ConfidenceMedium
}

// DetermineConfidenceFromResult calculates confidence level from a CostResult.
//
// Uses TotalCost > 0 as indicator of billing data (from GetActualCost API calls).
// The isExternal parameter indicates if the resource was imported.
//
// Parameters:
//   - result: CostResult containing cost data
//   - isExternal: true if the resource was imported (pulumi:external=true)
func DetermineConfidenceFromResult(result CostResult, isExternal bool) Confidence {
	hasBillingData := result.TotalCost > 0
	return DetermineConfidence(hasBillingData, isExternal)
}
