package analyzer

import (
	"context"
	"strings"
	"testing"

	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/rshade/pulumicost-core/internal/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

// mockCostCalculator implements the CostCalculator interface for testing.
type mockCostCalculator struct {
	results []engine.CostResult
	err     error
}

func (m *mockCostCalculator) GetProjectedCost(
	_ context.Context,
	_ []engine.ResourceDescriptor,
) ([]engine.CostResult, error) {
	return m.results, m.err
}

func TestNewServer(t *testing.T) {
	calc := &mockCostCalculator{}
	server := NewServer(calc, "1.0.0")

	require.NotNil(t, server)
	assert.Equal(t, "1.0.0", server.version)
}

func TestServer_AnalyzeStack(t *testing.T) {
	tests := []struct {
		name          string
		resources     []*pulumirpc.AnalyzerResource
		calcResults   []engine.CostResult
		calcErr       error
		wantDiagCount int
		wantContains  []string
	}{
		{
			name: "single resource with cost",
			resources: []*pulumirpc.AnalyzerResource{
				{
					Type: "aws:ec2/instance:Instance",
					Urn:  "urn:pulumi:dev::myapp::aws:ec2/instance:Instance::webserver",
					Name: "webserver",
				},
			},
			calcResults: []engine.CostResult{
				{
					ResourceType: "aws:ec2/instance:Instance",
					ResourceID:   "webserver",
					Adapter:      "local-spec",
					Currency:     "USD",
					Monthly:      8.45,
					Hourly:       0.0116,
				},
			},
			wantDiagCount: 2, // 1 per-resource + 1 summary
			wantContains:  []string{"$8.45 USD", "Total Estimated Monthly Cost"},
		},
		{
			name: "multiple resources",
			resources: []*pulumirpc.AnalyzerResource{
				{
					Type: "aws:ec2/instance:Instance",
					Urn:  "urn:pulumi:dev::myapp::aws:ec2/instance:Instance::web1",
					Name: "web1",
				},
				{
					Type: "aws:rds/instance:Instance",
					Urn:  "urn:pulumi:dev::myapp::aws:rds/instance:Instance::db1",
					Name: "db1",
				},
			},
			calcResults: []engine.CostResult{
				{
					ResourceType: "aws:ec2/instance:Instance",
					ResourceID:   "web1",
					Adapter:      "local-spec",
					Currency:     "USD",
					Monthly:      8.45,
				},
				{
					ResourceType: "aws:rds/instance:Instance",
					ResourceID:   "db1",
					Adapter:      "local-spec",
					Currency:     "USD",
					Monthly:      50.00,
				},
			},
			wantDiagCount: 3, // 2 per-resource + 1 summary
			wantContains:  []string{"$8.45 USD", "$50.00 USD", "2 resources analyzed"},
		},
		{
			name:          "empty resources",
			resources:     []*pulumirpc.AnalyzerResource{},
			calcResults:   []engine.CostResult{},
			wantDiagCount: 1, // summary only
			wantContains:  []string{"$0.00 USD", "0 resources analyzed"},
		},
		{
			name: "resource with zero cost (unsupported)",
			resources: []*pulumirpc.AnalyzerResource{
				{
					Type: "custom:component:Widget",
					Urn:  "urn:pulumi:dev::myapp::custom:component:Widget::widget1",
					Name: "widget1",
				},
			},
			calcResults: []engine.CostResult{
				{
					ResourceType: "custom:component:Widget",
					ResourceID:   "widget1",
					Adapter:      "none",
					Currency:     "USD",
					Monthly:      0,
					Notes:        "Unsupported resource type",
				},
			},
			wantDiagCount: 2, // 1 per-resource + 1 summary
			wantContains:  []string{"Unsupported resource type"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := &mockCostCalculator{
				results: tt.calcResults,
				err:     tt.calcErr,
			}
			server := NewServer(calc, "0.1.0")

			req := &pulumirpc.AnalyzeStackRequest{
				Resources: tt.resources,
			}

			resp, err := server.AnalyzeStack(context.Background(), req)

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Len(t, resp.GetDiagnostics(), tt.wantDiagCount)

			// Verify expected content is present in diagnostics
			var messages []string
			for _, diag := range resp.GetDiagnostics() {
				messages = append(messages, diag.GetMessage())
			}
			allMessages := strings.Join(messages, " ")
			for _, want := range tt.wantContains {
				assert.Contains(t, allMessages, want)
			}
		})
	}
}

func TestServer_AnalyzeStack_WithProperties(t *testing.T) {
	// Test that properties are correctly passed through to the cost calculator
	props, err := structpb.NewStruct(map[string]interface{}{
		"instanceType": "t3.micro",
		"ami":          "ami-0123456789abcdef0",
	})
	require.NoError(t, err)

	resources := []*pulumirpc.AnalyzerResource{
		{
			Type:       "aws:ec2/instance:Instance",
			Urn:        "urn:pulumi:dev::myapp::aws:ec2/instance:Instance::web",
			Name:       "web",
			Properties: props,
		},
	}

	calc := &mockCostCalculator{
		results: []engine.CostResult{
			{
				ResourceType: "aws:ec2/instance:Instance",
				ResourceID:   "web",
				Adapter:      "local-spec",
				Currency:     "USD",
				Monthly:      7.59,
			},
		},
	}
	server := NewServer(calc, "0.1.0")

	req := &pulumirpc.AnalyzeStackRequest{Resources: resources}
	resp, err := server.AnalyzeStack(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.GetDiagnostics(), 2)
}

func TestServer_AnalyzeStack_DiagnosticFields(t *testing.T) {
	// Verify diagnostic fields are correctly set
	resources := []*pulumirpc.AnalyzerResource{
		{
			Type: "aws:ec2/instance:Instance",
			Urn:  "urn:pulumi:dev::myapp::aws:ec2/instance:Instance::web",
			Name: "web",
		},
	}

	calc := &mockCostCalculator{
		results: []engine.CostResult{
			{
				ResourceType: "aws:ec2/instance:Instance",
				ResourceID:   "web",
				Adapter:      "local-spec",
				Currency:     "USD",
				Monthly:      10.00,
			},
		},
	}
	server := NewServer(calc, "1.2.3")

	req := &pulumirpc.AnalyzeStackRequest{Resources: resources}
	resp, err := server.AnalyzeStack(context.Background(), req)

	require.NoError(t, err)
	require.Len(t, resp.GetDiagnostics(), 2)

	// Check per-resource diagnostic
	resourceDiag := resp.GetDiagnostics()[0]
	assert.Equal(t, "cost-estimate", resourceDiag.GetPolicyName())
	assert.Equal(t, "pulumicost", resourceDiag.GetPolicyPackName())
	assert.Equal(t, "1.2.3", resourceDiag.GetPolicyPackVersion())
	assert.Equal(t, "urn:pulumi:dev::myapp::aws:ec2/instance:Instance::web", resourceDiag.GetUrn())
	assert.Equal(t, pulumirpc.EnforcementLevel_ADVISORY, resourceDiag.GetEnforcementLevel())

	// Check summary diagnostic
	summaryDiag := resp.GetDiagnostics()[1]
	assert.Equal(t, "stack-cost-summary", summaryDiag.GetPolicyName())
	assert.Empty(t, summaryDiag.GetUrn()) // Stack-level has no URN
}

func TestServer_GetAnalyzerInfo(t *testing.T) {
	calc := &mockCostCalculator{}
	server := NewServer(calc, "0.2.0")

	resp, err := server.GetAnalyzerInfo(context.Background(), &emptypb.Empty{})

	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, "pulumicost", resp.GetName())
	assert.Equal(t, "PulumiCost Analyzer", resp.GetDisplayName())
	assert.Equal(t, "0.2.0", resp.GetVersion())
	assert.NotEmpty(t, resp.GetDescription())

	// Check policies are defined
	require.Len(t, resp.GetPolicies(), 2)

	// Check cost-estimate policy
	costPolicy := resp.GetPolicies()[0]
	assert.Equal(t, "cost-estimate", costPolicy.GetName())
	assert.Equal(t, pulumirpc.EnforcementLevel_ADVISORY, costPolicy.GetEnforcementLevel())

	// Check stack-cost-summary policy
	summaryPolicy := resp.GetPolicies()[1]
	assert.Equal(t, "stack-cost-summary", summaryPolicy.GetName())
	assert.Equal(t, pulumirpc.EnforcementLevel_ADVISORY, summaryPolicy.GetEnforcementLevel())
}

