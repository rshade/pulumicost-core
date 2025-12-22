package cli_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/rshade/pulumicost-core/test/integration/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectedCost_FilterByType(t *testing.T) {
	h := helpers.NewCLIHelper(t)
	planFile := filepath.Join("..", "..", "..", "test", "fixtures", "plans", "multi-resource-plan.json")

	// Execute projected cost command with type filter
	output, err := h.Execute(
		"cost", "projected", "--pulumi-json", planFile, "--filter",
		"type=aws:ec2/instance:Instance", "--output", "json",
	)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	resources := result["resources"].([]interface{})

	// Verify filtered results
	assert.NotEmpty(t, resources, "Expected matches for filter: type=aws:ec2/instance:Instance. Output: %s", output)
	for _, r := range resources {
		res := r.(map[string]interface{})
		// The key in JSON output is "resourceType"
		assert.Equal(t, "aws:ec2/instance:Instance", res["resourceType"])
	}
}

func TestProjectedCost_FilterByTypeSubstring(t *testing.T) {
	h := helpers.NewCLIHelper(t)
	planFile := filepath.Join("..", "..", "..", "test", "fixtures", "plans", "multi-resource-plan.json")

	// Filter by "bucket" substring
	output, err := h.Execute(
		"cost", "projected", "--pulumi-json", planFile,
		"--filter", "type=bucket", "--output", "json",
	)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	resources := result["resources"].([]interface{})
	assert.NotEmpty(t, resources)
	for _, r := range resources {
		res := r.(map[string]interface{})
		assert.Contains(t, res["resourceType"], "Bucket")
	}
}

func TestProjectedCost_FilterByProvider(t *testing.T) {
	h := helpers.NewCLIHelper(t)
	planFile := filepath.Join("..", "..", "..", "test", "fixtures", "plans", "multi-resource-plan.json")

	// Filter by "azure" provider
	output, err := h.Execute(
		"cost", "projected", "--pulumi-json", planFile,
		"--filter", "provider=azure", "--output", "json",
	)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	resources := result["resources"].([]interface{})
	assert.NotEmpty(t, resources)
	for _, r := range resources {
		res := r.(map[string]interface{})
		typeStr := res["resourceType"].(string)
		assert.Contains(t, typeStr, "azure")
	}
}

func TestActualCost_FilterByTag(t *testing.T) {
	// Skip for now if no plugin support or mock data
	// The cost actual command relies on plugins. Without a plugin that supports actual cost *and* tags, this might fail or return empty.
	// But we can test the filtering logic if we assume the engine does filtering *before* or *after* plugin call?
	// Actually, for GetActualCost, the engine passes filters to the plugin usually, or filters results?
	// Let's check engine.go... Filter by tags happens in GetActualCostWithOptions.
	// It calls MatchesTags.
	// So we can test this even if the plugin returns mock data, as long as we have a plan with resources that have tags.
	// But `cost actual` usually requires resource IDs from the plan.

	h := helpers.NewCLIHelper(t)
	planFile := filepath.Join("..", "..", "..", "test", "fixtures", "plans", "multi-resource-plan.json")

	// Filter by env=prod tag
	// Note: 'cost actual' needs --from/--to usually, but defaults exist.
	// We'll use defaults.
	output, err := h.Execute(
		"cost", "actual", "--pulumi-json", planFile,
		"--filter", "tag:env=prod", "--output", "json",
	)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	// Since we don't have a real actual cost plugin configured that returns data for these IDs,
	// we expect results (placeholders) but filtered to only those with the tag.
	resources := result["resources"].([]interface{})
	assert.NotEmpty(t, resources)

	for _, r := range resources {
		res := r.(map[string]interface{})
		// Verify these resources are the ones with env=prod in the fixture
		// i-123... (aws_instance) has env=prod
		// vm-azure-1 has env=prod
		// db-1 has env=prod
		// vm-gcp-1 has env=staging
		// my-bucket has env=dev

		id := res["resourceId"].(string)
		// assert that ID is one of the expected ones
		if id != "i-1234567890abcdef0" && id != "vm-azure-1" && id != "db-1" {
			t.Errorf("Found unexpected resource ID in filtered output: %s", id)
		}
	}
}

func TestActualCost_FilterByTagAndType(t *testing.T) {
	h := helpers.NewCLIHelper(t)
	planFile := filepath.Join("..", "..", "..", "test", "fixtures", "plans", "multi-resource-plan.json")

	// Filter by env=prod AND type=aws
	// Note: CLI supports one filter flag string. Does it support multiple expressions?
	// The implementation split by "=". It doesn't seem to support AND logic in one flag unless we pass multiple flags?
	// Cobra flags can be slice. But `filter` is defined as `StringVar`.
	// So probably only one filter at a time?
	// Wait, the spec says "Filter resources (tag:key=value, type=*)".
	// The implementation `matchesFilter` handles ONE key/value pair.
	// `FilterResources` loops through resources.

	// If we want multiple filters, we might need multiple flags if supported, or logic change.
	// Looking at `internal/cli/cost_projected.go`: `cmd.Flags().StringVar(&filter, ...)` -> Single string.
	// So currently only one filter condition is supported per command execution.
	// User Story 2 says "Actual Cost Filtering by Tags".
	// Scenario 2: "cost actual with both a group-by and a filter".

	// Let's test that specific scenario then.

	output, err := h.Execute(
		"cost", "actual", "--pulumi-json", planFile, "--filter",
		"tag:env=prod", "--group-by", "type", "--output", "json",
	)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	resources := result["resources"].([]interface{})
	assert.NotEmpty(t, resources)

	// With group-by type, we should see aggregated results for types that have env=prod resources
	// aws:ec2/instance:Instance (prod)
	// azure:compute/virtualMachine:VirtualMachine (prod)
	// aws:rds/instance:Instance (prod)

	foundEC2 := false
	foundAzure := false
	foundRDS := false

	for _, r := range resources {
		res := r.(map[string]interface{})
		rType := res["resourceType"].(string)
		if rType == "aws:ec2/instance:Instance" {
			foundEC2 = true
		}
		if rType == "azure:compute/virtualMachine:VirtualMachine" {
			foundAzure = true
		}
		if rType == "aws:rds/instance:Instance" {
			foundRDS = true
		}

		// Should NOT see gcp or s3
		if rType == "gcp:compute/instance:Instance" {
			t.Error("Found GCP resource, should be filtered out")
		}
		if rType == "aws:s3/bucket:Bucket" {
			t.Error("Found S3 bucket, should be filtered out")
		}
	}

	assert.True(t, foundEC2, "Should find EC2")
	assert.True(t, foundAzure, "Should find Azure VM")
	assert.True(t, foundRDS, "Should find RDS")
}
