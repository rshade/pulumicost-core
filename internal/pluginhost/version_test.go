package pluginhost

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareSpecVersions(t *testing.T) {
	tests := []struct {
		name          string
		coreVersion   string
		pluginVersion string
		wantResult    CompatibilityResult
		wantErr       bool
	}{
		{
			name:          "Same version",
			coreVersion:   "0.4.14",
			pluginVersion: "0.4.14",
			wantResult:    Compatible,
			wantErr:       false,
		},
		{
			name:          "Compatible minor update",
			coreVersion:   "0.4.14",
			pluginVersion: "0.4.15",
			wantResult:    Compatible,
			wantErr:       false,
		},
		{
			name:          "Core newer minor (backward compatible)",
			coreVersion:   "0.4.14",
			pluginVersion: "0.4.10",
			wantResult:    Compatible,
			wantErr:       false,
		},
		{
			name:          "Major mismatch (plugin newer)",
			coreVersion:   "0.4.14",
			pluginVersion: "1.0.0",
			wantResult:    MajorMismatch,
			wantErr:       false,
		},
		{
			name:          "Major mismatch (core newer)",
			coreVersion:   "1.0.0",
			pluginVersion: "0.4.14",
			wantResult:    MajorMismatch,
			wantErr:       false,
		},
		{
			name:          "Invalid core version",
			coreVersion:   "invalid",
			pluginVersion: "0.4.14",
			wantResult:    Invalid,
			wantErr:       true,
		},
		{
			name:          "Invalid plugin version",
			coreVersion:   "0.4.14",
			pluginVersion: "invalid",
			wantResult:    Invalid,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompareSpecVersions(tt.coreVersion, tt.pluginVersion)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantResult, got)
		})
	}
}
