package recorder

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	pbc "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func testLogger() zerolog.Logger {
	return zerolog.New(os.Stderr).With().Timestamp().Logger().Level(zerolog.Disabled)
}

func TestNewRecorderPlugin(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{
		OutputDir:    tmpDir,
		MockResponse: false,
	}

	plugin := NewRecorderPlugin(cfg, testLogger())

	require.NotNil(t, plugin)
	assert.Equal(t, "recorder", plugin.Name())
	assert.NotNil(t, plugin.recorder)
	assert.Nil(t, plugin.mocker) // Mock mode disabled
}

func TestNewRecorderPlugin_WithMockMode(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{
		OutputDir:    tmpDir,
		MockResponse: true,
	}

	plugin := NewRecorderPlugin(cfg, testLogger())

	require.NotNil(t, plugin)
	assert.NotNil(t, plugin.mocker) // Mock mode enabled
}

func TestRecorderPlugin_Name(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{OutputDir: tmpDir}
	plugin := NewRecorderPlugin(cfg, testLogger())

	assert.Equal(t, "recorder", plugin.Name())
}

func TestRecorderPlugin_GetProjectedCost_MockDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{
		OutputDir:    tmpDir,
		MockResponse: false,
	}
	plugin := NewRecorderPlugin(cfg, testLogger())

	req := &pbc.GetProjectedCostRequest{
		Resource: &pbc.ResourceDescriptor{
			ResourceType: "aws:ec2:Instance",
			Provider:     "aws",
			Sku:          "t3.medium",
			Region:       "us-east-1",
		},
	}

	resp, err := plugin.GetProjectedCost(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, float64(0), resp.GetCostPerMonth())
	assert.Equal(t, "USD", resp.GetCurrency())
	assert.Contains(t, resp.GetBillingDetail(), "mock responses disabled")
}

func TestRecorderPlugin_GetProjectedCost_MockEnabled(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{
		OutputDir:    tmpDir,
		MockResponse: true,
	}
	plugin := NewRecorderPlugin(cfg, testLogger())

	req := &pbc.GetProjectedCostRequest{
		Resource: &pbc.ResourceDescriptor{
			ResourceType: "aws:ec2:Instance",
			Provider:     "aws",
			Sku:          "t3.medium",
			Region:       "us-east-1",
		},
	}

	resp, err := plugin.GetProjectedCost(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Greater(t, resp.GetCostPerMonth(), float64(0))
	assert.Equal(t, "USD", resp.GetCurrency())
	assert.Contains(t, resp.GetBillingDetail(), "Mock cost")
}

func TestRecorderPlugin_GetActualCost_MockDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{
		OutputDir:    tmpDir,
		MockResponse: false,
	}
	plugin := NewRecorderPlugin(cfg, testLogger())

	now := time.Now()
	req := &pbc.GetActualCostRequest{
		ResourceId: "arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0",
		Start:      timestamppb.New(now.Add(-24 * time.Hour)),
		End:        timestamppb.New(now),
	}

	resp, err := plugin.GetActualCost(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.GetResults())
}

func TestRecorderPlugin_GetActualCost_MockEnabled(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{
		OutputDir:    tmpDir,
		MockResponse: true,
	}
	plugin := NewRecorderPlugin(cfg, testLogger())

	now := time.Now()
	req := &pbc.GetActualCostRequest{
		ResourceId: "arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0",
		Start:      timestamppb.New(now.Add(-24 * time.Hour)),
		End:        timestamppb.New(now),
	}

	resp, err := plugin.GetActualCost(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.GetResults(), 1)
	assert.Equal(t, "recorder-mock", resp.GetResults()[0].GetSource())
	assert.Greater(t, resp.GetResults()[0].GetCost(), float64(0))
}

func TestRecorderPlugin_GetPricingSpec(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{OutputDir: tmpDir}
	plugin := NewRecorderPlugin(cfg, testLogger())

	req := &pbc.GetPricingSpecRequest{
		Resource: &pbc.ResourceDescriptor{
			ResourceType: "aws:ec2:Instance",
			Provider:     "aws",
		},
	}

	resp, err := plugin.GetPricingSpec(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestRecorderPlugin_EstimateCost_MockDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{
		OutputDir:    tmpDir,
		MockResponse: false,
	}
	plugin := NewRecorderPlugin(cfg, testLogger())

	req := &pbc.EstimateCostRequest{
		ResourceType: "aws:ec2:Instance",
	}

	resp, err := plugin.EstimateCost(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, float64(0), resp.GetCostMonthly())
	assert.Equal(t, "USD", resp.GetCurrency())
}

func TestRecorderPlugin_EstimateCost_MockEnabled(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{
		OutputDir:    tmpDir,
		MockResponse: true,
	}
	plugin := NewRecorderPlugin(cfg, testLogger())

	req := &pbc.EstimateCostRequest{
		ResourceType: "aws:ec2:Instance",
	}

	resp, err := plugin.EstimateCost(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Greater(t, resp.GetCostMonthly(), float64(0))
	assert.Equal(t, "USD", resp.GetCurrency())
}

func TestRecorderPlugin_Shutdown(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{OutputDir: tmpDir}
	plugin := NewRecorderPlugin(cfg, testLogger())

	// Should not panic
	plugin.Shutdown()
}

func TestRecorderPlugin_ThreadSafety(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{
		OutputDir:    tmpDir,
		MockResponse: true,
	}
	plugin := NewRecorderPlugin(cfg, testLogger())

	// Run multiple concurrent requests
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := &pbc.GetProjectedCostRequest{
				Resource: &pbc.ResourceDescriptor{
					ResourceType: "aws:ec2:Instance",
					Provider:     "aws",
				},
			}
			_, _ = plugin.GetProjectedCost(context.Background(), req)
		}()
	}

	wg.Wait()

	// No panic = thread-safe
}
