package engine

import (
	"context"
	"testing"
	"time"

	"github.com/rshade/pulumicost-core/internal/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngine_GetProjectedCost(t *testing.T) {
	tests := []struct {
		name        string
		resources   []engine.ResourceDescriptor
		expectError bool
	}{
		{
			name: "simple aws ec2 instance",
			resources: []engine.ResourceDescriptor{
				{
					ID:       "i-1234567890abcdef0",
					Type:     "aws_instance",
					Provider: "aws",
					Properties: map[string]interface{}{
						"instance_type": "t3.micro",
						"region":        "us-east-1",
					},
				},
			},
			expectError: false,
		},
		{
			name:        "empty resources",
			resources:   []engine.ResourceDescriptor{},
			expectError: false,
		},
		{
			name: "unknown resource type",
			resources: []engine.ResourceDescriptor{
				{
					ID:         "unknown-123",
					Type:       "unknown_resource",
					Provider:   "unknown",
					Properties: map[string]interface{}{},
				},
			},
			expectError: false, // Should not error, just return "none" adapter result
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create engine with no clients and no loader to test fallback behavior
			eng := engine.New(nil, nil)

			results, err := eng.GetProjectedCost(context.Background(), tt.resources)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, results)

			// Should have same number of results as resources (or at least handle gracefully)
			if len(tt.resources) > 0 {
				assert.Len(t, results, len(tt.resources))

				// Should have "none" adapter when no plugins available
				for _, result := range results {
					assert.Equal(t, "none", result.Adapter)
					assert.Equal(t, "USD", result.Currency)
				}
			}
		})
	}
}

func TestEngine_GetActualCost(t *testing.T) {
	tests := []struct {
		name        string
		resources   []engine.ResourceDescriptor
		expectError bool
	}{
		{
			name: "valid resources",
			resources: []engine.ResourceDescriptor{
				{
					ID:       "i-1234567890abcdef0",
					Type:     "aws_instance",
					Provider: "aws",
				},
			},
			expectError: false,
		},
		{
			name:        "empty resources",
			resources:   []engine.ResourceDescriptor{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eng := engine.New(nil, nil)

			from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			to := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

			results, err := eng.GetActualCost(context.Background(), tt.resources, from, to)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, results)
		})
	}
}

func TestEngine_PluginFallback(t *testing.T) {
	t.Run("falls back when no plugins available", func(t *testing.T) {
		eng := engine.New(nil, nil)

		resources := []engine.ResourceDescriptor{
			{
				ID:       "i-1234567890abcdef0",
				Type:     "aws_instance",
				Provider: "aws",
				Properties: map[string]interface{}{
					"instance_type": "t3.micro",
				},
			},
		}

		results, err := eng.GetProjectedCost(context.Background(), resources)
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.Len(t, results, 1)

		// Should get fallback result with "none" adapter
		result := results[0]
		assert.Equal(t, "none", result.Adapter)
		assert.Equal(t, "USD", result.Currency)
		assert.Contains(t, result.Notes, "No pricing information available")
	})
}

func TestEngine_ErrorHandling(t *testing.T) {
	t.Run("handles unknown resource gracefully", func(t *testing.T) {
		eng := engine.New(nil, nil)

		resources := []engine.ResourceDescriptor{
			{
				ID:       "unknown-resource-123",
				Type:     "unknown_resource_type",
				Provider: "unknown_provider",
				Properties: map[string]interface{}{
					"unknown_field": "unknown_value",
				},
			},
		}

		results, err := eng.GetProjectedCost(context.Background(), resources)
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.Len(t, results, 1)

		// Should handle gracefully with fallback result
		result := results[0]
		assert.Equal(t, "none", result.Adapter)
		assert.Equal(t, "USD", result.Currency)
		assert.Contains(t, result.Notes, "No pricing information available")
	})

	t.Run("handles empty resource list", func(t *testing.T) {
		eng := engine.New(nil, nil)

		results, err := eng.GetProjectedCost(context.Background(), []engine.ResourceDescriptor{})
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.Empty(t, results)
	})
}
