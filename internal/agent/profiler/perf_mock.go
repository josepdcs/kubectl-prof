package profiler

import (
	"time"

	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/flamegraph"
	"github.com/stretchr/testify/mock"
)

// mockPerfManager is a testify-based mock that implements PerfManager
type mockPerfManager struct {
	mock.Mock
}

func newMockPerfManager() *mockPerfManager {
	return &mockPerfManager{}
}

func (m *mockPerfManager) invoke(j *job.ProfilingJob, pid string) (error, time.Duration) {
	args := m.Called(j, pid)
	var err error
	var d time.Duration
	if a := args.Get(0); a != nil {
		err, _ = a.(error)
	}
	if a := args.Get(1); a != nil {
		d, _ = a.(time.Duration)
	}
	return err, d
}

func (m *mockPerfManager) runPerfRecord(j *job.ProfilingJob, pid string) error {
	args := m.Called(j, pid)
	if a := args.Get(0); a != nil {
		if err, ok := a.(error); ok {
			return err
		}
	}
	return nil
}

func (m *mockPerfManager) runPerfScript(j *job.ProfilingJob, pid string) error {
	args := m.Called(j, pid)
	if a := args.Get(0); a != nil {
		if err, ok := a.(error); ok {
			return err
		}
	}
	return nil
}

func (m *mockPerfManager) foldPerfOutput(j *job.ProfilingJob, pid string) (error, string) {
	args := m.Called(j, pid)
	var err error
	var out string
	if a := args.Get(0); a != nil {
		err, _ = a.(error)
	}
	if a := args.Get(1); a != nil {
		out, _ = a.(string)
	}
	return err, out
}

func (m *mockPerfManager) handleFlamegraph(j *job.ProfilingJob, fg flamegraph.FrameGrapher, raw string, out string) error {
	args := m.Called(j, fg, raw, out)
	if a := args.Get(0); a != nil {
		if err, ok := a.(error); ok {
			return err
		}
	}
	return nil
}
