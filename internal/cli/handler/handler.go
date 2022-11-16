package handler

import (
	"fmt"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/result"
	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
)

type EventHandler struct {
	Job      *batchv1.Job
	Target   *config.TargetConfig
	LogLevel api.LogLevel
}

func NewEventHandler(job *batchv1.Job, cfg *config.TargetConfig, level api.LogLevel) *EventHandler {
	return &EventHandler{
		Job:      job,
		Target:   cfg,
		LogLevel: level,
	}
}

func (h *EventHandler) Handle(events chan string, done chan bool, resultFile chan result.File) {
	for eventString := range events {
		event, err := api.ParseEvent(eventString)
		if err != nil {
			fmt.Printf("Got invalid event: %s\n", err)
		} else {
			switch eventType := event.(type) {
			case *api.ErrorData:
				fmt.Printf("Error: %s ", eventType.Reason)
				fmt.Printf("❌\n")
				done <- true
			case *api.ResultData:
				resultFile <- result.File{
					FileName:  eventType.File,
					Timestamp: eventType.Time,
				}
			case *api.ProgressData:
				h.reportProgress(eventType, done)
			case *api.LogData:
				h.logger(eventType)
			default:
				fmt.Printf("Unrecognized event type: %T!\n", eventType)
			}
		}
	}
}

func (h *EventHandler) reportProgress(data *api.ProgressData, done chan bool) {
	if data.Stage == api.Started {
		fmt.Printf("Profiling ...")
	} else if data.Stage == api.Ended {
		done <- true
	}
}

// logger print log
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
