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
	ResourceType string             `json:"resourceType"`
	ResourceID   string             `json:"resourceId"`
	Adapter      string             `json:"adapter"`
	Currency     string             `json:"currency"`
	Monthly      float64            `json:"monthly"`
	Hourly       float64            `json:"hourly"`
	Notes        string             `json:"notes"`
	Breakdown    map[string]float64 `json:"breakdown"`
	// Actual cost specific fields
	TotalCost   float64   `json:"totalCost,omitempty"`
	DailyCosts  []float64 `json:"dailyCosts,omitempty"`
	CostPeriod  string    `json:"costPeriod,omitempty"`
	StartDate   time.Time `json:"startDate,omitempty"`
	EndDate     time.Time `json:"endDate,omitempty"`
}

type ActualCostRequest struct {
	Resources []ResourceDescriptor
	From      time.Time
	To        time.Time
	Adapter   string
	GroupBy   string
	Tags      map[string]string
}

type GroupBy string

const (
	GroupByResource GroupBy = "resource"
	GroupByType     GroupBy = "type"
	GroupByProvider GroupBy = "provider"
	GroupByDate     GroupBy = "date"
	GroupByNone     GroupBy = ""
)

type ProjectedCostRequest struct {
	Resources []ResourceDescriptor
	SpecDir   string
	Adapter   string
}

type PricingSpec struct {
	Provider string                 `yaml:"provider"`
	Service  string                 `yaml:"service"`
	SKU      string                 `yaml:"sku"`
	Currency string                 `yaml:"currency"`
	Pricing  map[string]interface{} `yaml:"pricing"`
	Metadata map[string]interface{} `yaml:"metadata,omitempty"`
}
