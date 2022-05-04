package handler

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/josepdcs/kubectl-prof/pkg/cli/config"
	"github.com/josepdcs/kubectl-prof/pkg/cli/kubernetes"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	log "github.com/sirupsen/logrus"
	"io/ioutil"

	"github.com/josepdcs/kubectl-prof/api"
	batchv1 "k8s.io/api/batch/v1"
)

type EventHandler struct {
	Job      *batchv1.Job
	Target   *config.TargetConfig
	Deleter  kubernetes.Deleter
	LogLevel api.LogLevel
}

func NewEventHandler(job *batchv1.Job, cfg *config.TargetConfig, del kubernetes.Deleter, level api.LogLevel) *EventHandler {
	return &EventHandler{
		Job:      job,
		Target:   cfg,
		Deleter:  del,
		LogLevel: level,
	}
}

func (h *EventHandler) Handle(events chan string, done chan bool, ctx context.Context) {
	for eventString := range events {
		event, err := api.ParseEvent(eventString)
		if err != nil {
			fmt.Printf("Got invalid event: %s\n", err)
		} else {
			switch eventType := event.(type) {
			case *api.ErrorData:
				fmt.Printf("Error: %s\n", eventType.Reason)
			case *api.OutputData:
				h.writeEncodedFile(eventType.EncodedData)
			case *api.ProgressData:
				h.reportProgress(eventType, done, ctx)
			case *api.LogData:
				h.logger(eventType)
			default:
				fmt.Printf("Unrecognized event type: %T!\n", eventType)
			}
		}
	}
}

func (h *EventHandler) writeEncodedFile(encoded string) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		fmt.Printf("Failed to decode result data: %v\n", err)
		return
	}

	c, err := compressor.Get(h.Target.Compressor)
	if err != nil {
		fmt.Printf("Failed to get compressor: %v\n", err)
	}

	decoded, err = c.Decode(decoded)
	if err != nil {
		fmt.Printf("Failed to decode snappy result data: %v\n", err)
	}

	err = ioutil.WriteFile(h.Target.FileName, decoded, 0777)
	if err != nil {
		fmt.Printf("Failed to write result file: %v\n", err)
	}
}

func (h *EventHandler) reportProgress(data *api.ProgressData, done chan bool, ctx context.Context) {
	if data.Stage == api.Started {
		fmt.Printf("Profiling ...")
	} else if data.Stage == api.Ended {
		_ = h.Deleter.DeleteProfilingJob(h.Job, ctx)
		fmt.Printf("✔\nResult profiling data saved to: %s 🔥\n", h.Target.FileName)
		done <- true
	}
}

//logger print log
func (h *EventHandler) logger(data *api.LogData) {
	if api.LogLevel(data.Level) == h.LogLevel {
		fmt.Print("\n")
		switch data.Level {
		case string(api.InfoLevel):
			log.Infof("%s", data.Msg)
		case string(api.WarnLevel):
			log.Warnf("%s", data.Msg)
		case string(api.DebugLevel):
			log.Debugf("%s", data.Msg)
		case string(api.ErrorLevel):
			log.Errorf("%s", data.Msg)
		default:
			log.Tracef("%s", data.Msg)
		}
		fmt.Printf("Still profiling ...")
	}
}
