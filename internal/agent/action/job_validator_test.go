package action

import (
	"testing"
	"time"

	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/stretchr/testify/assert"
)

func Test_validateProfilingTool(t *testing.T) {
	type args struct {
		profilingTool string
		outputType    string
		job           *job.ProfilingJob
	}
	tests := []struct {
		name   string
		args   args
		assert func(t *testing.T, job *job.ProfilingJob)
	}{
		{
			name: "Get default tool",
			args: args{
				profilingTool: "",
				outputType:    "jfr",
				job: &job.ProfilingJob{
					Language: api.Java,
				},
			},
			assert: func(t *testing.T, job *job.ProfilingJob) {
				assert.Equal(t, api.Jcmd, job.Tool)
			},
		},
		{
			name: "Get default tool when not supported given tool",
			args: args{
				profilingTool: "other",
				job: &job.ProfilingJob{
					Language: api.Java,
				},
			},
			assert: func(t *testing.T, job *job.ProfilingJob) {
				assert.Equal(t, api.Jcmd, job.Tool)
			},
		},
		{
			name: "Get default tool when not valid given tool",
			args: args{
				profilingTool: string(api.Bpf),
				job: &job.ProfilingJob{
					Language: api.Java,
				},
			},
			assert: func(t *testing.T, job *job.ProfilingJob) {
				assert.Equal(t, api.Jcmd, job.Tool)
			},
		},
		{
			name: "Get tool",
			args: args{
				profilingTool: string(api.Jcmd),
				job: &job.ProfilingJob{
					Language: api.Java,
				},
			},
			assert: func(t *testing.T, job *job.ProfilingJob) {
				assert.Equal(t, api.Jcmd, job.Tool)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given & When
			validateProfilingTool(tt.args.profilingTool, tt.args.outputType, tt.args.job)

			// Then
			tt.assert(t, tt.args.job)
		})
	}
}

func Test_validateOutputType(t *testing.T) {
	type args struct {
		outputType string
		job        *job.ProfilingJob
	}
	tests := []struct {
		name   string
		args   args
		assert func(t *testing.T, job *job.ProfilingJob)
	}{
		{
			name: "Get default output type",
			args: args{
				outputType: "",
				job: &job.ProfilingJob{
					Tool: api.Jcmd,
				},
			},
			assert: func(t *testing.T, job *job.ProfilingJob) {
				assert.Equal(t, api.Jfr, job.OutputType)
			},
		},
		{
			name: "Get default output type when not supported output given",
			args: args{
				outputType: "other",
				job: &job.ProfilingJob{
					Tool: api.Jcmd,
				},
			},
			assert: func(t *testing.T, job *job.ProfilingJob) {
				assert.Equal(t, api.Jfr, job.OutputType)
			},
		},
		{
			name: "Get default output type when not valid output given",
			args: args{
				outputType: string(api.Jfr),
				job: &job.ProfilingJob{
					Tool: api.Bpf,
				},
			},
			assert: func(t *testing.T, job *job.ProfilingJob) {
				assert.Equal(t, api.FlameGraph, job.OutputType)
			},
		},
		{
			name: "Get output type",
			args: args{
				outputType: string(api.Jfr),
				job: &job.ProfilingJob{
					Tool: api.Jcmd,
				},
			},
			assert: func(t *testing.T, job *job.ProfilingJob) {
				assert.Equal(t, api.Jfr, job.OutputType)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given & When
			validateOutputType(tt.args.outputType, tt.args.job)

			// Then
			tt.assert(t, tt.args.job)
		})
	}
}

