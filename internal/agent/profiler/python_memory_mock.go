package profiler

import (
	"time"

	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/stretchr/testify/mock"
)

// mockMemrayManager is a testify-based mock that implements MemrayManager.
type mockMemrayManager struct {
	mock.Mock
}

func newMockMemrayManager() *mockMemrayManager {
	return &mockMemrayManager{}
}

func (m *mockMemrayManager) invoke(j *job.ProfilingJob, pid string) (error, time.Duration) {
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
