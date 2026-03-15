/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package api provides the core types, constants, and functions for the kubectl-prof profiling tool.
// It defines the supported programming languages, profiling tools, output types, events, and runtime
// configurations used throughout the profiling process.
package api

import (
	"time"

	jsoniter "github.com/json-iterator/go"
)

// EventType represents the type of an event, used to categorize and handle events in a structured manner.
type EventType string

// ProgressStage represents the stage of a profiling job, used to categorize progress events.
type ProgressStage string

const (
	Progress EventType = "progress" // Progress indicates an event type representing the progress of a given operation or task.
	Result   EventType = "result"   // Result indicates an event type representing the result of a profiling operation.
	Notice   EventType = "notice"   // Notice indicates an event type representing a notice or a message.
	Log      EventType = "log"      // Log indicates an event type representing a log message.
	Error    EventType = "error"    // Error indicates an event type representing an error.

	Started ProgressStage = "started" // Started indicates the start of a profiling job.
	Ended   ProgressStage = "ended"   // Ended indicates the end of a profiling job.
)

// Event represents an event emitted by the profiler.
type Event struct {
	Type EventType            `json:"type"`
	Data *jsoniter.RawMessage `json:"data"`
}

// ErrorData represents an error event.
type ErrorData struct {
	Reason string `json:"reason"`
}

// ResultData represents a profiling result event.
type ResultData struct {
	Time            time.Time   `json:"time"`
	ResultType      OutputType  `json:"result-type"`
	File            string      `json:"file,omitempty"`
	FileSizeInBytes int64       `json:"file-size-in-bytes,omitempty"`
	Checksum        string      `json:"checksum,omitempty"`
	CompressorType  string      `json:"compressor-type,omitempty"`
	Chunks          []ChunkData `json:"chunks,omitempty"`
}

// ChunkData represents a profiling result chunk.
type ChunkData struct {
	File            string `json:"file"`
	FileSizeInBytes int64  `json:"file-size-in-bytes"`
	Checksum        string `json:"checksum"`
}

// ProgressData represents a profiling progress event.
type ProgressData struct {
	Time  time.Time     `json:"time"`
	Stage ProgressStage `json:"stage"`
}

// NoticeData represents a profiling notice event.
type NoticeData struct {
	Time time.Time `json:"time"`
	Msg  string    `json:"msg"`
}

// LogData represents a profiling log event.
type LogData struct {
	Time  time.Time `json:"time"`
	Level string    `json:"level"`
	Msg   string    `json:"msg"`
}

// typeToData maps an event type to its corresponding data structure.
var typeToData = map[EventType]interface{}{
	Error:    &ErrorData{},
	Result:   &ResultData{},
	Progress: &ProgressData{},
	Notice:   &NoticeData{},
	Log:      &LogData{},
}

// GetDataStructByType returns the data structure corresponding to the given event type.
func GetDataStructByType(t EventType) interface{} {
	return typeToData[t]
}

// ParseEvent parses the given event string into its corresponding data structure.
func ParseEvent(eventString string) (interface{}, error) {
	event := &Event{}
	err := jsoniter.Unmarshal([]byte(eventString), event)
	if err != nil {
		return nil, err
	}

	eventData := GetDataStructByType(event.Type)
	err = jsoniter.Unmarshal(*event.Data, eventData)
	return eventData, err
}
