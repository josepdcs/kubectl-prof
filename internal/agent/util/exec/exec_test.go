package exec

import (
	"testing"

	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommand(t *testing.T) {
	cmd := NewCommander()
	c := cmd.Command("ls", "/tmp")
	assert.NotEmpty(t, c)
}

func TestExecute(t *testing.T) {
	log.SetPrintLogs(true)
	exitCode, output, err := Execute(Command("ls", "/tmp"))
	assert.Equal(t, 0, exitCode)
	assert.NotEmpty(t, output)
	require.NoError(t, err)
}

func TestExecuteWhenError(t *testing.T) {
	log.SetPrintLogs(true)
	exitCode, output, err := Execute(Command("ls", "/other"))
	assert.Equal(t, 2, exitCode)
	assert.NotEmpty(t, output)
	require.Error(t, err)
}

func TestExecuteWhenSilentCommand(t *testing.T) {
	log.SetPrintLogs(true)
	exitCode, output, err := Silent().Execute(SilentCommand("ls", "/tmp"))
	assert.Equal(t, 0, exitCode)
	assert.NotEmpty(t, output)
	require.NoError(t, err)
}
