package analyzer

import (
	"testing"

	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/rshade/pulumicost-core/internal/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCostToDiagnostic(t *testing.T) {
	tests := []struct {
		name         string
		cost         engine.CostResult
		urn          string
		version      string
		wantSeverity pulumirpc.PolicySeverity
		wantContains string
	}{
		{
			name: "successful cost calculation",
			cost: engine.CostResult{
				ResourceType: "aws:ec2/instance:Instance",
				ResourceID:   "webserver",
				Adapter:      "local-spec",
				Currency:     "USD",
				Monthly:      8.45,
				Hourly:       0.0116,
			},
			urn:          "urn:pulumi:dev::myapp::aws:ec2/instance:Instance::webserver",
			version:      "0.1.0",
			wantSeverity: pulumirpc.PolicySeverity_POLICY_SEVERITY_LOW,
			wantContains: "$8.45 USD",
		},
		{
			name: "zero cost with notes (fallback)",
			cost: engine.CostResult{
				ResourceType: "aws:ec2/instance:Instance",
				ResourceID:   "webserver",
				Adapter:      "none",
				Currency:     "USD",
				Monthly:      0,
				Notes:        "Unable to estimate: unsupported resource type",
			},
			urn:          "urn:pulumi:dev::myapp::aws:ec2/instance:Instance::webserver",
			version:      "0.1.0",
			wantSeverity: pulumirpc.PolicySeverity_POLICY_SEVERITY_MEDIUM,
			wantContains: "Unable to estimate",
		},
		{
			name: "zero cost no notes",
			cost: engine.CostResult{
				ResourceType: "aws:ec2/instance:Instance",
				ResourceID:   "webserver",
				Adapter:      "none",
				Currency:     "USD",
				Monthly:      0,
			},
			urn:          "urn:pulumi:dev::myapp::aws:ec2/instance:Instance::webserver",
			version:      "0.1.0",
			wantSeverity: pulumirpc.PolicySeverity_POLICY_SEVERITY_LOW,
			wantContains: "Unable to estimate cost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diag := CostToDiagnostic(tt.cost, tt.urn, tt.version)

			require.NotNil(t, diag)
			assert.Equal(t, policyNameCost, diag.GetPolicyName())
			assert.Equal(t, policyPackName, diag.GetPolicyPackName())
			assert.Equal(t, tt.version, diag.GetPolicyPackVersion())
			assert.Equal(t, tt.urn, diag.GetUrn())
			assert.Equal(t, pulumirpc.EnforcementLevel_ADVISORY, diag.GetEnforcementLevel())
			assert.Equal(t, tt.wantSeverity, diag.GetSeverity())
			assert.Contains(t, diag.GetMessage(), tt.wantContains)
		})
	}
}

func TestStackSummaryDiagnostic(t *testing.T) {
	tests := []struct {
		name         string
		costs        []engine.CostResult
		version      string
		wantTotal    string
		wantAnalyzed string
	}{
		{
			name: "multiple resources",
			costs: []engine.CostResult{
				{Monthly: 8.45, Currency: "USD"},
				{Monthly: 25.00, Currency: "USD"},
				{Monthly: 0.50, Currency: "USD"},
			},
			version:      "0.1.0",
			wantTotal:    "$33.95 USD",
			wantAnalyzed: "3 resources analyzed",
		},
		{
			name: "mixed costs (some zero)",
			costs: []engine.CostResult{
				{Monthly: 10.00, Currency: "USD"},
				{Monthly: 0, Currency: "USD"},
				{Monthly: 20.00, Currency: "USD"},
			},
			version:      "0.1.0",
			wantTotal:    "$30.00 USD",
			wantAnalyzed: "2 resources analyzed",
		},
		{
			name:         "empty costs",
			costs:        []engine.CostResult{},
			version:      "0.1.0",
			wantTotal:    "$0.00 USD",
			wantAnalyzed: "0 resources analyzed",
		},
		{
			name:         "nil costs",
			costs:        nil,
			version:      "0.1.0",
			wantTotal:    "$0.00 USD",
			wantAnalyzed: "0 resources analyzed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diag := StackSummaryDiagnostic(tt.costs, tt.version)

			require.NotNil(t, diag)
			assert.Equal(t, policyNameSum, diag.GetPolicyName())
			assert.Equal(t, policyPackName, diag.GetPolicyPackName())
			assert.Equal(t, tt.version, diag.GetPolicyPackVersion())
			assert.Empty(t, diag.GetUrn()) // Stack-level has no URN
			assert.Equal(t, pulumirpc.EnforcementLevel_ADVISORY, diag.GetEnforcementLevel())
			assert.Equal(t, pulumirpc.PolicySeverity_POLICY_SEVERITY_LOW, diag.GetSeverity())
			assert.Contains(t, diag.GetMessage(), tt.wantTotal)
			assert.Contains(t, diag.GetMessage(), tt.wantAnalyzed)
		})
	}
}

