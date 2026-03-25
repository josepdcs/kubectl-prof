package cmd

import (
	"testing"

	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/stretchr/testify/assert"
)

func TestValidateFlags(t *testing.T) {
	type args struct {
		flags  *profilingFlags
		target *config.TargetConfig
		job    *config.JobConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid flags",
			args: args{
				flags: &profilingFlags{
					lang:            string(api.Go),
					runtime:         string(api.Containerd),
					event:           string(api.Cpu),
					logLevel:        string(api.InfoLevel),
					compressorType:  "gzip",
					imagePullPolicy: "Always",
				},
				target: &config.TargetConfig{},
				job:    &config.JobConfig{},
			},
			wantErr: false,
		},
		{
			name: "invalid language",
			args: args{
				flags: &profilingFlags{
					lang: "invalid",
				},
				target: &config.TargetConfig{},
				job:    &config.JobConfig{},
			},
			wantErr: true,
		},
		{
			name: "empty language",
			args: args{
				flags: &profilingFlags{
					lang: "",
				},
				target: &config.TargetConfig{},
				job:    &config.JobConfig{},
			},
			wantErr: true,
		},
		{
			name: "invalid runtime",
			args: args{
				flags: &profilingFlags{
					lang:    string(api.Go),
					runtime: "invalid",
				},
				target: &config.TargetConfig{},
				job:    &config.JobConfig{},
			},
			wantErr: true,
		},
		{
			name: "invalid event",
			args: args{
				flags: &profilingFlags{
					lang:    string(api.Go),
					runtime: string(api.Containerd),
					event:   "invalid",
				},
				target: &config.TargetConfig{},
				job:    &config.JobConfig{},
			},
			wantErr: true,
		},
		{
			name: "invalid log level",
			args: args{
				flags: &profilingFlags{
					lang:     string(api.Go),
					runtime:  string(api.Containerd),
					logLevel: "invalid",
				},
				target: &config.TargetConfig{},
				job:    &config.JobConfig{},
			},
			wantErr: true,
		},
		{
			name: "invalid image pull policy",
			args: args{
				flags: &profilingFlags{
					lang:            string(api.Go),
					runtime:         string(api.Containerd),
					imagePullPolicy: "invalid",
				},
				target: &config.TargetConfig{},
				job:    &config.JobConfig{},
			},
			wantErr: true,
		},
		{
			name: "invalid pid",
			args: args{
				flags: &profilingFlags{
					lang:    string(api.Go),
					runtime: string(api.Containerd),
				},
				target: &config.TargetConfig{
					ExtraTargetOptions: config.ExtraTargetOptions{
						PID: "abc",
					},
				},
				job: &config.JobConfig{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFlags(tt.args.flags, tt.args.target, tt.args.job)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
