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

	resources, ok := result["resources"].([]interface{})
	require.True(t, ok, "expected resources to be an array")

	// Verify filtered results
	assert.NotEmpty(t, resources, "Expected matches for filter: type=aws:ec2/instance:Instance. Output: %s", output)
	for _, r := range resources {
		res, ok := r.(map[string]interface{})
		require.True(t, ok, "expected resource to be an object")
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

	resources, ok := result["resources"].([]interface{})
	require.True(t, ok, "expected resources to be an array")
	assert.NotEmpty(t, resources)
	for _, r := range resources {
		res, ok := r.(map[string]interface{})
		require.True(t, ok, "expected resource to be an object")
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

	resources, ok := result["resources"].([]interface{})
	require.True(t, ok, "expected resources to be an array")
	assert.NotEmpty(t, resources)
	for _, r := range resources {
		res, ok := r.(map[string]interface{})
		require.True(t, ok, "expected resource to be an object")
		typeStr, ok := res["resourceType"].(string)
		require.True(t, ok, "expected resourceType to be a string")
		assert.Contains(t, typeStr, "azure")
	}
}

func TestActualCost_FilterByTag(t *testing.T) {
	h := helpers.NewCLIHelper(t)
	planFile := filepath.Join("..", "..", "..", "test", "fixtures", "plans", "multi-resource-plan.json")

	// Filter by env=prod tag
	output, err := h.Execute(
		"cost", "actual", "--pulumi-json", planFile,
		"--filter", "tag:env=prod", "--output", "json",
	)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	// Expect results filtered to only those with the tag
	resources, ok := result["resources"].([]interface{})
	require.True(t, ok, "expected resources to be an array")
	assert.NotEmpty(t, resources)

	for _, r := range resources {
		res, ok := r.(map[string]interface{})
		require.True(t, ok, "expected resource to be an object")

		id, ok := res["resourceId"].(string)
		require.True(t, ok, "expected resourceId to be a string")
		// Only these resources have env=prod in the fixture
		assert.True(t, id == "i-1234567890abcdef0" || id == "vm-azure-1" || id == "db-1",
			"Found unexpected resource ID in filtered output: %s", id)
	}
}

func TestActualCost_FilterByTagAndType(t *testing.T) {
	h := helpers.NewCLIHelper(t)
	planFile := filepath.Join("..", "..", "..", "test", "fixtures", "plans", "multi-resource-plan.json")

	// Test filter combined with group-by
	output, err := h.Execute(
		"cost", "actual", "--pulumi-json", planFile, "--filter",
		"tag:env=prod", "--group-by", "type", "--output", "json",
	)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	resources, ok := result["resources"].([]interface{})
	require.True(t, ok, "expected resources to be an array")
	assert.NotEmpty(t, resources)

	// With group-by type, we should see aggregated results for types that have env=prod resources
	foundEC2 := false
	foundAzure := false
	foundRDS := false

	for _, r := range resources {
		res, ok := r.(map[string]interface{})
		require.True(t, ok, "expected resource to be an object")
		rType, ok := res["resourceType"].(string)
		require.True(t, ok, "expected resourceType to be a string")

		switch rType {
		case "aws:ec2/instance:Instance":
			foundEC2 = true
		case "azure:compute/virtualMachine:VirtualMachine":
			foundAzure = true
		case "aws:rds/instance:Instance":
			foundRDS = true
		case "gcp:compute/instance:Instance":
			t.Error("Found GCP resource, should be filtered out")
		case "aws:s3/bucket:Bucket":
			t.Error("Found S3 bucket, should be filtered out")
		}
	}

	assert.True(t, foundEC2, "Should find EC2")
	assert.True(t, foundAzure, "Should find Azure VM")
	assert.True(t, foundRDS, "Should find RDS")
}
