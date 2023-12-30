package profiler

import (
	"bytes"
	"fmt"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	"github.com/josepdcs/kubectl-prof/internal/agent/testdata"
	executil "github.com/josepdcs/kubectl-prof/internal/agent/util/exec"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/flamegraph"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/pkg/errors"
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
	InvokeInvokedTimes() int
	HandleProfilingResultInvokedTimes() int
	PublishResultInvokedTimes() int
	WithHandleProfilingResultError() MockPythonManager
}

type mockPythonManager struct {
	invokeInvokedTimes                int
	handleProfilingResultInvokedTimes int
	publishResultInvokedTimes         int
	withHandleProfilingResultError    bool
}

// NewMockPythonManager instances an empty MockPythonManager util for unit tests
func NewMockPythonManager() MockPythonManager {
	return &mockPythonManager{}
}

func (m *mockPythonManager) invoke(*job.ProfilingJob, string, string) (error, string, time.Duration) {
	m.invokeInvokedTimes++
	fmt.Println("fake invoke")
	return nil, "", 0
}

func (m *mockPythonManager) handleProfilingResult(*job.ProfilingJob, flamegraph.FrameGrapher, string) error {
	m.handleProfilingResultInvokedTimes++
	if m.withHandleProfilingResultError {
		return errors.New("fake handleProfilingResult with error")
	}
	fmt.Println("fake handleProfilingResult")
	return nil
}

func (m *mockPythonManager) publishResult(compressor.Type, string, api.OutputType) error {
	m.publishResultInvokedTimes++
	fmt.Println("fake publish result")
	return nil
}

func (m *mockPythonManager) InvokeInvokedTimes() int {
	return m.invokeInvokedTimes
}

func (m *mockPythonManager) HandleProfilingResultInvokedTimes() int {
	return m.handleProfilingResultInvokedTimes
}

func (m *mockPythonManager) PublishResultInvokedTimes() int {
	return m.publishResultInvokedTimes
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
				assert.Nil(t, err)
				assert.Equal(t, []string{"PID_ContainerID"}, fields.PythonProfiler.targetPIDs)
			},
		},
		{
			name: "should setup with given PID",
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
							PID:              "PID_ContainerID",
						},
					}
			},
			when: func(fields fields, args args) error {
				return fields.PythonProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
				assert.Equal(t, []string{"PID_ContainerID"}, fields.PythonProfiler.targetPIDs)
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
				assert.Empty(t, fields.PythonProfiler.targetPIDs)
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
				return fields{
						PythonProfiler: &PythonProfiler{
							PythonManager: NewMockPythonManager(),
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
				fields.PythonProfiler.delay = 0
				fields.PythonProfiler.targetPIDs = []string{"1000", "2000"}
				return fields.PythonProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.PythonProfiler.PythonManager.(MockPythonManager)
				assert.Nil(t, err)
				assert.Equal(t, 2, mock.InvokeInvokedTimes())
				assert.Equal(t, 1, mock.HandleProfilingResultInvokedTimes())
				assert.Equal(t, 1, mock.PublishResultInvokedTimes())
			},
		},
		{
			name: "should fail when handle profiling result fail",
			given: func() (fields, args) {
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
				fields.PythonProfiler.delay = 0
				fields.PythonProfiler.targetPIDs = []string{"1000", "2000"}
				return fields.PythonProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.PythonProfiler.PythonManager.(MockPythonManager)
				assert.NotNil(t, err)
				assert.Equal(t, 2, mock.InvokeInvokedTimes())
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
				f := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.svg")
				g := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.svg"+
					compressor.GetExtensionFileByCompressor[compressor.Gzip])
				assert.False(t, file.Exists(f))
				assert.False(t, file.Exists(g))
				assert.Nil(t, err)
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

func Test_pythonManager_invoke(t *testing.T) {
	type fields struct {
		PythonProfiler *PythonProfiler
	}
	type args struct {
		job      *job.ProfilingJob
		pid      string
		fileName string
	}
	tests := []struct {
		name  string
		given func() (fields, args)
		when  func(fields, args) (error, string, time.Duration)
		then  func(t *testing.T, err error, fileName string)
		after func()
	}{
		{
			name: "should invoke",
			given: func() (fields, args) {
				pySpyCommander = executil.NewFakeCommander(exec.Command("ls", "/tmp"))
				return fields{
						PythonProfiler: NewPythonProfiler(),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.FakeLang,
						},
						pid:      "1000",
						fileName: filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt"),
					}
			},
			when: func(fields fields, args args) (error, string, time.Duration) {
				return fields.PythonProfiler.invoke(args.job, args.pid, args.fileName)
			},
			then: func(t *testing.T, err error, fileName string) {
				assert.Nil(t, err)
				assert.Equal(t, filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt.1000"), fileName)
			},
		},
		{
			name: "should invoke when thread dump",
			given: func() (fields, args) {
				pySpyCommander = executil.NewFakeCommander(exec.Command("ls", "/tmp"))
				return fields{
						PythonProfiler: NewPythonProfiler(),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.ThreadDump,
							Language:         api.FakeLang,
						},
						pid:      "1000",
						fileName: filepath.Join(common.TmpDir(), config.ProfilingPrefix+"threaddump.txt"),
					}
			},
			when: func(fields fields, args args) (error, string, time.Duration) {
				return fields.PythonProfiler.invoke(args.job, args.pid, args.fileName)
			},
			then: func(t *testing.T, err error, fileName string) {
				assert.Nil(t, err)
				assert.Equal(t, filepath.Join(common.TmpDir(), config.ProfilingPrefix+"threaddump.txt.1000"), fileName)
			},
		},
		{
			name: "should invoke fail when command fail",
			given: func() (fields, args) {
				pySpyCommander = executil.NewFakeCommander(&exec.Cmd{})
				return fields{
						PythonProfiler: NewPythonProfiler(),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.FakeLang,
						},
						pid:      "1000",
						fileName: filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt"),
					}
			},
			when: func(fields fields, args args) (error, string, time.Duration) {
				return fields.PythonProfiler.invoke(args.job, args.pid, args.fileName)
			},
			then: func(t *testing.T, err error, fileName string) {
				require.Error(t, err)
				assert.Equal(t, "", fileName)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			fields, args := tt.given()

			// When
			err, fileName, _ := tt.when(fields, args)

			// Then
			tt.then(t, err, fileName)

			if tt.after != nil {
				tt.after()
			}
		})
	}
}

