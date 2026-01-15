package plugin

import (
	"github.com/rshade/finfocus/internal/proto"
)

// ResponseScenario represents a pre-configured response scenario for testing.
type ResponseScenario string

const (
	// ScenarioSuccess represents a successful cost calculation scenario.
	ScenarioSuccess ResponseScenario = "success"

	// ScenarioPartialData represents a scenario with some missing cost data.
	ScenarioPartialData ResponseScenario = "partial"

	// ScenarioHighCost represents a scenario with expensive resources.
	ScenarioHighCost ResponseScenario = "high_cost"

	// ScenarioZeroCost represents a scenario with zero-cost resources.
	ScenarioZeroCost ResponseScenario = "zero_cost"

	// ScenarioMultiCurrency represents a scenario with mixed currencies.
	ScenarioMultiCurrency ResponseScenario = "multi_currency"
)

// ConfigureScenario applies a pre-defined response scenario to the mock plugin.
// This is a convenience method for common testing scenarios.
func (m *MockPlugin) ConfigureScenario(scenario ResponseScenario) {
	m.Reset()

	switch scenario {
	case ScenarioSuccess:
		m.configureSuccessScenario()
	case ScenarioPartialData:
		m.configurePartialDataScenario()
	case ScenarioHighCost:
		m.configureHighCostScenario()
	case ScenarioZeroCost:
		m.configureZeroCostScenario()
	case ScenarioMultiCurrency:
		m.configureMultiCurrencyScenario()
	}
}

// configureSuccessScenario sets up typical successful responses for common AWS resources.
func (m *MockPlugin) configureSuccessScenario() {
	// EC2 t3.micro instance
	m.SetProjectedCostResponse("aws:ec2/instance:Instance", &proto.CostResult{
		Currency:    "USD",
		MonthlyCost: 7.30,
		HourlyCost:  0.01,
		Notes:       "t3.micro on-demand pricing",
		CostBreakdown: map[string]float64{
			"compute": 7.30,
		},
	})

	// S3 bucket (standard storage)
	m.SetProjectedCostResponse("aws:s3/bucket:Bucket", &proto.CostResult{
		Currency:    "USD",
		MonthlyCost: 2.30,
		HourlyCost:  0.00315,
		Notes:       "Standard storage, 100GB",
		CostBreakdown: map[string]float64{
			"storage": 2.30,
		},
	})

	// RDS db.t3.micro instance
	m.SetProjectedCostResponse("aws:rds/instance:Instance", &proto.CostResult{
		Currency:    "USD",
		MonthlyCost: 12.41,
		HourlyCost:  0.017,
		Notes:       "db.t3.micro single-AZ",
		CostBreakdown: map[string]float64{
			"compute": 12.41,
		},
	})

	// Lambda function
	m.SetProjectedCostResponse("aws:lambda/function:Function", &proto.CostResult{
		Currency:    "USD",
		MonthlyCost: 0.20,
		HourlyCost:  0.000274,
		Notes:       "128MB, 1M requests/month",
		CostBreakdown: map[string]float64{
			"compute":  0.17,
			"requests": 0.03,
		},
	})
}

// configurePartialDataScenario simulates a scenario where some resources have no cost data.
func (m *MockPlugin) configurePartialDataScenario() {
	// Only configure some resources
	m.SetProjectedCostResponse("aws:ec2/instance:Instance", &proto.CostResult{
		Currency:    "USD",
		MonthlyCost: 7.30,
		HourlyCost:  0.01,
		Notes:       "t3.micro on-demand pricing",
		CostBreakdown: map[string]float64{
			"compute": 7.30,
		},
	})
	// aws:s3/bucket:Bucket intentionally not configured to simulate missing data
}

// configureHighCostScenario simulates expensive resources for testing cost warnings.
func (m *MockPlugin) configureHighCostScenario() {
	// High-end EC2 instance
	m.SetProjectedCostResponse("aws:ec2/instance:Instance", &proto.CostResult{
		Currency:    "USD",
		MonthlyCost: 2500.00,
		HourlyCost:  3.424,
		Notes:       "p3.8xlarge GPU instance",
		CostBreakdown: map[string]float64{
			"compute": 2500.00,
		},
	})

	// Large RDS instance
	m.SetProjectedCostResponse("aws:rds/instance:Instance", &proto.CostResult{
		Currency:    "USD",
		MonthlyCost: 1200.00,
		HourlyCost:  1.644,
		Notes:       "db.r5.4xlarge multi-AZ",
		CostBreakdown: map[string]float64{
			"compute": 1200.00,
		},
	})
}

// configureZeroCostScenario simulates free-tier or zero-cost resources.
func (m *MockPlugin) configureZeroCostScenario() {
	m.SetProjectedCostResponse("aws:s3/bucket:Bucket", &proto.CostResult{
		Currency:    "USD",
		MonthlyCost: 0.00,
		HourlyCost:  0.00,
		Notes:       "Free tier eligible",
		CostBreakdown: map[string]float64{
			"storage": 0.00,
		},
	})

	m.SetProjectedCostResponse("aws:lambda/function:Function", &proto.CostResult{
		Currency:    "USD",
		MonthlyCost: 0.00,
		HourlyCost:  0.00,
		Notes:       "Within free tier limits",
		CostBreakdown: map[string]float64{
			"compute":  0.00,
			"requests": 0.00,
		},
	})
}

// configureMultiCurrencyScenario simulates mixed currency responses for testing aggregation.
func (m *MockPlugin) configureMultiCurrencyScenario() {
	m.SetProjectedCostResponse("aws:ec2/instance:Instance", &proto.CostResult{
		Currency:    "USD",
		MonthlyCost: 7.30,
		HourlyCost:  0.01,
		Notes:       "US region pricing",
		CostBreakdown: map[string]float64{
			"compute": 7.30,
		},
	})

	m.SetProjectedCostResponse("aws:rds/instance:Instance", &proto.CostResult{
		Currency:    "EUR",
		MonthlyCost: 11.50,
		HourlyCost:  0.0158,
		Notes:       "EU region pricing",
		CostBreakdown: map[string]float64{
			"compute": 11.50,
		},
	})
}

// ConfigureActualCostScenario sets up actual cost responses for testing historical data.
func (m *MockPlugin) ConfigureActualCostScenario(resourceID string, totalCost float64, breakdown map[string]float64) {
	m.SetActualCostResponse(resourceID, &proto.ActualCostResult{
		Currency:      "USD",
		TotalCost:     totalCost,
		CostBreakdown: breakdown,
	})
}

// QuickResponse creates a simple cost response with the given values (convenience method for tests).
func QuickResponse(currency string, monthly, hourly float64) *proto.CostResult {
	return &proto.CostResult{
		Currency:    currency,
		MonthlyCost: monthly,
		HourlyCost:  hourly,
		CostBreakdown: map[string]float64{
			"total": monthly,
		},
	}
}

// QuickActualResponse creates a simple actual cost response (convenience method for tests).
func QuickActualResponse(currency string, total float64) *proto.ActualCostResult {
	return &proto.ActualCostResult{
		Currency:  currency,
		TotalCost: total,
		CostBreakdown: map[string]float64{
			"total": total,
		},
	}
}
