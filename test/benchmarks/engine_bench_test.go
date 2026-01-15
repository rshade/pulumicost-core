// Package benchmarks_test provides performance benchmarks for the finfocus engine.
package benchmarks_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rshade/finfocus/internal/engine"
)

// BenchmarkEngine_GetProjectedCost_Single benchmarks the performance of GetProjectedCost
// with a single resource to establish baseline performance.
func BenchmarkEngine_GetProjectedCost_Single(b *testing.B) {
	b.ReportAllocs()
	eng := engine.New(nil, nil)

	resources := []engine.ResourceDescriptor{
		{
			ID:       "test-resource",
			Type:     "aws_instance",
			Provider: "aws",
			Properties: map[string]interface{}{
				"instance_type": "t3.micro",
				"region":        "us-east-1",
			},
		},
	}

	b.ResetTimer()
	for range b.N {
		_, err := eng.GetProjectedCost(context.Background(), resources)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEngine_GetProjectedCost_Multiple benchmarks the performance of GetProjectedCost
// with multiple resources (batch of 10) to evaluate batching performance.
func BenchmarkEngine_GetProjectedCost_Multiple(b *testing.B) {
	b.ReportAllocs()
	eng := engine.New(nil, nil)

	// Create 10 resources for batch testing
	resources := make([]engine.ResourceDescriptor, 10)
	for i := range 10 {
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
	for range b.N {
		_, err := eng.GetProjectedCost(context.Background(), resources)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEngine_GetProjectedCost_Large benchmarks the performance of GetProjectedCost
// with a large batch of resources (100) to identify performance at scale.
func BenchmarkEngine_GetProjectedCost_Large(b *testing.B) {
	b.ReportAllocs()
	eng := engine.New(nil, nil)

	// Create 100 resources for large batch testing
	resources := make([]engine.ResourceDescriptor, 100)
	for i := range 100 {
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
	for range b.N {
		_, err := eng.GetProjectedCost(context.Background(), resources)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEngine_GetActualCost_Single benchmarks the performance of GetActualCost
// with a single resource to establish baseline performance for actual cost queries.
func BenchmarkEngine_GetActualCost_Single(b *testing.B) {
	b.ReportAllocs()
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
	for range b.N {
		_, err := eng.GetActualCost(context.Background(), resources, from, to)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEngine_GetActualCost_Multiple benchmarks the performance of GetActualCost
// with multiple resources (batch of 10) to evaluate batching for actual cost queries.
func BenchmarkEngine_GetActualCost_Multiple(b *testing.B) {
	b.ReportAllocs()
	eng := engine.New(nil, nil)

	// Create 10 resources for batch testing
	resources := make([]engine.ResourceDescriptor, 10)
	for i := range 10 {
		resources[i] = engine.ResourceDescriptor{
			ID:       fmt.Sprintf("resource-%d", i),
			Type:     "aws_instance",
			Provider: "aws",
		}
	}

	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	b.ResetTimer()
	for range b.N {
		_, err := eng.GetActualCost(context.Background(), resources, from, to)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEngine_ResourceDescriptor_Allocation benchmarks the memory allocation performance
// when creating and initializing batches of ResourceDescriptor structures.
func BenchmarkEngine_ResourceDescriptor_Allocation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		resources := make([]engine.ResourceDescriptor, 100)
		for j := range 100 {
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

// BenchmarkEngine_GetProjectedCost_Concurrent benchmarks the thread-safe concurrent
// performance of GetProjectedCost when accessed by multiple goroutines in parallel.
func BenchmarkEngine_GetProjectedCost_Concurrent(b *testing.B) {
	b.ReportAllocs()
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

// BenchmarkEngine_GetProjectedCost_WithTimeout benchmarks the performance of GetProjectedCost
// when using context timeout to measure cancellation overhead.
func BenchmarkEngine_GetProjectedCost_WithTimeout(b *testing.B) {
	b.ReportAllocs()
	eng := engine.New(nil, nil)

	resources := []engine.ResourceDescriptor{
		{
			ID:       "i-1234567890abcdef0",
			Type:     "aws_instance",
			Provider: "aws",
		},
	}

	b.ResetTimer()
	for range b.N {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		_, err := eng.GetProjectedCost(ctx, resources)
		cancel()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEngine_Properties_Conversion benchmarks the performance of converting
// resource properties from interface{} maps to string maps for serialization.
func BenchmarkEngine_Properties_Conversion(b *testing.B) {
	b.ReportAllocs()
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
	for range b.N {
		// Simulate property conversion (from internal/engine/engine.go:187)
		result := make(map[string]string)
		for k, v := range properties {
			result[k] = fmt.Sprintf("%v", v)
		}
		// Prevent compiler optimization
		_ = result
	}
}

// BenchmarkEngine_GetProjectedCost_NoClients benchmarks the performance of GetProjectedCost
// fallback behavior when no plugins are available, testing error handling performance.
func BenchmarkEngine_GetProjectedCost_NoClients(b *testing.B) {
	b.ReportAllocs()
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
	for range b.N {
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

// ============================================================================
// Enterprise Scale Benchmarks (1K, 10K, 100K resources)
// These benchmarks test performance at enterprise deployment scales.
// ============================================================================

// createResources creates n ResourceDescriptor instances for benchmarking.
func createResources(n int) []engine.ResourceDescriptor {
	resources := make([]engine.ResourceDescriptor, n)
	for i := range n {
		resources[i] = engine.ResourceDescriptor{
			ID:       fmt.Sprintf("i-%08d", i),
			Type:     "aws:ec2:Instance",
			Provider: "aws",
			Properties: map[string]interface{}{
				"instance_type": "t3.micro",
				"region":        "us-east-1",
			},
		}
	}
	return resources
}

// BenchmarkEngine_GetProjectedCost_1K benchmarks projected cost with 1,000 resources.
func BenchmarkEngine_GetProjectedCost_1K(b *testing.B) {
	b.ReportAllocs()
	eng := engine.New(nil, nil)
	resources := createResources(1000)

	b.ResetTimer()
	for range b.N {
		_, err := eng.GetProjectedCost(context.Background(), resources)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEngine_GetProjectedCost_10K benchmarks projected cost with 10,000 resources.
func BenchmarkEngine_GetProjectedCost_10K(b *testing.B) {
	b.ReportAllocs()
	eng := engine.New(nil, nil)
	resources := createResources(10000)

	b.ResetTimer()
	for range b.N {
		_, err := eng.GetProjectedCost(context.Background(), resources)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEngine_GetProjectedCost_100K benchmarks projected cost with 100,000 resources.
func BenchmarkEngine_GetProjectedCost_100K(b *testing.B) {
	b.ReportAllocs()
	eng := engine.New(nil, nil)
	resources := createResources(100000)

	b.ResetTimer()
	for range b.N {
		_, err := eng.GetProjectedCost(context.Background(), resources)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEngine_GetActualCost_1K benchmarks actual cost with 1,000 resources.
func BenchmarkEngine_GetActualCost_1K(b *testing.B) {
	b.ReportAllocs()
	eng := engine.New(nil, nil)
	resources := createResources(1000)
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	b.ResetTimer()
	for range b.N {
		_, err := eng.GetActualCost(context.Background(), resources, from, to)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEngine_GetActualCost_10K benchmarks actual cost with 10,000 resources.
func BenchmarkEngine_GetActualCost_10K(b *testing.B) {
	b.ReportAllocs()
	eng := engine.New(nil, nil)
	resources := createResources(10000)
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	b.ResetTimer()
	for range b.N {
		_, err := eng.GetActualCost(context.Background(), resources, from, to)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEngine_CrossProviderAggregation_1K benchmarks cross-provider aggregation with 1K results.
func BenchmarkEngine_CrossProviderAggregation_1K(b *testing.B) {
	b.ReportAllocs()
	results := make([]engine.CostResult, 1000)
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	providers := []string{"aws", "azure", "gcp"}

	for i := range 1000 {
		results[i] = engine.CostResult{
			ResourceType: fmt.Sprintf("%s:compute:Instance", providers[i%3]),
			ResourceID:   fmt.Sprintf("resource-%d", i),
			Monthly:      float64(100 + i%50),
			Currency:     "USD",
			StartDate:    startDate.Add(time.Duration(i%30) * 24 * time.Hour),
		}
	}

	b.ResetTimer()
	for range b.N {
		_, err := engine.CreateCrossProviderAggregation(results, engine.GroupByDaily)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEngine_CrossProviderAggregation_10K benchmarks cross-provider aggregation with 10K results.
func BenchmarkEngine_CrossProviderAggregation_10K(b *testing.B) {
	b.ReportAllocs()
	results := make([]engine.CostResult, 10000)
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	providers := []string{"aws", "azure", "gcp"}

	for i := range 10000 {
		results[i] = engine.CostResult{
			ResourceType: fmt.Sprintf("%s:compute:Instance", providers[i%3]),
			ResourceID:   fmt.Sprintf("resource-%d", i),
			Monthly:      float64(100 + i%50),
			Currency:     "USD",
			StartDate:    startDate.Add(time.Duration(i%30) * 24 * time.Hour),
		}
	}

	b.ResetTimer()
	for range b.N {
		_, err := engine.CreateCrossProviderAggregation(results, engine.GroupByDaily)
		if err != nil {
			b.Fatal(err)
		}
	}
}
