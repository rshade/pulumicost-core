//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AnalyzerDiagnostic represents an expected diagnostic pattern in preview output.
type AnalyzerDiagnostic struct {
	Pattern  string
	Required bool
	Count    int // 0 means any number > 0 if Required is true, or specific count
}

func validateDiagnostics(t *testing.T, output string, diagnostics []AnalyzerDiagnostic) {
	for _, d := range diagnostics {
		count := strings.Count(output, d.Pattern)
		if d.Required {
			assert.Greater(t, count, 0, "Expected diagnostic pattern not found: %s", d.Pattern)
		}
		if d.Count > 0 {
			assert.Equal(t, d.Count, count,
				"Expected %d occurrences of pattern '%s', got %d", d.Count, d.Pattern, count)
		}
	}
}

// ensureAnalyzerConfig sets up the policy pack directory for the analyzer.
// This creates a directory with:
//   - PulumiPolicy.yaml (required for policy pack loading)
//   - The pulumicost binary renamed to pulumi-analyzer-policy-pulumicost
//
// Returns the absolute path to the policy pack directory.
func ensureAnalyzerConfig(t *testing.T, projectDir string) string {
	binaryPath := findPulumicostBinary()
	require.NotEmpty(t, binaryPath, "pulumicost binary not found")

	// Create a directory for the policy pack
	policyPackDir := filepath.Join(projectDir, "policy-pack")
	err := os.MkdirAll(policyPackDir, 0755)
	require.NoError(t, err)

	// Create PulumiPolicy.yaml (required for policy pack loading)
	policyYaml := `runtime: pulumicost
name: pulumicost
version: 0.0.0-dev
`
	err = os.WriteFile(filepath.Join(policyPackDir, "PulumiPolicy.yaml"), []byte(policyYaml), 0644)
	require.NoError(t, err)

	// Determine destination binary name (pulumi-analyzer-policy-pulumicost)
	// Preserve extension if present (e.g., .exe)
	ext := filepath.Ext(binaryPath)
	destBinaryName := "pulumi-analyzer-policy-pulumicost" + ext
	destBinaryPath := filepath.Join(policyPackDir, destBinaryName)

	// Copy the binary
	input, err := os.ReadFile(binaryPath)
	require.NoError(t, err)
	err = os.WriteFile(destBinaryPath, input, 0755)
	require.NoError(t, err)

	absPolicyPackDir, err := filepath.Abs(policyPackDir)
	require.NoError(t, err)

	return absPolicyPackDir
}

func skipIfPulumiNotInstalled(t *testing.T) {
	if _, err := exec.LookPath("pulumi"); err != nil {
		t.Skip("Pulumi CLI not installed, skipping analyzer E2E test")
	}
}

// runAnalyzerPreview sets up the analyzer project and runs pulumi preview, returning the output.
func runAnalyzerPreview(t *testing.T, tc *TestContext) string {
	skipIfPulumiNotInstalled(t)
	ctx := context.Background()

	// Install aws-public plugin (assumed needed for all tests for now)
	pm := NewPluginManager(t)
	err := pm.EnsurePluginInstalled(ctx, "aws-public")
	require.NoError(t, err, "Failed to install aws-public plugin")
	defer pm.DeferPluginCleanup(ctx, "aws-public")()

	// Prepare project
	projectPath := "projects/analyzer"
	absProjectPath, err := filepath.Abs(projectPath)
	require.NoError(t, err)
	tc.ProjectDir = absProjectPath

	workDir, err := os.MkdirTemp("", "pulumi-e2e-analyzer-")
	require.NoError(t, err)
	tc.WorkDir = workDir
	tc.PreviewJSON = filepath.Join(workDir, "preview.json")

	err = tc.CopyDir(absProjectPath, workDir)
	require.NoError(t, err)

	// Set up policy pack and get its path
	policyPackDir := ensureAnalyzerConfig(t, workDir)

	// Init Pulumi
	stateDir := filepath.Join(workDir, ".pulumi-state")
	err = os.MkdirAll(stateDir, 0755)
	require.NoError(t, err)

	// Add policy pack directory to PATH so Pulumi can find the binary
	pathEnv := policyPackDir + string(os.PathListSeparator) + os.Getenv("PATH")
	env := []string{
		"PULUMI_BACKEND_URL=file://" + stateDir,
		"PATH=" + pathEnv,
	}
	err = tc.runCmdWithEnv(ctx, workDir, env, "pulumi", "login", "--local")
	require.NoError(t, err)
	err = tc.runCmdWithEnv(ctx, workDir, env, "pulumi", "stack", "init", tc.StackName)
	require.NoError(t, err)
	err = tc.runCmdWithEnv(ctx, workDir, env, "pulumi", "config", "set", "aws:region", tc.Region)
	require.NoError(t, err)

	// Run preview with --policy-pack flag to activate the analyzer
	outputBytes, err := tc.runCmdOutput(ctx, workDir, env, "pulumi", "preview", "--policy-pack", policyPackDir)
	require.NoError(t, err, "pulumi preview failed")
	return string(outputBytes)
}

