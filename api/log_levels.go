package api

import "github.com/samber/lo"

// LogLevel represents a logging level for the profiling agent.
type LogLevel string

const (
	InfoLevel  LogLevel = "info"  // InfoLevel represents informational messages.
	WarnLevel  LogLevel = "warn"  // WarnLevel represents warning messages.
	DebugLevel LogLevel = "debug" // DebugLevel represents debug messages for troubleshooting.
	TraceLevel LogLevel = "trace" // TraceLevel represents detailed trace messages for fine-grained debugging.
	ErrorLevel LogLevel = "error" // ErrorLevel represents error messages.
	PanicLevel LogLevel = "panic" // PanicLevel represents panic messages indicating critical failures.
)

var (
	// logLevels contains all supported log levels.
	logLevels = []LogLevel{InfoLevel, WarnLevel, DebugLevel, TraceLevel, ErrorLevel, PanicLevel}
)

// AvailableLogLevels returns the list of all supported log levels.
func AvailableLogLevels() []LogLevel {
	return logLevels
}

// IsSupportedLogLevel checks if the given log level string is a supported log level.
// It returns true if the log level is in the list of available log levels.
func IsSupportedLogLevel(logLevel string) bool {
	return lo.Contains(AvailableLogLevels(), LogLevel(logLevel))
}
