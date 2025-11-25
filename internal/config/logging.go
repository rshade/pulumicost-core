package config

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Logger is the global zerolog logger instance.
//
//nolint:gochecknoglobals // Logger is intentionally global for application-wide structured logging
var Logger zerolog.Logger

// InitLogger initializes the global zerolog logger with the specified configuration.
// It configures output format, level, and optional file logging.
func InitLogger(level string, logToFile bool) error {
	// Parse log level
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	// Set up output writers
	var writers []io.Writer

	// Console writer with human-readable format
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}
	writers = append(writers, consoleWriter)

	// File writer if enabled
	if logToFile {
		if logDirErr := EnsureLogDir(); logDirErr != nil {
			return logDirErr
		}

		cfg := GetGlobalConfig()
		logPath := cfg.Logging.File
		if logPath == "" {
			logPath = "/tmp/pulumicost.log"
		}

		logFile, fileErr := os.OpenFile(
			logPath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0600,
		)
		if fileErr != nil {
			return fileErr
		}
		writers = append(writers, logFile)
	}

	// Create multi-writer
	multi := zerolog.MultiLevelWriter(writers...)

	// Initialize logger
	Logger = zerolog.New(multi).
		Level(lvl).
		With().
		Timestamp().
		Caller().
		Logger()

	return nil
}

// SetLogLevel sets the global log level.
func SetLogLevel(level string) {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	Logger = Logger.Level(lvl)
}

// GetLogger returns the global logger instance.
func GetLogger() zerolog.Logger {
	return Logger
}

//nolint:gochecknoinits // init is required for default logger initialization
func init() {
	// Default to info level, console only
	_ = InitLogger("info", false)
}
