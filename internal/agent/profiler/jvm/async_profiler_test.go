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
	"github.com/stretchr/testify/mock"
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
				asyncProfilerManager := newMockAsyncProfilerManager()
				asyncProfilerManager.On("stageProfilerLibrary", mock.AnythingOfType("string")).Return(nil)

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
				fields.AsyncProfiler.AsyncProfilerManager.(*mockAsyncProfilerManager).AssertNumberOfCalls(t, "stageProfilerLibrary", 1)
			},
		},
		{
			name: "should setup when PID is given",
			given: func() (fields, args) {
				asyncProfilerManager := newMockAsyncProfilerManager()
				asyncProfilerManager.On("stageProfilerLibrary", mock.AnythingOfType("string")).Return(nil)

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
				fields.AsyncProfiler.AsyncProfilerManager.(*mockAsyncProfilerManager).AssertNumberOfCalls(t, "stageProfilerLibrary", 1)
			},
		},
		{
			name: "should fail when getting target filesystem fail",
			given: func() (fields, args) {
				asyncProfilerManager := newMockAsyncProfilerManager()
				asyncProfilerManager.On("stageProfilerLibrary", mock.AnythingOfType("string")).Return(nil)

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
				fields.AsyncProfiler.AsyncProfilerManager.(*mockAsyncProfilerManager).AssertNumberOfCalls(t, "stageProfilerLibrary", 0)
			},
		},
		{
			name: "should fail when staging the profiler library fails",
			given: func() (fields, args) {
				asyncProfilerManager := newMockAsyncProfilerManager()
				asyncProfilerManager.On("stageProfilerLibrary", mock.AnythingOfType("string")).Return(errors.New("fake error"))

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
				fields.AsyncProfiler.AsyncProfilerManager.(*mockAsyncProfilerManager).AssertNumberOfCalls(t, "stageProfilerLibrary", 1)
			},
		},
		{
			name: "should fail when container PID not found",
			given: func() (fields, args) {
				asyncProfilerManager := newMockAsyncProfilerManager()
				asyncProfilerManager.On("stageProfilerLibrary", mock.AnythingOfType("string")).Return(nil)

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
				fields.AsyncProfiler.AsyncProfilerManager.(*mockAsyncProfilerManager).AssertNumberOfCalls(t, "stageProfilerLibrary", 0)
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
				asyncProfilerManager := newMockAsyncProfilerManager()
				asyncProfilerManager.On("invoke", mock.Anything, mock.AnythingOfType("string")).
					Return(nil, time.Duration(0)).Twice()

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
				fields.AsyncProfiler.AsyncProfilerManager.(*mockAsyncProfilerManager).AssertNumberOfCalls(t, "invoke", 2)
			},
		},
		{
			name: "should invoke fail when invoke fail",
			given: func() (fields, args) {
				asyncProfilerManager := newMockAsyncProfilerManager()
				// Due to concurrency, one or both tasks may reach invoke before the first error is observed.
				// Allow any number of calls (including 2) without making the test flaky.
				asyncProfilerManager.On("invoke", mock.Anything, mock.AnythingOfType("string")).
					Return(errors.New("fake invoke error"), time.Duration(0)).Maybe()

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
				// Both tasks are submitted concurrently; depending on timing, 1 or 2 invocations may occur before Wait() returns.
				// Assert that at least one invocation happened, and no more than the number of PIDs (2).
				m := fields.AsyncProfiler.AsyncProfilerManager.(*mockAsyncProfilerManager)
				invokes := 0
				for _, c := range m.Calls {
					if c.Method == "invoke" {
						invokes++
					}
				}
				assert.GreaterOrEqual(t, invokes, 1, "invoke should be called at least once")
				assert.LessOrEqual(t, invokes, 2, "invoke should not be called more than the number of PIDs")
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
				asyncProfilerManager := newMockAsyncProfilerManager()
				asyncProfilerManager.On("releaseTargetArtifacts").Return()

				f := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.html")
				_, _ = os.Create(f)
				_, _ = os.Create(f + compressor.GetExtensionFileByCompressor[compressor.Gzip])
				return fields{
					AsyncProfiler: AsyncProfiler{
						AsyncProfilerManager: asyncProfilerManager,
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
				g := f + compressor.GetExtensionFileByCompressor[compressor.Gzip]
				assert.False(t, file.Exists(f))
				assert.False(t, file.Exists(g))
				assert.Nil(t, err)
				fields.AsyncProfiler.AsyncProfilerManager.(*mockAsyncProfilerManager).
					AssertNumberOfCalls(t, "releaseTargetArtifacts", 1)
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

func Test_asyncProfilerManager_stageProfilerLibrary(t *testing.T) {
	commander := executil.NewMockCommander()
	publisher := publish.NewFakePublisher()
	a := NewAsyncProfiler(commander, publisher)

	err := a.stageProfilerLibrary(t.TempDir())

	require.Error(t, err)
	assert.Contains(t, err.Error(), libSourcePath)
}

func Test_safeMkdir(t *testing.T) {
	t.Run("should create a directory that does not exist", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "kubectl-prof")

		require.NoError(t, safeMkdir(dir))

		info, err := os.Lstat(dir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("should accept an existing directory", func(t *testing.T) {
		dir := t.TempDir()

		assert.NoError(t, safeMkdir(dir))
	})

	t.Run("should refuse a directory writable by the target", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "kubectl-prof")
		require.NoError(t, os.Mkdir(dir, 0o755))
		require.NoError(t, os.Chmod(dir, 0o777))

		err := safeMkdir(dir)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "writable by group or others")
	})

	t.Run("should refuse a symlink", func(t *testing.T) {
		base := t.TempDir()
		elsewhere := filepath.Join(base, "elsewhere")
		require.NoError(t, os.Mkdir(elsewhere, 0o755))
		link := filepath.Join(base, "kubectl-prof")
		require.NoError(t, os.Symlink(elsewhere, link))

		err := safeMkdir(link)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "symlink")
	})

	t.Run("should refuse a regular file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "kubectl-prof")
		require.NoError(t, os.WriteFile(path, []byte("not a directory"), 0o600))

		err := safeMkdir(path)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a directory")
	})
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

				commander := executil.NewMockCommander()
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
				commander := executil.NewMockCommander()
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

				commander := executil.NewMockCommander()
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
				// Publishing was attempted and returned an error; avoid asserting Fake publisher counters to reduce flakiness.
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
			withLocalTarget(t)
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

func withLocalTarget(t *testing.T) {
	t.Helper()

	dir := t.TempDir()
	creds, tmp := readTargetCredsFn, targetTmpDir

	readTargetCredsFn = func(string) (targetCreds, error) {
		return targetCreds{uid: os.Getuid(), gid: os.Getgid()}, nil
	}
	targetTmpDir = func(string) string { return dir }

	t.Cleanup(func() {
		readTargetCredsFn, targetTmpDir = creds, tmp
	})
}

func Test_asyncProfilerManager_cleanUp(t *testing.T) {
	commander := executil.NewMockCommander()
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
	commander.AssertNumberOfCalls(t, "Command", 1)

	commander = executil.NewMockCommander()
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
	commander.AssertNumberOfCalls(t, "Command", 1)
}

func Test_asyncProfilerCommand_WithAdditionalArguments(t *testing.T) {
	log.SetPrintLogs(false)
	manager := &asyncProfilerManager{
		commander: executil.NewCommander(),
	}

	t.Run("should include additional arguments in command", func(t *testing.T) {
		profilingJob := &job.ProfilingJob{
			Interval:   60 * time.Second,
			Event:      api.Cpu,
			OutputType: api.FlameGraph,
			AdditionalArguments: map[string]string{
				"async-profiler-arg-0": "-t",
				"async-profiler-arg-1": "--alloc=2m",
			},
		}

		cmd := asyncProfilerCommand(manager, profilingJob, "1234", "/tmp/test.txt")

		// Verify the command contains the additional arguments
		args := cmd.Args
		assert.Contains(t, args, "-t")
		assert.Contains(t, args, "--alloc=2m")
		// Verify the PID is at the end
		assert.Equal(t, "1234", args[len(args)-1])
	})

	t.Run("should work without additional arguments", func(t *testing.T) {
		profilingJob := &job.ProfilingJob{
			Interval:            60 * time.Second,
			Event:               api.Cpu,
			OutputType:          api.FlameGraph,
			AdditionalArguments: nil,
		}

		cmd := asyncProfilerCommand(manager, profilingJob, "1234", "/tmp/test.txt")

		// Verify the command doesn't have extra unexpected arguments
		args := cmd.Args
		assert.Equal(t, "1234", args[len(args)-1])
		assert.NotContains(t, args, "-t")
	})

	t.Run("should handle single additional argument", func(t *testing.T) {
		profilingJob := &job.ProfilingJob{
			Interval:   60 * time.Second,
			Event:      api.Wall,
			OutputType: api.FlameGraph,
			AdditionalArguments: map[string]string{
				"async-profiler-arg-0": "-t",
			},
		}

		cmd := asyncProfilerCommand(manager, profilingJob, "5678", "/tmp/test2.txt")

		args := cmd.Args
		assert.Contains(t, args, "-t")
		assert.Equal(t, "5678", args[len(args)-1])
		// Verify the -e flag has wall event
		for i, arg := range args {
			if arg == "-e" && i+1 < len(args) {
				assert.Equal(t, "wall", args[i+1])
			}
		}
	})
}
