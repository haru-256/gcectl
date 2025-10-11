package log

import (
	"os"

	"github.com/charmbracelet/log"
)

// DefaultLogger is the global logger instance used throughout the application.
// It is initialized automatically when the package is loaded.
var DefaultLogger Logger

func init() {
	DefaultLogger = NewLogger()
}

// Logger defines the interface for logging operations throughout the application.
// This interface abstracts the concrete logger implementation (charmbracelet/log),
// allowing for easier testing and potential logger replacement in the future.
//
// All methods come in two variants:
//   - Non-formatted (e.g., Debug): Takes a simple message string
//   - Formatted (e.g., Debugf): Takes a format string and arguments (printf-style)
type Logger interface {
	Debug(msg string)
	Debugf(format string, args ...any)
	Info(msg string)
	Infof(format string, args ...any)
	Warn(msg string)
	Warnf(format string, args ...any)
	Error(msg string)
	Errorf(format string, args ...any)
	Fatal(msg string)
	Fatalf(format string, args ...any)
}

// charmLogger is a concrete implementation of the Logger interface.
// It wraps the charmbracelet/log logger to provide the interface methods.
type charmLogger struct {
	*log.Logger
}

// Debug outputs a debug-level message.
// Debug messages are typically used for detailed diagnostic information.
func (l *charmLogger) Debug(msg string) {
	l.Logger.Debug(msg)
}

// Info outputs an informational message.
// Info messages are used for general informational messages about application flow.
func (l *charmLogger) Info(msg string) {
	l.Logger.Info(msg)
}

// Warn outputs a warning message.
// Warning messages indicate potentially harmful situations that don't prevent execution.
func (l *charmLogger) Warn(msg string) {
	l.Logger.Warn(msg)
}

// Error outputs an error message.
// Error messages indicate error conditions that occurred but allow continued execution.
func (l *charmLogger) Error(msg string) {
	l.Logger.Error(msg)
}

// Fatal outputs a fatal error message and exits the program.
// Fatal should be used only for unrecoverable errors that require immediate termination.
func (l *charmLogger) Fatal(msg string) {
	l.Logger.Fatal(msg)
}

// NewLogger creates and returns a new Logger instance.
//
// The logger is configured with:
//   - Output to stderr
//   - Log level from GCE_COMMANDS_LOG_LEVEL environment variable (default: INFO)
//   - Caller reporting enabled (shows source file and line number)
//   - Timestamp reporting enabled
//
// This function is typically called once during application initialization.
//
// Environment variables:
//   - GCE_COMMANDS_LOG_LEVEL: Sets the log level (INFO or DEBUG, default: INFO)
//
// Returns:
//   - Logger: A new logger instance ready for use
func NewLogger() Logger {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		Level:           getLevel(),
		ReportCaller:    true,
		ReportTimestamp: true,
	})
	return &charmLogger{Logger: logger}
}

func getLevel() log.Level {
	level := os.Getenv("GCE_COMMANDS_LOG_LEVEL")
	if level == "" {
		level = "INFO"
	}
	switch level {
	case "INFO":
		return log.InfoLevel
	case "DEBUG":
		return log.DebugLevel
	default:
		return log.InfoLevel // 不明な値の場合はINFOをデフォルトとする
	}
}
