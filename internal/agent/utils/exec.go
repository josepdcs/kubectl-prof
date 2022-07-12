package utils

import (
	"github.com/josepdcs/kubectl-prof/api"
	"os/exec"
	"strings"
)

func Command(name string, arg ...string) *exec.Cmd {
	var builder strings.Builder
	builder.WriteString(name)
	builder.WriteString(" ")
	for _, str := range arg {
		builder.WriteString(str)
		builder.WriteString(" ")
	}
	PublishLogEvent(api.DebugLevel, strings.TrimSpace(builder.String()))

	return exec.Command(name, arg...)
}

func ExecuteCommand(cmd *exec.Cmd) (int, string, error) {
	exitCode := 0
	output, err := cmd.CombinedOutput()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
	}

	return exitCode, string(output), err
}
