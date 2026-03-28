package profiler

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func Test_detectTargetPythonVersion_regexes(t *testing.T) {
	// Test libpythonRe against various maps content
	t.Run("libpythonRe matches", func(t *testing.T) {
		tests := []struct {
			input string
			want  string
		}{
			{"/usr/lib/x86_64-linux-gnu/libpython3.11.so.1.0", "3.11"},
			{"/usr/lib/libpython3.10.so", "3.10"},
			{"/usr/lib/libpython3.13.so.1.0", "3.13"},
			{"/usr/lib/libpython3.9.so", "3.9"},
		}
		for _, tt := range tests {
			m := libpythonRe.FindStringSubmatch(tt.input)
			require.NotNil(t, m, "expected match for %q", tt.input)
			assert.Equal(t, tt.want, m[1])
		}
	})

	t.Run("libpythonRe does not match non-python libs", func(t *testing.T) {
		inputs := []string{
			"/usr/lib/libc.so.6",
			"/usr/lib/libssl.so.1.1",
			"/usr/bin/python3",
		}
		for _, input := range inputs {
			m := libpythonRe.FindStringSubmatch(input)
			assert.Nil(t, m, "expected no match for %q", input)
		}
	})

	// Test pythonBinaryVersionRe against various binary names
	t.Run("pythonBinaryVersionRe matches", func(t *testing.T) {
		tests := []struct {
			input string
			want  string
		}{
			{"python3.11", "3.11"},
			{"python3.10", "3.10"},
			{"python3.13", "3.13"},
		}
		for _, tt := range tests {
			m := pythonBinaryVersionRe.FindStringSubmatch(tt.input)
			require.NotNil(t, m, "expected match for %q", tt.input)
			assert.Equal(t, tt.want, m[1])
		}
	})

	t.Run("pythonBinaryVersionRe does not match versionless names", func(t *testing.T) {
		inputs := []string{
			"python3",
			"python",
			"node",
		}
		for _, input := range inputs {
			m := pythonBinaryVersionRe.FindStringSubmatch(input)
			assert.Nil(t, m, "expected no match for %q", input)
		}
	})
}

func Test_detectTargetPythonVersion(t *testing.T) {
	realDetect := detectTargetPythonVersion
	defer func() { detectTargetPythonVersion = realDetect }()

	t.Run("should return error for non-existent PID", func(t *testing.T) {
		// Use the real implementation with a PID that doesn't exist
		_, err := realDetect("999999999")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not detect Python version")
		assert.Contains(t, err.Error(), "statically-linked Python not supported")
	})

	t.Run("should detect version from maps via override", func(t *testing.T) {
		detectTargetPythonVersion = func(pid string) (string, error) {
			return "3.11", nil
		}
		ver, err := detectTargetPythonVersion("1234")
		require.NoError(t, err)
		assert.Equal(t, "3.11", ver)
	})

	t.Run("should propagate error via override", func(t *testing.T) {
		detectTargetPythonVersion = func(pid string) (string, error) {
			return "", fmt.Errorf("could not detect Python version for PID %s: statically-linked Python not supported", pid)
		}
		_, err := detectTargetPythonVersion("1234")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "statically-linked Python not supported")
	})
}

