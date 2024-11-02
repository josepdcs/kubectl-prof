package handler

import (
	"testing"

	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/cli"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
)

func TestEventHandler_reportProgress(t *testing.T) {
	type fields struct {
		Target   *config.TargetConfig
		LogLevel api.LogLevel
	}
	type args struct {
		data *api.ProgressData
		done chan bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "Started stage",
			fields: fields{},
			args: args{
				data: &api.ProgressData{
					Stage: api.Started,
				},
			},
		},
		{
			name:   "Ended stage",
			fields: fields{},
			args: args{
				data: &api.ProgressData{
					Stage: api.Ended,
				},
				done: make(chan bool, 1),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &EventHandler{
				target:  tt.fields.Target,
				printer: cli.NewPrinter(false),
			}
			h.reportProgress(tt.args.data, tt.args.done)
			if tt.args.done != nil {
				t.Logf("From done chan: %v", <-tt.args.done)
			}
		})
	}
}
