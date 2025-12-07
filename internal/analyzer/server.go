package analyzer

import (
	"context"
	"sync"

	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/rshade/pulumicost-core/internal/engine"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	// defaultVersion is used when no version is provided.
	defaultVersion = "0.0.0-dev"

	// analyzerDisplayName is the human-readable name for the analyzer.
	analyzerDisplayName = "PulumiCost Analyzer"

	// analyzerDescription provides a description of the analyzer's purpose.
	analyzerDescription = "Provides real-time cost estimation for Pulumi infrastructure resources during preview operations."
)

// CostCalculator is the interface for calculating projected costs.
//
// This interface abstracts the cost calculation engine, allowing for
// easier testing and decoupling from the concrete engine implementation.
type CostCalculator interface {
	GetProjectedCost(ctx context.Context, resources []engine.ResourceDescriptor) ([]engine.CostResult, error)
}

// Server implements the Pulumi Analyzer gRPC service for cost estimation.
//
// The Server processes resources during `pulumi preview` and returns
// cost diagnostics that appear in the Pulumi CLI output. It coordinates
// with the cost calculation engine to estimate resource costs and
// formats the results as AnalyzeDiagnostic messages.
//
// Key behaviors:
//   - All diagnostics use ADVISORY enforcement (never blocks deployments)
//   - Reports both per-resource costs and a stack-level summary
//   - Handles unsupported resources gracefully with informative messages
type Server struct {
	pulumirpc.UnimplementedAnalyzerServer

	calculator CostCalculator
	version    string

	// Stack context from ConfigureStack RPC
	stackName    string
	projectName  string
	organization string
	dryRun       bool

	// Cancellation support
	cancelMu sync.Mutex
	canceled bool
}

// NewServer creates a new Analyzer server with the given cost calculator.
//
// Parameters:
//   - calculator: The cost calculation engine to use for estimating costs
//   - version: The version string for this analyzer plugin
//
// version is an empty string, it defaults to "0.0.0-dev".
func NewServer(calculator CostCalculator, version string) *Server {
	if version == "" {
		version = defaultVersion
	}
	return &Server{
		calculator: calculator,
		version:    version,
	}
}

// AnalyzeStack analyzes all resources in a stack and returns cost diagnostics.
//
// This method is called by the Pulumi engine at the end of a successful
// preview or update. It receives the complete list of resources and returns
// diagnostics with cost estimates.
//
// The response includes:
//   - Per-resource diagnostics with individual cost estimates
//   - A stack-level summary diagnostic with total costs
//
// All diagnostics use ADVISORY enforcement per FR-005.
func (s *Server) AnalyzeStack(
	ctx context.Context,
	req *pulumirpc.AnalyzeStackRequest,
) (*pulumirpc.AnalyzeResponse, error) {
	// Map Pulumi resources to internal format
	resources := MapResources(req.GetResources())

	// Calculate costs using the engine
	costs, err := s.calculator.GetProjectedCost(ctx, resources)
	if err != nil {
		// Log error but don't fail - return empty diagnostics
		// This ensures preview continues even if cost calculation fails
		costs = []engine.CostResult{}
	}

	// Build diagnostics list
	diagnostics := make([]*pulumirpc.AnalyzeDiagnostic, 0, len(costs)+1)

	// Add per-resource diagnostics
	for i, cost := range costs {
		urn := ""
		if i < len(req.GetResources()) {
			urn = req.GetResources()[i].GetUrn()
		}
		diag := CostToDiagnostic(cost, urn, s.version)
		diagnostics = append(diagnostics, diag)
	}

	// Add stack summary diagnostic
	summary := StackSummaryDiagnostic(costs, s.version)
	diagnostics = append(diagnostics, summary)

	return &pulumirpc.AnalyzeResponse{
		Diagnostics: diagnostics,
	}, nil
}

// GetAnalyzerInfo returns metadata about this analyzer.
//
// This method provides information about the policies contained in this
// analyzer, including their enforcement levels and descriptions.
func (s *Server) GetAnalyzerInfo(
	_ context.Context,
	_ *emptypb.Empty,
) (*pulumirpc.AnalyzerInfo, error) {
	return &pulumirpc.AnalyzerInfo{
		Name:        policyPackName,
		DisplayName: analyzerDisplayName,
		Version:     s.version,
		Description: analyzerDescription,
		Policies: []*pulumirpc.PolicyInfo{
			{
				Name:             policyNameCost,
				DisplayName:      "Cost Estimate",
				Description:      "Provides estimated monthly cost for individual resources",
				EnforcementLevel: pulumirpc.EnforcementLevel_ADVISORY,
			},
			{
				Name:             policyNameSum,
				DisplayName:      "Stack Cost Summary",
				Description:      "Provides total estimated monthly cost across all resources in the stack",
				EnforcementLevel: pulumirpc.EnforcementLevel_ADVISORY,
			},
		},
		SupportsConfig: false,
	}, nil
}

// GetPluginInfo returns generic information about this plugin.
//
// This method returns the plugin version which is used by the Pulumi
// engine for version compatibility checks.
func (s *Server) GetPluginInfo(
	_ context.Context,
	_ *emptypb.Empty,
) (*pulumirpc.PluginInfo, error) {
	return &pulumirpc.PluginInfo{
		Version: s.version,
	}, nil
}

// Handshake establishes connection with the Pulumi engine.
//
// This method is called immediately after the plugin starts. It receives
// the engine's gRPC address and directory information. For the cost
// analyzer, we simply acknowledge the handshake as we don't need to
// make callbacks to the engine.
func (s *Server) Handshake(
	_ context.Context,
	_ *pulumirpc.AnalyzerHandshakeRequest,
) (*pulumirpc.AnalyzerHandshakeResponse, error) {
	// Store engine address if needed for future callbacks (not used in MVP)
	// The handshake is primarily informational - we just acknowledge it
	return &pulumirpc.AnalyzerHandshakeResponse{}, nil
}

// ConfigureStack receives stack context before analysis begins.
//
// This method is called before AnalyzeStack to provide context about
// the stack being analyzed. The information is stored for logging
// and diagnostic purposes.
func (s *Server) ConfigureStack(
	_ context.Context,
	req *pulumirpc.AnalyzerStackConfigureRequest,
) (*pulumirpc.AnalyzerStackConfigureResponse, error) {
	// Store stack context for logging and diagnostic enrichment
	s.stackName = req.GetStack()
	s.projectName = req.GetProject()
	s.organization = req.GetOrganization()
	s.dryRun = req.GetDryRun()

	return &pulumirpc.AnalyzerStackConfigureResponse{}, nil
}

// Cancel signals graceful shutdown of the analyzer.
//
// This method is called when the Pulumi engine is shutting down or
// when the user cancels the operation. It sets the canceled flag
// which can be checked by long-running operations.
func (s *Server) Cancel(
	_ context.Context,
	_ *emptypb.Empty,
) (*emptypb.Empty, error) {
	s.cancelMu.Lock()
	defer s.cancelMu.Unlock()
	s.canceled = true
	return &emptypb.Empty{}, nil
}

// IsCanceled returns true if the Cancel RPC has been called.
//
// This method is thread-safe and can be called concurrently.
func (s *Server) IsCanceled() bool {
	s.cancelMu.Lock()
	defer s.cancelMu.Unlock()
	return s.canceled
}