package action

import (
	"testing"

	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/jvm"
	"github.com/stretchr/testify/assert"
)

func TestNewProfile(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
		assert  func(t *testing.T, p profiler.Profiler, job *job.ProfilingJob)
	}{
		{
			name: "New action",
			args: map[string]any{
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
			assert: func(t *testing.T, p profiler.Profiler, job *job.ProfilingJob) {
				assert.IsType(t, profiler.NewMockProfiler(), p)
			},
		},
		{
			name: "New profilejvm heap dump",
			args: map[string]any{
				PrintLogs:                  true,
				Duration:                   "60s",
				Interval:                   "5s",
				JobId:                      "JobId",
				TargetPodUID:               "TargetPodUID",
				TargetContainerID:          "cri-o://TargetContainerID",
				Filename:                   "Filename",
				Lang:                       string(api.Java),
				ProfilingTool:              string(api.Jcmd),
				OutputType:                 string(api.HeapDump),
				EventType:                  "",
				CompressorType:             "",
				TargetContainerRuntime:     "crio",
				TargetContainerRuntimePath: "/my/path",
				OutputSplitInChunkSize:     "50M",
			},
			assert: func(t *testing.T, p profiler.Profiler, job *job.ProfilingJob) {
				assert.IsType(t, &jvm.JcmdProfiler{}, p)
				assert.Equal(t, api.HeapDump, job.OutputType)
				assert.Equal(t, "50M", job.OutputSplitInChunkSize)
			},
		},
		{
			name: "New profilenode heap snapshot",
			args: map[string]any{
				PrintLogs:                  true,
				Duration:                   "60s",
				Interval:                   "5s",
				JobId:                      "JobId",
				TargetPodUID:               "TargetPodUID",
				TargetContainerID:          "cri-o://TargetContainerID",
				Filename:                   "Filename",
				Lang:                       string(api.Node),
				ProfilingTool:              string(api.NodeDummy),
				OutputType:                 string(api.HeapSnapshot),
				EventType:                  "",
				CompressorType:             "",
				TargetContainerRuntime:     "crio",
				TargetContainerRuntimePath: "/my/path",
				OutputSplitInChunkSize:     "50M",
				NodeHeapSnapshotSignal:     12,
			},
			assert: func(t *testing.T, p profiler.Profiler, job *job.ProfilingJob) {
				assert.IsType(t, &profiler.NodeDummyProfiler{}, p)
				assert.Equal(t, api.HeapSnapshot, job.OutputType)
				assert.Equal(t, "50M", job.OutputSplitInChunkSize)
				assert.Equal(t, 12, job.NodeHeapSnapshotSignal)
			},
		},
		{
			name: "New profileempty duration",
			args: map[string]any{
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
			assert: func(t *testing.T, p profiler.Profiler, job *job.ProfilingJob) {
				assert.IsType(t, profiler.NewMockProfiler(), p)
			},
		},
		{
			name: "New profilefail when wrong duration",
			args: map[string]any{
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
			assert: func(t *testing.T, p profiler.Profiler, job *job.ProfilingJob) {
				assert.Empty(t, p)
			},
		},
		{
			name: "New profilefail when unsupported container runtime",
			args: map[string]any{
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
			assert: func(t *testing.T, p profiler.Profiler, job *job.ProfilingJob) {
				assert.Empty(t, p)
			},
		},
		{
			name: "New profilefail when unsupported language",
			args: map[string]any{
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
			assert: func(t *testing.T, p profiler.Profiler, job *job.ProfilingJob) {
				assert.Empty(t, p)
			},
		},
		{
			name: "New profilefail when unsupported event",
			args: map[string]any{
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
			assert: func(t *testing.T, p profiler.Profiler, job *job.ProfilingJob) {
				assert.Empty(t, p)
			},
		},
		{
			name: "New profilefail when unsupported compressor",
			args: map[string]any{
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
			assert: func(t *testing.T, p profiler.Profiler, job *job.ProfilingJob) {
				assert.Empty(t, p)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			var p profiler.Profiler
			var j *job.ProfilingJob
			var err error

			// When
			p, j, err = NewProfile(tt.args)

			// Then
			if err != nil && !tt.wantErr {
				t.Errorf("EventLn() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.assert(t, p, j)
		})
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
		verify  func(mock profiler.Profiler) bool
	}{
		{
			name: "Run action",
			args: map[string]any{
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
			args: map[string]any{
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
			args: map[string]any{
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
			args: map[string]any{
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
			p, profilingJob, err = NewProfile(tt.args)
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
