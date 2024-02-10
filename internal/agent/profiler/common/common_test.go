package common

import (
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func TestGetResultFileWithPID(t *testing.T) {
	type args struct {
		targetDir string
		job       *job.ProfilingJob
	}
	tests := []struct {
		name  string
		given func() args
		when  func(args) string
		then  func(t *testing.T, result string)
	}{
		{
			name: "should return Jfr",
			given: func() args {
				return args{
					targetDir: TmpDir(),
					job: &job.ProfilingJob{
						Tool:       api.Jcmd,
						OutputType: api.Jfr,
						Iteration:  1,
					},
				}
			},
			when: func(args args) string {
				return GetResultFile(args.targetDir, args.job.Tool, args.job.OutputType, "PID", args.job.Iteration)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, filepath.Join(TmpDir(), config.ProfilingPrefix+"jfr-PID-1.jfr"), result)
			},
		},
		{
			name: "should return ThreadDump",
			given: func() args {
				return args{
					targetDir: TmpDir(),
					job: &job.ProfilingJob{
						Tool:       api.Jcmd,
						OutputType: api.ThreadDump,
						Iteration:  1,
					},
				}
			},
			when: func(args args) string {
				return GetResultFile(args.targetDir, args.job.Tool, args.job.OutputType, "PID", args.job.Iteration)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, filepath.Join(TmpDir(), config.ProfilingPrefix+"threaddump-PID-1.txt"), result)
			},
		},
		{
			name: "should return HeapDump",
			given: func() args {
				return args{
					targetDir: TmpDir(),
					job: &job.ProfilingJob{
						Tool:       api.Jcmd,
						OutputType: api.HeapDump,
						Iteration:  1,
					},
				}
			},
			when: func(args args) string {
				return GetResultFile(args.targetDir, args.job.Tool, args.job.OutputType, "PID", args.job.Iteration)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, filepath.Join(TmpDir(), config.ProfilingPrefix+"heapdump-PID-1.hprof"), result)
			},
		},
		{
			name: "should return HeapHistogram",
			given: func() args {
				return args{
					targetDir: TmpDir(),
					job: &job.ProfilingJob{
						Tool:       api.Jcmd,
						OutputType: api.HeapHistogram,
						Iteration:  1,
					},
				}
			},
			when: func(args args) string {
				return GetResultFile(args.targetDir, args.job.Tool, args.job.OutputType, "PID", args.job.Iteration)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, filepath.Join(TmpDir(), config.ProfilingPrefix+"heaphistogram-PID-1.txt"), result)
			},
		},
		{
			name: "should return Flat",
			given: func() args {
				return args{
					targetDir: TmpDir(),
					job: &job.ProfilingJob{
						Tool:       api.AsyncProfiler,
						OutputType: api.Flat,
						Iteration:  1,
					},
				}
			},
			when: func(args args) string {
				return GetResultFile(args.targetDir, args.job.Tool, args.job.OutputType, "PID", args.job.Iteration)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, filepath.Join(TmpDir(), config.ProfilingPrefix+"flat-PID-1.txt"), result)
			},
		},
		{
			name: "should return Traces",
			given: func() args {
				return args{
					targetDir: TmpDir(),
					job: &job.ProfilingJob{
						Tool:       api.AsyncProfiler,
						OutputType: api.Traces,
						Iteration:  1,
					},
				}
			},
			when: func(args args) string {
				return GetResultFile(args.targetDir, args.job.Tool, args.job.OutputType, "PID", args.job.Iteration)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, filepath.Join(TmpDir(), config.ProfilingPrefix+"traces-PID-1.txt"), result)
			},
		},
		{
			name: "should return Collapsed",
			given: func() args {
				return args{
					targetDir: TmpDir(),
					job: &job.ProfilingJob{
						Tool:       api.AsyncProfiler,
						OutputType: api.Collapsed,
						Iteration:  1,
					},
				}
			},
			when: func(args args) string {
				return GetResultFile(args.targetDir, args.job.Tool, args.job.OutputType, "PID", args.job.Iteration)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, filepath.Join(TmpDir(), config.ProfilingPrefix+"collapsed-PID-1.txt"), result)
			},
		},
		{
			name: "should return Collapsed",
			given: func() args {
				return args{
					targetDir: TmpDir(),
					job: &job.ProfilingJob{
						Tool:       api.AsyncProfiler,
						OutputType: api.Raw,
						Iteration:  1,
					},
				}
			},
			when: func(args args) string {
				return GetResultFile(args.targetDir, args.job.Tool, args.job.OutputType, "PID", args.job.Iteration)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, filepath.Join(TmpDir(), config.ProfilingPrefix+"raw-PID-1.txt"), result)
			},
		},
		{
			name: "should return Tree",
			given: func() args {
				return args{
					targetDir: TmpDir(),
					job: &job.ProfilingJob{
						Tool:       api.AsyncProfiler,
						OutputType: api.Tree,
						Iteration:  1,
					},
				}
			},
			when: func(args args) string {
				return GetResultFile(args.targetDir, args.job.Tool, args.job.OutputType, "PID", args.job.Iteration)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, filepath.Join(TmpDir(), config.ProfilingPrefix+"tree-PID-1.html"), result)
			},
		},
		{
			name: "should return FlameGraph",
			given: func() args {
				return args{
					targetDir: TmpDir(),
					job: &job.ProfilingJob{
						Tool:       api.AsyncProfiler,
						OutputType: api.FlameGraph,
						Iteration:  1,
					},
				}
			},
			when: func(args args) string {
				return GetResultFile(args.targetDir, args.job.Tool, args.job.OutputType, "PID", args.job.Iteration)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, filepath.Join(TmpDir(), config.ProfilingPrefix+"flamegraph-PID-1.html"), result)
			},
		},
		{
			name: "should return SpeedScope",
			given: func() args {
				return args{
					targetDir: TmpDir(),
					job: &job.ProfilingJob{
						Tool:       api.Pyspy,
						OutputType: api.SpeedScope,
						Iteration:  1,
					},
				}
			},
			when: func(args args) string {
				return GetResultFile(args.targetDir, args.job.Tool, args.job.OutputType, "PID", args.job.Iteration)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, filepath.Join(TmpDir(), config.ProfilingPrefix+"speedscope-PID-1.json"), result)
			},
		},
		{
			name: "should return ThreadDump when PySpy",
			given: func() args {
				return args{
					targetDir: TmpDir(),
					job: &job.ProfilingJob{
						Tool:       api.Pyspy,
						OutputType: api.ThreadDump,
						Iteration:  1,
					},
				}
			},
			when: func(args args) string {
				return GetResultFile(args.targetDir, args.job.Tool, args.job.OutputType, "PID", args.job.Iteration)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, filepath.Join(TmpDir(), config.ProfilingPrefix+"threaddump-PID-1.txt"), result)
			},
		},
		{
			name: "should return FlameGraph when PySpy",
			given: func() args {
				return args{
					targetDir: TmpDir(),
					job: &job.ProfilingJob{
						Tool:       api.Pyspy,
						OutputType: api.FlameGraph,
						Iteration:  1,
					},
				}
			},
			when: func(args args) string {
				return GetResultFile(args.targetDir, args.job.Tool, args.job.OutputType, "PID", args.job.Iteration)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, filepath.Join(TmpDir(), config.ProfilingPrefix+"flamegraph-PID-1.svg"), result)
			},
		},
		{
			name: "should return FlameGraph when default",
			given: func() args {
				return args{
					targetDir: TmpDir(),
					job: &job.ProfilingJob{
						Tool:       api.Bpf,
						OutputType: api.FlameGraph,
						Iteration:  1,
					},
				}
			},
			when: func(args args) string {
				return GetResultFile(args.targetDir, args.job.Tool, args.job.OutputType, "PID", args.job.Iteration)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, filepath.Join(TmpDir(), config.ProfilingPrefix+"flamegraph-PID-1.svg"), result)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			args := tt.given()

			// When
			result := tt.when(args)

			// Then
			tt.then(t, result)

		})
	}
}

