package handler

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/josepdcs/kubectl-perf/pkg/cli/config"
	"github.com/josepdcs/kubectl-perf/pkg/cli/kubernetes"
	log "github.com/sirupsen/logrus"
	"io/ioutil"

	"github.com/josepdcs/kubectl-perf/api"
	batchv1 "k8s.io/api/batch/v1"
)

type EventHandler struct {
	Job     *batchv1.Job
	Target  *config.TargetConfig
	Deleter kubernetes.Deleter
}

func NewEventHandler(job *batchv1.Job, cfg *config.TargetConfig, del kubernetes.Deleter) *EventHandler {
	return &EventHandler{
		Job:     job,
		Target:  cfg,
		Deleter: del,
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
			case *api.FlameGraphData:
				h.createFlameGraph(eventType)
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

func (h *EventHandler) createFlameGraph(data *api.FlameGraphData) {
	decodedData, err := base64.StdEncoding.DecodeString(data.EncodedFile)
	if err != nil {
		fmt.Printf("Failed to decode flamegraph: %v\n", err)
		return
	}

	err = ioutil.WriteFile(h.Target.FileName, decodedData, 0777)
	if err != nil {
		fmt.Printf("Failed to write flamegraph file: %v\n", err)
	}
}

func (h *EventHandler) reportProgress(data *api.ProgressData, done chan bool, ctx context.Context) {
	if data.Stage == api.Started {
		fmt.Printf("Profiling ...")
	} else if data.Stage == api.Ended {
		_ = h.Deleter.DeleteProfilingJob(h.Job, h.Target, ctx)
		fmt.Printf("✔\nProfiled as FrameGraph saved to: %s 🔥\n", h.Target.FileName)
		done <- true
	}
}

//logger print log
func (h *EventHandler) logger(data *api.LogData) {
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
	fmt.Printf("Profiling ...")
}
