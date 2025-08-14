package spec

import (
	"fmt"
)

func ValidateSpec(spec *PricingSpec) error {
	if spec.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if spec.Service == "" {
		return fmt.Errorf("service is required")
	}
	if spec.SKU == "" {
		return fmt.Errorf("SKU is required")
	}
	if spec.Currency == "" {
		return fmt.Errorf("currency is required")
	}
	if len(spec.Pricing) == 0 {
		return fmt.Errorf("pricing information is required")
	}
	return nil
}