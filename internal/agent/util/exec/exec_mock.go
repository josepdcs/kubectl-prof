package exec

import (
    "os/exec"

    "github.com/stretchr/testify/mock"
)

// MockCommander is a testify-based mock implementing Commander
type MockCommander struct {
    mock.Mock
}

// NewMockCommander creates a new MockCommander
func NewMockCommander() *MockCommander {
    return &MockCommander{}
}

// Command mocks Commander.Command. Arguments are ignored for matching convenience in tests.
func (m *MockCommander) Command(name string, args ...string) *exec.Cmd {
    called := m.Called()
    if len(called) > 0 {
        if cmd, ok := called.Get(0).(*exec.Cmd); ok {
            return cmd
        }
    }
    return nil
}

// Execute mocks Commander.Execute. The provided *exec.Cmd argument is ignored for matching convenience.
func (m *MockCommander) Execute(cmd *exec.Cmd) (int, []byte, error) {
    called := m.Called()
    var (
        code int
        out []byte
        err error
    )
    if len(called) > 0 && called.Get(0) != nil {
        code, _ = called.Get(0).(int)
    }
    if len(called) > 1 && called.Get(1) != nil {
        out, _ = called.Get(1).([]byte)
    }
    if len(called) > 2 && called.Get(2) != nil {
        err, _ = called.Get(2).(error)
    }
    return code, out, err
}
