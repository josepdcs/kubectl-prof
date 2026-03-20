package profiler

import (
	"time"

	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/flamegraph"
	"github.com/stretchr/testify/mock"
)

// mockPhpspyManager is a testify-based mock that implements PhpspyManager
type mockPhpspyManager struct {
	mock.Mock
}

func newMockPhpspyManager() *mockPhpspyManager {
	return &mockPhpspyManager{}
}

func (m *mockPhpspyManager) invoke(j *job.ProfilingJob, pid string) (error, time.Duration) {
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

func (m *mockPhpspyManager) collapsePhpspyOutput(j *job.ProfilingJob, pid string) (error, string) {
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

func (m *mockPhpspyManager) handleFlamegraph(j *job.ProfilingJob, fg flamegraph.FrameGrapher, raw string, out string) error {
	args := m.Called(j, fg, raw, out)
	if a := args.Get(0); a != nil {
		if err, ok := a.(error); ok {
			return err
		}
	}
	return nil
}
