package profiler

import (
	"bytes"
	"fmt"
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
	"github.com/josepdcs/kubectl-prof/internal/agent/util/flamegraph"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/publish"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPhpspyProfiler_SetUp(t *testing.T) {
	type fields struct {
		PhpspyProfiler *PhpspyProfiler
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
						PhpspyProfiler: &PhpspyProfiler{
							PhpspyManager: newMockPhpspyManager(),
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
				return fields.PhpspyProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
				assert.Equal(t, []string{"PID_ContainerID"}, fields.PhpspyProfiler.targetPIDs)
			},
		},
		{
			name: "should setup with given PID",
			given: func() (fields, args) {
				return fields{
						PhpspyProfiler: &PhpspyProfiler{
							PhpspyManager: newMockPhpspyManager(),
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
				return fields.PhpspyProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
				assert.Equal(t, []string{"PID_ContainerID"}, fields.PhpspyProfiler.targetPIDs)
			},
		},
		{
			name: "should fail when container PID not found",
			given: func() (fields, args) {
				return fields{
						PhpspyProfiler: &PhpspyProfiler{
							PhpspyManager: newMockPhpspyManager(),
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
				return fields.PhpspyProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.NotNil(t, err)
				assert.Empty(t, fields.PhpspyProfiler.targetPIDs)
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

func TestPhpspyProfiler_Invoke(t *testing.T) {
	type fields struct {
		PhpspyProfiler *PhpspyProfiler
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
				phpspyManager := newMockPhpspyManager()
				phpspyManager.On("invoke", mock.Anything, mock.AnythingOfType("string")).
					Return(nil, time.Duration(0)).
					Twice()

				return fields{
						PhpspyProfiler: &PhpspyProfiler{
							PhpspyManager: phpspyManager,
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
				fields.PhpspyProfiler.delay = 0
				fields.PhpspyProfiler.targetPIDs = []string{"1000", "2000"}
				return fields.PhpspyProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
				fields.PhpspyProfiler.PhpspyManager.(*mockPhpspyManager).AssertNumberOfCalls(t, "invoke", 2)
			},
		},
		{
			name: "should invoke fail when invoke fail",
			given: func() (fields, args) {
				phpspyManager := newMockPhpspyManager()
				phpspyManager.On("invoke", mock.Anything, mock.AnythingOfType("string")).
					Return(errors.New("fake invoke error"), time.Duration(0)).
					Once()

				return fields{
						PhpspyProfiler: &PhpspyProfiler{
							PhpspyManager: phpspyManager,
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
				fields.PhpspyProfiler.delay = 0
				fields.PhpspyProfiler.targetPIDs = []string{"1000", "2000"}
				return fields.PhpspyProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				require.Error(t, err)
				assert.EqualError(t, err, "fake invoke error")
				fields.PhpspyProfiler.PhpspyManager.(*mockPhpspyManager).AssertNumberOfCalls(t, "invoke", 1)
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

func TestPhpspyProfiler_CleanUp(t *testing.T) {
	type fields struct {
		PhpspyProfiler *PhpspyProfiler
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
						PhpspyProfiler: &PhpspyProfiler{
							PhpspyManager: newMockPhpspyManager(),
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
				return fields.PhpspyProfiler.CleanUp(args.job)
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

func Test_phpspyManager_invoke(t *testing.T) {
	type fields struct {
		PhpspyProfiler *PhpspyProfiler
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
				file.Write(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw-1000-1.txt"), b.String())
				file.Write(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000-1.svg"), b.String())

				commander := executil.NewMockCommander()
				commander.On("Command").Return(exec.Command("ls", common.TmpDir())).Once()
				commander.On("Command").Return(exec.Command("ls", common.TmpDir())).Once()
				publisher := publish.NewFakePublisher()
				publisher.On("Do").Return(nil)

				return fields{
						PhpspyProfiler: NewPhpspyProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.FakeLang,
							Tool:             api.Phpspy,
							Compressor:       compressor.None,
							Iteration:        1,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.PhpspyProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				assert.Nil(t, err)
				assert.True(t, file.Exists(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000-1.svg")))
				assert.True(t, fields.PhpspyProfiler.PhpspyManager.(*phpspyManager).publisher.(*publish.Fake).On("Do").InvokedTimes() == 1)
			},
			after: func() {
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw-1000-1.txt"))
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000-1.svg"))
			},
		},
		{
			name: "should invoke fail when command fail",
			given: func() (fields, args) {
				commander := executil.NewMockCommander()
				commander.On("Command").Return(&exec.Cmd{})
				publisher := publish.NewFakePublisher()
				publisher.On("Do").Return(nil)

				return fields{
						PhpspyProfiler: NewPhpspyProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.PHP,
							Tool:             api.Phpspy,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.PhpspyProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				require.Error(t, err)
				assert.True(t, fields.PhpspyProfiler.PhpspyManager.(*phpspyManager).publisher.(*publish.Fake).On("Do").InvokedTimes() == 0)
			},
		},
		{
			name: "should invoke fail when folding phpspy output fails",
			given: func() (fields, args) {
				log.SetPrintLogs(true)
				commander := executil.NewMockCommander()
				commander.On("Command").Return(exec.Command("ls", common.TmpDir())).Once()
				commander.On("Command").Return(&exec.Cmd{}).Once()
				publisher := publish.NewFakePublisher()
				publisher.On("Do").Return(nil)

				return fields{
						PhpspyProfiler: NewPhpspyProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.PHP,
							Tool:             api.Phpspy,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.PhpspyProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				require.Error(t, err)
				assert.ErrorContains(t, err, "folding phpspy output failed")
				assert.True(t, fields.PhpspyProfiler.PhpspyManager.(*phpspyManager).publisher.(*publish.Fake).On("Do").InvokedTimes() == 0)
			},
		},
		{
			name: "should invoke return nil when fail handle flamegraph",
			given: func() (fields, args) {
				log.SetPrintLogs(true)
				commander := executil.NewMockCommander()
				commander.On("Command").Return(exec.Command("ls", common.TmpDir())).Once()
				commander.On("Command").Return(exec.Command("ls", common.TmpDir())).Once()
				publisher := publish.NewFakePublisher()
				publisher.On("Do").Return(nil)

				return fields{
						PhpspyProfiler: NewPhpspyProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.PHP,
							Tool:             api.Phpspy,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.PhpspyProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				require.NoError(t, err)
				assert.True(t, fields.PhpspyProfiler.PhpspyManager.(*phpspyManager).publisher.(*publish.Fake).On("Do").InvokedTimes() == 0)
			},
		},
		{
			name: "should invoke fail when fail publish",
			given: func() (fields, args) {
				log.SetPrintLogs(true)
				var b bytes.Buffer
				b.Write([]byte("test"))
				file.Write(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw-1000-1.txt"), b.String())
				file.Write(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000-1.svg"), b.String())

				commander := executil.NewMockCommander()
				commander.On("Command").Return(exec.Command("ls", common.TmpDir())).Once()
				commander.On("Command").Return(exec.Command("ls", common.TmpDir())).Once()
				publisher := publish.NewFakePublisher()
				publisher.On("Do").Return(errors.New("fake publisher with error"))

				return fields{
						PhpspyProfiler: NewPhpspyProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.FakeLang,
							Tool:             api.Phpspy,
							Compressor:       compressor.None,
							Iteration:        1,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.PhpspyProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				require.Error(t, err)
				assert.ErrorContains(t, err, "fake publisher with error")
				assert.True(t, file.Exists(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000-1.svg")))
				assert.True(t, fields.PhpspyProfiler.PhpspyManager.(*phpspyManager).publisher.(*publish.Fake).On("Do").InvokedTimes() == 1)
			},
			after: func() {
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw-1000-1.txt"))
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000-1.svg"))
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

func Test_phpspyManager_collapsePhpspyOutput(t *testing.T) {
	type fields struct {
		PhpspyProfiler *PhpspyProfiler
	}
	type args struct {
		job *job.ProfilingJob
		pid string
	}
	tests := []struct {
		name  string
		given func() (fields, args)
		when  func(fields, args) (error, string)
		then  func(t *testing.T, fields fields, err error, fileName string)
		after func()
	}{
		{
			name: "should collapse phpspy output",
			given: func() (fields, args) {
				log.SetPrintLogs(true)
				var b bytes.Buffer
				b.Write([]byte("test phpspy output"))
				rawOutputFile := fmt.Sprintf(phpSpyRawOutputFile, "1000", 1)
				_ = os.WriteFile(rawOutputFile, b.Bytes(), 0644)

				commander := executil.NewMockCommander()
				commander.On("Command").Return(exec.Command("ls", common.TmpDir()))
				publisher := publish.NewFakePublisher()

				return fields{
						PhpspyProfiler: NewPhpspyProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.PHP,
							Tool:             api.Phpspy,
							Iteration:        1,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, string) {
				return fields.PhpspyProfiler.collapsePhpspyOutput(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error, fileName string) {
				assert.Nil(t, err)
				assert.True(t, file.Exists(fileName))
			},
			after: func() {
				rawOutputFile := fmt.Sprintf(phpSpyRawOutputFile, "1000", 1)
				_ = file.Remove(rawOutputFile)
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw-1000-1.txt"))
			},
		},
		{
			name: "should fail when collapse phpspy output fails",
			given: func() (fields, args) {
				log.SetPrintLogs(true)
				commander := executil.NewMockCommander()
				commander.On("Command").Return(&exec.Cmd{})
				publisher := publish.NewFakePublisher()

				return fields{
						PhpspyProfiler: NewPhpspyProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.PHP,
							Tool:             api.Phpspy,
							Iteration:        1,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, string) {
				return fields.PhpspyProfiler.collapsePhpspyOutput(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error, fileName string) {
				require.Error(t, err)
				assert.ErrorContains(t, err, "could not collapse phpspy output")
				assert.Empty(t, fileName)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			fields, args := tt.given()

			// When
			err, fileName := tt.when(fields, args)

			// Then
			tt.then(t, fields, err, fileName)

			if tt.after != nil {
				tt.after()
			}
		})
	}
}

func Test_phpspyManager_handleFlamegraph(t *testing.T) {
	type fields struct {
		PhpspyProfiler *PhpspyProfiler
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
				commander := executil.NewMockCommander()
				commander.On("Command").Return(exec.Command("ls", common.TmpDir()))
				publisher := publish.NewFakePublisher()

				return fields{
						PhpspyProfiler: NewPhpspyProfiler(commander, publisher),
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
				return fields.PhpspyProfiler.handleFlamegraph(args.job, args.flameGrapher, args.fileName, args.resultFileName)
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
				commander := executil.NewMockCommander()
				commander.On("Command").Return(exec.Command("ls", common.TmpDir()))
				publisher := publish.NewFakePublisher()

				return fields{
						PhpspyProfiler: NewPhpspyProfiler(commander, publisher),
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
				return fields.PhpspyProfiler.handleFlamegraph(args.job, args.flameGrapher, args.fileName, args.resultFileName)
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
				commander := executil.NewFakeCommander()
				commander.On("Command").Return(exec.Command("ls", common.TmpDir()))
				publisher := publish.NewFakePublisher()

				return fields{
						PhpspyProfiler: NewPhpspyProfiler(commander, publisher),
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
				return fields.PhpspyProfiler.handleFlamegraph(args.job, args.flameGrapher, args.fileName, args.resultFileName)
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
