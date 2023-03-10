package api

import "github.com/samber/lo"

type LogLevel string

const (
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	DebugLevel LogLevel = "debug"
	ErrorLevel LogLevel = "error"
)

var (
	logLevels = []LogLevel{InfoLevel, WarnLevel, DebugLevel, ErrorLevel}
)

func AvailableLogLevels() []LogLevel {
	return logLevels
}

func IsSupportedLogLevel(logLevel string) bool {
	return lo.Contains(AvailableLogLevels(), LogLevel(logLevel))
}
