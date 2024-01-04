package profiler

import (
	"bytes"
	"fmt"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	executil "github.com/josepdcs/kubectl-prof/internal/agent/util/exec"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/flamegraph"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/publish"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
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
	HandleFlamegraphInvokedTimes() int
	PublishResultInvokedTimes() int
	WithHandleFlamegraphError() MockPythonManager
	WithInvokeError() MockPythonManager
}

type mockPythonManager struct {
	invokeInvokedTimes           int
	handleFlamegraphInvokedTimes int
	publishResultInvokedTimes    int
	withHandleFlamegraphError    bool
	withInvokeError              bool
}

// NewMockPythonManager instances an empty MockPythonManager util for unit tests
func NewMockPythonManager() MockPythonManager {
	return &mockPythonManager{}
}

func (m *mockPythonManager) invoke(*job.ProfilingJob, string) (error, time.Duration) {
	m.invokeInvokedTimes++
	if m.withInvokeError {
		return errors.New("fake invoke with error"), 0
	}
	fmt.Println("fake invoke")
	return nil, 0
}

func (m *mockPythonManager) handleFlamegraph(*job.ProfilingJob, flamegraph.FrameGrapher, string, string) error {
	m.handleFlamegraphInvokedTimes++
	if m.withHandleFlamegraphError {
		return errors.New("fake handleFlamegraph with error")
	}
	fmt.Println("fake handleFlamegraph")
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

func (m *mockPythonManager) HandleFlamegraphInvokedTimes() int {
	return m.handleFlamegraphInvokedTimes
}

func (m *mockPythonManager) PublishResultInvokedTimes() int {
	return m.publishResultInvokedTimes
}

func (m *mockPythonManager) WithHandleFlamegraphError() MockPythonManager {
	m.withHandleFlamegraphError = true
	return m
}

func (m *mockPythonManager) WithInvokeError() MockPythonManager {
	m.withInvokeError = true
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
			name: "should invoke",
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
			},
		},
		{
			name: "should invoke fail when invoke fail",
			given: func() (fields, args) {
				return fields{
						PythonProfiler: &PythonProfiler{
							PythonManager: NewMockPythonManager().WithInvokeError(),
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
				require.Error(t, err)
				assert.Equal(t, 1, mock.InvokeInvokedTimes())
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
		job *job.ProfilingJob
		pid string
	}
	tests := []struct {
		name  string
		given func() (fields, args)
		when  func(fields, args) (error, time.Duration)
		then  func(t *testing.T, err error)
		after func()
	}{
		{
			name: "should invoke",
			given: func() (fields, args) {
				log.SetPrintLogs(true)
				var b bytes.Buffer
				b.Write([]byte("test"))
				file.Write(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw-1000.txt"), b.String())
				file.Write(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000.svg"), b.String())

				return fields{
						PythonProfiler: NewPythonProfiler(executil.NewFakeCommander(exec.Command("ls", "/tmp")),
							publish.NewPublisherFake(nil)),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.FakeLang,
							Tool:             api.Pyspy,
							Compressor:       compressor.None,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.PythonProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, err error) {
				assert.Nil(t, err)
				assert.True(t, file.Exists(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000.svg")))
			},
			after: func() {
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw-1000.txt"))
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000.svg"))
			},
		},
		{
			name: "should invoke when thread dump",
			given: func() (fields, args) {
				return fields{
						PythonProfiler: NewPythonProfiler(executil.NewFakeCommander(exec.Command("ls", "/tmp")),
							publish.NewPublisherFake(nil)),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.ThreadDump,
							Language:         api.FakeLang,
							Compressor:       compressor.None,
							Tool:             api.Pyspy,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.PythonProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, err error) {
				assert.Nil(t, err)
				assert.True(t, file.Exists(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"threaddump-1000.txt")))
			},
			after: func() {
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"threaddump-1000.txt"))
			},
		},
		{
			name: "should invoke fail when command fail",
			given: func() (fields, args) {
				return fields{
						PythonProfiler: NewPythonProfiler(executil.NewFakeCommander(&exec.Cmd{}),
							publish.NewPublisherFake(nil)),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.Python,
							Tool:             api.Pyspy,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.PythonProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name: "should invoke return nil when fail handle flamegraph",
			given: func() (fields, args) {
				log.SetPrintLogs(true)
				return fields{
						PythonProfiler: NewPythonProfiler(executil.NewFakeCommander(exec.Command("ls", "/tmp")),
							publish.NewPublisherFake(nil)),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.Python,
							Tool:             api.Pyspy,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.PythonProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, err error) {
				require.NoError(t, err)
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
			tt.then(t, err)

			if tt.after != nil {
				tt.after()
			}
		})
	}
}

func Test_pythonManager_handleFlamegraph(t *testing.T) {
	type fields struct {
		PythonProfiler *PythonProfiler
	}
	type args struct {
		job            *job.ProfilingJob
		flameGrapher   flamegraph.FrameGrapher
		fileName       string
		resultFileName string
	}
	tests := []struct {
		name  string
		given func() (fields, args)
		when  func(fields, args) error
		then  func(t *testing.T, err error, flameGrapher flamegraph.FrameGrapher)
		after func()
	}{
		{
			name: "should handle flamegraph",
			given: func() (fields, args) {
				var b bytes.Buffer
				b.Write([]byte("test"))
				_ = os.WriteFile(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt"), b.Bytes(), 0644)
				return fields{
						PythonProfiler: NewPythonProfiler(executil.NewFakeCommander(exec.Command("ls", "/tmp")),
							publish.NewPublisherFake(nil)),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.FakeLang,
						},
						flameGrapher:   flamegraph.NewFlameGrapherFake(),
						fileName:       filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt"),
						resultFileName: filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.svg"),
					}
			},
			when: func(fields fields, args args) error {
				return fields.PythonProfiler.handleFlamegraph(args.job, args.flameGrapher, args.fileName, args.resultFileName)
			},
			then: func(t *testing.T, err error, flameGrapher flamegraph.FrameGrapher) {
				assert.True(t, flameGrapher.(*flamegraph.FlameGrapherFake).StackSamplesToFlameGraphInvoked)
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
						PythonProfiler: NewPythonProfiler(executil.NewFakeCommander(exec.Command("ls", "/tmp")),
							publish.NewPublisherFake(nil)),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         "other",
						},
						flameGrapher:   flamegraph.NewFlameGrapherFakeWithError(),
						fileName:       filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt"),
						resultFileName: filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.svg"),
					}
			},
			when: func(fields fields, args args) error {
				return fields.PythonProfiler.handleFlamegraph(args.job, args.flameGrapher, args.fileName, args.resultFileName)
			},
			then: func(t *testing.T, err error, flameGrapher flamegraph.FrameGrapher) {
				assert.True(t, flameGrapher.(*flamegraph.FlameGrapherFakeWithError).StackSamplesToFlameGraphInvoked)
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
						PythonProfiler: NewPythonProfiler(executil.NewFakeCommander(exec.Command("ls", "/tmp")),
							publish.NewPublisherFake(nil)),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         "other",
						},
						flameGrapher:   flamegraph.NewFlameGrapherFake(),
						fileName:       filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt"),
						resultFileName: filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.svg"),
					}
			},
			when: func(fields fields, args args) error {
				return fields.PythonProfiler.handleFlamegraph(args.job, args.flameGrapher, args.fileName, args.resultFileName)
			},
			then: func(t *testing.T, err error, flameGrapher flamegraph.FrameGrapher) {
				assert.False(t, flameGrapher.(*flamegraph.FlameGrapherFake).StackSamplesToFlameGraphInvoked)
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
			tt.then(t, err, args.flameGrapher)

			if tt.after != nil {
				tt.after()
			}
		})
	}
}
