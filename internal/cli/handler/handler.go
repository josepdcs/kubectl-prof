package handler

import (
	"fmt"

	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/cli"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/result"
)

type EventHandler struct {
	target  *config.TargetConfig
	printer cli.Printer
}

func NewEventHandler(cfg *config.TargetConfig, printer cli.Printer) *EventHandler {
	return &EventHandler{
		target:  cfg,
		printer: printer,
	}
}

func (h *EventHandler) Handle(events chan string, done chan bool, resultFile chan result.File) {
	for eventString := range events {
		event, _ := api.ParseEvent(eventString)
		switch eventType := event.(type) {
		case *api.ErrorData:
			h.printer.Print(fmt.Sprintf("Error: %s ", eventType.Reason))
			h.printer.Print("❌\n")
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
		case *api.NoticeData:
			h.printer.Print(fmt.Sprintf("⚠️ %s\n", eventType.Msg))
			h.printer.Print("Profiling ... 🔬\n")
		default:
		}
	}
}

func (h *EventHandler) reportProgress(data *api.ProgressData, done chan bool) {
	if data.Stage == api.Started {
		h.printer.Print("Profiling ... 🔬\n")
	} else if data.Stage == api.Ended {
		done <- true
	}
}
