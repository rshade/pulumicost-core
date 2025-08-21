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
	Provider string                 `yaml:"provider"`
	Service  string                 `yaml:"service"`
	SKU      string                 `yaml:"sku"`
	Currency string                 `yaml:"currency"`
	Pricing  map[string]interface{} `yaml:"pricing"`
	Metadata map[string]interface{} `yaml:"metadata,omitempty"`
}

type CostSummary struct {
	TotalMonthly float64                   `json:"totalMonthly"`
	TotalHourly  float64                   `json:"totalHourly"`
	Currency     string                    `json:"currency"`
	ByProvider   map[string]float64        `json:"byProvider"`
	ByService    map[string]float64        `json:"byService"`
	ByAdapter    map[string]float64        `json:"byAdapter"`
	Resources    []CostResult              `json:"resources"`
}

type AggregatedResults struct {
	Summary   CostSummary  `json:"summary"`
	Resources []CostResult `json:"resources"`
}
