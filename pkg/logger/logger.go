// Package logger provides a structured logging solution using zerolog.
//
// The package configures a global logger instance with support for configurable
// log levels via the LOG_LEVEL environment variable. It produces JSON-formatted
// logs suitable for production environments and log aggregation systems.
//
// Supported log levels: debug, info, warn, error (default: info)
//
// Example usage:
//
//	logger.Init()
//	logger.Info().Msg("Application started")
//	logger.Debug().Str("key", "value").Msg("Debug information")
package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// log holds the global logger instance used throughout the application.
var log zerolog.Logger

// Init initializes the global logger with the configuration specified
// by the LOG_LEVEL environment variable. If LOG_LEVEL is not set or
// contains an invalid value, it defaults to "info" level.
//
// The logger outputs structured JSON to stdout with timestamps in RFC3339 format.
func Init() {
	// Configure zerolog to use RFC3339 timestamps for consistency
	zerolog.TimeFieldFormat = time.RFC3339

	// Parse log level from environment variable
	level := parseLogLevel(os.Getenv("LOG_LEVEL"))
	zerolog.SetGlobalLevel(level)

	// Create logger with timestamp and caller information
	log = zerolog.New(os.Stdout).
		With().
		Timestamp().
		Caller().
		Logger()
}

// parseLogLevel converts a string log level to a zerolog.Level.
// Supported values: "debug", "info", "warn", "error".
// Defaults to InfoLevel if the value is unrecognized or empty.
func parseLogLevel(levelStr string) zerolog.Level {
	switch strings.ToLower(strings.TrimSpace(levelStr)) {
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

// Debug returns a zerolog.Event for logging at debug level.
// Debug logs are intended for detailed troubleshooting information.
func Debug() *zerolog.Event {
	return log.Debug()
}

// Info returns a zerolog.Event for logging at info level.
// Info logs are intended for general operational information.
func Info() *zerolog.Event {
	return log.Info()
}

// Warn returns a zerolog.Event for logging at warn level.
// Warn logs indicate potentially harmful situations.
func Warn() *zerolog.Event {
	return log.Warn()
}

// Error returns a zerolog.Event for logging at error level.
// Error logs indicate error conditions that should be addressed.
func Error() *zerolog.Event {
	return log.Error()
}

// Fatal returns a zerolog.Event for logging at fatal level.
// Fatal logs indicate severe errors; the application will exit after logging.
func Fatal() *zerolog.Event {
	return log.Fatal()
}

// With creates a child logger with additional context fields.
// This is useful for adding request-specific or operation-specific context.
func With() zerolog.Context {
	return log.With()
}

// Logger returns the underlying zerolog.Logger instance for advanced usage.
func Logger() zerolog.Logger {
	return log
}
