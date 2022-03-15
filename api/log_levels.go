package api

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
