package profiler

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
	// Save the real implementation so after() functions can restore it.
	realHandleReportInMountNs := handleReportInMountNs

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
				var b bytes.Buffer
				b.Write([]byte("test"))
				file.Write(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw-1000-1.bin"), b.String())

				// Override handleReportInMountNs to write the result file directly,
				// avoiding /proc/<pid>/root access which is unavailable in tests.
				handleReportInMountNs = func(_ executil.Commander, _ *job.ProfilingJob, _, _, resultFileName string) error {
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
				assert.False(t, file.Exists(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw-1000-1.bin")))
				assert.True(t, fields.MemrayProfiler.MemrayManager.(*memrayManager).publisher.(*publish.Fake).On("Do").InvokedTimes() == 1)
			},
			after: func() {
				handleReportInMountNs = realHandleReportInMountNs
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw-1000-1.bin"))
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000-1.html"))
			},
		},
		{
			name: "should invoke for summary",
			given: func() (fields, args) {
				log.SetPrintLogs(true)

				commander := executil.NewMockCommander()
				// two separate calls: attach command then report command — each needs its own exec.Cmd
				commander.On("Command").Return(exec.Command("ls", common.TmpDir())).Once()
				commander.On("Command").Return(exec.Command("echo", "summary output")).Once()
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
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw-1000-1.bin"))
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"summary-1000-1.txt"))
			},
		},
		{
			name: "should invoke for tree",
			given: func() (fields, args) {
				log.SetPrintLogs(true)

				commander := executil.NewMockCommander()
				// two separate calls: attach command then report command — each needs its own exec.Cmd
				commander.On("Command").Return(exec.Command("ls", common.TmpDir())).Once()
				commander.On("Command").Return(exec.Command("echo", "tree output")).Once()
				publisher := publish.NewFakePublisher()
				publisher.On("Do").Return(nil)

				return fields{
						MemrayProfiler: NewMemrayProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.Tree,
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
				treeFile := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"tree-1000-1.txt")
				assert.True(t, file.Exists(treeFile))
				assert.Contains(t, file.Read(treeFile), "tree output")
				assert.True(t, fields.MemrayProfiler.MemrayManager.(*memrayManager).publisher.(*publish.Fake).On("Do").InvokedTimes() == 1)
			},
			after: func() {
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw-1000-1.bin"))
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"tree-1000-1.txt"))
			},
		},
		{
			name: "should invoke fail when command fails",
			given: func() (fields, args) {
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
		},
		{
			name: "should invoke fail when report command fails",
			given: func() (fields, args) {
				log.SetPrintLogs(true)

				commander := executil.NewMockCommander()
				// first command (attach) succeeds, second command (report) fails
				commander.On("Command").Return(exec.Command("ls", common.TmpDir())).Once()
				commander.On("Command").Return(&exec.Cmd{}).Once()
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
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw-1000-1.bin"))
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000-1.html"))
			},
		},
		{
			name: "should invoke fail when publish fails",
			given: func() (fields, args) {
				log.SetPrintLogs(true)
				var b bytes.Buffer
				b.Write([]byte("test"))
				file.Write(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw-1000-1.bin"), b.String())

				// Override handleReportInMountNs to write the result file directly.
				handleReportInMountNs = func(_ executil.Commander, _ *job.ProfilingJob, _, _, resultFileName string) error {
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
				handleReportInMountNs = realHandleReportInMountNs
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw-1000-1.bin"))
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

