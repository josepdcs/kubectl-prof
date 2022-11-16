package log

import (
	"fmt"
	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/josepdcs/kubectl-prof/api"
	jsoniter "github.com/json-iterator/go"
	"sync"
	"time"
)

// Logger the structure with the needed fields
// Print logs can be enabled.
type Logger struct {
	mu        sync.Mutex // ensures atomic writes; protects the following fields
	printLogs bool
}

// New instances new default Logger
func New() *Logger {
	return &Logger{printLogs: false}
}

var std = New()

// Default the default Logger
func Default() *Logger {
	return std
}

// SetPrintLogs enables or disables print logs on standard output
// Default is false
func (l *Logger) SetPrintLogs(printLogs bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.printLogs = printLogs
}

// SetPrintLogs enables or disables print logs on standard output
func SetPrintLogs(printLogs bool) {
	std.SetPrintLogs(printLogs)
}

// PrintLogs returns if print logs on standard output is enabled
func PrintLogs() bool {
	return std.printLogs
}

func (l *Logger) EventLn(eventType api.EventType, data interface{}) error {
	eventData, err := jsoniter.Marshal(data)
	if err != nil {
		return err
	}

	rawEventData := jsoniter.RawMessage(eventData)
	event := api.Event{Type: eventType, Data: &rawEventData}
	str, _ := jsoniter.MarshalToString(event)

	// print on standard output if allowed
	if l.isPrintLogsAllowed(eventType) {
		fmt.Println(str)
	}

	return nil
}

func (l *Logger) isPrintLogsAllowed(eventType api.EventType) bool {
	if eventType != api.Log {
		return true
	}
	return l.printLogs
}

func EventLn(eventType api.EventType, data interface{}) error {
	return std.EventLn(eventType, data)
}

func (l *Logger) ErrorLn(err error) {
	data := &api.ErrorData{Reason: err.Error()}
	_ = l.EventLn(api.Error, data)
}

func ErrorLn(err error) {
	std.ErrorLn(err)
}

func (l *Logger) PrintLogLn(level api.LogLevel, msg string) {
	if stringUtils.IsNotBlank(msg) {
		_ = l.EventLn(
			api.Log,
			&api.LogData{
				Time:  time.Now(),
				Level: string(level),
				Msg:   msg},
		)
	}
}

func PrintLogLn(level api.LogLevel, msg string) {
	std.PrintLogLn(level, msg)
}

func InfoLogLn(msg string) {
	std.PrintLogLn(api.InfoLevel, msg)
}

func DebugLogLn(msg string) {
	std.PrintLogLn(api.DebugLevel, msg)
}

func WarningLogLn(msg string) {
	std.PrintLogLn(api.WarnLevel, msg)
}

func ErrorLogLn(msg string) {
	std.PrintLogLn(api.ErrorLevel, msg)
}
