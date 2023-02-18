package common

import (
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func TestGetResultFile(t *testing.T) {
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
					},
				}
			},
			when: func(args args) string {
				return GetResultFile(args.targetDir, args.job)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, filepath.Join(TmpDir(), config.ProfilingPrefix+"jfr.jfr"), result)
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
					},
				}
			},
			when: func(args args) string {
				return GetResultFile(args.targetDir, args.job)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, filepath.Join(TmpDir(), config.ProfilingPrefix+"threaddump.txt"), result)
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
					},
				}
			},
			when: func(args args) string {
				return GetResultFile(args.targetDir, args.job)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, filepath.Join(TmpDir(), config.ProfilingPrefix+"heapdump.hprof"), result)
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
					},
				}
			},
			when: func(args args) string {
				return GetResultFile(args.targetDir, args.job)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, filepath.Join(TmpDir(), config.ProfilingPrefix+"heaphistogram.txt"), result)
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
					},
				}
			},
			when: func(args args) string {
				return GetResultFile(args.targetDir, args.job)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, filepath.Join(TmpDir(), config.ProfilingPrefix+"flat.txt"), result)
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
					},
				}
			},
			when: func(args args) string {
				return GetResultFile(args.targetDir, args.job)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, filepath.Join(TmpDir(), config.ProfilingPrefix+"traces.txt"), result)
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
					},
				}
			},
			when: func(args args) string {
				return GetResultFile(args.targetDir, args.job)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, filepath.Join(TmpDir(), config.ProfilingPrefix+"collapsed.txt"), result)
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
					},
				}
			},
			when: func(args args) string {
				return GetResultFile(args.targetDir, args.job)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, filepath.Join(TmpDir(), config.ProfilingPrefix+"tree.html"), result)
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
					},
				}
			},
			when: func(args args) string {
				return GetResultFile(args.targetDir, args.job)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, filepath.Join(TmpDir(), config.ProfilingPrefix+"flamegraph.html"), result)
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
					},
				}
			},
			when: func(args args) string {
				return GetResultFile(args.targetDir, args.job)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, filepath.Join(TmpDir(), config.ProfilingPrefix+"speedscope.json"), result)
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
					},
				}
			},
			when: func(args args) string {
				return GetResultFile(args.targetDir, args.job)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, filepath.Join(TmpDir(), config.ProfilingPrefix+"threaddump.txt"), result)
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
					},
				}
			},
			when: func(args args) string {
				return GetResultFile(args.targetDir, args.job)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, filepath.Join(TmpDir(), config.ProfilingPrefix+"flamegraph.svg"), result)
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
					},
				}
			},
			when: func(args args) string {
				return GetResultFile(args.targetDir, args.job)
			},
			then: func(t *testing.T, result string) {
				assert.Equal(t, filepath.Join(TmpDir(), config.ProfilingPrefix+"flamegraph.svg"), result)
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

	result := GetResultFile("/tmp", &job.ProfilingJob{
		Tool:       api.AsyncProfiler,
		OutputType: api.FlameGraph,
	})

	assert.Equal(t, "/tmp/agent-flamegraph.html", result)
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
