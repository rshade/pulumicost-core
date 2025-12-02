package logging

import (
	"context"
	"crypto/rand"
	"io"
	"os"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"
)

// traceIDKey is a private type for context keys to avoid collisions.
type traceIDKey struct{}

// Config holds logging configuration settings.
type Config struct {
	Level      string // Log level: trace, debug, info, warn, error
	Format     string // Output format: json, console, text
	Output     string // Output destination: stderr, stdout, file
	File       string // File path when Output is "file"
	Caller     bool   // Include file:line in output
	StackTrace bool   // Include stack trace on errors
}

// TracingHook implements zerolog.Hook to automatically inject trace_id from context.
type TracingHook struct{}

// Run implements zerolog.Hook interface.
// It extracts trace_id from the event's context and adds it to the log entry.
func (h TracingHook) Run(e *zerolog.Event, _ zerolog.Level, _ string) {
	ctx := e.GetCtx()
	if ctx == nil {
		return
	}
	if traceID, ok := ctx.Value(traceIDKey{}).(string); ok && traceID != "" {
		e.Str("trace_id", traceID)
	}
}

// LoggingConfig is an alias for Config for backward compatibility.
//
//nolint:revive // Backward compatibility alias intentionally uses stuttering name
type LoggingConfig = Config

// NewLogger creates a new zerolog logger with the provided configuration.
// The logger writes to stderr by default and includes a TracingHook for automatic
// trace ID injection.
func NewLogger(cfg Config) zerolog.Logger {
	writer := createWriter(cfg)
	return newLoggerWithWriter(cfg, writer)
}

// NewLoggerWithWriter creates a logger writing to the specified writer.
// This is primarily used for testing to capture log output.
func NewLoggerWithWriter(cfg Config, writer io.Writer) zerolog.Logger {
	return newLoggerWithWriter(cfg, writer)
}

// newLoggerWithWriter is the internal implementation for creating loggers.
func newLoggerWithWriter(cfg Config, writer io.Writer) zerolog.Logger {
	// Apply format transformation
	output := writer
	if cfg.Format == "console" || cfg.Format == "text" {
		output = zerolog.ConsoleWriter{
			Out:        writer,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		}
	}

	// Create base logger with timestamp
	logger := zerolog.New(output).
		With().
		Timestamp().
		Logger().
		Hook(TracingHook{}).
		Level(parseLevel(cfg.Level))

	// Add caller info if configured
	if cfg.Caller {
		logger = logger.With().Caller().Logger()
	}

	return logger
}

// createWriter creates the appropriate io.Writer based on configuration.
func createWriter(cfg LoggingConfig) io.Writer {
	switch cfg.Output {
	case "stdout":
		return os.Stdout
	case "file":
		if cfg.File != "" {
			file, err := os.OpenFile(cfg.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				// Fall back to stderr and log warning
				warnLogger := zerolog.New(os.Stderr)
				warnLogger.Warn().
					Err(err).
					Str("file", cfg.File).
					Msg("cannot open log file, falling back to stderr")
				return os.Stderr
			}
			return file
		}
		return os.Stderr
	default:
		// Default to stderr (keeps stdout clean for command output)
		return os.Stderr
	}
}

// parseLevel converts a string log level to zerolog.Level.
// Invalid levels default to InfoLevel.
func parseLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

// GenerateTraceID creates a new ULID-format trace identifier.
// ULIDs are lexicographically sortable and monotonic within the same millisecond.
func GenerateTraceID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String()
}

// GetOrGenerateTraceID returns a trace ID from environment, context, or generates a new one.
// Priority: PULUMICOST_TRACE_ID env var > context > generate new.
func GetOrGenerateTraceID(ctx context.Context) string {
	// 1. Check environment variable (external injection)
	if envTraceID := os.Getenv("PULUMICOST_TRACE_ID"); envTraceID != "" {
		return envTraceID
	}

	// 2. Check context
	if traceID := TraceIDFromContext(ctx); traceID != "" {
		return traceID
	}

	// 3. Generate new
	return GenerateTraceID()
}

// ContextWithTraceID stores a trace ID in the context.
func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey{}, traceID)
}

// TraceIDFromContext extracts the trace ID from context.
// Returns empty string if no trace ID is stored.
func TraceIDFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey{}).(string); ok {
		return traceID
	}
	return ""
}

// FromContext returns a logger from context, creating a default if none exists.
// The returned logger will have the TracingHook applied for automatic trace ID injection.
func FromContext(ctx context.Context) *zerolog.Logger {
	logger := zerolog.Ctx(ctx)
	if logger.GetLevel() == zerolog.Disabled {
		// No logger in context, create a default one
		defaultLogger := zerolog.New(os.Stderr).With().Timestamp().Logger().Hook(TracingHook{})
		return &defaultLogger
	}
	return logger
}

// sensitivePatterns defines keys that should have their values redacted.
//
//nolint:gochecknoglobals // Package-level constant patterns for sensitive key detection
var sensitivePatterns = []string{
	"api_key", "apikey", "api-key",
	"password", "passwd", "pwd",
	"secret", "token",
	"credential", "cred",
	"private_key", "privatekey",
	"auth", "authorization", "bearer",
}

// isSensitiveKey checks if a key name contains sensitive patterns.
func isSensitiveKey(key string) bool {
	lower := strings.ToLower(key)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

// SafeStr adds a string field to the event, redacting sensitive values.
// Use this when logging potentially sensitive key-value pairs.
func SafeStr(e *zerolog.Event, key, value string) *zerolog.Event {
	if isSensitiveKey(key) {
		return e.Str(key, "[REDACTED]")
	}
	return e.Str(key, value)
}

// ComponentLogger creates a sub-logger with the component field set.
func ComponentLogger(logger zerolog.Logger, component string) zerolog.Logger {
	return logger.With().Str("component", component).Logger()
}
