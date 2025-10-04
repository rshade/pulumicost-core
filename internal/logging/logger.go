// Package logging provides structured logging infrastructure for PulumiCost.
//
// This package implements a comprehensive logging system with:
// - Structured logging using Go's slog package
// - Multiple output destinations (console, file, syslog)
// - Configurable log levels (debug, info, warn, error)
// - Context-based request tracing with request IDs
// - JSON and text output formats
//
// Usage:
//
//	logger := logging.NewFromConfig(config.Logging)
//	logger.Info("operation completed", "resource", "aws:ec2:Instance", "cost", 10.50)
//	logger.WithRequestID(ctx).Debug("processing resource", "type", resourceType)
package logging

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/rshade/pulumicost-core/internal/config"
)

// contextKey is a private type for context keys to avoid collisions.
type contextKey string

const (
	requestIDKey      contextKey  = "requestID"
	logFilePermission os.FileMode = 0600
)

// Logger wraps slog.Logger with additional functionality.
type Logger struct {
	*slog.Logger

	closers []io.Closer
}

// NewFromConfig creates a new logger from configuration.
func NewFromConfig(cfg config.LoggingConfig) (*Logger, error) {
	// If outputs are configured, use them
	if len(cfg.Outputs) > 0 {
		return newMultiOutputLogger(cfg)
	}

	// Fallback to legacy single file configuration
	return newLegacyLogger(cfg)
}

// newMultiOutputLogger creates a logger with multiple outputs.
func newMultiOutputLogger(cfg config.LoggingConfig) (*Logger, error) {
	var writers []io.Writer
	var closers []io.Closer

	for _, output := range cfg.Outputs {
		writer, closer, err := createOutput(output, cfg.Format)
		if err != nil {
			// Close any already opened writers
			for _, c := range closers {
				_ = c.Close()
			}
			return nil, fmt.Errorf("creating output %s: %w", output.Type, err)
		}

		writers = append(writers, writer)
		if closer != nil {
			closers = append(closers, closer)
		}
	}

	// Use first output's level or global level
	level := parseLevel(cfg.Level)
	if len(cfg.Outputs) > 0 && cfg.Outputs[0].Level != "" {
		level = parseLevel(cfg.Outputs[0].Level)
	}

	// Use first output's format or global format
	format := cfg.Format
	if len(cfg.Outputs) > 0 && cfg.Outputs[0].Format != "" {
		format = cfg.Outputs[0].Format
	}

	// Combine all writers
	multiWriter := io.MultiWriter(writers...)

	var handler slog.Handler
	if format == "json" {
		handler = slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
			Level: level,
		})
	} else {
		handler = slog.NewTextHandler(multiWriter, &slog.HandlerOptions{
			Level: level,
		})
	}

	return &Logger{
		Logger:  slog.New(handler),
		closers: closers,
	}, nil
}

// newLegacyLogger creates a logger using legacy single-file configuration.
func newLegacyLogger(cfg config.LoggingConfig) (*Logger, error) {
	var writer io.Writer = os.Stdout
	var closers []io.Closer

	// If file path specified, open file for writing
	if cfg.File != "" {
		file, err := openSecureLogFile(cfg.File)
		if err != nil {
			return nil, err
		}
		writer = file
		closers = append(closers, file)
	}

	level := parseLevel(cfg.Level)

	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{
			Level: level,
		})
	} else {
		handler = slog.NewTextHandler(writer, &slog.HandlerOptions{
			Level: level,
		})
	}

	return &Logger{
		Logger:  slog.New(handler),
		closers: closers,
	}, nil
}

// openSecureLogFile opens a log file with security checks and proper permissions.
func openSecureLogFile(path string) (*os.File, error) {
	// Ensure log directory exists
	logDir := filepath.Dir(path)
	if err := os.MkdirAll(logDir, 0700); err != nil {
		return nil, fmt.Errorf("creating log directory: %w", err)
	}

	// Check for symlink before opening
	if info, err := os.Lstat(path); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, errors.New("log file cannot be a symlink")
		}
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, logFilePermission)
	if err != nil {
		return nil, fmt.Errorf("opening log file: %w", err)
	}

	// Ensure permissions even for existing files
	if chmodErr := file.Chmod(logFilePermission); chmodErr != nil {
		_ = file.Close()
		return nil, fmt.Errorf("setting log file permissions: %w", chmodErr)
	}

	return file, nil
}

// createOutput creates an output writer based on configuration.
func createOutput(output config.LogOutput, _ string) (io.Writer, io.Closer, error) {
	switch output.Type {
	case "console":
		return os.Stdout, nil, nil

	case "file":
		if output.Path == "" {
			return nil, nil, errors.New("file output requires path")
		}

		file, err := openSecureLogFile(output.Path)
		if err != nil {
			return nil, nil, err
		}

		return file, file, nil

	case "syslog":
		// For now, syslog is not implemented - would require platform-specific code
		return nil, nil, errors.New("syslog output not yet implemented")

	default:
		return nil, nil, fmt.Errorf("unknown output type: %s", output.Type)
	}
}

// parseLevel converts string log level to slog.Level.
func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithRequestID returns a logger with request ID from context.
func (l *Logger) WithRequestID(ctx context.Context) *Logger {
	if reqID, ok := ctx.Value(requestIDKey).(string); ok {
		return &Logger{
			Logger:  l.Logger.With("request_id", reqID),
			closers: l.closers,
		}
	}
	return l
}

// WithContext returns a logger with contextual fields from context.
func (l *Logger) WithContext(ctx context.Context) *Logger {
	return l.WithRequestID(ctx)
}

// With returns a logger with additional key-value pairs.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		Logger:  l.Logger.With(args...),
		closers: l.closers,
	}
}

// WithComponent returns a logger with component name.
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		Logger:  l.Logger.With("component", component),
		closers: l.closers,
	}
}

// ContextWithRequestID adds request ID to context.
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestIDFromContext extracts request ID from context.
func RequestIDFromContext(ctx context.Context) (string, bool) {
	reqID, ok := ctx.Value(requestIDKey).(string)
	return reqID, ok
}

// Default returns a default logger for cases where config is not available.
func Default() *Logger {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return &Logger{Logger: slog.New(handler)}
}

// Close closes all underlying writers that need cleanup.
// It should be called when the logger is no longer needed to prevent resource leaks.
func (l *Logger) Close() error {
	var errs []error
	for _, closer := range l.closers {
		if err := closer.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
