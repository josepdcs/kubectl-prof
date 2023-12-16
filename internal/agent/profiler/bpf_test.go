package profiler

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/flamegraph"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

type MockBpfManager interface {
	BpfManager
	HandleProfilingResultInvokedTimes() int
	PublishResultInvokedTimes() int
	CleanUpInvokedTimes() int
	WithHandleProfilingResultError() MockBpfManager
}

type mockBpfManager struct {
	handleProfilingResultInvokedTimes int
	publishResultInvokedTimes         int
	cleanUpInvokedTimes               int
	withHandleProfilingResultError    bool
}

// NewMockBpfManager instances an empty MockBpfManager util for unit tests
func NewMockBpfManager() MockBpfManager {
	return &mockBpfManager{}
}

func (m *mockBpfManager) handleProfilingResult(*job.ProfilingJob, flamegraph.FrameGrapher, string, bytes.Buffer) error {
	m.handleProfilingResultInvokedTimes++
	if m.withHandleProfilingResultError {
		return errors.New("fake handleProfilingResult with error")
	}
	fmt.Println("fake handleProfilingResult")
	return nil
}

func (m *mockBpfManager) publishResult(compressor.Type, string, api.OutputType) error {
	m.publishResultInvokedTimes++
	fmt.Println("fake publish result")
	return nil
}

func (m *mockBpfManager) cleanUp(*exec.Cmd) {
	m.cleanUpInvokedTimes++
	fmt.Println("fake cleanUp")
}

func (m *mockBpfManager) HandleProfilingResultInvokedTimes() int {
	return m.handleProfilingResultInvokedTimes
}

func (m *mockBpfManager) PublishResultInvokedTimes() int {
	return m.publishResultInvokedTimes
}

func (m *mockBpfManager) CleanUpInvokedTimes() int {
	return m.cleanUpInvokedTimes
}

func (m *mockBpfManager) WithHandleProfilingResultError() MockBpfManager {
	m.withHandleProfilingResultError = true
	return m
}