func TestServer_GetPluginInfo(t *testing.T) {
	calc := &mockCostCalculator{}
	server := NewServer(calc, "1.5.0")

	resp, err := server.GetPluginInfo(context.Background(), &emptypb.Empty{})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "1.5.0", resp.GetVersion())
}

func TestServer_GetPluginInfo_EmptyVersion(t *testing.T) {
	calc := &mockCostCalculator{}
	server := NewServer(calc, "")

	resp, err := server.GetPluginInfo(context.Background(), &emptypb.Empty{})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "0.0.0-dev", resp.GetVersion()) // Should use default
}

func TestServer_Handshake(t *testing.T) {
	rootDir := "/home/user/.pulumi"
	programDir := "/home/user/myproject"

	tests := []struct {
		name     string
		request  *pulumirpc.AnalyzerHandshakeRequest
		wantErr  bool
		checkCtx bool
	}{
		{
			name: "standard handshake with engine address",
			request: &pulumirpc.AnalyzerHandshakeRequest{
				EngineAddress:    "localhost:12345",
				RootDirectory:    &rootDir,
				ProgramDirectory: &programDir,
			},
			wantErr: false,
		},
		{
			name: "handshake with minimal data",
			request: &pulumirpc.AnalyzerHandshakeRequest{
				EngineAddress: "127.0.0.1:9999",
			},
			wantErr: false,
		},
		{
			name:    "handshake with nil request fields",
			request: &pulumirpc.AnalyzerHandshakeRequest{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := &mockCostCalculator{}
			server := NewServer(calc, "1.0.0")

			resp, err := server.Handshake(context.Background(), tt.request)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
		})
	}
}