func Test_setAsyncProfilerArgs(t *testing.T) {
	type args struct {
		args map[string]any
		job  *job.ProfilingJob
	}
	tests := []struct {
		name   string
		args   args
		assert func(t *testing.T, job *job.ProfilingJob)
	}{
		{
			name: "should set single async-profiler argument",
			args: args{
				args: map[string]any{
					AsyncProfilerArg: []string{"-t"},
				},
				job: &job.ProfilingJob{},
			},
			assert: func(t *testing.T, job *job.ProfilingJob) {
				assert.NotNil(t, job.AdditionalArguments)
				assert.Equal(t, "-t", job.AdditionalArguments["async-profiler-arg-0"])
				assert.Equal(t, 1, len(job.AdditionalArguments))
			},
		},
		{
			name: "should set multiple async-profiler arguments in order",
			args: args{
				args: map[string]any{
					AsyncProfilerArg: []string{"-t", "--alloc=2m", "--lock=10ms"},
				},
				job: &job.ProfilingJob{},
			},
			assert: func(t *testing.T, job *job.ProfilingJob) {
				assert.NotNil(t, job.AdditionalArguments)
				assert.Equal(t, "-t", job.AdditionalArguments["async-profiler-arg-0"])
				assert.Equal(t, "--alloc=2m", job.AdditionalArguments["async-profiler-arg-1"])
				assert.Equal(t, "--lock=10ms", job.AdditionalArguments["async-profiler-arg-2"])
				assert.Equal(t, 3, len(job.AdditionalArguments))
			},
		},
		{
			name: "should not set anything when args is nil",
			args: args{
				args: map[string]any{
					AsyncProfilerArg: nil,
				},
				job: &job.ProfilingJob{},
			},
			assert: func(t *testing.T, job *job.ProfilingJob) {
				assert.Nil(t, job.AdditionalArguments)
			},
		},
		{
			name: "should not set anything when args is empty slice",
			args: args{
				args: map[string]any{
					AsyncProfilerArg: []string{},
				},
				job: &job.ProfilingJob{},
			},
			assert: func(t *testing.T, job *job.ProfilingJob) {
				assert.Nil(t, job.AdditionalArguments)
			},
		},
		{
			name: "should not override existing additional arguments",
			args: args{
				args: map[string]any{
					AsyncProfilerArg: []string{"-t"},
					"existing":       "value",
				},
				job: &job.ProfilingJob{
					AdditionalArguments: map[string]string{
						"existing": "value",
					},
				},
			},
			assert: func(t *testing.T, job *job.ProfilingJob) {
				assert.NotNil(t, job.AdditionalArguments)
				assert.Equal(t, "-t", job.AdditionalArguments["async-profiler-arg-0"])
				assert.Equal(t, "value", job.AdditionalArguments["existing"])
				assert.Equal(t, 2, len(job.AdditionalArguments))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setAsyncProfilerArgs(tt.args.args, tt.args.job)
			tt.assert(t, tt.args.job)
		})
	}
}

func TestValidateJob(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		verify  func(t *testing.T, j *job.ProfilingJob)
		wantErr bool
	}{
		{
			name: "Full validation success",
			args: map[string]any{
				JobId:                      "job-1",
				Duration:                   "10s",
				Interval:                   "5s",
				TargetContainerRuntime:     string(api.Containerd),
				TargetContainerRuntimePath: "/var/run/containerd/containerd.sock",
				TargetPodUID:               "pod-uid",
				TargetContainerID:          "container-id",
				Lang:                       string(api.Java),
				EventType:                  string(api.Cpu),
				CompressorType:             string(compressor.Gzip),
				ProfilingTool:              string(api.Jcmd),
				OutputType:                 string(api.HeapDump),
				Filename:                   "result.hprof",
				OutputSplitInChunkSize:     "10M",
			},
			verify: func(t *testing.T, j *job.ProfilingJob) {
				assert.Equal(t, "job-1", j.UID)
				assert.Equal(t, 10*time.Second, j.Duration)
				assert.Equal(t, 5*time.Second, j.Interval)
				assert.Equal(t, api.Containerd, j.ContainerRuntime)
				assert.Equal(t, "/var/run/containerd/containerd.sock", j.ContainerRuntimePath)
				assert.Equal(t, "pod-uid", j.PodUID)
				assert.Equal(t, "container-id", j.ContainerID)
				assert.Equal(t, api.Java, j.Language)
				assert.Equal(t, api.Cpu, j.Event)
				assert.Equal(t, compressor.Type(compressor.Gzip), j.Compressor)
				assert.Equal(t, api.Jcmd, j.Tool)
				assert.Equal(t, api.HeapDump, j.OutputType)
				assert.Equal(t, "result.hprof", j.FileName)
				assert.Equal(t, "10M", j.OutputSplitInChunkSize)
			},
			wantErr: false,
		},
		{
			name: "Default values",
			args: map[string]any{
				JobId:                      "job-2",
				Duration:                   "",
				Interval:                   "",
				TargetContainerRuntime:     "",
				TargetContainerRuntimePath: "",
				TargetPodUID:               "",
				TargetContainerID:          "",
				Lang:                       string(api.FakeLang),
				EventType:                  "",
				CompressorType:             "",
				ProfilingTool:              "",
				OutputType:                 "",
				Filename:                   "",
				OutputSplitInChunkSize:     "",
			},
			verify: func(t *testing.T, j *job.ProfilingJob) {
				assert.Equal(t, defaultDuration, j.Duration)
				assert.Equal(t, defaultDuration, j.Interval)
				assert.Equal(t, defaultContainerRuntime, j.ContainerRuntime)
				assert.Equal(t, api.FakeLang, j.Language)
				assert.Equal(t, defaultEventType, j.Event)
				assert.Equal(t, compressor.Type(defaultCompressor), j.Compressor)
				assert.Equal(t, "", j.OutputSplitInChunkSize)
			},
			wantErr: false,
		},
		{
			name: "Invalid duration",
			args: map[string]any{
				Duration: "invalid",
			},
			wantErr: true,
		},
		{
			name: "Invalid interval",
			args: map[string]any{
				Duration: "10s",
				Interval: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &job.ProfilingJob{}
			err := validateJob(tt.args, j)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.verify != nil {
					tt.verify(t, j)
				}
			}
		})
	}
}
