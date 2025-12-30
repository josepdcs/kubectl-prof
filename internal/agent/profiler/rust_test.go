package profiler

import (
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
	"github.com/josepdcs/kubectl-prof/internal/agent/util/publish"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRustProfiler_SetUp(t *testing.T) {
	type fields struct {
		RustProfiler *RustProfiler
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
						RustProfiler: &RustProfiler{
							RustManager: newMockRustManager(),
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
				return fields.RustProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
				assert.Equal(t, []string{"PID_ContainerID"}, fields.RustProfiler.targetPIDs)
			},
		},
		{
			name: "should setup with given PID",
			given: func() (fields, args) {
				return fields{
						RustProfiler: &RustProfiler{
							RustManager: newMockRustManager(),
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
				return fields.RustProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
				assert.Equal(t, []string{"PID_ContainerID"}, fields.RustProfiler.targetPIDs)
			},
		},
		{
			name: "should fail when container PID not found",
			given: func() (fields, args) {
				return fields{
						RustProfiler: &RustProfiler{
							RustManager: newMockRustManager(),
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
				return fields.RustProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.NotNil(t, err)
				assert.Empty(t, fields.RustProfiler.targetPIDs)
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

func TestRustProfiler_Invoke(t *testing.T) {
	type fields struct {
		RustProfiler *RustProfiler
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
				rustManager := newMockRustManager()
				rustManager.On("invoke", mock.Anything, mock.AnythingOfType("string")).
					Return(nil, time.Duration(0)).
					Twice()

				return fields{
						RustProfiler: &RustProfiler{
							RustManager: rustManager,
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
				fields.RustProfiler.delay = 0
				fields.RustProfiler.targetPIDs = []string{"1000", "2000"}
				return fields.RustProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
				fields.RustProfiler.RustManager.(*mockRustManager).AssertNumberOfCalls(t, "invoke", 2)
			},
		},
		{
			name: "should invoke fail when invoke fail",
			given: func() (fields, args) {
				rustManager := newMockRustManager()
				rustManager.On("invoke", mock.Anything, mock.AnythingOfType("string")).
					Return(errors.New("fake invoke error"), time.Duration(0)).
					Once()

				return fields{
						RustProfiler: &RustProfiler{
							RustManager: rustManager,
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
				fields.RustProfiler.delay = 0
				fields.RustProfiler.targetPIDs = []string{"1000", "2000"}
				return fields.RustProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				require.Error(t, err)
				assert.EqualError(t, err, "fake invoke error")
				fields.RustProfiler.RustManager.(*mockRustManager).AssertNumberOfCalls(t, "invoke", 1)
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

func TestRustProfiler_CleanUp(t *testing.T) {
	type fields struct {
		RustProfiler *RustProfiler
	}
	type args struct {
		job *job.ProfilingJob
	}
	tests := []struct {
		name  string
		given func() (fields, args)
		when  func(fields, args) error
		then  func(t *testing.T, err error)
	}{
		{
			name: "should clean up",
			given: func() (fields, args) {
				f := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.svg")
				_, _ = os.Create(f)
				_, _ = os.Create(f + compressor.GetExtensionFileByCompressor[compressor.Gzip])
				return fields{
						RustProfiler: &RustProfiler{
							RustManager: newMockRustManager(),
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
				return fields.RustProfiler.CleanUp(args.job)
			},
			then: func(t *testing.T, err error) {
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
			tt.then(t, err)
		})
	}
}

func Test_rustManager_invoke(t *testing.T) {
	type fields struct {
		RustProfiler *RustProfiler
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
				fileName := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000-0.svg")

				commander := executil.NewMockCommander()
				// Create a mock command that simulates flamegraph behavior:
				// Creates the output file and sleeps briefly, then exits when receives SIGINT
				script := fmt.Sprintf(`
#!/bin/sh
touch '%s'
trap 'exit 0' INT TERM
sleep 0.5 &
wait $!
`, fileName)
				commander.On("Command").Return(exec.Command("sh", "-c", script))
				publisher := publish.NewFakePublisher()
				publisher.On("Do").Return(nil)

				return fields{
						RustProfiler: NewRustProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							Interval:         100 * time.Millisecond, // Send SIGINT after 0.1 second
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.FakeLang,
							Tool:             api.CargoFlame,
							Compressor:       compressor.None,
							Iteration:        0,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.RustProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				assert.Nil(t, err)
				fileName := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000-0.svg")
				assert.True(t, file.Exists(fileName))
				assert.True(t, fields.RustProfiler.RustManager.(*rustManager).publisher.(*publish.Fake).On("Do").InvokedTimes() == 1)
				fields.RustProfiler.RustManager.(*rustManager).commander.(*executil.MockCommander).AssertNumberOfCalls(t, "Command", 1)
			},
			after: func() {
				fileName := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000-0.svg")
				_ = file.Remove(fileName)
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
						RustProfiler: NewRustProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.Rust,
							Tool:             api.CargoFlame,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.RustProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				require.Error(t, err)
				assert.True(t, fields.RustProfiler.RustManager.(*rustManager).publisher.(*publish.Fake).On("Do").InvokedTimes() == 0)
				fields.RustProfiler.RustManager.(*rustManager).commander.(*executil.MockCommander).AssertNumberOfCalls(t, "Command", 1)
			},
		},
		{
			name: "should fail when publisher fail",
			given: func() (fields, args) {
				fileName := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000-0.svg")

				commander := executil.NewMockCommander()
				// Create a mock command that creates the output file
				script := fmt.Sprintf(`
#!/bin/sh
touch '%s'
trap 'exit 0' INT TERM
sleep 0.5 &
wait $!
`, fileName)
				commander.On("Command").Return(exec.Command("sh", "-c", script))
				publisher := publish.NewFakePublisher()
				publisher.On("Do").Return(errors.New("fake publisher with error"))

				return fields{
						RustProfiler: NewRustProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							Interval:         100 * time.Millisecond, // Send SIGINT after 0.1 second
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.FakeLang,
							Tool:             api.CargoFlame,
							Compressor:       compressor.None,
							Iteration:        0,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.RustProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				require.Error(t, err)
				assert.ErrorContains(t, err, "fake publisher with error")
				fileName := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000-0.svg")
				assert.True(t, file.Exists(fileName))
				assert.True(t, fields.RustProfiler.RustManager.(*rustManager).publisher.(*publish.Fake).On("Do").InvokedTimes() == 1)
				fields.RustProfiler.RustManager.(*rustManager).commander.(*executil.MockCommander).AssertNumberOfCalls(t, "Command", 1)

			},
			after: func() {
				fileName := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000-0.svg")
				_ = file.Remove(fileName)
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

func Test_rustCommand(t *testing.T) {
	type args struct {
		job      *job.ProfilingJob
		pid      string
		fileName string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "should build rust command with flamegraph output",
			args: args{
				job: &job.ProfilingJob{
					Interval:   time.Duration(10) * time.Second,
					OutputType: api.FlameGraph,
				},
				pid:      "1000",
				fileName: "/tmp/flamegraph.svg",
			},
			want: []string{cargoFlameLocation, "-p", "1000", "-o", "/tmp/flamegraph.svg", "--palette", "rust", "--title", "Flamegraph for PID 1000"},
		},
		{
			name: "should build rust command with different interval",
			args: args{
				job: &job.ProfilingJob{
					Interval:   time.Duration(30) * time.Second,
					OutputType: api.FlameGraph,
				},
				pid:      "2000",
				fileName: "/tmp/output.svg",
			},
			want: []string{cargoFlameLocation, "-p", "2000", "-o", "/tmp/output.svg", "--palette", "rust", "--title", "Flamegraph for PID 2000"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commander := executil.NewCommander()
			cmd := rustCommand(commander, tt.args.job, tt.args.pid, tt.args.fileName)
			got := append([]string{cmd.Path}, cmd.Args[1:]...)
			assert.Equal(t, tt.want, got)
		})
	}
}
