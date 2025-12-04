package e2e

// PricingReference holds expected monthly costs for resources.
// Key: Resource type or description. Value: Expected monthly cost in USD.
var PricingReference = map[string]float64{
	"t3.micro": 7.59, // ~$0.0104/hour * 730 hours
	"gp3":      0.64, // 8GB * $0.08/GB-month
}

// GetExpectedCost returns the expected monthly cost for a given resource key.
func GetExpectedCost(key string) (float64, bool) {
	val, ok := PricingReference[key]
	return val, ok
}