func TestBpfProfiler_SetUp(t *testing.T) {
	type fields struct {
		BpfProfiler *BpfProfiler
	}
	type args struct {
		job *job.ProfilingJob
	}
	tests := []struct {
		name  string
		given func() (fields, args)
		when  func(fields, args) error
		then  func(t *testing.T, err error, fields fields)
	}{
		{
			name: "should setup",
			given: func() (fields, args) {
				return fields{
						BpfProfiler: &BpfProfiler{
							BpfManager: NewMockBpfManager(),
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
						},
					}
			},
			when: func(fields fields, args args) error {
				return fields.BpfProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
				assert.Equal(t, "PID_ContainerID", fields.BpfProfiler.targetPID)
			},
		},
		{
			name: "should setup when PID is provided",
			given: func() (fields, args) {
				return fields{
						BpfProfiler: &BpfProfiler{
							BpfManager: NewMockBpfManager(),
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							PID:              "PID_ContainerID",
						},
					}
			},
			when: func(fields fields, args args) error {
				return fields.BpfProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
				assert.Equal(t, "PID_ContainerID", fields.BpfProfiler.targetPID)
			},
		},
		{
			name: "should fail when container PID not found",
			given: func() (fields, args) {
				return fields{
						BpfProfiler: &BpfProfiler{
							BpfManager: NewMockBpfManager(),
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainerWithPIDResultError,
							ContainerID:      "ContainerID",
						},
					}
			},
			when: func(fields fields, args args) error {
				return fields.BpfProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.NotNil(t, err)
				assert.Empty(t, fields.BpfProfiler.targetPID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			fields, args := tt.given()

			// When
			err := tt.when(fields, args)

			// Then
			tt.then(t, err, fields)
		})
	}
}

func TestBpfProfiler_Invoke(t *testing.T) {
	type fields struct {
		BpfProfiler *BpfProfiler
	}
	type args struct {
		job *job.ProfilingJob
	}
	tests := []struct {
		name  string
		given func() (fields, args)
		when  func(fields, args) (error, time.Duration)
		then  func(t *testing.T, err error, fields fields)
	}{
		{
			name: "should publish result",
			given: func() (fields, args) {
				bccProfilerCommand = func(job *job.ProfilingJob, pid string) *exec.Cmd {
					return exec.Command("ls", "/tmp")
				}
				return fields{
						BpfProfiler: &BpfProfiler{
							BpfManager: NewMockBpfManager(),
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
						},
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.BpfProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.BpfProfiler.BpfManager.(MockBpfManager)
				assert.Nil(t, err)
				assert.Equal(t, 1, mock.HandleProfilingResultInvokedTimes())
				assert.Equal(t, 1, mock.PublishResultInvokedTimes())
			},
		},
		{
			name: "should publish result when ThreadDump output type",
			given: func() (fields, args) {
				bccProfilerCommand = func(job *job.ProfilingJob, pid string) *exec.Cmd {
					return exec.Command("ls", "/tmp")
				}
				return fields{
						BpfProfiler: &BpfProfiler{
							BpfManager: NewMockBpfManager(),
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         time.Duration(1) * time.Second,
							Interval:         time.Duration(1) * time.Second,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.ThreadDump,
						},
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.BpfProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.BpfProfiler.BpfManager.(MockBpfManager)
				assert.Nil(t, err)
				assert.Equal(t, 1, mock.HandleProfilingResultInvokedTimes())
				assert.Equal(t, 1, mock.PublishResultInvokedTimes())
			},
		},
		{
			name: "should fail when fail exec command",
			given: func() (fields, args) {
				bccProfilerCommand = func(job *job.ProfilingJob, pid string) *exec.Cmd {
					return &exec.Cmd{}
				}
				return fields{
						BpfProfiler: &BpfProfiler{
							BpfManager: NewMockBpfManager(),
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
						},
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.BpfProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.BpfProfiler.BpfManager.(MockBpfManager)
				assert.NotNil(t, err)
				assert.Equal(t, 0, mock.HandleProfilingResultInvokedTimes())
				assert.Equal(t, 0, mock.PublishResultInvokedTimes())
			},
		},
		{
			name: "should fail when handle profiling result fail",
			given: func() (fields, args) {
				bccProfilerCommand = func(job *job.ProfilingJob, pid string) *exec.Cmd {
					return exec.Command("ls", "/tmp")
				}
				return fields{
						BpfProfiler: &BpfProfiler{
							BpfManager: NewMockBpfManager().WithHandleProfilingResultError(),
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
						},
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.BpfProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.BpfProfiler.BpfManager.(MockBpfManager)
				assert.NotNil(t, err)
				assert.Equal(t, 0, mock.PublishResultInvokedTimes())
			},
		},
		{
			name: "should fail when profile fail for ThreadDump output type",
			given: func() (fields, args) {
				bccProfilerCommand = func(job *job.ProfilingJob, pid string) *exec.Cmd {
					return exec.Command("ls", "/tmp")
				}
				return fields{
						BpfProfiler: &BpfProfiler{
							BpfManager: NewMockBpfManager().WithHandleProfilingResultError(),
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         time.Duration(2) * time.Second,
							Interval:         time.Duration(1) * time.Second,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.ThreadDump,
						},
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.BpfProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.BpfProfiler.BpfManager.(MockBpfManager)
				assert.NotNil(t, err)
				assert.Equal(t, 0, mock.PublishResultInvokedTimes())
			},
		},
		{
			name: "should fail when profile fail for ThreadDump output type",
			given: func() (fields, args) {
				bccProfilerCommand = func(job *job.ProfilingJob, pid string) *exec.Cmd {
					return exec.Command("ls", "/tmp")
				}
				return fields{
						BpfProfiler: &BpfProfiler{
							BpfManager: NewMockBpfManager().WithHandleProfilingResultError(),
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         time.Duration(2) * time.Second,
							Interval:         time.Duration(1) * time.Second,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.ThreadDump,
						},
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.BpfProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.BpfProfiler.BpfManager.(MockBpfManager)
				assert.NotNil(t, err)
				assert.Equal(t, 0, mock.PublishResultInvokedTimes())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			fields, args := tt.given()

			// When
			err, _ := tt.when(fields, args)

			// Then
			tt.then(t, err, fields)
		})
	}
}

func TestBpfProfiler_CleanUp(t *testing.T) {
	type fields struct {
		BpfProfiler *BpfProfiler
	}
	type args struct {
		job *job.ProfilingJob
	}
	tests := []struct {
		name  string
		given func() (fields, args)
		when  func(fields, args) error
		then  func(t *testing.T, err error, fields fields)
	}{
		{
			name: "should cleanup",
			given: func() (fields, args) {
				f := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.svg")
				_, _ = os.Create(f)
				_, _ = os.Create(f + compressor.GetExtensionFileByCompressor[compressor.Gzip])
				return fields{
						BpfProfiler: &BpfProfiler{
							BpfManager: NewMockBpfManager(),
							cmd:        exec.Command("ls", "/tmp"),
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
						},
					}
			},
			when: func(fields fields, args args) error {
				return fields.BpfProfiler.CleanUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.BpfProfiler.BpfManager.(MockBpfManager)
				f := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.svg")
				g := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.svg"+
					compressor.GetExtensionFileByCompressor[compressor.Gzip])
				assert.False(t, file.Exists(f))
				assert.False(t, file.Exists(g))
				assert.Nil(t, err)
				assert.Equal(t, 1, mock.CleanUpInvokedTimes())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			fields, args := tt.given()

			// When
			err := tt.when(fields, args)

			// Then
			tt.then(t, err, fields)
		})
	}
}

func Test_bpfManager_handleProfilingResult(t *testing.T) {
	type fields struct {
		BpfProfiler *BpfProfiler
	}
	type args struct {
		job          *job.ProfilingJob
		flameGrapher flamegraph.FrameGrapher
		fileName     string
		out          bytes.Buffer
	}
	tests := []struct {
		name  string
		given func() (fields, args)
		when  func(fields, args) error
		then  func(t *testing.T, err error, fields fields)
		after func()
	}{
		{
			name: "should handle thread dump profiler result",
			given: func() (fields, args) {
				var b bytes.Buffer
				b.Write([]byte("test"))
				return fields{
						BpfProfiler: NewBpfProfiler(),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.ThreadDump,
						},
						fileName: filepath.Join(common.TmpDir(), config.ProfilingPrefix+"threaddump.txt"),
						out:      b,
					}
			},
			when: func(fields fields, args args) error {
				return fields.BpfProfiler.handleProfilingResult(args.job, args.flameGrapher, args.fileName, args.out)
			},
			then: func(t *testing.T, err error, fields fields) {
				b, err := os.ReadFile(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"threaddump.txt"))
				assert.Equal(t, "test", string(b))
				assert.Nil(t, err)
			},
			after: func() {
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"threaddump.txt"))
			},
		},
		{
			name: "should fail when unable write for thread dump profiler result",
			given: func() (fields, args) {
				var b bytes.Buffer
				b.Write([]byte("test"))
				return fields{
						BpfProfiler: NewBpfProfiler(),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.ThreadDump,
						},
						fileName: filepath.Join("/", config.ProfilingPrefix+"threaddump.txt"),
						out:      b,
					}
			},
			when: func(fields fields, args args) error {
				return fields.BpfProfiler.handleProfilingResult(args.job, args.flameGrapher, args.fileName, args.out)
			},
			then: func(t *testing.T, err error, fields fields) {
				require.Error(t, err)
			},
		},
		{
			name: "should handle flamegraph profiler result",
			given: func() (fields, args) {
				var b bytes.Buffer
				b.Write([]byte("test"))
				return fields{
						BpfProfiler: NewBpfProfiler(),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.FakeLang,
						},
						flameGrapher: flamegraph.NewFlameGrapherFake(),
						fileName:     filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt"),
						out:          b,
					}
			},
			when: func(fields fields, args args) error {
				return fields.BpfProfiler.handleProfilingResult(args.job, args.flameGrapher, args.fileName, args.out)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
			},
		},
		{
			name: "should fail handle flamegraph profiler result",
			given: func() (fields, args) {
				var b bytes.Buffer
				b.Write([]byte("test"))
				return fields{
						BpfProfiler: NewBpfProfiler(),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         "other",
						},
						flameGrapher: flamegraph.NewFlameGrapherFakeWithError(),
						fileName:     filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt"),
						out:          b,
					}
			},
			when: func(fields fields, args args) error {
				return fields.BpfProfiler.handleProfilingResult(args.job, args.flameGrapher, args.fileName, args.out)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.EqualError(t, err, "could not convert raw format to flamegraph: StackSamplesToFlameGraph with error")
			},
		},
		{
			name: "should fail handle flamegraph profiler result when no stacks found",
			given: func() (fields, args) {
				return fields{
						BpfProfiler: NewBpfProfiler(),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         "other",
						},
						flameGrapher: flamegraph.NewFlameGrapherFake(),
						fileName:     filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt"),
					}
			},
			when: func(fields fields, args args) error {
				return fields.BpfProfiler.handleProfilingResult(args.job, args.flameGrapher, args.fileName, args.out)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.EqualError(t, err, "unable to generate flamegraph: no stacks found (maybe due low cpu load)")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			fields, args := tt.given()

			// When
			err := tt.when(fields, args)

			// Then
			tt.then(t, err, fields)

			if tt.after != nil {
				tt.after()
			}
		})
	}
}

func Test_bpfManager_bccProfilerCommand(t *testing.T) {
	// Given
	j := &job.ProfilingJob{
		Interval: 10 * time.Second,
	}

	// When
	result := bccProfilerCommand(j, "1000")

	// Then
	assert.NotNil(t, result)
}
