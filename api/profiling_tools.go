package api

import jsoniter "github.com/json-iterator/go"

type ProfilingTool string

const (
	AsyncProfiler ProfilingTool = "async-profiler"
	Jcmd          ProfilingTool = "jcmd"
	Pyspy         ProfilingTool = "pyspy"
	Bpf           ProfilingTool = "bpf"
	Perf          ProfilingTool = "perf"
	Rbspy         ProfilingTool = "rbspy"
	FakeTool      ProfilingTool = "fake"
)

var (
	ProfilingTools = []ProfilingTool{AsyncProfiler, Jcmd, Pyspy, Bpf, Perf, Rbspy}
)

func AvailableProfilingTools() []ProfilingTool {
	return ProfilingTools
}

func IsSupportedProfilingTool(profilingTool string) bool {
	return containsProfilingTool(ProfilingTool(profilingTool), AvailableProfilingTools())
}

func containsProfilingTool(profilingTool ProfilingTool, profilingTools []ProfilingTool) bool {
	for _, current := range profilingTools {
		if profilingTool == current {
			return true
		}
	}
	return false
}

// GetProfilingTool Gets profiling tool related to the programming language and output event type.
var GetProfilingTool = func(l ProgrammingLanguage, e EventType) ProfilingTool {
	switch l {
	case Java:
		switch e {
		case Jfr, ThreadDump, HeapDump, HeapHistogram:
			return Jcmd
		case FlameGraph, Flat, Traces, Collapsed, Tree:
			return AsyncProfiler
		}
	case Python:
		return Pyspy
	case Go, Node, Clang, ClangPlusPlus:
		return Bpf
	case Ruby:
		return Rbspy
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
	Node:          {Bpf, Perf},
	Clang:         {Bpf, Perf},
	ClangPlusPlus: {Bpf, Perf},
	Ruby:          {Rbspy},
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
