package api

import "github.com/samber/lo"

type LogLevel string

const (
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	DebugLevel LogLevel = "debug"
	TraceLevel LogLevel = "trace"
	ErrorLevel LogLevel = "error"
	PanicLevel LogLevel = "panic"
)

var (
	logLevels = []LogLevel{InfoLevel, WarnLevel, DebugLevel, TraceLevel, ErrorLevel, PanicLevel}
)

func AvailableLogLevels() []LogLevel {
	return logLevels
}

func IsSupportedLogLevel(logLevel string) bool {
	return lo.Contains(AvailableLogLevels(), LogLevel(logLevel))
}
