package api

import (
	"encoding/json"
	"fmt"
)

func PublishError(err error) {
	data := &ErrorData{Reason: err.Error()}
	_ = PublishEvent(Error, data)
}

func PublishEvent(eventType EventType, data interface{}) error {
	eventData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	rawEventData := json.RawMessage(eventData)
	event := Event{Type: eventType, Data: &rawEventData}

	eventString, err := json.Marshal(event)
	if err != nil {
		return err
	}

	fmt.Println(string(eventString))
	return nil
}

func ParseEvent(eventString string) (interface{}, error) {
	event := &Event{}
	err := json.Unmarshal([]byte(eventString), event)
	if err != nil {
		return nil, err
	}

	eventData := GetDataStructByType(event.Type)
	err = json.Unmarshal(*event.Data, eventData)
	return eventData, err
}
