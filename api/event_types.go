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

package api

import (
	jsoniter "github.com/json-iterator/go"
	"time"
)

type EventType string
type ProgressStage string

const (
	Error      EventType = "error"
	FlameGraph EventType = "flamegraph"
	Jfr        EventType = "jfr"
	Progress   EventType = "progress"
	Log        EventType = "log"

	Started ProgressStage = "started"
	Ended   ProgressStage = "ended"
)

type Event struct {
	Type EventType            `json:"type"`
	Data *jsoniter.RawMessage `json:"data"`
}

type ErrorData struct {
	Reason string `json:"reason"`
}

type FlameGraphData struct {
	EncodedFile string `json:"encoded_file"`
}

type JfrData struct {
	EncodedFile string `json:"encoded_file"`
}

type ProgressData struct {
	Time  time.Time     `json:"time"`
	Stage ProgressStage `json:"stage"`
}

type LogData struct {
	Time  time.Time `json:"time"`
	Level string    `json:"level"`
	Msg   string    `json:"msg"`
}

var typeToData = map[EventType]interface{}{
	Error:      &ErrorData{},
	FlameGraph: &FlameGraphData{},
	Jfr:        &JfrData{},
	Progress:   &ProgressData{},
	Log:        &LogData{},
}

func GetDataStructByType(t EventType) interface{} {
	return typeToData[t]
}
