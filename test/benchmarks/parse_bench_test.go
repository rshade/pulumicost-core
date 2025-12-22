package benchmarks_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// BenchmarkParse_PulumiPlan benchmarks parsing of a typical Pulumi plan JSON.
func BenchmarkParse_PulumiPlan(b *testing.B) {
	// Generate a simple plan string
	resources := make([]string, 100)
	for i := 0; i < 100; i++ {
		resources[i] = fmt.Sprintf(
			`{"type": "aws:ec2/instance:Instance", "id": "i-%d", "inputs": {"instanceType": "t3.micro"}}`,
			i,
		)
	}
	jsonStr := fmt.Sprintf(`{"resourceChanges": [%s]}`, strings.Join(resources, ","))
	data := []byte(jsonStr)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParse_LargePlan benchmarks parsing of a large Pulumi plan JSON (10k resources).
func BenchmarkParse_LargePlan(b *testing.B) {
	// Generate a large plan string
	count := 10000
	resources := make([]string, count)
	for i := 0; i < count; i++ {
		resources[i] = fmt.Sprintf(
			`{"type": "aws:ec2/instance:Instance", "id": "i-%d", "inputs": {"instanceType": "t3.micro"}}`,
			i,
		)
	}
	jsonStr := fmt.Sprintf(`{"resourceChanges": [%s]}`, strings.Join(resources, ","))
	data := []byte(jsonStr)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			b.Fatal(err)
		}
	}
}
