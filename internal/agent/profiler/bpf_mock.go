package profiler

import (
    "time"

    "github.com/josepdcs/kubectl-prof/internal/agent/job"
    "github.com/josepdcs/kubectl-prof/internal/agent/util/flamegraph"
    "github.com/stretchr/testify/mock"
)

// mockBpfManager is a testify-based mock that implements BpfManager
type mockBpfManager struct {
    mock.Mock
}

func newMockBpfManager() *mockBpfManager {
    return &mockBpfManager{}
}

func (m *mockBpfManager) invoke(j *job.ProfilingJob, pid string) (error, time.Duration) {
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

func (m *mockBpfManager) handleFlamegraph(j *job.ProfilingJob, fg flamegraph.FrameGrapher, raw string, out string) error {
    args := m.Called(j, fg, raw, out)
    if a := args.Get(0); a != nil {
        if err, ok := a.(error); ok {
            return err
        }
    }
    return nil
}
