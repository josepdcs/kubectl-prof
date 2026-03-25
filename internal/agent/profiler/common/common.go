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

// TmpDir provides the system's temporary directory as a string by invoking os.TempDir().
var TmpDir = func() string {
	return os.TempDir()
}

// GetResultFile constructs and returns the complete file path for a profiling result based on the given parameters.
func GetResultFile(targetDir string, tool api.ProfilingTool, outputType api.OutputType, pid string, iteration int) string {
	prefix := filepath.Join(targetDir, config.ProfilingPrefix)
	return fmt.Sprint(prefix, string(outputType), "-", pid, "-", iteration, GetFileExtension(tool, outputType))
}

// toolExtensionMap maps profiling tools to their supported output types and their corresponding file extensions.
var toolExtensionMap = map[api.ProfilingTool]map[api.OutputType]string{
	api.Jcmd: {
		api.Jfr:           ".jfr",
		api.ThreadDump:    ".txt",
		api.HeapHistogram: ".txt",
		api.Flat:          ".txt",
		api.Traces:        ".txt",
		api.Collapsed:     ".txt",
		api.SpeedScope:    ".txt",
		api.Raw:           ".txt",
		api.HeapDump:      ".hprof",
	},
	api.AsyncProfiler: {
		api.Jfr:           ".jfr",
		api.ThreadDump:    ".txt",
		api.HeapHistogram: ".txt",
		api.Flat:          ".txt",
		api.Traces:        ".txt",
		api.Collapsed:     ".txt",
		api.SpeedScope:    ".txt",
		api.Raw:           ".txt",
		api.HeapDump:      ".hprof",
	},
	api.Pyspy: {
		api.SpeedScope: ".json",
		api.ThreadDump: ".txt",
		api.Raw:        ".txt",
	},
	api.Rbspy: {
		api.SpeedScope:    ".json",
		api.Callgrind:     ".out",
		api.Summary:       ".txt",
		api.SummaryByLine: ".txt",
	},
	api.Bpf: {
		api.Raw: ".txt",
	},
	api.Btf: {
		api.Raw: ".txt",
	},
	api.Perf: {
		api.Raw: ".txt",
	},
	api.Phpspy: {
		api.Raw: ".txt",
	},
	api.NodeDummy: {
		api.HeapSnapshot: ".heapsnapshot",
	},
	api.DotnetTrace: {
		api.SpeedScope: ".json",
		api.Raw:        ".nettrace",
	},
	api.DotnetGcdump: {
		api.Gcdump: ".gcdump",
	},
	api.DotnetCounters: {
		api.Counters: ".json",
	},
	api.DotnetDump: {
		api.Dump: ".dmp",
	},
}

// toolDefaultExtensionMap maps profiling tools to their default file extensions.
// This is used when a specific output type does not have a defined extension in toolExtensionMap.
var toolDefaultExtensionMap = map[api.ProfilingTool]string{
	api.Jcmd:           ".html",
	api.AsyncProfiler:  ".html",
	api.Pyspy:          ".svg",
	api.Rbspy:          ".svg",
	api.Bpf:            ".svg",
	api.Btf:            ".svg",
	api.Perf:           ".svg",
	api.Phpspy:         ".svg",
	api.NodeDummy:      ".svg",
	api.DotnetTrace:    ".nettrace",
	api.DotnetGcdump:   ".gcdump",
	api.DotnetCounters: ".json",
	api.DotnetDump:     ".dmp",
}

// GetFileExtension returns the file extension for a given profiling tool and output type.
func GetFileExtension(tool api.ProfilingTool, outputType api.OutputType) string {
	if extensions, ok := toolExtensionMap[tool]; ok {
		if ext, ok := extensions[outputType]; ok {
			return ext
		}
	}
	if ext, ok := toolDefaultExtensionMap[tool]; ok {
		return ext
	}
	return ".svg"
}
