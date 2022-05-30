package utils

import (
	"bufio"
	"fmt"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"os"
	"time"
)

func Publish(c api.Compressor, file string, eventType api.EventType) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(f)
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	comp, err := compressor.Get(c)
	if err != nil {
		return err
	}
	compressed, err := comp.Encode(content)
	if err != nil {
		return err
	}

	resultFile := file + api.GetExtensionFileByCompressor[c]
	err = ioutil.WriteFile(resultFile, compressed, 0644)
	if err != nil {
		return fmt.Errorf("could not save compressed file %s, error: %w", resultFile, err)
	}

	return PublishEvent(
		api.Result,
		api.ResultData{
			ResultType:     eventType,
			File:           resultFile,
			CompressorType: string(c),
		},
	)
}

func PublishError(err error) {
	data := &api.ErrorData{Reason: err.Error()}
	_ = PublishEvent(api.Error, data)
}

func PublishEvent(eventType api.EventType, data interface{}) error {
	eventData, err := jsoniter.Marshal(data)
	if err != nil {
		return err
	}

	rawEventData := jsoniter.RawMessage(eventData)
	event := api.Event{Type: eventType, Data: &rawEventData}

	str, err := jsoniter.MarshalToString(event)
	if err != nil {
		return err
	}

	fmt.Println(str)
	return nil
}

func PublishLogEvent(level api.LogLevel, msg string) {
	if len(msg) > 0 {
		_ = PublishEvent(
			api.Log,
			&api.LogData{
				Time:  time.Now(),
				Level: string(level),
				Msg:   fmt.Sprint(msg)},
		)
	}
}
