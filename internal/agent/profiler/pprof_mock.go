package profiler

import (
	"time"

	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/stretchr/testify/mock"
)

// mockPprofManager is a testify-based mock that implements PprofManager
type mockPprofManager struct {
	mock.Mock
}

func newMockPprofManager() *mockPprofManager {
	return &mockPprofManager{}
}

func (m *mockPprofManager) invoke(j *job.ProfilingJob) (error, time.Duration) {
	args := m.Called(j)
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
