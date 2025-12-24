package ingest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rshade/pulumicost-core/internal/ingest"
)

// getLoadPulumiPlanTestData returns test data for LoadPulumiPlan function tests.
func getLoadPulumiPlanTestData() []struct {
	name     string
	content  string
	wantErr  bool
	errMsg   string
	validate func(*testing.T, *ingest.PulumiPlan)
} {
	return []struct {
		name     string
		content  string
		wantErr  bool
		errMsg   string
		validate func(*testing.T, *ingest.PulumiPlan)
	}{
		{
			name: "valid_simple_plan",
			content: `{
				"steps": [
					{
						"op": "create",
						"urn": "urn:pulumi:dev::my-app::aws:ec2/instance:Instance::web-server",
						"type": "aws:ec2/instance:Instance",
						"provider": "urn:pulumi:dev::my-app::pulumi:providers:aws::default_1_0_0::04da6b54-80e4-46f7-96ec-b56ff0331ba9",
						"inputs": {
							"ami": "ami-0c02fb55956c7d316",
							"instanceType": "t3.micro"
						},
						"outputs": {}
					}
				]
			}`,
			wantErr: false,
			validate: func(t *testing.T, plan *ingest.PulumiPlan) {
				if len(plan.Steps) != 1 {
					t.Errorf("expected 1 step, got %d", len(plan.Steps))
				}
				step := plan.Steps[0]
				if step.Op != "create" {
					t.Errorf("expected op 'create', got '%s'", step.Op)
				}
				if step.Type != "aws:ec2/instance:Instance" {
					t.Errorf("expected type 'aws:ec2/instance:Instance', got '%s'", step.Type)
				}
			},
		},
		{
			name: "valid_multi_step_plan",
			content: `{
				"steps": [
					{
						"op": "create",
						"urn": "urn:pulumi:dev::my-app::aws:ec2/instance:Instance::web-server",
						"type": "aws:ec2/instance:Instance",
						"provider": "urn:pulumi:dev::my-app::pulumi:providers:aws::default",
						"inputs": {"instanceType": "t3.micro"},
						"outputs": {}
					},
					{
						"op": "update",
						"urn": "urn:pulumi:dev::my-app::aws:s3/bucket:Bucket::assets",
						"type": "aws:s3/bucket:Bucket",
						"provider": "urn:pulumi:dev::my-app::pulumi:providers:aws::default",
						"inputs": {"bucket": "my-bucket"},
						"outputs": {}
					}
				]
			}`,
			wantErr: false,
			validate: func(t *testing.T, plan *ingest.PulumiPlan) {
				if len(plan.Steps) != 2 {
					t.Errorf("expected 2 steps, got %d", len(plan.Steps))
				}
				if plan.Steps[0].Op != "create" {
					t.Errorf("expected first step op 'create', got '%s'", plan.Steps[0].Op)
				}
				if plan.Steps[1].Op != "update" {
					t.Errorf("expected second step op 'update', got '%s'", plan.Steps[1].Op)
				}
			},
		},
		{
			name: "empty_plan",
			content: `{
				"steps": []
			}`,
			wantErr: false,
			validate: func(t *testing.T, plan *ingest.PulumiPlan) {
				if len(plan.Steps) != 0 {
					t.Errorf("expected 0 steps, got %d", len(plan.Steps))
				}
			},
		},
		{
			name:    "invalid_json",
			content: `{"steps": [`,
			wantErr: true,
			errMsg:  "parsing plan JSON",
		},
		{
			name:    "empty_file",
			content: "",
			wantErr: true,
			errMsg:  "parsing plan JSON",
		},
		{
			name: "missing_steps_field",
			content: `{
				"other": "value"
			}`,
			wantErr: false,
			validate: func(t *testing.T, plan *ingest.PulumiPlan) {
				if len(plan.Steps) != 0 {
					t.Errorf("expected 0 steps when steps field missing, got %d", len(plan.Steps))
				}
			},
		},
	}
}

// TestLoadPulumiPlan tests loading and parsing Pulumi plan JSON files.
func TestLoadPulumiPlan(t *testing.T) {
	tests := getLoadPulumiPlanTestData()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "plan.json")

			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			// Test LoadPulumiPlan
			plan, err := ingest.LoadPulumiPlan(tmpFile)

			if tt.wantErr {
				if err == nil {
					t.Errorf("LoadPulumiPlan() expected error, got nil")
					return
				}
				if tt.errMsg != "" && !containsString(err.Error(), tt.errMsg) {
					t.Errorf(
						"LoadPulumiPlan() error = %v, want error containing %v",
						err,
						tt.errMsg,
					)
				}
				return
			}

			if err != nil {
				t.Errorf("LoadPulumiPlan() unexpected error = %v", err)
				return
			}

			if plan == nil {
				t.Errorf("LoadPulumiPlan() returned nil plan")
				return
			}

			if tt.validate != nil {
				tt.validate(t, plan)
			}
		})
	}
}

