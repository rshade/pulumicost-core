package main

import (
	"testing"
)

func TestExtractRunID(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid run id",
			body:     "Workflow run https://github.com/rshade/pulumicost-core/actions/runs/1234567890 failed",
			expected: "1234567890",
			wantErr:  false,
		},
		{
			name:     "no run id",
			body:     "Some other issue description",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractRunID(tt.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractRunID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("extractRunID() = %v, want %v", got, tt.expected)
			}
		})
	}
}
