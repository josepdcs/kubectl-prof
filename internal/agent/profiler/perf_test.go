package profiler

import (
	"bytes"
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

// FakePerfManager is an interface that wraps the PerfManager interface
type FakePerfManager interface {
	Return(fakeReturnValues ...interface{}) *fakePerfManager
	On(methodName string) *fakePerfManager
	InvokedTimes(methodName string) int

	PerfManager
}

// fakePerfManager is a fake implementation of the PerfManager interface
type fakePerfManager struct {
	fakeReturnValues []interface{}
	methodName       string
	invokes          map[string]int
}

// newFakePerfManager returns a new fake perf manager
func newFakePerfManager() FakePerfManager {
	return &fakePerfManager{
		invokes: make(map[string]int),
	}
}

func (p *fakePerfManager) Return(fakeReturnValues ...interface{}) *fakePerfManager {
	p.fakeReturnValues = fakeReturnValues
	return p
}

func (p *fakePerfManager) On(methodName string) *fakePerfManager {
	p.methodName = methodName
	return p
}

func (p *fakePerfManager) InvokedTimes(methodName string) int {
	return p.invokes[methodName]
}

func (p *fakePerfManager) invoke(job *job.ProfilingJob, pid string) (error, time.Duration) {
	p.invokes["invoke"]++
	if p.methodName == "invoke" && p.fakeReturnValues != nil && len(p.fakeReturnValues) > 0 {
		return p.fakeReturnValues[0].(error), p.fakeReturnValues[1].(time.Duration)
	}
	return nil, 0
}

func (p *fakePerfManager) runPerfRecord(job *job.ProfilingJob, pid string) error {
	p.invokes["runPerfRecord"]++
	if p.methodName == "runPerfRecord" && p.fakeReturnValues != nil && len(p.fakeReturnValues) > 0 {
		return p.fakeReturnValues[0].(error)
	}
	return nil
}

func (p *fakePerfManager) runPerfScript(job *job.ProfilingJob, pid string) error {
	p.invokes["runPerfScript"]++
	if p.methodName == "runPerfScript" && p.fakeReturnValues != nil && len(p.fakeReturnValues) > 0 {
		return p.fakeReturnValues[0].(error)
	}
	return nil
}

func (p *fakePerfManager) foldPerfOutput(job *job.ProfilingJob, pid string) (error, string) {
	p.invokes["foldPerfOutput"]++
	if p.methodName == "foldPerfOutput" && p.fakeReturnValues != nil && len(p.fakeReturnValues) > 0 {
		return p.fakeReturnValues[0].(error), p.fakeReturnValues[1].(string)
	}
	return nil, ""
}

func (p *fakePerfManager) handleFlamegraph(job *job.ProfilingJob, f flamegraph.FrameGrapher, fileName string, pid string) error {
	p.invokes["handleFlamegraph"]++
	if p.methodName == "handleFlamegraph" && p.fakeReturnValues != nil && len(p.fakeReturnValues) > 0 {
		return p.fakeReturnValues[0].(error)
	}
	return nil
}

func TestPerfProfiler_SetUp(t *testing.T) {
	type fields struct {
		PerfProfiler *PerfProfiler
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
						PerfProfiler: &PerfProfiler{
							PerfManager: newFakePerfManager(),
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
				return fields.PerfProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
				assert.Equal(t, []string{"PID_ContainerID"}, fields.PerfProfiler.targetPIDs)
			},
		},
		{
			name: "should setup with given PID",
			given: func() (fields, args) {
				return fields{
						PerfProfiler: &PerfProfiler{
							PerfManager: newFakePerfManager(),
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
				return fields.PerfProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
				assert.Equal(t, []string{"PID_ContainerID"}, fields.PerfProfiler.targetPIDs)
			},
		},
		{
			name: "should fail when container PID not found",
			given: func() (fields, args) {
				return fields{
						PerfProfiler: &PerfProfiler{
							PerfManager: newFakePerfManager(),
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
				return fields.PerfProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.NotNil(t, err)
				assert.Empty(t, fields.PerfProfiler.targetPIDs)
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

func TestPerfProfiler_Invoke(t *testing.T) {
	type fields struct {
		PerfProfiler *PerfProfiler
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
						PerfProfiler: &PerfProfiler{
							PerfManager: newFakePerfManager(),
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
				fields.PerfProfiler.delay = 0
				fields.PerfProfiler.targetPIDs = []string{"1000", "2000"}
				return fields.PerfProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
				assert.Equal(t, 2, fields.PerfProfiler.PerfManager.(FakePerfManager).InvokedTimes("invoke"))
			},
		},
		{
			name: "should invoke fail when invoke fail",
			given: func() (fields, args) {
				return fields{
						PerfProfiler: &PerfProfiler{
							PerfManager: newFakePerfManager().On("invoke").Return(errors.New("fake invoke error"), time.Duration(0)),
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
				fields.PerfProfiler.delay = 0
				fields.PerfProfiler.targetPIDs = []string{"1000", "2000"}
				return fields.PerfProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				require.Error(t, err)
				assert.Equal(t, 1, fields.PerfProfiler.PerfManager.(FakePerfManager).InvokedTimes("invoke"))
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

func TestPerfProfiler_CleanUp(t *testing.T) {
	type fields struct {
		PerfProfiler *PerfProfiler
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
						PerfProfiler: &PerfProfiler{
							PerfManager: newFakePerfManager(),
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
				return fields.PerfProfiler.CleanUp(args.job)
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

func Test_perfManager_invoke(t *testing.T) {
	type fields struct {
		PerfProfiler *PerfProfiler
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
				// mock commander.Command return exec.Command("ls", common.TmpDir())
				commander.Return(exec.Command("ls", common.TmpDir())).Return(exec.Command("ls", common.TmpDir())).Return(exec.Command("ls", common.TmpDir())).On("Command")
				publisher := publish.NewFakePublisher()

				return fields{
						PerfProfiler: NewPerfProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.FakeLang,
							Tool:             api.Perf,
							Compressor:       compressor.None,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.PerfProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				assert.Nil(t, err)
				assert.True(t, file.Exists(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000.svg")))
				assert.True(t, fields.PerfProfiler.PerfManager.(*perfManager).publisher.(*publish.Fake).InvokedTimes("Do") == 1)
			},
			after: func() {
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw-1000.txt"))
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000.svg"))
			},
		},
		{
			name: "should invoke fail when perf record fail",
			given: func() (fields, args) {
				commander := executil.NewFakeCommander()
				commander.Return(&exec.Cmd{}).On("Command")
				publisher := publish.NewFakePublisher()

				return fields{
						PerfProfiler: NewPerfProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.Go,
							Tool:             api.Perf,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.PerfProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				require.Error(t, err)
				assert.ErrorContains(t, err, "perf record failed")
				assert.True(t, fields.PerfProfiler.PerfManager.(*perfManager).publisher.(*publish.Fake).InvokedTimes("Do") == 0)
			},
		},
		{
			name: "should invoke fail when perf script fail",
			given: func() (fields, args) {
				commander := executil.NewFakeCommander()
				commander.Return(exec.Command("ls", common.TmpDir())).Return(&exec.Cmd{}).On("Command")
				publisher := publish.NewFakePublisher()

				return fields{
						PerfProfiler: NewPerfProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.Go,
							Tool:             api.Perf,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.PerfProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				require.Error(t, err)
				assert.ErrorContains(t, err, "perf script failed")
				assert.True(t, fields.PerfProfiler.PerfManager.(*perfManager).publisher.(*publish.Fake).InvokedTimes("Do") == 0)
			},
		},
		{
			name: "should invoke fail when folding perf output fail",
			given: func() (fields, args) {
				commander := executil.NewFakeCommander()
				commander.Return(exec.Command("ls", common.TmpDir())).Return(exec.Command("ls", common.TmpDir())).Return(&exec.Cmd{}).On("Command")
				publisher := publish.NewFakePublisher()

				return fields{
						PerfProfiler: NewPerfProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.Go,
							Tool:             api.Perf,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.PerfProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				require.Error(t, err)
				assert.ErrorContains(t, err, "folding perf output failed")
				assert.True(t, fields.PerfProfiler.PerfManager.(*perfManager).publisher.(*publish.Fake).InvokedTimes("Do") == 0)
			},
		},
		{
			name: "should invoke return nil when fail handle flamegraph",
			given: func() (fields, args) {
				log.SetPrintLogs(true)
				commander := executil.NewFakeCommander()
				// mock commander.Command return exec.Command("ls", common.TmpDir())
				commander.Return(exec.Command("ls", common.TmpDir())).Return(exec.Command("ls", common.TmpDir())).Return(exec.Command("ls", common.TmpDir())).On("Command")
				publisher := publish.NewFakePublisher()

				return fields{
						PerfProfiler: NewPerfProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.Go,
							Tool:             api.Perf,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.PerfProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				require.NoError(t, err)
				assert.True(t, fields.PerfProfiler.PerfManager.(*perfManager).publisher.(*publish.Fake).InvokedTimes("Do") == 0)
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
				commander.Return(exec.Command("ls", common.TmpDir())).Return(exec.Command("ls", common.TmpDir())).Return(exec.Command("ls", common.TmpDir())).On("Command")
				publisher := publish.NewFakePublisher()
				// mock publisher.Do return error
				publisher.Return(errors.New("fake publisher with error")).On("Do")

				return fields{
						PerfProfiler: NewPerfProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.FakeLang,
							Tool:             api.Perf,
							Compressor:       compressor.None,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.PerfProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				require.Error(t, err)
				assert.ErrorContains(t, err, "fake publisher with error")
				assert.True(t, file.Exists(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000.svg")))
				assert.True(t, fields.PerfProfiler.PerfManager.(*perfManager).publisher.(*publish.Fake).InvokedTimes("Do") == 1)
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

func Test_perfManager_handleFlamegraph(t *testing.T) {
	type fields struct {
		PerfProfiler *PerfProfiler
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
				b.Write([]byte("testtesttesttesttesttesttesttesttesttesttesttesttest"))
				_ = os.WriteFile(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt"), b.Bytes(), 0644)

				commander := executil.NewFakeCommander()
				commander.(*executil.Fake).Return(exec.Command("ls", common.TmpDir()))
				publisher := publish.NewFakePublisher()

				return fields{
						PerfProfiler: NewPerfProfiler(commander, publisher),
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
				return fields.PerfProfiler.handleFlamegraph(args.job, args.flameGrapher, args.fileName, args.resultFileName)
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
				b.Write([]byte("testtesttesttesttesttesttesttesttesttesttesttesttest"))
				_ = os.WriteFile(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt"), b.Bytes(), 0644)

				commander := executil.NewFakeCommander()
				commander.(*executil.Fake).Return(exec.Command("ls", common.TmpDir()))
				publisher := publish.NewFakePublisher()

				return fields{
						PerfProfiler: NewPerfProfiler(commander, publisher),
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
				return fields.PerfProfiler.handleFlamegraph(args.job, args.flameGrapher, args.fileName, args.resultFileName)
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
				commander.(*executil.Fake).Return(exec.Command("ls", common.TmpDir()))
				publisher := publish.NewFakePublisher()

				return fields{
						PerfProfiler: NewPerfProfiler(commander, publisher),
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
				return fields.PerfProfiler.handleFlamegraph(args.job, args.flameGrapher, args.fileName, args.resultFileName)
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
