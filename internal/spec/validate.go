package spec

import (
	"errors"
)

func ValidateSpec(spec *PricingSpec) error {
	if spec.Provider == "" {
		return errors.New("provider is required")
	}
	if spec.Service == "" {
		return errors.New("service is required")
	}
	if spec.SKU == "" {
		return errors.New("SKU is required")
	}
	if spec.Currency == "" {
		return errors.New("currency is required")
	}
	if len(spec.Pricing) == 0 {
		return errors.New("pricing information is required")
	}
	return nil
}
