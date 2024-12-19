package api

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/samber/lo"
)

type ProfilingTool string

const (
	AsyncProfiler ProfilingTool = "async-profiler"
	Jcmd          ProfilingTool = "jcmd"
	Pyspy         ProfilingTool = "pyspy"
	Bpf           ProfilingTool = "bpf"
	Perf          ProfilingTool = "perf"
	Rbspy         ProfilingTool = "rbspy"
	NodeDummy     ProfilingTool = "node-dummy"
	FakeTool      ProfilingTool = "fake"
)

var (
	profilingTools = []ProfilingTool{AsyncProfiler, Jcmd, Pyspy, Bpf, Perf, Rbspy, NodeDummy}
)

func AvailableProfilingTools() []ProfilingTool {
	return profilingTools
}

func IsSupportedProfilingTool(profilingTool string) bool {
	return lo.Contains(AvailableProfilingTools(), ProfilingTool(profilingTool))
}

// GetProfilingTool Gets profiling tool related to the programming language and output event type.
var GetProfilingTool = func(l ProgrammingLanguage, o OutputType) ProfilingTool {
	switch l {
	case Java:
		switch o {
		case Jfr, ThreadDump, HeapDump, HeapHistogram:
			return Jcmd
		case FlameGraph, Flat, Traces, Collapsed, Tree, Raw:
			return AsyncProfiler
		}
	case Python:
		return Pyspy
	case Go, Rust:
		return Bpf
	case Clang, ClangPlusPlus:
		return Bpf
	case Ruby:
		return Rbspy
	case Node:
		switch o {
		case FlameGraph, Raw:
			return Bpf
		case HeapSnapshot, HeapDump:
			return NodeDummy
		}
	}

	// return the default according programming language
	return GetProfilingToolsByProgrammingLanguage[l][0]
}

// GetProfilingToolsByProgrammingLanguage Gets profiling tool related to the programming language.
// The first one is considered the default
var GetProfilingToolsByProgrammingLanguage = map[ProgrammingLanguage][]ProfilingTool{
	Java:          {Jcmd, AsyncProfiler},
	Python:        {Pyspy},
	Go:            {Bpf},
	Node:          {Bpf, Perf, NodeDummy},
	Clang:         {Bpf, Perf},
	ClangPlusPlus: {Bpf, Perf},
	Ruby:          {Rbspy},
	Rust:          {Bpf, Perf},
	FakeLang:      {FakeTool},
}

func AvailableProfilingToolsString() string {
	out, _ := jsoniter.Marshal(GetProfilingToolsByProgrammingLanguage)
	return string(out)
}

// IsValidProfilingTool Identifies if given ProfilingTool is valid for the also given ProgrammingLanguage
func IsValidProfilingTool(tool ProfilingTool, language ProgrammingLanguage) bool {
	for _, current := range GetProfilingToolsByProgrammingLanguage[language] {
		if tool == current {
			return true
		}
	}
	return false
}
