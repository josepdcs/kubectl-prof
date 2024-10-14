package common

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
)

const (
	MinimumRawSize = 50
)

var TmpDir = func() string {
	return os.TempDir()
}

func GetResultFile(targetDir string, tool api.ProfilingTool, outputType api.OutputType, pid string, iteration int) string {
	prefix := filepath.Join(targetDir, config.ProfilingPrefix)
	return fmt.Sprint(prefix, string(outputType), "-", pid, "-", iteration, GetFileExtension(tool, outputType))
}

func GetFileExtension(tool api.ProfilingTool, outputType api.OutputType) string {
	switch tool {
	case api.Jcmd, api.AsyncProfiler:
		switch outputType {
		case api.Jfr:
			return ".jfr"
		case api.ThreadDump, api.HeapHistogram, api.Flat, api.Traces, api.Collapsed, api.SpeedScope, api.Raw:
			return ".txt"
		case api.HeapDump:
			return ".hprof"
		default:
			// api.FlameGraph
			return ".html"
		}
	case api.Pyspy:
		switch outputType {
		case api.SpeedScope:
			return ".json"
		case api.ThreadDump, api.Raw:
			return ".txt"
		default:
			// api.FlameGraph
			return ".svg"
		}
	case api.Rbspy:
		switch outputType {
		case api.SpeedScope:
			return ".json"
		case api.Callgrind:
			return ".out"
		case api.Summary, api.SummaryByLine:
			return ".txt"
		default:
			// api.FlameGraph
			return ".svg"
		}
	case api.Bpf, api.Perf:
		switch outputType {
		case api.Raw:
			return ".txt"
		default:
			// api.FlameGraph
			return ".svg"
		}
	default:
		// api.FlameGraph
		return ".svg"
	}
}
