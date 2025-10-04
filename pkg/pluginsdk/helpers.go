package pluginsdk

import (
	"context"
	"errors"
	"fmt"

	pbc "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
)

const hoursPerMonth = 730.0

// ResourceMatcher helps plugins determine if they support a resource.
type ResourceMatcher struct {
	supportedProviders map[string]bool
	supportedTypes     map[string]bool
}

// NewResourceMatcher creates a ResourceMatcher with initialized empty maps for supported providers and supported resource types.
func NewResourceMatcher() *ResourceMatcher {
	return &ResourceMatcher{
		supportedProviders: make(map[string]bool),
		supportedTypes:     make(map[string]bool),
	}
}

// AddProvider adds a supported provider (e.g., "aws", "azure", "gcp").
func (rm *ResourceMatcher) AddProvider(provider string) {
	rm.supportedProviders[provider] = true
}

// AddResourceType adds a supported resource type (e.g., "aws:ec2:Instance").
func (rm *ResourceMatcher) AddResourceType(resourceType string) {
	rm.supportedTypes[resourceType] = true
}

// Supports checks if a resource is supported by this plugin.
func (rm *ResourceMatcher) Supports(resource *pbc.ResourceDescriptor) bool {
	if rm == nil || resource == nil {
		return false
	}

	if len(rm.supportedProviders) > 0 {
		if !rm.supportedProviders[resource.GetProvider()] {
			return false
		}
	}

	if len(rm.supportedTypes) > 0 {
		if !rm.supportedTypes[resource.GetResourceType()] {
			return false
		}
	}

	return true
}

// CostCalculator provides utilities for cost calculations.
type CostCalculator struct{}

// NewCostCalculator returns a new CostCalculator for performing cost conversions and creating cost responses.
func NewCostCalculator() *CostCalculator {
	return &CostCalculator{}
}

// HourlyToMonthly converts hourly cost to monthly cost (730 hours).
func (cc *CostCalculator) HourlyToMonthly(hourlyCost float64) float64 {
	return hourlyCost * hoursPerMonth
}

// MonthlyToHourly converts monthly cost to hourly cost (730 hours).
func (cc *CostCalculator) MonthlyToHourly(monthlyCost float64) float64 {
	return monthlyCost / hoursPerMonth
}

// CreateProjectedCostResponse creates a standard projected cost response.
// unitPrice is expected to be an hourly rate; CostPerMonth is derived using 730 hours.
func (cc *CostCalculator) CreateProjectedCostResponse(
	currency string,
	unitPrice float64,
	billingDetail string,
) *pbc.GetProjectedCostResponse {
	if cc == nil {
		return &pbc.GetProjectedCostResponse{
			Currency:      currency,
			UnitPrice:     unitPrice,
			CostPerMonth:  unitPrice * hoursPerMonth,
			BillingDetail: billingDetail,
		}
	}
	return &pbc.GetProjectedCostResponse{
		Currency:      currency,
		UnitPrice:     unitPrice,
		CostPerMonth:  cc.HourlyToMonthly(unitPrice),
		BillingDetail: billingDetail,
	}
}

// CreateActualCostResponse creates a standard actual cost response.
func (cc *CostCalculator) CreateActualCostResponse(
	results []*pbc.ActualCostResult,
) *pbc.GetActualCostResponse {
	return &pbc.GetActualCostResponse{
		Results: results,
	}
}

// NotSupportedError returns an error indicating the specified resource type and provider are not supported.
// The formatted message includes the resource's ResourceType and Provider.
func NotSupportedError(resource *pbc.ResourceDescriptor) error {
	return fmt.Errorf(
		"resource type %s from provider %s is not supported",
		resource.GetResourceType(),
		resource.GetProvider(),
	)
}

// NoDataError returns a standard error when no cost data is available.
func NoDataError(resourceID string) error {
	return fmt.Errorf("no cost data available for resource %s", resourceID)
}

// BasePlugin provides common functionality for plugin implementations.
type BasePlugin struct {
	name    string
	matcher *ResourceMatcher
	calc    *CostCalculator
}

// NewBasePlugin creates a new BasePlugin with the given name and initializes its
// ResourceMatcher and CostCalculator for shared plugin functionality.
func NewBasePlugin(name string) *BasePlugin {
	return &BasePlugin{
		name:    name,
		matcher: NewResourceMatcher(),
		calc:    NewCostCalculator(),
	}
}

// Name returns the plugin name.
func (bp *BasePlugin) Name() string {
	return bp.name
}

// Matcher returns the resource matcher.
func (bp *BasePlugin) Matcher() *ResourceMatcher {
	return bp.matcher
}

// Calculator returns the cost calculator.
func (bp *BasePlugin) Calculator() *CostCalculator {
	return bp.calc
}

// GetProjectedCost provides a default implementation that returns not supported.
func (bp *BasePlugin) GetProjectedCost(
	_ context.Context,
	req *pbc.GetProjectedCostRequest,
) (*pbc.GetProjectedCostResponse, error) {
	if req == nil {
		return nil, errors.New("GetProjectedCostRequest cannot be nil")
	}
	resource := req.GetResource()
	if resource == nil {
		return nil, errors.New("resource cannot be nil")
	}
	return nil, NotSupportedError(resource)
}

// GetActualCost provides a default implementation that returns no data.
func (bp *BasePlugin) GetActualCost(
	_ context.Context,
	req *pbc.GetActualCostRequest,
) (*pbc.GetActualCostResponse, error) {
	if req == nil {
		return nil, errors.New("GetActualCostRequest cannot be nil")
	}
	return nil, NoDataError(req.GetResourceId())
}