func Test_pythonManager_handleProfilingResult(t *testing.T) {
	type fields struct {
		PythonProfiler *PythonProfiler
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
			name: "should handle flamegraph profiler result",
			given: func() (fields, args) {
				var b bytes.Buffer
				b.Write([]byte("test"))
				_ = os.WriteFile(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt"), b.Bytes(), 0644)
				return fields{
						PythonProfiler: NewPythonProfiler(),
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
				return fields.PythonProfiler.handleProfilingResult(args.job, args.flameGrapher, args.fileName)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
			},
			after: func() {
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt"))
			},
		},
		{
			name: "should fail handle flamegraph profiler result",
			given: func() (fields, args) {
				var b bytes.Buffer
				b.Write([]byte("test"))
				_ = os.WriteFile(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt"), b.Bytes(), 0644)
				return fields{
						PythonProfiler: NewPythonProfiler(),
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
					}
			},
			when: func(fields fields, args args) error {
				return fields.PythonProfiler.handleProfilingResult(args.job, args.flameGrapher, args.fileName)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.EqualError(t, err, "could not convert raw format to flamegraph: StackSamplesToFlameGraph with error")
			},
			after: func() {
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt"))
			},
		},
		{
			name: "should fail handle flamegraph profiler result when no stacks found",
			given: func() (fields, args) {
				var b bytes.Buffer
				_ = os.WriteFile(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt"), b.Bytes(), 0644)
				return fields{
						PythonProfiler: NewPythonProfiler(),
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
				return fields.PythonProfiler.handleProfilingResult(args.job, args.flameGrapher, args.fileName)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.EqualError(t, err, "unable to generate flamegraph: no stacks found (maybe due low cpu load)")
			},
			after: func() {
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt"))
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

func Test_pythonManager_publishResult(t *testing.T) {
	p := NewPythonProfiler()
	err := p.publishResult(compressor.Gzip, testdata.ResultTestDataDir()+"/flamegraph.svg", api.FlameGraph)
	assert.Nil(t, err)
}
