package logging_test

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rshade/pulumicost-core/internal/config"
	"github.com/rshade/pulumicost-core/internal/logging"
)

func TestNewFromConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.LoggingConfig
		wantErr bool
	}{
		{
			name: "text format with console output",
			cfg: config.LoggingConfig{
				Level:  "info",
				Format: "text",
				Outputs: []config.LogOutput{
					{Type: "console", Level: "info", Format: "text"},
				},
			},
			wantErr: false,
		},
		{
			name: "json format with console output",
			cfg: config.LoggingConfig{
				Level:  "debug",
				Format: "json",
				Outputs: []config.LogOutput{
					{Type: "console", Level: "debug", Format: "json"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := logging.NewFromConfig(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFromConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && logger == nil {
				t.Error("NewFromConfig() returned nil logger")
			}
			if logger != nil {
				defer logger.Close()
			}
		})
	}
}

func TestLogger_WithRequestID(t *testing.T) {
	logger := logging.Default()

	ctx := logging.ContextWithRequestID(context.Background(), "test-req-123")

	loggerWithReq := logger.WithRequestID(ctx)
	if loggerWithReq == nil {
		t.Error("WithRequestID() returned nil")
	}
}

func TestLogger_WithComponent(t *testing.T) {
	logger := logging.Default()

	componentLogger := logger.WithComponent("engine")
	if componentLogger == nil {
		t.Error("WithComponent() returned nil")
	}
}

func TestContextWithRequestID(t *testing.T) {
	ctx := context.Background()
	reqID := "test-request-id"

	ctx = logging.ContextWithRequestID(ctx, reqID)

	retrieved, ok := logging.RequestIDFromContext(ctx)
	if !ok {
		t.Error("RequestIDFromContext() failed to retrieve request ID")
	}
	if retrieved != reqID {
		t.Errorf("RequestIDFromContext() = %v, want %v", retrieved, reqID)
	}
}

func TestDefault(t *testing.T) {
	logger := logging.Default()
	if logger == nil {
		t.Error("Default() returned nil")
	}
}

func TestLogger_LogLevels(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create logger with text handler pointing to buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := &logging.Logger{Logger: slog.New(handler)}

	// Test different log levels
	logger.Debug("debug message", "key", "value")
	logger.Info("info message", "key", "value")
	logger.Warn("warn message", "key", "value")
	logger.Error("error message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "debug message") {
		t.Error("Debug message not logged")
	}
	if !strings.Contains(output, "info message") {
		t.Error("Info message not logged")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Warn message not logged")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error message not logged")
	}
}

func TestNewFromConfig_FileOutput(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "text",
		File:   logFile,
	}

	logger, err := logging.NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("NewFromConfig() error = %v", err)
	}
	defer logger.Close()

	logger.Info("test message")

	// Verify file was created
	if _, statErr := os.Stat(logFile); os.IsNotExist(statErr) {
		t.Error("Log file was not created")
	}
}

func TestLogger_Close(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "close-test.log")

	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "text",
		File:   logFile,
	}

	logger, err := logging.NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("NewFromConfig() error = %v", err)
	}

	logger.Info("test message before close")

	// Close should not error
	closeErr := logger.Close()
	if closeErr != nil {
		t.Errorf("Close() error = %v", closeErr)
	}

	// Verify file exists and contains data
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(content), "test message before close") {
		t.Error("Log file does not contain expected message")
	}
}

func TestLogger_Close_ConsoleOutput(t *testing.T) {
	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "text",
		Outputs: []config.LogOutput{
			{Type: "console", Level: "info", Format: "text"},
		},
	}

	logger, err := logging.NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("NewFromConfig() error = %v", err)
	}

	// Close should not error even with no closers
	closeErr := logger.Close()
	if closeErr != nil {
		t.Errorf("Close() error = %v", closeErr)
	}
}
