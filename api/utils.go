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
	"fmt"
	jsoniter "github.com/json-iterator/go"
)

func PublishError(err error) {
	data := &ErrorData{Reason: err.Error()}
	_ = PublishEvent(Error, data)
}

func PublishEvent(eventType EventType, data interface{}) error {
	eventData, err := jsoniter.Marshal(data)
	if err != nil {
		return err
	}

	rawEventData := jsoniter.RawMessage(eventData)
	event := Event{Type: eventType, Data: &rawEventData}

	bytes, err := jsoniter.Marshal(event)
	if err != nil {
		return err
	}

	fmt.Println(string(bytes))
	return nil
}

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
