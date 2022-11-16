package common

import (
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"path/filepath"
)

var TmpDir = func() string {
	return "/tmp"
}

func GetResultFile(targetDir string, job *job.ProfilingJob) string {
	prefix := filepath.Join(targetDir, config.ProfilingPrefix)
	return prefix + string(job.OutputType) + GetFileExtension(job.Tool, job.OutputType)
}

func GetFileExtension(tool api.ProfilingTool, OutputType api.EventType) string {
	switch tool {
	case api.Jcmd, api.AsyncProfiler:
		switch OutputType {
		case api.Jfr:
			return ".jfr"
		case api.ThreadDump, api.HeapHistogram, api.Flat, api.Traces, api.Collapsed:
			return ".txt"
		case api.HeapDump:
			return ".hprof"
		default:
			// api.FlameGraph
			return ".html"
		}
	case api.Pyspy:
		switch OutputType {
		case api.SpeedScope:
			return ".json"
		case api.ThreadDump:
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