func TestLoadPulumiPlan_FileErrors(t *testing.T) {
	t.Run("nonexistent_file", func(t *testing.T) {
		_, err := ingest.LoadPulumiPlan("/nonexistent/path/file.json")
		if err == nil {
			t.Error("LoadPulumiPlan() expected error for nonexistent file, got nil")
		}
		if !containsString(err.Error(), "reading plan file") {
			t.Errorf("LoadPulumiPlan() error = %v, want error containing 'reading plan file'", err)
		}
	})
}

// getPulumiPlanGetResourcesTestData returns test data for PulumiPlan GetResources method tests.
func getPulumiPlanGetResourcesTestData() []struct {
	name      string
	plan      *ingest.PulumiPlan
	wantCount int
	validate  func(*testing.T, []ingest.PulumiResource)
} {
	return []struct {
		name      string
		plan      *ingest.PulumiPlan
		wantCount int
		validate  func(*testing.T, []ingest.PulumiResource)
	}{
		{
			name: "mixed_operations",
			plan: &ingest.PulumiPlan{
				Steps: []ingest.PulumiStep{
					{
						Op:   "create",
						URN:  "urn:pulumi:dev::app::aws:ec2/instance:Instance::web",
						Type: "aws:ec2/instance:Instance",
						Inputs: map[string]interface{}{
							"instanceType": "t3.micro",
						},
					},
					{
						Op:   "delete",
						URN:  "urn:pulumi:dev::app::aws:s3/bucket:Bucket::old",
						Type: "aws:s3/bucket:Bucket",
						Inputs: map[string]interface{}{
							"bucket": "old-bucket",
						},
					},
					{
						Op:   "update",
						URN:  "urn:pulumi:dev::app::aws:rds/instance:Instance::db",
						Type: "aws:rds/instance:Instance",
						Inputs: map[string]interface{}{
							"dbInstanceClass": "db.t3.micro",
						},
					},
					{
						Op:   "same",
						URN:  "urn:pulumi:dev::app::aws:s3/bucket:Bucket::assets",
						Type: "aws:s3/bucket:Bucket",
						Inputs: map[string]interface{}{
							"bucket": "assets-bucket",
						},
					},
				},
			},
			wantCount: 3, // delete operations should be excluded
			validate: func(t *testing.T, resources []ingest.PulumiResource) {
				ops := make(map[string]bool)
				for _, r := range resources {
					// Check that we can extract operation from the steps that created these resources
					for _, step := range []string{"create", "update", "same"} {
						if containsString(r.URN, step) {
							ops[step] = true
						}
					}
				}
				// Should not contain any delete operations
				for _, r := range resources {
					if containsString(r.URN, "old") {
						t.Error("GetResources() should not include deleted resources")
					}
				}
			},
		},
		{
			name: "provider_extraction",
			plan: &ingest.PulumiPlan{
				Steps: []ingest.PulumiStep{
					{
						Op:   "create",
						URN:  "urn:pulumi:dev::app::aws:ec2/instance:Instance::web",
						Type: "aws:ec2/instance:Instance",
						Inputs: map[string]interface{}{
							"instanceType": "t3.micro",
						},
					},
					{
						Op:   "create",
						URN:  "urn:pulumi:dev::app::azure:compute/virtualMachine:VirtualMachine::vm",
						Type: "azure:compute/virtualMachine:VirtualMachine",
						Inputs: map[string]interface{}{
							"vmSize": "Standard_B1s",
						},
					},
				},
			},
			wantCount: 2,
			validate: func(t *testing.T, resources []ingest.PulumiResource) {
				providers := make(map[string]bool)
				for _, r := range resources {
					providers[r.Provider] = true
				}
				if !providers["aws"] {
					t.Error("GetResources() should extract 'aws' provider")
				}
				if !providers["azure"] {
					t.Error("GetResources() should extract 'azure' provider")
				}
			},
		},
		{
			name: "complex_inputs",
			plan: &ingest.PulumiPlan{
				Steps: []ingest.PulumiStep{
					{
						Op:   "create",
						URN:  "urn:pulumi:dev::app::aws:ec2/instance:Instance::web",
						Type: "aws:ec2/instance:Instance",
						Inputs: map[string]interface{}{
							"instanceType": "t3.micro",
							"tags": map[string]interface{}{
								"Name":        "Web Server",
								"Environment": "dev",
							},
							"userData": "#!/bin/bash\necho 'Hello World'",
							"count":    float64(1),
							"enabled":  true,
						},
					},
				},
			},
			wantCount: 1,
			validate: func(t *testing.T, resources []ingest.PulumiResource) {
				r := resources[0]

				// Check string value
				if r.Inputs["instanceType"] != "t3.micro" {
					t.Errorf("expected instanceType 't3.micro', got %v", r.Inputs["instanceType"])
				}

				// Check nested map
				tags, ok := r.Inputs["tags"].(map[string]interface{})
				if !ok {
					t.Error("expected tags to be map[string]interface{}")
				} else if tags["Name"] != "Web Server" {
					t.Errorf("expected Name tag 'Web Server', got %v", tags["Name"])
				}

				// Check boolean
				if r.Inputs["enabled"] != true {
					t.Errorf("expected enabled true, got %v", r.Inputs["enabled"])
				}

				// Check number
				if r.Inputs["count"] != float64(1) {
					t.Errorf("expected count 1, got %v", r.Inputs["count"])
				}
			},
		},
		{
			name: "resource_ordering_preservation",
			plan: &ingest.PulumiPlan{
				Steps: []ingest.PulumiStep{
					{
						Op:   "create",
						URN:  "urn:pulumi:dev::app::aws:s3/bucket:Bucket::bucket",
						Type: "aws:s3/bucket:Bucket",
						Inputs: map[string]interface{}{
							"bucket": "my-bucket",
						},
					},
					{
						Op:   "create",
						URN:  "urn:pulumi:dev::app::aws:s3/bucketPolicy:BucketPolicy::policy",
						Type: "aws:s3/bucketPolicy:BucketPolicy",
						Inputs: map[string]interface{}{
							"bucket": "${aws:s3/bucket:Bucket.bucket.id}",
							"policy": "{}",
						},
					},
					{
						Op:   "create",
						URN:  "urn:pulumi:dev::app::aws:ec2/instance:Instance::web",
						Type: "aws:ec2/instance:Instance",
						Inputs: map[string]interface{}{
							"instanceType": "t3.micro",
							"userData":     "aws s3 cp s3://${aws:s3/bucket:Bucket.bucket.id}/script.sh",
						},
					},
				},
			},
			wantCount: 3,
			validate: func(t *testing.T, resources []ingest.PulumiResource) {
				// Verify that resources are returned in the same order as defined in steps
				expectedOrder := []string{
					"urn:pulumi:dev::app::aws:s3/bucket:Bucket::bucket",
					"urn:pulumi:dev::app::aws:s3/bucketPolicy:BucketPolicy::policy",
					"urn:pulumi:dev::app::aws:ec2/instance:Instance::web",
				}

				for i, expected := range expectedOrder {
					if i < len(resources) && resources[i].URN != expected {
						t.Errorf(
							"resource ordering not preserved: expected %s at position %d, got %s",
							expected,
							i,
							resources[i].URN,
						)
					}
				}

				// Verify that dependency references in properties are preserved
				policyResource := resources[1]
				bucketRef, ok := policyResource.Inputs["bucket"].(string)
				if !ok || !containsString(bucketRef, "bucket.id") {
					t.Error("dependency reference in bucket policy not preserved")
				}

				webResource := resources[2]
				userData, ok := webResource.Inputs["userData"].(string)
				if !ok || !containsString(userData, "bucket.id") {
					t.Error("dependency reference in EC2 user data not preserved")
				}
			},
		},
		{
			name:      "empty_plan",
			plan:      &ingest.PulumiPlan{Steps: []ingest.PulumiStep{}},
			wantCount: 0,
		},
	}
}

// TestPulumiPlan_GetResources tests the GetResources method of PulumiPlan.
func TestPulumiPlan_GetResources(t *testing.T) {
	tests := getPulumiPlanGetResourcesTestData()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resources := tt.plan.GetResources()

			if len(resources) != tt.wantCount {
				t.Errorf(
					"GetResources() returned %d resources, want %d",
					len(resources),
					tt.wantCount,
				)
			}

			if tt.validate != nil {
				tt.validate(t, resources)
			}
		})
	}
}

// Helper function to check if a string contains a substring.
func containsString(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
