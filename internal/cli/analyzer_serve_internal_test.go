package cli

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestGetAnalyzerLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     zerolog.Level
	}{
		{
			name:     "default level when env not set",
			envValue: "",
			want:     zerolog.InfoLevel,
		},
		{
			name:     "debug level from env",
			envValue: "debug",
			want:     zerolog.DebugLevel,
		},
		{
			name:     "warn level from env",
			envValue: "warn",
			want:     zerolog.WarnLevel,
		},
		{
			name:     "error level from env",
			envValue: "error",
			want:     zerolog.ErrorLevel,
		},
		{
			name:     "trace level from env",
			envValue: "trace",
			want:     zerolog.TraceLevel,
		},
		{
			name:     "invalid level falls back to info",
			envValue: "invalid-level",
			want:     zerolog.InfoLevel,
		},
		{
			name:     "uppercase level works",
			envValue: "DEBUG",
			want:     zerolog.DebugLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				t.Setenv("FINFOCUS_LOG_LEVEL", tt.envValue)
			}

			got := getAnalyzerLogLevel()
			assert.Equal(t, tt.want, got)
		})
	}
}
