package profiler

import (
	"bufio"
	"github.com/agrison/go-commons-lang/stringUtils"
	"strings"
)

// addProcessPIDLegend adds the process PID to each line of the input string and returns the result.
// If the input string is empty, it returns the input string.
// If an error occurs while scanning the input string, it returns the input string.
func addProcessPIDLegend(input string, pid string) string {
	if stringUtils.IsBlank(input) {
		return input
	}
	var sb strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		sb.WriteString("process: ")
		sb.WriteString(pid)
		sb.WriteString(";")
		sb.WriteString(scanner.Text())
		sb.WriteString("\n")
	}
	// if scanner encountered an error, return the input
	if err := scanner.Err(); err != nil {
		return input
	}
	return sb.String()
}
