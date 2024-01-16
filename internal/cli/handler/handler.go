package handler

import (
	"fmt"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/result"
)

type EventHandler struct {
	Target   *config.TargetConfig
	LogLevel api.LogLevel
}

func NewEventHandler(cfg *config.TargetConfig, level api.LogLevel) *EventHandler {
	return &EventHandler{
		Target:   cfg,
		LogLevel: level,
	}
}

func (h *EventHandler) Handle(events chan string, done chan bool, resultFile chan result.File) {
	for eventString := range events {
		event, _ := api.ParseEvent(eventString)
		switch eventType := event.(type) {
		case *api.ErrorData:
			fmt.Printf("Error: %s ", eventType.Reason)
			fmt.Printf("❌\n")
			done <- true
		case *api.ResultData:
			resultFile <- result.File{
				FileName:        eventType.File,
				FileSizeInBytes: eventType.FileSizeInBytes,
				Checksum:        eventType.Checksum,
				Chunks:          eventType.Chunks,
				Timestamp:       eventType.Time,
			}
		case *api.ProgressData:
			h.reportProgress(eventType, done)
		default:
		}
	}
}

func (*EventHandler) reportProgress(data *api.ProgressData, done chan bool) {
	if data.Stage == api.Started {
		fmt.Printf("Profiling ... ✔\n")
	} else if data.Stage == api.Ended {
		done <- true
	}
}
