package profile

import (
	"testing"

	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler"
	"github.com/stretchr/testify/assert"
)

func TestNewAction(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]interface{}
		wantErr bool
		assert  func(t *testing.T, p profiler.Profiler)
	}{
		{
			name: "New action",
			args: map[string]interface{}{
				PrintLogs:                  true,
				Duration:                   "60s",
				Interval:                   "5s",
				JobId:                      "JobId",
				TargetPodUID:               "TargetPodUID",
				TargetContainerID:          "cri-o://TargetContainerID",
				Filename:                   "Filename",
				Lang:                       string(api.FakeLang),
				ProfilingTool:              string(api.Jcmd),
				OutputType:                 string(api.Jfr),
				EventType:                  "",
				CompressorType:             "",
				TargetContainerRuntime:     "crio",
				TargetContainerRuntimePath: "/my/path",
			},
			assert: func(t *testing.T, p profiler.Profiler) {
				assert.IsType(t, profiler.NewMockProfiler(), p)
			},
		},
		{
			name: "New action empty duration",
			args: map[string]interface{}{
				PrintLogs:                  true,
				Duration:                   "",
				Interval:                   "",
				JobId:                      "JobId",
				TargetPodUID:               "TargetPodUID",
				TargetContainerID:          "cri-o://TargetContainerID",
				Filename:                   "Filename",
				Lang:                       string(api.FakeLang),
				ProfilingTool:              string(api.Jcmd),
				OutputType:                 string(api.Jfr),
				EventType:                  "",
				CompressorType:             "",
				TargetContainerRuntime:     "containerd",
				TargetContainerRuntimePath: "/my/path",
			},
			assert: func(t *testing.T, p profiler.Profiler) {
				assert.IsType(t, profiler.NewMockProfiler(), p)
			},
		},
		{
			name: "New action fail when wrong duration",
			args: map[string]interface{}{
				PrintLogs:                  true,
				Duration:                   "duration_wrong",
				Interval:                   "5s",
				JobId:                      "JobId",
				TargetPodUID:               "TargetPodUID",
				TargetContainerID:          "cri-o://TargetContainerID",
				Filename:                   "Filename",
				Lang:                       string(api.FakeLang),
				ProfilingTool:              string(api.Jcmd),
				OutputType:                 string(api.Jfr),
				EventType:                  "",
				CompressorType:             "",
				TargetContainerRuntime:     "crio",
				TargetContainerRuntimePath: "/my/path",
			},
			wantErr: true,
			assert: func(t *testing.T, p profiler.Profiler) {
				assert.Empty(t, p)
			},
		},
		{
			name: "New action fail when unsupported container runtime",
			args: map[string]interface{}{
				PrintLogs:                  true,
				Duration:                   "60s",
				Interval:                   "5s",
				JobId:                      "JobId",
				TargetPodUID:               "TargetPodUID",
				TargetContainerID:          "cri-o://TargetContainerID",
				Filename:                   "Filename",
				Lang:                       string(api.FakeLang),
				ProfilingTool:              string(api.Jcmd),
				OutputType:                 string(api.Jfr),
				EventType:                  "",
				CompressorType:             "",
				TargetContainerRuntime:     "unsupported",
				TargetContainerRuntimePath: "/my/path",
			},
			wantErr: true,
			assert: func(t *testing.T, p profiler.Profiler) {
				assert.Empty(t, p)
			},
		},
		{
			name: "New action fail when unsupported language",
			args: map[string]interface{}{
				PrintLogs:                  true,
				Duration:                   "60s",
				Interval:                   "5s",
				JobId:                      "JobId",
				TargetPodUID:               "TargetPodUID",
				TargetContainerID:          "cri-o://TargetContainerID",
				Filename:                   "Filename",
				Lang:                       "unsupported",
				ProfilingTool:              string(api.Jcmd),
				OutputType:                 string(api.Jfr),
				EventType:                  "",
				CompressorType:             "",
				TargetContainerRuntime:     "crio",
				TargetContainerRuntimePath: "/my/path",
			},
			wantErr: true,
			assert: func(t *testing.T, p profiler.Profiler) {
				assert.Empty(t, p)
			},
		},
		{
			name: "New action fail when unsupported event",
			args: map[string]interface{}{
				PrintLogs:                  true,
				Duration:                   "60s",
				Interval:                   "5s",
				JobId:                      "JobId",
				TargetPodUID:               "TargetPodUID",
				TargetContainerID:          "cri-o://TargetContainerID",
				Filename:                   "Filename",
				Lang:                       string(api.FakeLang),
				ProfilingTool:              string(api.Jcmd),
				OutputType:                 string(api.Jfr),
				EventType:                  "unsupported",
				CompressorType:             "",
				TargetContainerRuntime:     "crio",
				TargetContainerRuntimePath: "/my/path",
			},
			wantErr: true,
			assert: func(t *testing.T, p profiler.Profiler) {
				assert.Empty(t, p)
			},
		},
		{
			name: "New action fail when unsupported compressor",
			args: map[string]interface{}{
				PrintLogs:                  true,
				Duration:                   "60s",
				Interval:                   "5s",
				JobId:                      "JobId",
				TargetPodUID:               "TargetPodUID",
				TargetContainerID:          "cri-o://TargetContainerID",
				Filename:                   "Filename",
				Lang:                       string(api.FakeLang),
				ProfilingTool:              string(api.Jcmd),
				OutputType:                 string(api.Jfr),
				EventType:                  "",
				CompressorType:             "unsupported",
				TargetContainerRuntime:     "crio",
				TargetContainerRuntimePath: "/my/path",
			},
			wantErr: true,
			assert: func(t *testing.T, p profiler.Profiler) {
				assert.Empty(t, p)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			var p profiler.Profiler
			var err error

			// When
			p, _, err = NewAction(tt.args)

			// Then
			if err != nil && !tt.wantErr {
				t.Errorf("EventLn() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.assert(t, p)
		})
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]interface{}
		wantErr bool
		verify  func(mock profiler.Profiler) bool
	}{
		{
			name: "Run action",
			args: map[string]interface{}{
				PrintLogs:                  true,
				Duration:                   "60s",
				Interval:                   "60s",
				JobId:                      "JobId",
				TargetPodUID:               "TargetPodUID",
				TargetContainerID:          "cri-o://TargetContainerID",
				Filename:                   "Filename",
				Lang:                       string(api.FakeLang),
				ProfilingTool:              string(api.Jcmd),
				OutputType:                 string(api.Jfr),
				EventType:                  "",
				CompressorType:             "",
				TargetContainerRuntime:     "crio",
				TargetContainerRuntimePath: "/my/path",
			},
			verify: func(p profiler.Profiler) bool {
				mock := p.(profiler.MockProfiler)
				return mock.SetUpInvokedTimes() == 1 && mock.InvokeInvokedTimes() == 1 && mock.CleanUpInvokedTimes() == 0
			},
		},
		{
			name: "Run action with window smaller than duration",
			args: map[string]interface{}{
				PrintLogs:                  true,
				Duration:                   "2s",
				Interval:                   "1s",
				JobId:                      "JobId",
				TargetPodUID:               "TargetPodUID",
				TargetContainerID:          "cri-o://TargetContainerID",
				Filename:                   "Filename",
				Lang:                       string(api.FakeLang),
				ProfilingTool:              string(api.Jcmd),
				OutputType:                 string(api.Jfr),
				EventType:                  "",
				CompressorType:             "",
				TargetContainerRuntime:     "crio",
				TargetContainerRuntimePath: "/my/path",
			},
			verify: func(p profiler.Profiler) bool {
				mock := p.(profiler.MockProfiler)
				return mock.SetUpInvokedTimes() == 1 && mock.InvokeInvokedTimes() == 2 && mock.CleanUpInvokedTimes() == 0
			},
		},
		{
			name: "Run action fail when setup profiler fail",
			args: map[string]interface{}{
				PrintLogs:                  true,
				Duration:                   "60s",
				Interval:                   "60s",
				JobId:                      "JobId",
				TargetPodUID:               "WithSetupError",
				TargetContainerID:          "cri-o://WithSetupError",
				Filename:                   "Filename",
				Lang:                       string(api.FakeLang),
				ProfilingTool:              string(api.Jcmd),
				OutputType:                 string(api.Jfr),
				EventType:                  "",
				CompressorType:             "",
				TargetContainerRuntime:     "crio",
				TargetContainerRuntimePath: "/my/path",
			},
			wantErr: true,
			verify: func(p profiler.Profiler) bool {
				mock := p.(profiler.MockProfiler)
				return mock.SetUpInvokedTimes() == 1 && mock.InvokeInvokedTimes() == 0 && mock.CleanUpInvokedTimes() == 0
			},
		},
		{
			name: "Run action fail when invoke profiler fail",
			args: map[string]interface{}{
				PrintLogs:                  true,
				Duration:                   "60s",
				Interval:                   "60s",
				JobId:                      "JobId",
				TargetPodUID:               "WithInvokeError",
				TargetContainerID:          "cri-o://WithInvokeError",
				Filename:                   "Filename",
				Lang:                       string(api.FakeLang),
				ProfilingTool:              string(api.Jcmd),
				OutputType:                 string(api.Jfr),
				EventType:                  "",
				CompressorType:             "",
				TargetContainerRuntime:     "crio",
				TargetContainerRuntimePath: "/my/path",
			},
			wantErr: true,
			verify: func(p profiler.Profiler) bool {
				mock := p.(profiler.MockProfiler)
				return mock.SetUpInvokedTimes() == 1 && mock.InvokeInvokedTimes() == 1 && mock.CleanUpInvokedTimes() == 0
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			var p profiler.Profiler
			var profilingJob *job.ProfilingJob
			var err error

			// When
			p, profilingJob, err = NewAction(tt.args)
			err = Run(p, profilingJob)

			// Then
			if err != nil && !tt.wantErr {
				t.Errorf("EventLn() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.verify != nil && !tt.verify(p) {
				t.Errorf("Error verifying behaviour: %s", tt.name)
			}
		})
	}
}

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
