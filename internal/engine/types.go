package engine

import (
	"time"
)

type ResourceDescriptor struct {
	Type       string
	ID         string
	Provider   string
	Properties map[string]interface{}
}

type CostResult struct {
	ResourceType string
	ResourceID   string
	Adapter      string
	Currency     string
	Monthly      float64
	Hourly       float64
	Notes        string
	Breakdown    map[string]float64
}

type ActualCostRequest struct {
	Resources []ResourceDescriptor
	From      time.Time
	To        time.Time
	Adapter   string
}

type ProjectedCostRequest struct {
	Resources []ResourceDescriptor
	SpecDir   string
	Adapter   string
}

type PricingSpec struct {
	Provider  string                 `yaml:"provider"`
	Service   string                 `yaml:"service"`
	SKU       string                 `yaml:"sku"`
	Currency  string                 `yaml:"currency"`
	Pricing   map[string]interface{} `yaml:"pricing"`
	Metadata  map[string]interface{} `yaml:"metadata,omitempty"`
}