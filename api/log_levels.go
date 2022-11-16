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

func IsSupportedLogLevel(event string) bool {
	return containsLogLevel(LogLevel(event), AvailableLogLevels())
}

func containsLogLevel(e LogLevel, events []LogLevel) bool {
	for _, current := range events {
		if e == current {
			return true
		}
	}
	return false
}