func TestMemrayProfiler_SetUp(t *testing.T) {
	type fields struct {
		MemrayProfiler *MemrayProfiler
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
						MemrayProfiler: &MemrayProfiler{
							MemrayManager: newMockMemrayManager(),
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
				return fields.MemrayProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
				assert.Equal(t, []string{"PID_ContainerID"}, fields.MemrayProfiler.targetPIDs)
			},
		},
		{
			name: "should setup with given PID",
			given: func() (fields, args) {
				return fields{
						MemrayProfiler: &MemrayProfiler{
							MemrayManager: newMockMemrayManager(),
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
				return fields.MemrayProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
				assert.Equal(t, []string{"PID_ContainerID"}, fields.MemrayProfiler.targetPIDs)
			},
		},
		{
			name: "should fail when container PID not found",
			given: func() (fields, args) {
				return fields{
						MemrayProfiler: &MemrayProfiler{
							MemrayManager: newMockMemrayManager(),
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
				return fields.MemrayProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.NotNil(t, err)
				assert.Empty(t, fields.MemrayProfiler.targetPIDs)
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

func TestMemrayProfiler_Invoke(t *testing.T) {
	type fields struct {
		MemrayProfiler *MemrayProfiler
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
				memrayManager := newMockMemrayManager()
				memrayManager.On("invoke", mock.Anything, mock.AnythingOfType("string")).
					Return(nil, time.Duration(0)).
					Twice()

				return fields{
						MemrayProfiler: &MemrayProfiler{
							MemrayManager: memrayManager,
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
				fields.MemrayProfiler.delay = 0
				fields.MemrayProfiler.targetPIDs = []string{"1000", "2000"}
				return fields.MemrayProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
				fields.MemrayProfiler.MemrayManager.(*mockMemrayManager).AssertNumberOfCalls(t, "invoke", 2)
			},
		},
		{
			name: "should invoke fail when invoke fails",
			given: func() (fields, args) {
				memrayManager := newMockMemrayManager()
				memrayManager.On("invoke", mock.Anything, mock.AnythingOfType("string")).
					Return(errors.New("fake invoke error"), time.Duration(0)).
					Once()

				return fields{
						MemrayProfiler: &MemrayProfiler{
							MemrayManager: memrayManager,
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
				fields.MemrayProfiler.delay = 0
				fields.MemrayProfiler.targetPIDs = []string{"1000"}
				return fields.MemrayProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				require.Error(t, err)
				assert.EqualError(t, err, "fake invoke error")
				fields.MemrayProfiler.MemrayManager.(*mockMemrayManager).AssertNumberOfCalls(t, "invoke", 1)
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

func TestMemrayProfiler_CleanUp(t *testing.T) {
	type fields struct {
		MemrayProfiler *MemrayProfiler
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
				f := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.html")
				_, _ = os.Create(f)
				_, _ = os.Create(f + compressor.GetExtensionFileByCompressor[compressor.Gzip])
				return fields{
						MemrayProfiler: &MemrayProfiler{
							MemrayManager: newMockMemrayManager(),
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
				return fields.MemrayProfiler.CleanUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				f := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.html")
				g := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.html"+
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

func Test_memrayManager_invoke(t *testing.T) {
	// Save real implementations so after() functions can restore them.
	realHandleReport := handleReport
	realStageMemrayPackage := stageMemrayPackage
	realCopyDir := copyDir
	realRawFileInTargetPath := rawFileInTargetPath
	realCheckRawFileExists := checkRawFileExists
	realCopyRawFileFromTarget := copyRawFileFromTarget
	realPidExists := pidExists
	realCopyFile := copyFile

	// Default test overrides: no-op stageMemrayPackage, use local path for raw file, always find raw file,
	// no-op copy (raw file is already local in tests).
	stubTestHelpers := func() {
		stageMemrayPackage = func(_ string) (func(), error) { return func() {}, nil }
		copyDir = func(_, _ string) error { return nil }
		rawFileInTargetPath = func(_ string, rawFileName string) string { return rawFileName }
		checkRawFileExists = func(_ string) (os.FileInfo, error) { return nil, nil }
		copyRawFileFromTarget = func(src string, dst string) error { return nil }
		pidExists = func(_ string) bool { return true }
		copyFile = realCopyFile
	}
	restoreTestHelpers := func() {
		handleReport = realHandleReport
		stageMemrayPackage = realStageMemrayPackage
		copyDir = realCopyDir
		rawFileInTargetPath = realRawFileInTargetPath
		checkRawFileExists = realCheckRawFileExists
		copyRawFileFromTarget = realCopyRawFileFromTarget
		pidExists = realPidExists
		copyFile = realCopyFile
	}

	type fields struct {
		MemrayProfiler *MemrayProfiler
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
			name: "should invoke for flamegraph",
			given: func() (fields, args) {
				log.SetPrintLogs(true)
				stubTestHelpers()
				var b bytes.Buffer
				b.Write([]byte("test"))
				file.Write(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"memray-raw-1000-1.bin"), b.String())

				// Override handleReport to write the result file directly,
				// avoiding /proc/<pid>/root access which is unavailable in tests.
				handleReport = func(_ executil.Commander, _ *job.ProfilingJob, _, _, resultFileName string) error {
					return os.WriteFile(resultFileName, []byte("test flamegraph"), 0600)
				}
				commander := executil.NewMockCommander()
				// attach command only — report is handled by the overridden var
				commander.On("Command").Return(exec.Command("ls", common.TmpDir())).Once()
				publisher := publish.NewFakePublisher()
				publisher.On("Do").Return(nil)

				return fields{
						MemrayProfiler: NewMemrayProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.FakeLang,
							Tool:             api.Memray,
							Compressor:       compressor.None,
							Iteration:        1,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.MemrayProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				assert.Nil(t, err)
				assert.False(t, file.Exists(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"memray-raw-1000-1.bin")))
				assert.True(t, fields.MemrayProfiler.MemrayManager.(*memrayManager).publisher.(*publish.Fake).On("Do").InvokedTimes() == 1)
			},
			after: func() {
				restoreTestHelpers()
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"memray-raw-1000-1.bin"))
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000-1.html"))
			},
		},
		{
			name: "should invoke for summary",
			given: func() (fields, args) {
				log.SetPrintLogs(true)
				stubTestHelpers()
				file.Write(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"memray-raw-1000-1.bin"), "test")

				// Override handleReport to write report file directly.
				handleReport = func(_ executil.Commander, _ *job.ProfilingJob, _, _, resultFileName string) error {
					return os.WriteFile(resultFileName, []byte("summary output"), 0600)
				}
				commander := executil.NewMockCommander()
				commander.On("Command").Return(exec.Command("ls", common.TmpDir())).Once()
				publisher := publish.NewFakePublisher()
				publisher.On("Do").Return(nil)

				return fields{
						MemrayProfiler: NewMemrayProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.Summary,
							Language:         api.FakeLang,
							Tool:             api.Memray,
							Compressor:       compressor.None,
							Iteration:        1,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.MemrayProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				assert.Nil(t, err)
				summaryFile := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"summary-1000-1.txt")
				assert.True(t, file.Exists(summaryFile))
				assert.Contains(t, file.Read(summaryFile), "summary output")
				assert.True(t, fields.MemrayProfiler.MemrayManager.(*memrayManager).publisher.(*publish.Fake).On("Do").InvokedTimes() == 1)
			},
			after: func() {
				restoreTestHelpers()
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"memray-raw-1000-1.bin"))
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"summary-1000-1.txt"))
			},
		},
		{
			name: "should skip PID that exits during staging",
			given: func() (fields, args) {
				stubTestHelpers()
				// Stage fails because PID is gone, and pidExists confirms it.
				stageMemrayPackage = func(_ string) (func(), error) {
					return nil, fmt.Errorf("could not detect target Python version")
				}
				pidExists = func(_ string) bool { return false }
				commander := executil.NewMockCommander()
				publisher := publish.NewFakePublisher()
				publisher.On("Do").Return(nil)

				return fields{
						MemrayProfiler: NewMemrayProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.FakeLang,
							Tool:             api.Memray,
							Compressor:       compressor.None,
							Iteration:        1,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.MemrayProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				assert.Nil(t, err, "should skip exited PID, not return error")
				assert.True(t, fields.MemrayProfiler.MemrayManager.(*memrayManager).publisher.(*publish.Fake).On("Do").InvokedTimes() == 0)
			},
			after: func() {
				restoreTestHelpers()
			},
		},
		{
			name: "should fail staging when PID still exists",
			given: func() (fields, args) {
				stubTestHelpers()
				stageMemrayPackage = func(_ string) (func(), error) {
					return nil, fmt.Errorf("could not stage memray package: permission denied")
				}
				pidExists = func(_ string) bool { return true }
				commander := executil.NewMockCommander()
				publisher := publish.NewFakePublisher()
				publisher.On("Do").Return(nil)

				return fields{
						MemrayProfiler: NewMemrayProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.FakeLang,
							Tool:             api.Memray,
							Compressor:       compressor.None,
							Iteration:        1,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.MemrayProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				require.Error(t, err, "should fail when PID exists but staging fails")
				assert.Contains(t, err.Error(), "could not stage memray package")
			},
			after: func() {
				restoreTestHelpers()
			},
		},
		{
			name: "should invoke fail when command fails",
			given: func() (fields, args) {
				stubTestHelpers()
				commander := executil.NewMockCommander()
				commander.On("Command").Return(&exec.Cmd{})
				publisher := publish.NewFakePublisher()
				publisher.On("Do").Return(nil)

				return fields{
						MemrayProfiler: NewMemrayProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.Python,
							Tool:             api.Memray,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.MemrayProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				require.Error(t, err)
				assert.True(t, fields.MemrayProfiler.MemrayManager.(*memrayManager).publisher.(*publish.Fake).On("Do").InvokedTimes() == 0)
			},
			after: func() {
				restoreTestHelpers()
			},
		},
		{
			name: "should invoke fail when report command fails",
			given: func() (fields, args) {
				log.SetPrintLogs(true)
				stubTestHelpers()
				file.Write(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"memray-raw-1000-1.bin"), "test")

				// Override handleReport to simulate report failure.
				handleReport = func(_ executil.Commander, _ *job.ProfilingJob, _, _, _ string) error {
					return errors.New("report generation failed")
				}
				commander := executil.NewMockCommander()
				commander.On("Command").Return(exec.Command("ls", common.TmpDir())).Once()
				publisher := publish.NewFakePublisher()
				publisher.On("Do").Return(nil)

				return fields{
						MemrayProfiler: NewMemrayProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.FakeLang,
							Tool:             api.Memray,
							Compressor:       compressor.None,
							Iteration:        1,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.MemrayProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				require.Error(t, err)
				assert.True(t, fields.MemrayProfiler.MemrayManager.(*memrayManager).publisher.(*publish.Fake).On("Do").InvokedTimes() == 0)
			},
			after: func() {
				restoreTestHelpers()
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"memray-raw-1000-1.bin"))
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000-1.html"))
			},
		},
		{
			name: "should invoke fail when raw file copy from target fails",
			given: func() (fields, args) {
				log.SetPrintLogs(true)
				stubTestHelpers()
				file.Write(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"memray-raw-1000-1.bin"), "test")

				// Override handleReport to not be called — instead override copyRawFileFromTarget to fail.
				handleReport = realHandleReport
				copyRawFileFromTarget = func(src string, dst string) error {
					return fmt.Errorf("could not read raw file from target at %s: permission denied", src)
				}
				commander := executil.NewMockCommander()
				commander.On("Command").Return(exec.Command("ls", common.TmpDir())).Once()
				publisher := publish.NewFakePublisher()
				publisher.On("Do").Return(nil)

				return fields{
						MemrayProfiler: NewMemrayProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.FakeLang,
							Tool:             api.Memray,
							Compressor:       compressor.None,
							Iteration:        1,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.MemrayProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				require.Error(t, err)
				assert.ErrorContains(t, err, "could not read raw file from target")
				assert.True(t, fields.MemrayProfiler.MemrayManager.(*memrayManager).publisher.(*publish.Fake).On("Do").InvokedTimes() == 0)
			},
			after: func() {
				restoreTestHelpers()
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"memray-raw-1000-1.bin"))
			},
		},
		{
			name: "should invoke fail when publish fails",
			given: func() (fields, args) {
				log.SetPrintLogs(true)
				stubTestHelpers()
				file.Write(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"memray-raw-1000-1.bin"), "test")

				// Override handleReport to write the result file directly.
				handleReport = func(_ executil.Commander, _ *job.ProfilingJob, _, _, resultFileName string) error {
					return os.WriteFile(resultFileName, []byte("test flamegraph"), 0600)
				}

				commander := executil.NewMockCommander()
				// attach command only — report is handled by the overridden var
				commander.On("Command").Return(exec.Command("ls", common.TmpDir())).Once()
				publisher := publish.NewFakePublisher()
				publisher.On("Do").Return(errors.New("fake publisher with error"))

				return fields{
						MemrayProfiler: NewMemrayProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.FakeLang,
							Tool:             api.Memray,
							Compressor:       compressor.None,
							Iteration:        1,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.MemrayProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				require.Error(t, err)
				assert.ErrorContains(t, err, "fake publisher with error")
				assert.True(t, fields.MemrayProfiler.MemrayManager.(*memrayManager).publisher.(*publish.Fake).On("Do").InvokedTimes() == 1)
			},
			after: func() {
				restoreTestHelpers()
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"memray-raw-1000-1.bin"))
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000-1.html"))
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

func Test_copyDir(t *testing.T) {
	t.Run("should copy directory tree", func(t *testing.T) {
		// Create source directory with nested files.
		srcDir := filepath.Join(os.TempDir(), "test-copydir-src")
		dstDir := filepath.Join(os.TempDir(), "test-copydir-dst")
		defer os.RemoveAll(srcDir)
		defer os.RemoveAll(dstDir)

		require.NoError(t, os.MkdirAll(filepath.Join(srcDir, "sub"), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("file-a"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(srcDir, "sub", "b.txt"), []byte("file-b"), 0644))

		// Copy.
		err := copyDir(srcDir, dstDir)
		require.NoError(t, err)

		// Verify.
		data, err := os.ReadFile(filepath.Join(dstDir, "a.txt"))
		require.NoError(t, err)
		assert.Equal(t, "file-a", string(data))

		data, err = os.ReadFile(filepath.Join(dstDir, "sub", "b.txt"))
		require.NoError(t, err)
		assert.Equal(t, "file-b", string(data))
	})

	t.Run("should return error for non-existent source", func(t *testing.T) {
		err := copyDir("/non/existent/path", "/tmp/dst")
		require.Error(t, err)
	})
}

func Test_stageMemrayPackage(t *testing.T) {
	realStageMemrayPackage := stageMemrayPackage
	realDetect := detectTargetPythonVersion
	realCopyDir := copyDir
	defer func() {
		stageMemrayPackage = realStageMemrayPackage
		detectTargetPythonVersion = realDetect
		copyDir = realCopyDir
	}()

	t.Run("should stage successfully via override", func(t *testing.T) {
		cleanupCalled := false
		stageMemrayPackage = func(pid string) (func(), error) {
			assert.Equal(t, "1234", pid)
			return func() { cleanupCalled = true }, nil
		}

		cleanup, err := stageMemrayPackage("1234")
		require.NoError(t, err)
		require.NotNil(t, cleanup)
		cleanup()
		assert.True(t, cleanupCalled)
	})

	t.Run("should return error when Python version detection fails", func(t *testing.T) {
		stageMemrayPackage = func(pid string) (func(), error) {
			return nil, fmt.Errorf("could not detect target Python version for memray staging: version unknown")
		}

		_, err := stageMemrayPackage("1234")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not detect target Python version")
	})

	t.Run("should select correct version based on target Python", func(t *testing.T) {
		var selectedVersion string
		stageMemrayPackage = func(pid string) (func(), error) {
			ver, err := detectTargetPythonVersion(pid)
			if err != nil {
				return nil, err
			}
			selectedVersion = ver
			return func() {}, nil
		}
		detectTargetPythonVersion = func(pid string) (string, error) {
			return "3.11", nil
		}

		cleanup, err := stageMemrayPackage("1234")
		require.NoError(t, err)
		require.NotNil(t, cleanup)
		assert.Equal(t, "3.11", selectedVersion)
	})

	t.Run("cleanup should remove staged directory", func(t *testing.T) {
		stagingDir := filepath.Join(os.TempDir(), "test-staging-cleanup")
		require.NoError(t, os.MkdirAll(filepath.Join(stagingDir, "memray"), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(stagingDir, "memray", "__init__.py"), []byte("# memray"), 0644))

		stageMemrayPackage = func(pid string) (func(), error) {
			return func() {
				_ = os.RemoveAll(stagingDir)
			}, nil
		}

		cleanup, err := stageMemrayPackage("1234")
		require.NoError(t, err)
		cleanup()

		_, err = os.Stat(stagingDir)
		assert.True(t, os.IsNotExist(err), "staging directory should be removed after cleanup")
	})

	t.Run("should propagate copyDir errors", func(t *testing.T) {
		stageMemrayPackage = func(pid string) (func(), error) {
			err := copyDir("/non/existent/src", "/tmp/dst")
			if err != nil {
				return nil, fmt.Errorf("could not stage memray package: %w", err)
			}
			return func() {}, nil
		}
		copyDir = realCopyDir

		_, err := stageMemrayPackage("1234")
		require.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "could not stage memray package"))
	})
}
