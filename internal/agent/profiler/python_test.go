package profiler

import (
	"bytes"
	"fmt"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
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

type MockPythonManager interface {
	PythonManager
	HandleProfilingResultInvokedTimes() int
	PublishResultInvokedTimes() int
	CleanUpInvokedTimes() int
	WithHandleProfilingResultError() MockPythonManager
}

type mockPythonManager struct {
	handleProfilingResultInvokedTimes int
	publishResultInvokedTimes         int
	cleanUpInvokedTimes               int
	withHandleProfilingResultError    bool
}

// NewMockPythonManager instances an empty MockPythonManager util for unit tests
func NewMockPythonManager() MockPythonManager {
	return &mockPythonManager{}
}

func (m *mockPythonManager) handleProfilingResult(*job.ProfilingJob, string, bytes.Buffer) error {
	m.handleProfilingResultInvokedTimes++
	if m.withHandleProfilingResultError {
		return fmt.Errorf("fake handleProfilingResult with error")
	}
	fmt.Println("fake handleProfilingResult")
	return nil
}

func (m *mockPythonManager) publishResult(compressor.Type, string, api.OutputType) error {
	m.publishResultInvokedTimes++
	fmt.Println("fake publish result")
	return nil
}

func (m *mockPythonManager) cleanUp(*exec.Cmd) {
	m.cleanUpInvokedTimes++
	fmt.Println("fake cleanUp")
}

func (m *mockPythonManager) HandleProfilingResultInvokedTimes() int {
	return m.handleProfilingResultInvokedTimes
}

func (m *mockPythonManager) PublishResultInvokedTimes() int {
	return m.publishResultInvokedTimes
}

func (m *mockPythonManager) CleanUpInvokedTimes() int {
	return m.cleanUpInvokedTimes
}

func (m *mockPythonManager) WithHandleProfilingResultError() MockPythonManager {
	m.withHandleProfilingResultError = true
	return m
}

func TestPythonProfiler_SetUp(t *testing.T) {
	type fields struct {
		PythonProfiler *PythonProfiler
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
						PythonProfiler: &PythonProfiler{
							PythonManager: NewMockPythonManager(),
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
				return fields.PythonProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				//mock := fields.PythonProfiler.PythonManager.(MockPythonManager)
				assert.Nil(t, err)
				assert.Equal(t, "PID_ContainerID", fields.PythonProfiler.targetPID)
			},
		},
		{
			name: "should fail when container PID not found",
			given: func() (fields, args) {
				return fields{
						PythonProfiler: &PythonProfiler{
							PythonManager: NewMockPythonManager(),
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
				return fields.PythonProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.NotNil(t, err)
				assert.Empty(t, fields.PythonProfiler.targetPID)
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

func TestPythonProfiler_Invoke(t *testing.T) {
	type fields struct {
		PythonProfiler *PythonProfiler
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
				pythonCommand = func(job *job.ProfilingJob, pid string, fileName string) *exec.Cmd {
					return exec.Command("ls", "/tmp")
				}
				return fields{
						PythonProfiler: &PythonProfiler{
							PythonManager: NewMockPythonManager(),
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
				return fields.PythonProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.PythonProfiler.PythonManager.(MockPythonManager)
				assert.Nil(t, err)
				assert.Equal(t, 1, mock.HandleProfilingResultInvokedTimes())
				assert.Equal(t, 1, mock.PublishResultInvokedTimes())
			},
		},
		{
			name: "should publish result when ThreadDump output type",
			given: func() (fields, args) {
				pythonCommand = func(job *job.ProfilingJob, pid string, fileName string) *exec.Cmd {
					return exec.Command("ls", "/tmp")
				}
				return fields{
						PythonProfiler: &PythonProfiler{
							PythonManager: NewMockPythonManager(),
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
				return fields.PythonProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.PythonProfiler.PythonManager.(MockPythonManager)
				assert.Nil(t, err)
				assert.Equal(t, 1, mock.HandleProfilingResultInvokedTimes())
				assert.Equal(t, 1, mock.PublishResultInvokedTimes())
			},
		},
		{
			name: "should fail when fail exec command",
			given: func() (fields, args) {
				pythonCommand = func(job *job.ProfilingJob, pid string, fileName string) *exec.Cmd {
					return &exec.Cmd{}
				}
				return fields{
						PythonProfiler: &PythonProfiler{
							PythonManager: NewMockPythonManager(),
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
				return fields.PythonProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.PythonProfiler.PythonManager.(MockPythonManager)
				assert.NotNil(t, err)
				assert.Equal(t, 0, mock.HandleProfilingResultInvokedTimes())
				assert.Equal(t, 0, mock.PublishResultInvokedTimes())
			},
		},
		{
			name: "should fail when handle profiling result fail",
			given: func() (fields, args) {
				pythonCommand = func(job *job.ProfilingJob, pid string, fileName string) *exec.Cmd {
					return exec.Command("ls", "/tmp")
				}
				return fields{
						PythonProfiler: &PythonProfiler{
							PythonManager: NewMockPythonManager().WithHandleProfilingResultError(),
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
				return fields.PythonProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.PythonProfiler.PythonManager.(MockPythonManager)
				assert.NotNil(t, err)
				assert.Equal(t, 0, mock.PublishResultInvokedTimes())
			},
		},
		{
			name: "should fail when profile fail for ThreadDump output type",
			given: func() (fields, args) {
				pythonCommand = func(job *job.ProfilingJob, pid string, fileName string) *exec.Cmd {
					return exec.Command("ls", "/tmp")
				}
				return fields{
						PythonProfiler: &PythonProfiler{
							PythonManager: NewMockPythonManager().WithHandleProfilingResultError(),
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
				return fields.PythonProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.PythonProfiler.PythonManager.(MockPythonManager)
				assert.NotNil(t, err)
				assert.Equal(t, 0, mock.PublishResultInvokedTimes())
			},
		},
		{
			name: "should fail when profile fail for ThreadDump output type",
			given: func() (fields, args) {
				pythonCommand = func(job *job.ProfilingJob, pid string, fileName string) *exec.Cmd {
					return exec.Command("ls", "/tmp")
				}
				return fields{
						PythonProfiler: &PythonProfiler{
							PythonManager: NewMockPythonManager().WithHandleProfilingResultError(),
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
				return fields.PythonProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.PythonProfiler.PythonManager.(MockPythonManager)
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

func TestPythonProfiler_CleanUp(t *testing.T) {
	type fields struct {
		PythonProfiler *PythonProfiler
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
						PythonProfiler: &PythonProfiler{
							PythonManager: NewMockPythonManager(),
							cmd:           exec.Command("ls", "/tmp"),
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
				return fields.PythonProfiler.CleanUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.PythonProfiler.PythonManager.(MockPythonManager)
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

func Test_pythonManager_handleProfilingResult(t *testing.T) {
	type fields struct {
		PythonProfiler *PythonProfiler
	}
	type args struct {
		job      *job.ProfilingJob
		fileName string
		out      bytes.Buffer
	}
	tests := []struct {
		name  string
		given func() (fields, args)
		when  func(fields, args) error
		then  func(t *testing.T, err error, fields fields)
		after func()
	}{
		{
			name: "should handle profiling result",
			given: func() (fields, args) {
				var b bytes.Buffer
				b.Write([]byte("test"))
				return fields{
						PythonProfiler: NewPythonProfiler(),
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
				return fields.PythonProfiler.handleProfilingResult(args.job, args.fileName, args.out)
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
			name: "should fail when unable write",
			given: func() (fields, args) {
				var b bytes.Buffer
				b.Write([]byte("test"))
				return fields{
						PythonProfiler: NewPythonProfiler(),
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
				return fields.PythonProfiler.handleProfilingResult(args.job, args.fileName, args.out)
			},
			then: func(t *testing.T, err error, fields fields) {
				require.Error(t, err)
			},
		},
		{
			name: "no thread dump",
			given: func() (fields, args) {
				return fields{
						PythonProfiler: NewPythonProfiler(),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
						},
						fileName: filepath.Join(common.TmpDir(), config.ProfilingPrefix+"threaddump.txt"),
					}
			},
			when: func(fields fields, args args) error {
				return fields.PythonProfiler.handleProfilingResult(args.job, args.fileName, args.out)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Equal(t, false, file.Exists(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"threaddump.txt")))
				assert.Nil(t, err)
			},
			after: func() {
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"threaddump.txt"))
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