func TestServer_ConfigureStack(t *testing.T) {
	tests := []struct {
		name    string
		request *pulumirpc.AnalyzerStackConfigureRequest
		wantErr bool
	}{
		{
			name: "full stack configuration",
			request: &pulumirpc.AnalyzerStackConfigureRequest{
				Stack:        "dev",
				Project:      "my-infrastructure",
				Organization: "myorg",
				DryRun:       true,
				Config:       map[string]string{"aws:region": "us-east-1"},
				Tags:         map[string]string{"environment": "development"},
			},
			wantErr: false,
		},
		{
			name: "minimal stack configuration",
			request: &pulumirpc.AnalyzerStackConfigureRequest{
				Stack:   "prod",
				Project: "api",
			},
			wantErr: false,
		},
		{
			name:    "empty configuration",
			request: &pulumirpc.AnalyzerStackConfigureRequest{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := &mockCostCalculator{}
			server := NewServer(calc, "1.0.0")

			resp, err := server.ConfigureStack(context.Background(), tt.request)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
		})
	}
}

func TestServer_ConfigureStack_StoresContext(t *testing.T) {
	calc := &mockCostCalculator{}
	server := NewServer(calc, "1.0.0")

	req := &pulumirpc.AnalyzerStackConfigureRequest{
		Stack:        "staging",
		Project:      "web-app",
		Organization: "acme-corp",
		DryRun:       true,
	}

	_, err := server.ConfigureStack(context.Background(), req)
	require.NoError(t, err)

	// Verify server stored the stack context
	assert.Equal(t, "staging", server.stackName)
	assert.Equal(t, "web-app", server.projectName)
	assert.Equal(t, "acme-corp", server.organization)
	assert.True(t, server.dryRun)
}

func TestServer_Cancel(t *testing.T) {
	calc := &mockCostCalculator{}
	server := NewServer(calc, "1.0.0")

	// Initially not canceled
	assert.False(t, server.IsCanceled())

	// Call Cancel
	resp, err := server.Cancel(context.Background(), &emptypb.Empty{})

	require.NoError(t, err)
	require.NotNil(t, resp)

	// Now should be canceled
	assert.True(t, server.IsCanceled())
}

func TestServer_Cancel_ThreadSafe(t *testing.T) {
	calc := &mockCostCalculator{}
	server := NewServer(calc, "1.0.0")

	// Call Cancel concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := server.Cancel(context.Background(), &emptypb.Empty{})
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should be canceled
	assert.True(t, server.IsCanceled())
}
