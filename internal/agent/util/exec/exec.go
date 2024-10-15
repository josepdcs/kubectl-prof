package exec

import (
	"errors"
	"os/exec"
	"strings"

	"github.com/josepdcs/kubectl-prof/pkg/util/log"
)

type Commander interface {
	Command(name string, arg ...string) *exec.Cmd
	Execute(cmd *exec.Cmd) (int, []byte, error)
}

type commander struct {
	logCommand bool
}

func NewCommander() Commander {
	return &commander{
		logCommand: true,
	}
}

func NewSilentCommander() Commander {
	return &commander{
		logCommand: false,
	}
}

func (c commander) Command(name string, arg ...string) *exec.Cmd {
	var builder strings.Builder
	builder.WriteString(name)
	builder.WriteString(" ")
	for _, str := range arg {
		builder.WriteString(str)
		builder.WriteString(" ")
	}
	if c.logCommand {
		log.DebugLogLn(strings.TrimSpace(builder.String()))
	}

	return exec.Command(name, arg...)
}

func (c commander) Execute(cmd *exec.Cmd) (int, []byte, error) {
	exitCode := 0
	output, err := cmd.CombinedOutput()

	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			exitCode = exitError.ExitCode()
		}
	}

	return exitCode, output, err
}

var std = NewCommander()
var silent = NewSilentCommander()

// Default the default Commander
func Default() Commander {
	return std
}

// Silent the silent Commander
func Silent() Commander {
	return silent
}

func Command(name string, arg ...string) *exec.Cmd {
	return std.Command(name, arg...)
}

func SilentCommand(name string, arg ...string) *exec.Cmd {
	return silent.Command(name, arg...)
}

func Execute(cmd *exec.Cmd) (int, []byte, error) {
	return std.Execute(cmd)
}
