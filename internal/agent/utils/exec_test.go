package utils

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCommand(t *testing.T) {
	c := Command("ls", "/tmp")
	assert.NotEmpty(t, c)
}

func TestExecuteCommand(t *testing.T) {
	exitCode, output, err := ExecuteCommand(Command("ls", "/tmp"))
	assert.Equal(t, 0, exitCode)
	assert.NotEmpty(t, output)
	require.NoError(t, err)
}
