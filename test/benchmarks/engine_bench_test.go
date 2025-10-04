package benchmarks

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rshade/pulumicost-core/internal/engine"
)

// Benchmark projected cost calculations
func BenchmarkEngine_GetProjectedCost_Single(b *testing.B) {
	eng := engine.New(nil, nil)

	resources := []engine.ResourceDescriptor{
		{
			ID:       "i-1234567890abcdef0",
			Type:     "aws_instance",
			Provider: "aws",
			Properties: map[string]interface{}{
				"instance_type": "t3.micro",
				"region":        "us-east-1",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := eng.GetProjectedCost(context.Background(), resources)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEngine_GetProjectedCost_Multiple(b *testing.B) {
	eng := engine.New(nil, nil)

	// Create 10 resources for batch testing
	resources := make([]engine.ResourceDescriptor, 10)
	for i := 0; i < 10; i++ {
		resources[i] = engine.ResourceDescriptor{
			ID:       fmt.Sprintf("resource-%d", i),
			Type:     "aws_instance",
			Provider: "aws",
			Properties: map[string]interface{}{
				"instance_type": "t3.micro",
				"region":        "us-east-1",
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := eng.GetProjectedCost(context.Background(), resources)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEngine_GetProjectedCost_Large(b *testing.B) {
	eng := engine.New(nil, nil)

	// Create 100 resources for large batch testing
	resources := make([]engine.ResourceDescriptor, 100)
	for i := 0; i < 100; i++ {
		resources[i] = engine.ResourceDescriptor{
			ID:       fmt.Sprintf("resource-%d", i),
			Type:     "aws_instance",
			Provider: "aws",
			Properties: map[string]interface{}{
				"instance_type": "t3.micro",
				"region":        "us-east-1",
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := eng.GetProjectedCost(context.Background(), resources)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark actual cost calculations
func BenchmarkEngine_GetActualCost_Single(b *testing.B) {
	eng := engine.New(nil, nil)

	resources := []engine.ResourceDescriptor{
		{
			ID:       "i-1234567890abcdef0",
			Type:     "aws_instance",
			Provider: "aws",
		},
	}

	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := eng.GetActualCost(context.Background(), resources, from, to)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEngine_GetActualCost_Multiple(b *testing.B) {
	eng := engine.New(nil, nil)

	// Create 10 resources for batch testing
	resources := make([]engine.ResourceDescriptor, 10)
	for i := 0; i < 10; i++ {
		resources[i] = engine.ResourceDescriptor{
			ID:       fmt.Sprintf("resource-%d", i),
			Type:     "aws_instance",
			Provider: "aws",
		}
	}

	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := eng.GetActualCost(context.Background(), resources, from, to)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Memory allocation benchmarks
func BenchmarkEngine_ResourceDescriptor_Allocation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resources := make([]engine.ResourceDescriptor, 100)
		for j := 0; j < 100; j++ {
			resources[j] = engine.ResourceDescriptor{
				ID:       fmt.Sprintf("resource-%d", j),
				Type:     "aws_instance",
				Provider: "aws",
				Properties: map[string]interface{}{
					"instance_type": "t3.micro",
					"region":        "us-east-1",
					"tags": map[string]interface{}{
						"Environment": "production",
						"Name":        fmt.Sprintf("instance-%d", j),
					},
				},
			}
		}
		// Prevent compiler optimization
		_ = resources
	}
}

// Concurrent access benchmarks
func BenchmarkEngine_GetProjectedCost_Concurrent(b *testing.B) {
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

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := eng.GetProjectedCost(context.Background(), resources)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// Context cancellation benchmark
func BenchmarkEngine_GetProjectedCost_WithTimeout(b *testing.B) {
	eng := engine.New(nil, nil)

	resources := []engine.ResourceDescriptor{
		{
			ID:       "i-1234567890abcdef0",
			Type:     "aws_instance",
			Provider: "aws",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		_, err := eng.GetProjectedCost(ctx, resources)
		cancel()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Property serialization benchmark
func BenchmarkEngine_Properties_Conversion(b *testing.B) {
	properties := map[string]interface{}{
		"instance_type":     "t3.micro",
		"region":            "us-east-1",
		"availability_zone": "us-east-1a",
		"security_groups":   []string{"sg-1", "sg-2", "sg-3"},
		"tags": map[string]interface{}{
			"Environment": "production",
			"Name":        "test-instance",
			"Project":     "myapp",
		},
		"user_data":   "#!/bin/bash\necho 'Hello World'",
		"instance_id": "i-1234567890abcdef0",
		"private_ip":  "10.0.1.100",
		"public_ip":   "203.0.113.1",
		"subnet_id":   "subnet-12345",
		"vpc_id":      "vpc-67890",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate property conversion (from internal/engine/engine.go:187)
		result := make(map[string]string)
		for k, v := range properties {
			result[k] = fmt.Sprintf("%v", v)
		}
		// Prevent compiler optimization
		_ = result
	}
}

// Error handling performance
func BenchmarkEngine_GetProjectedCost_NoClients(b *testing.B) {
	// Test fallback performance when no plugins are available
	eng := engine.New(nil, nil)

	resources := []engine.ResourceDescriptor{
		{
			ID:       "i-1234567890abcdef0",
			Type:     "aws_instance",
			Provider: "aws",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results, err := eng.GetProjectedCost(context.Background(), resources)
		if err != nil {
			b.Fatal(err)
		}
		if len(results) != 1 {
			b.Fatalf("Expected 1 result, got %d", len(results))
		}
		if results[0].Adapter != "none" {
			b.Fatalf("Expected 'none' adapter, got %s", results[0].Adapter)
		}
	}
}
