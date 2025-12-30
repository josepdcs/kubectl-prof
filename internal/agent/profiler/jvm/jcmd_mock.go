package jvm

import (
    "bytes"
    "time"

    "github.com/josepdcs/kubectl-prof/api"
    "github.com/josepdcs/kubectl-prof/internal/agent/job"
    "github.com/josepdcs/kubectl-prof/pkg/util/compressor"
    "github.com/stretchr/testify/mock"
)

// mockJcmdManager is a testify-based mock that implements JcmdManager
type mockJcmdManager struct {
    mock.Mock
}

func newMockJcmdManager() *mockJcmdManager {
    return &mockJcmdManager{}
}

func (m *mockJcmdManager) removeTmpDir() error {
    args := m.Called()
    if a := args.Get(0); a != nil {
        if err, ok := a.(error); ok {
            return err
        }
    }
    return nil
}

func (m *mockJcmdManager) linkTmpDirToTargetTmpDir(targetTmpDir string) error {
    args := m.Called(targetTmpDir)
    if a := args.Get(0); a != nil {
        if err, ok := a.(error); ok {
            return err
        }
    }
    return nil
}

func (m *mockJcmdManager) copyJfrSettingsToTmpDir() error {
    args := m.Called()
    if a := args.Get(0); a != nil {
        if err, ok := a.(error); ok {
            return err
        }
    }
    return nil
}

func (m *mockJcmdManager) invoke(j *job.ProfilingJob, pid string) (error, time.Duration) {
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

func (m *mockJcmdManager) handleProfilingResult(j *job.ProfilingJob, fileName string, out bytes.Buffer, targetPID string) error {
    args := m.Called(j, fileName, out, targetPID)
    if a := args.Get(0); a != nil {
        if err, ok := a.(error); ok {
            return err
        }
    }
    return nil
}

func (m *mockJcmdManager) handleJcmdRecording(targetPID string, iteration int, outputType string) {
    _ = m.Called(targetPID, iteration, outputType)
}

func (m *mockJcmdManager) publishResult(c compressor.Type, fileName string, outputType api.OutputType, heapDumpSplitInChunkSize string) error {
    args := m.Called(c, fileName, outputType, heapDumpSplitInChunkSize)
    if a := args.Get(0); a != nil {
        if err, ok := a.(error); ok {
            return err
        }
    }
    return nil
}

func (m *mockJcmdManager) cleanUp(j *job.ProfilingJob, pid string) {
    _ = m.Called(j, pid)
}
