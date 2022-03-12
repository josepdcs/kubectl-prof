package handler

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/josepdcs/kubectl-profile/internal/cli/config"
	"github.com/josepdcs/kubectl-profile/internal/cli/kubernetes"
	log "github.com/sirupsen/logrus"
	"io/ioutil"

	"github.com/josepdcs/kubectl-profile/api"
	batchv1 "k8s.io/api/batch/v1"
)

type ApiEventsHandler struct {
	Job    *batchv1.Job
	Target *config.TargetConfig
}

func (h *ApiEventsHandler) Handle(events chan string, done chan bool, ctx context.Context) {
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

func (h *ApiEventsHandler) createFlameGraph(data *api.FlameGraphData) {
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

func (h *ApiEventsHandler) reportProgress(data *api.ProgressData, done chan bool, ctx context.Context) {
	if data.Stage == api.Started {
		fmt.Printf("Profiling ...\n")
	} else if data.Stage == api.Ended {
		_ = kubernetes.DeleteProfilingJob(h.Job, h.Target, ctx)
		fmt.Printf("✔\nFlameGraph saved to: %s 🔥\n", h.Target.FileName)
		done <- true
	}
}

//logger func config message
func (h *ApiEventsHandler) logger(data *api.LogData) {
	switch data.Level {
	case api.InfoLevel:
		log.Info(data.Msg)
	case api.WarnLevel:
		log.Warn(data.Msg)
	case api.DebugLevel:
		log.Debug(data.Msg)
	case api.ErrorLevel:
		log.Error(data.Msg)
	default:
		log.Trace(data.Msg)
	}
}