func TestGetFileExtension(t *testing.T) {
	type args struct {
		tool       api.ProfilingTool
		OutputType api.OutputType
	}
	tests := []struct {
		name  string
		given func() args
		when  func(args) string
		then  func(t *testing.T, result string)
	}{
		{
			name: "with jfr output type",
			given: func() args {
				return args{
					tool:       api.Jcmd,
					OutputType: api.Jfr,
				}
			},
			when: func(args args) string {
				return GetFileExtension(args.tool, args.OutputType)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, ".jfr", result)
			},
		},
		{
			name: "with ThreadDump output type",
			given: func() args {
				return args{
					tool:       api.Jcmd,
					OutputType: api.ThreadDump,
				}
			},
			when: func(args args) string {
				return GetFileExtension(args.tool, args.OutputType)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, ".txt", result)
			},
		},
		{
			name: "with HeapDump output type",
			given: func() args {
				return args{
					tool:       api.Jcmd,
					OutputType: api.HeapDump,
				}
			},
			when: func(args args) string {
				return GetFileExtension(args.tool, args.OutputType)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, ".hprof", result)
			},
		},
		{
			name: "with HeapHistogram output type",
			given: func() args {
				return args{
					tool:       api.Jcmd,
					OutputType: api.HeapHistogram,
				}
			},
			when: func(args args) string {
				return GetFileExtension(args.tool, args.OutputType)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, ".txt", result)
			},
		},
		{
			name: "with Flat output type",
			given: func() args {
				return args{
					tool:       api.AsyncProfiler,
					OutputType: api.Flat,
				}
			},
			when: func(args args) string {
				return GetFileExtension(args.tool, args.OutputType)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, ".txt", result)
			},
		},
		{
			name: "with Traces output type",
			given: func() args {
				return args{
					tool:       api.AsyncProfiler,
					OutputType: api.Traces,
				}
			},
			when: func(args args) string {
				return GetFileExtension(args.tool, args.OutputType)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, ".txt", result)
			},
		},
		{
			name: "with Collapsed output type",
			given: func() args {
				return args{
					tool:       api.AsyncProfiler,
					OutputType: api.Collapsed,
				}
			},
			when: func(args args) string {
				return GetFileExtension(args.tool, args.OutputType)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, ".txt", result)
			},
		},
		{
			name: "with Tree output type",
			given: func() args {
				return args{
					tool:       api.AsyncProfiler,
					OutputType: api.Tree,
				}
			},
			when: func(args args) string {
				return GetFileExtension(args.tool, args.OutputType)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, ".html", result)
			},
		},
		{
			name: "with FlameGraph output type, when async profiler",
			given: func() args {
				return args{
					tool:       api.AsyncProfiler,
					OutputType: api.FlameGraph,
				}
			},
			when: func(args args) string {
				return GetFileExtension(args.tool, args.OutputType)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, ".html", result)
			},
		},
		{
			name: "with SpeedScope output type",
			given: func() args {
				return args{
					tool:       api.Pyspy,
					OutputType: api.SpeedScope,
				}
			},
			when: func(args args) string {
				return GetFileExtension(args.tool, args.OutputType)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, ".json", result)
			},
		},
		{
			name: "with ThreadDump output type when PySpy",
			given: func() args {
				return args{
					tool:       api.Pyspy,
					OutputType: api.ThreadDump,
				}
			},
			when: func(args args) string {
				return GetFileExtension(args.tool, args.OutputType)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, ".txt", result)
			},
		},
		{
			name: "with FlameGraph when PySpy",
			given: func() args {
				return args{
					tool:       api.Pyspy,
					OutputType: api.FlameGraph,
				}
			},
			when: func(args args) string {
				return GetFileExtension(args.tool, args.OutputType)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, ".svg", result)
			},
		},
		{
			name: "with FlameGraph when default",
			given: func() args {
				return args{
					tool:       api.Bpf,
					OutputType: api.FlameGraph,
				}
			},
			when: func(args args) string {
				return GetFileExtension(args.tool, args.OutputType)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, ".svg", result)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			args := tt.given()

			// When
			result := tt.when(args)

			// Then
			tt.then(t, result)

		})
	}
}