func TestFormatCostMessage(t *testing.T) {
	tests := []struct {
		name string
		cost engine.CostResult
		want string
	}{
		{
			name: "has monthly cost",
			cost: engine.CostResult{
				Monthly:  25.50,
				Currency: "USD",
				Adapter:  "vantage",
			},
			want: "Estimated Monthly Cost: $25.50 USD (source: vantage)",
		},
		{
			name: "zero cost with notes",
			cost: engine.CostResult{
				Monthly: 0,
				Notes:   "Plugin returned no pricing data",
			},
			want: "Plugin returned no pricing data",
		},
		{
			name: "zero cost no notes",
			cost: engine.CostResult{
				Monthly: 0,
			},
			want: "Unable to estimate cost",
		},
		{
			name: "small cost",
			cost: engine.CostResult{
				Monthly:  0.01,
				Currency: "USD",
				Adapter:  "local-spec",
			},
			want: "Estimated Monthly Cost: $0.01 USD (source: local-spec)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatCostMessage(tt.cost)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCostToDiagnostic_EnforcementLevel(t *testing.T) {
	// Verify all cost diagnostics use ADVISORY (never ERROR in MVP)
	cost := engine.CostResult{
		ResourceType: "aws:ec2/instance:Instance",
		ResourceID:   "expensive-server",
		Monthly:      10000.00, // Very expensive
		Currency:     "USD",
	}

	diag := CostToDiagnostic(
		cost,
		"urn:pulumi:dev::myapp::aws:ec2/instance:Instance::expensive-server",
		"0.1.0",
	)

	// Must be ADVISORY per FR-005
	assert.Equal(t, pulumirpc.EnforcementLevel_ADVISORY, diag.GetEnforcementLevel())
	// Should never be ERROR, even for high costs
	assert.NotEqual(t, pulumirpc.EnforcementLevel_MANDATORY, diag.GetEnforcementLevel())
}

func TestStackSummaryDiagnostic_CurrencyHandling(t *testing.T) {
	// Test that currency is properly extracted from results
	costs := []engine.CostResult{
		{Monthly: 10.00, Currency: ""},
		{Monthly: 20.00, Currency: "EUR"},
		{Monthly: 30.00, Currency: "USD"},
	}

	diag := StackSummaryDiagnostic(costs, "0.1.0")

	// Should use the last non-empty currency
	assert.Contains(t, diag.GetMessage(), "USD")
}

// Phase 5 (US3) - Warning Diagnostic Tests

func TestWarningDiagnostic(t *testing.T) {
	tests := []struct {
		name    string
		message string
		urn     string
		version string
	}{
		{
			name:    "plugin timeout warning",
			message: "Cost estimation timed out, using fallback pricing",
			urn:     "urn:pulumi:dev::myapp::aws:ec2/instance:Instance::web",
			version: "0.1.0",
		},
		{
			name:    "network failure warning",
			message: "Network unavailable, using cached pricing specs",
			urn:     "urn:pulumi:dev::myapp::aws:rds/instance:Instance::db",
			version: "0.1.0",
		},
		{
			name:    "unsupported resource warning",
			message: "Unsupported resource type: custom:component:Widget",
			urn:     "urn:pulumi:dev::myapp::custom:component:Widget::w1",
			version: "0.1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diag := WarningDiagnostic(tt.message, tt.urn, tt.version)

			require.NotNil(t, diag)
			assert.Equal(t, policyNameCost, diag.GetPolicyName())
			assert.Equal(t, policyPackName, diag.GetPolicyPackName())
			assert.Equal(t, tt.version, diag.GetPolicyPackVersion())
			assert.Equal(t, tt.urn, diag.GetUrn())
			assert.Equal(t, pulumirpc.EnforcementLevel_ADVISORY, diag.GetEnforcementLevel())
			assert.Equal(t, pulumirpc.PolicySeverity_POLICY_SEVERITY_MEDIUM, diag.GetSeverity())
			assert.Equal(t, tt.message, diag.GetMessage())
		})
	}
}

func TestWarningDiagnostic_NoURN(t *testing.T) {
	// Stack-level warnings have no URN
	diag := WarningDiagnostic("Unable to connect to pricing API", "", "0.1.0")

	require.NotNil(t, diag)
	assert.Empty(t, diag.GetUrn())
	assert.Equal(t, pulumirpc.PolicySeverity_POLICY_SEVERITY_MEDIUM, diag.GetSeverity())
}

func TestCostToDiagnostic_ErrorInNotes(t *testing.T) {
	// When cost calculation fails, the error appears in Notes
	cost := engine.CostResult{
		ResourceType: "aws:lambda/function:Function",
		ResourceID:   "api-handler",
		Adapter:      "none",
		Currency:     "USD",
		Monthly:      0,
		Notes:        "ERROR: Plugin vantage failed: connection refused",
	}

	diag := CostToDiagnostic(
		cost,
		"urn:pulumi:dev::myapp::aws:lambda/function:Function::api-handler",
		"0.1.0",
	)

	// Should report the error in the message
	assert.Contains(t, diag.GetMessage(), "ERROR")
	assert.Contains(t, diag.GetMessage(), "Plugin vantage failed")
	// Should use MEDIUM severity for errors
	assert.Equal(t, pulumirpc.PolicySeverity_POLICY_SEVERITY_MEDIUM, diag.GetSeverity())
}
