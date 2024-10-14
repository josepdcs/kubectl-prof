package jvm

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	executil "github.com/josepdcs/kubectl-prof/internal/agent/util/exec"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/publish"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsyncProfiler_SetUp(t *testing.T) {
	type fields struct {
		AsyncProfiler *AsyncProfiler
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
				asyncProfilerManager := newFakeAsyncProfilerManager()
				asyncProfilerManager.On("removeTmpDir").Return(nil)
				asyncProfilerManager.On("linkTmpDirToTargetTmpDir").Return(nil)
				asyncProfilerManager.On("copyProfilerToTmpDir").Return(nil)

				return fields{
						AsyncProfiler: &AsyncProfiler{
							AsyncProfilerManager: asyncProfilerManager,
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
				return fields.AsyncProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
				assert.Equal(t, []string{"PID_ContainerID"}, fields.AsyncProfiler.targetPIDs)
				assert.Equal(t, 1, fields.AsyncProfiler.AsyncProfilerManager.(FakeAsyncProfilerManager).On("removeTmpDir").InvokedTimes())
				assert.Equal(t, 1, fields.AsyncProfiler.AsyncProfilerManager.(FakeAsyncProfilerManager).On("linkTmpDirToTargetTmpDir").InvokedTimes())
				assert.Equal(t, 1, fields.AsyncProfiler.AsyncProfilerManager.(FakeAsyncProfilerManager).On("copyProfilerToTmpDir").InvokedTimes())
			},
		},
		{
			name: "should setup when PID is given",
			given: func() (fields, args) {
				asyncProfilerManager := newFakeAsyncProfilerManager()
				asyncProfilerManager.On("removeTmpDir").Return(nil)
				asyncProfilerManager.On("linkTmpDirToTargetTmpDir").Return(nil)
				asyncProfilerManager.On("copyProfilerToTmpDir").Return(nil)

				return fields{
						AsyncProfiler: &AsyncProfiler{
							AsyncProfilerManager: asyncProfilerManager,
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
				return fields.AsyncProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
				assert.Equal(t, []string{"PID_ContainerID"}, fields.AsyncProfiler.targetPIDs)
				assert.Equal(t, 1, fields.AsyncProfiler.AsyncProfilerManager.(FakeAsyncProfilerManager).On("removeTmpDir").InvokedTimes())
				assert.Equal(t, 1, fields.AsyncProfiler.AsyncProfilerManager.(FakeAsyncProfilerManager).On("linkTmpDirToTargetTmpDir").InvokedTimes())
				assert.Equal(t, 1, fields.AsyncProfiler.AsyncProfilerManager.(FakeAsyncProfilerManager).On("copyProfilerToTmpDir").InvokedTimes())
			},
		},
		{
			name: "should fail when getting target filesystem fail",
			given: func() (fields, args) {
				asyncProfilerManager := newFakeAsyncProfilerManager()
				asyncProfilerManager.On("removeTmpDir").Return(nil)
				asyncProfilerManager.On("linkTmpDirToTargetTmpDir").Return(nil)
				asyncProfilerManager.On("copyProfilerToTmpDir").Return(nil)

				return fields{
						AsyncProfiler: &AsyncProfiler{
							AsyncProfilerManager: asyncProfilerManager,
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: "other",
							ContainerID:      "ContainerID",
							PID:              "PID_ContainerID",
						},
					}
			},
			when: func(fields fields, args args) error {
				return fields.AsyncProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.NotNil(t, err)
				assert.Equal(t, 0, fields.AsyncProfiler.AsyncProfilerManager.(FakeAsyncProfilerManager).On("removeTmpDir").InvokedTimes())
				assert.Equal(t, 0, fields.AsyncProfiler.AsyncProfilerManager.(FakeAsyncProfilerManager).On("linkTmpDirToTargetTmpDir").InvokedTimes())
				assert.Equal(t, 0, fields.AsyncProfiler.AsyncProfilerManager.(FakeAsyncProfilerManager).On("copyProfilerToTmpDir").InvokedTimes())
			},
		},
		{
			name: "should fail when removing tmp dir fail",
			given: func() (fields, args) {
				asyncProfilerManager := newFakeAsyncProfilerManager()
				asyncProfilerManager.On("removeTmpDir").Return(errors.New("fake error"))

				return fields{
						AsyncProfiler: &AsyncProfiler{
							AsyncProfilerManager: asyncProfilerManager,
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
				return fields.AsyncProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.NotNil(t, err)
				assert.EqualError(t, err, "fake error")
				assert.Equal(t, 1, fields.AsyncProfiler.AsyncProfilerManager.(FakeAsyncProfilerManager).On("removeTmpDir").InvokedTimes())
				assert.Equal(t, 0, fields.AsyncProfiler.AsyncProfilerManager.(FakeAsyncProfilerManager).On("linkTmpDirToTargetTmpDir").InvokedTimes())
				assert.Equal(t, 0, fields.AsyncProfiler.AsyncProfilerManager.(FakeAsyncProfilerManager).On("copyProfilerToTmpDir").InvokedTimes())
			},
		},
		{
			name: "should fail when link tmp dir to target tmp dir fail",
			given: func() (fields, args) {
				asyncProfilerManager := newFakeAsyncProfilerManager()
				asyncProfilerManager.On("removeTmpDir").Return(nil)
				asyncProfilerManager.On("linkTmpDirToTargetTmpDir").Return(errors.New("fake error"))

				return fields{
						AsyncProfiler: &AsyncProfiler{
							AsyncProfilerManager: asyncProfilerManager,
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
				return fields.AsyncProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.NotNil(t, err)
				assert.EqualError(t, err, "fake error")
				assert.Equal(t, 1, fields.AsyncProfiler.AsyncProfilerManager.(FakeAsyncProfilerManager).On("removeTmpDir").InvokedTimes())
				assert.Equal(t, 1, fields.AsyncProfiler.AsyncProfilerManager.(FakeAsyncProfilerManager).On("linkTmpDirToTargetTmpDir").InvokedTimes())
				assert.Equal(t, 0, fields.AsyncProfiler.AsyncProfilerManager.(FakeAsyncProfilerManager).On("copyProfilerToTmpDir").InvokedTimes())
			},
		},
		{
			name: "should fail when container PID not found",
			given: func() (fields, args) {
				asyncProfilerManager := newFakeAsyncProfilerManager()
				asyncProfilerManager.On("removeTmpDir").Return(nil)
				asyncProfilerManager.On("linkTmpDirToTargetTmpDir").Return(nil)

				return fields{
						AsyncProfiler: &AsyncProfiler{
							AsyncProfilerManager: asyncProfilerManager,
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
				return fields.AsyncProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.NotNil(t, err)
				assert.Equal(t, 1, fields.AsyncProfiler.AsyncProfilerManager.(FakeAsyncProfilerManager).On("removeTmpDir").InvokedTimes())
				assert.Equal(t, 1, fields.AsyncProfiler.AsyncProfilerManager.(FakeAsyncProfilerManager).On("linkTmpDirToTargetTmpDir").InvokedTimes())
				assert.Equal(t, 0, fields.AsyncProfiler.AsyncProfilerManager.(FakeAsyncProfilerManager).On("copyProfilerToTmpDir").InvokedTimes())
			},
		},
		{
			name: "should fail when copy profiler to tmp dir fail",
			given: func() (fields, args) {
				asyncProfilerManager := newFakeAsyncProfilerManager()
				asyncProfilerManager.On("removeTmpDir").Return(nil)
				asyncProfilerManager.On("linkTmpDirToTargetTmpDir").Return(nil)
				asyncProfilerManager.On("copyProfilerToTmpDir").Return(errors.New("fake error"))

				return fields{
						AsyncProfiler: &AsyncProfiler{
							AsyncProfilerManager: asyncProfilerManager,
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
				return fields.AsyncProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.NotNil(t, err)
				assert.EqualError(t, err, "fake error")
				assert.Equal(t, 1, fields.AsyncProfiler.AsyncProfilerManager.(FakeAsyncProfilerManager).On("removeTmpDir").InvokedTimes())
				assert.Equal(t, 1, fields.AsyncProfiler.AsyncProfilerManager.(FakeAsyncProfilerManager).On("linkTmpDirToTargetTmpDir").InvokedTimes())
				assert.Equal(t, 1, fields.AsyncProfiler.AsyncProfilerManager.(FakeAsyncProfilerManager).On("copyProfilerToTmpDir").InvokedTimes())
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

func TestAsyncProfiler_Invoke(t *testing.T) {
	type fields struct {
		AsyncProfiler AsyncProfiler
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
				asyncProfilerManager := newFakeAsyncProfilerManager()
				asyncProfilerManager.On("invoke").
					Return(nil, time.Duration(0)).
					Return(nil, time.Duration(0))

				return fields{
						AsyncProfiler: AsyncProfiler{
							AsyncProfilerManager: asyncProfilerManager,
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
				fields.AsyncProfiler.delay = 0
				fields.AsyncProfiler.targetPIDs = []string{"1000", "2000"}
				return fields.AsyncProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
				assert.Equal(t, 2, fields.AsyncProfiler.AsyncProfilerManager.(FakeAsyncProfilerManager).On("invoke").InvokedTimes())
			},
		},
		{
			name: "should invoke fail when invoke fail",
			given: func() (fields, args) {
				asyncProfilerManager := newFakeAsyncProfilerManager()
				asyncProfilerManager.On("invoke").Return(errors.New("fake invoke error"), time.Duration(0))

				return fields{
						AsyncProfiler: AsyncProfiler{
							AsyncProfilerManager: asyncProfilerManager,
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
				fields.AsyncProfiler.delay = 0
				fields.AsyncProfiler.targetPIDs = []string{"1000", "2000"}
				return fields.AsyncProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				require.Error(t, err)
				assert.EqualError(t, err, "fake invoke error")
				assert.Equal(t, 1, fields.AsyncProfiler.AsyncProfilerManager.(FakeAsyncProfilerManager).On("invoke").InvokedTimes())
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

func TestAsyncProfiler_CleanUp(t *testing.T) {
	type fields struct {
		AsyncProfiler AsyncProfiler
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
			name: "should clean up",
			given: func() (fields, args) {
				_ = os.Mkdir(filepath.Join(common.TmpDir(), "async-profiler"), os.ModePerm)
				f := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.html")
				_, _ = os.Create(f)
				_, _ = os.Create(f + compressor.GetExtensionFileByCompressor[compressor.Gzip])
				return fields{
						AsyncProfiler: AsyncProfiler{
							AsyncProfilerManager: newFakeAsyncProfilerManager(),
						},
					}, args{
						job: &job.ProfilingJob{
							UID:              "UID",
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							Compressor:       compressor.Gzip,
							Tool:             api.AsyncProfiler,
							OutputType:       api.FlameGraph,
						},
					}
			},
			when: func(fields fields, args args) error {
				return fields.AsyncProfiler.CleanUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				f := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.html")
				g := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.html"+
					compressor.GetExtensionFileByCompressor[compressor.Gzip])
				assert.False(t, file.Exists(f))
				assert.False(t, file.Exists(g))
				assert.False(t, file.Exists(filepath.Join(common.TmpDir(), "async-profiler")))
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

func Test_asyncProfilerManager_copyProfilerToTmpDir(t *testing.T) {
	commander := executil.NewFakeCommander()
	commander.On("Command").Return(exec.Command("ls", common.TmpDir()))
	publisher := publish.NewFakePublisher()
	a := NewAsyncProfiler(commander, publisher)
	assert.Nil(t, a.copyProfilerToTmpDir())
}

func Test_asyncProfilerManager_invoke(t *testing.T) {
	type fields struct {
		AsyncProfiler *AsyncProfiler
	}
	type args struct {
		job *job.ProfilingJob
		pid string
	}
	tests := []struct {
		name  string
		given func() (fields, args)
		when  func(fields, args) (error, time.Duration)
		then  func(t *testing.T, fields fields, err error)
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

				commander := executil.NewFakeCommander()
				commander.On("Command").Return(exec.Command("ls", common.TmpDir()))
				publisher := publish.NewFakePublisher()
				publisher.On("Do").Return(nil)

				return fields{
						AsyncProfiler: NewAsyncProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.FakeLang,
							Tool:             api.AsyncProfiler,
							Compressor:       compressor.None,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.AsyncProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				assert.Nil(t, err)
				assert.True(t, file.Exists(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000.svg")))
				assert.True(t, fields.AsyncProfiler.AsyncProfilerManager.(*asyncProfilerManager).publisher.(*publish.Fake).On("Do").InvokedTimes() == 1)
			},
			after: func() {
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw-1000.txt"))
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000.svg"))
			},
		},
		{
			name: "should invoke fail when command fail",
			given: func() (fields, args) {
				commander := executil.NewFakeCommander()
				commander.On("Command").Return(&exec.Cmd{})
				publisher := publish.NewFakePublisher()
				publisher.On("Do").Return(nil)

				return fields{
						AsyncProfiler: NewAsyncProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.Java,
							Tool:             api.AsyncProfiler,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.AsyncProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				require.Error(t, err)
				assert.True(t, fields.AsyncProfiler.AsyncProfilerManager.(*asyncProfilerManager).publisher.(*publish.Fake).On("Do").InvokedTimes() == 0)
			},
		},
		{
			name: "should invoke fail when publish result fail",
			given: func() (fields, args) {
				log.SetPrintLogs(true)
				var b bytes.Buffer
				b.Write([]byte("test"))
				file.Write(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw-1000.txt"), b.String())
				file.Write(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000.svg"), b.String())

				commander := executil.NewFakeCommander()
				// mock commander.Command return exec.Command("ls", common.TmpDir())
				commander.On("Command").Return(exec.Command("ls", common.TmpDir()))
				publisher := publish.NewFakePublisher()
				// mock publisher.Do return error
				publisher.On("Do").Return(errors.New("fake publisher with error"))

				return fields{
						AsyncProfiler: NewAsyncProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:             0,
							ContainerRuntime:     api.FakeContainer,
							ContainerRuntimePath: common.TmpDir(),
							ContainerID:          "ContainerID",
							OutputType:           api.FlameGraph,
							Language:             api.FakeLang,
							Tool:                 api.AsyncProfiler,
							Compressor:           compressor.None,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.AsyncProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				require.Error(t, err)
				assert.ErrorContains(t, err, "fake publisher with error")
				assert.True(t, file.Exists(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000.svg")))
				assert.True(t, fields.AsyncProfiler.AsyncProfilerManager.(*asyncProfilerManager).publisher.(*publish.Fake).On("Do").InvokedTimes() == 1)
			},
			after: func() {
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw-1000.txt"))
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000.svg"))
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
			tt.then(t, fields, err)

			if tt.after != nil {
				tt.after()
			}
		})
	}
}

func Test_asyncProfilerManager_cleanUp(t *testing.T) {
	commander := executil.NewFakeCommander()
	commander.On("Command").Return(exec.Command("ls", common.TmpDir()))
	publisher := publish.NewFakePublisher()
	a := NewAsyncProfiler(commander, publisher)
	a.cleanUp(&job.ProfilingJob{
		Duration:         0,
		ContainerRuntime: api.FakeContainer,
		ContainerID:      "ContainerID",
		OutputType:       api.FlameGraph,
		Language:         api.FakeLang,
		Tool:             api.AsyncProfiler,
		Compressor:       compressor.None,
	}, "1000")
	assert.True(t, commander.(*executil.Fake).On("Command").InvokedTimes() == 1)

	commander = executil.NewFakeCommander()
	commander.On("Command").Return(&exec.Cmd{})
	a = NewAsyncProfiler(commander, publisher)
	a.cleanUp(&job.ProfilingJob{
		Duration:         0,
		ContainerRuntime: api.FakeContainer,
		ContainerID:      "ContainerID",
		OutputType:       api.FlameGraph,
		Language:         api.FakeLang,
		Tool:             api.AsyncProfiler,
		Compressor:       compressor.None,
	}, "1000")
	assert.True(t, commander.(*executil.Fake).On("Command").InvokedTimes() == 1)
}