func TestAnalyzer_Handshake(t *testing.T) {
	tc := NewTestContext(t, "e2e-analyzer-handshake")
	ctx := context.Background()
	defer tc.Teardown(ctx)

	output := runAnalyzerPreview(t, tc)
	t.Logf("Preview output:\n%s", output)
	// Implicit success check in runAnalyzerPreview (require.NoError)
}

func TestAnalyzer_CostDiagnostics(t *testing.T) {
	tc := NewTestContext(t, "e2e-analyzer-diagnostics")
	ctx := context.Background()
	defer tc.Teardown(ctx)

	output := runAnalyzerPreview(t, tc)

	diagnostics := []AnalyzerDiagnostic{
		{Pattern: "Estimated Monthly Cost:", Required: true},
		{Pattern: "pulumicost", Required: true},
	}
	validateDiagnostics(t, output, diagnostics)
}

func TestAnalyzer_StackSummary(t *testing.T) {
	tc := NewTestContext(t, "e2e-analyzer-summary")
	ctx := context.Background()
	defer tc.Teardown(ctx)

	output := runAnalyzerPreview(t, tc)

	diagnostics := []AnalyzerDiagnostic{
		{Pattern: "Total Estimated Monthly Cost:", Required: true, Count: 1},
	}
	validateDiagnostics(t, output, diagnostics)
}

func TestAnalyzer_GracefulDegradation(t *testing.T) {
	skipIfPulumiNotInstalled(t)
	tc := NewTestContext(t, "e2e-analyzer-degradation")
	ctx := context.Background()
	defer tc.Teardown(ctx)

	// Install aws-public plugin
	pm := NewPluginManager(t)
	err := pm.EnsurePluginInstalled(ctx, "aws-public")
	require.NoError(t, err, "Failed to install aws-public plugin")
	defer pm.DeferPluginCleanup(ctx, "aws-public")()

	// Prepare project
	projectPath := "projects/analyzer"
	absProjectPath, err := filepath.Abs(projectPath)
	require.NoError(t, err)
	tc.ProjectDir = absProjectPath

	workDir, err := os.MkdirTemp("", "pulumi-e2e-analyzer-")
	require.NoError(t, err)
	tc.WorkDir = workDir
	tc.PreviewJSON = filepath.Join(workDir, "preview.json")

	err = tc.CopyDir(absProjectPath, workDir)
	require.NoError(t, err)

	// Set up policy pack and get its path
	policyPackDir := ensureAnalyzerConfig(t, workDir)

	// Modify Pulumi.yaml to use invalid instance type to trigger pricing failure
	yamlPath := filepath.Join(workDir, "Pulumi.yaml")
	content, err := os.ReadFile(yamlPath)
	require.NoError(t, err)
	newContent := strings.Replace(string(content), "t3.micro", "invalid.type.xlarge", 1)
	err = os.WriteFile(yamlPath, []byte(newContent), 0644)
	require.NoError(t, err)

	// Init Pulumi
	stateDir := filepath.Join(workDir, ".pulumi-state")
	err = os.MkdirAll(stateDir, 0755)
	require.NoError(t, err)

	// Add policy pack directory to PATH so Pulumi can find the binary
	pathEnv := policyPackDir + string(os.PathListSeparator) + os.Getenv("PATH")
	env := []string{
		"PULUMI_BACKEND_URL=file://" + stateDir,
		"PATH=" + pathEnv,
	}
	err = tc.runCmdWithEnv(ctx, workDir, env, "pulumi", "login", "--local")
	require.NoError(t, err)
	err = tc.runCmdWithEnv(ctx, workDir, env, "pulumi", "stack", "init", tc.StackName)
	require.NoError(t, err)
	err = tc.runCmdWithEnv(ctx, workDir, env, "pulumi", "config", "set", "aws:region", tc.Region)
	require.NoError(t, err)

	// Run preview with --policy-pack flag
	outputBytes, err := tc.runCmdOutput(ctx, workDir, env, "pulumi", "preview", "--policy-pack", policyPackDir)

	// We expect success (graceful degradation)
	require.NoError(t, err, "Preview should succeed even with pricing failure")
	output := string(outputBytes)

	t.Logf("Preview output:\n%s", output)

	// Verify analyzer didn't crash and output produced something
	assert.Contains(t, output, "pulumicost", "Analyzer should still run")
}
