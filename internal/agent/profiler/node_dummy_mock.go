package profiler

import (
    "time"

    "github.com/josepdcs/kubectl-prof/internal/agent/job"
    "github.com/stretchr/testify/mock"
)

// mockNodeDummyManager is a testify-based mock that implements NodeDummyManager
type mockNodeDummyManager struct {
    mock.Mock
}

func newMockNodeDummyManager() *mockNodeDummyManager {
    return &mockNodeDummyManager{}
}

func (m *mockNodeDummyManager) invoke(j *job.ProfilingJob, pid string, cwd string) (error, time.Duration) {
    args := m.Called(j, pid, cwd)
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
