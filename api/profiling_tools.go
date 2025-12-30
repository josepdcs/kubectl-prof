package api

import (
	"encoding/json"

	"github.com/samber/lo"
)

// ProfilingTool represents a profiling tool
type ProfilingTool string

const (
	AsyncProfiler ProfilingTool = "async-profiler"   // AsyncProfiler is a profiling tool used primarily for Java applications to output various performance metrics.
	Jcmd          ProfilingTool = "jcmd"             // Jcmd is a profiling tool used primarily for Java applications to output various performance metrics.
	Pyspy         ProfilingTool = "pyspy"            // Pyspy is a profiling tool used primarily for Python applications to output various performance metrics.
	Bpf           ProfilingTool = "bpf"              // Bpf is a profiling tool used primarily for C/C++ applications to output various performance metrics.
	Perf          ProfilingTool = "perf"             // Perf is a profiling tool used primarily for Linux applications to output various performance metrics.
	Rbspy         ProfilingTool = "rbspy"            // Rbspy is a profiling tool used primarily for Ruby applications to output various performance metrics.
	NodeDummy     ProfilingTool = "node-dummy"       // NodeDummy is a profiling tool used primarily for Node.js applications to output various performance metrics.
	CargoFlame    ProfilingTool = "cargo-flamegraph" // CargoFlame is a profiling tool used primarily for Rust applications to output various performance metrics.
	FakeTool      ProfilingTool = "fake"             // FakeTool is a profiling tool used primarily for testing purposes.
)

var (
	// profilingTools contains all supported profiling tools.
	profilingTools = []ProfilingTool{AsyncProfiler, Jcmd, Pyspy, Bpf, Perf, Rbspy, NodeDummy, CargoFlame}
)

// AvailableProfilingTools returns the list of all available profiling tools.
func AvailableProfilingTools() []ProfilingTool {
	return profilingTools
}

// IsSupportedProfilingTool checks if the given profiling tool string is supported.
// It returns true if the profiling tool is in the list of available profiling tools.
func IsSupportedProfilingTool(profilingTool string) bool {
	return lo.Contains(AvailableProfilingTools(), ProfilingTool(profilingTool))
}

// GetProfilingTool returns the appropriate profiling tool for the given programming language and output type.
// It selects the most suitable tool based on the combination of language and desired output format.
// If no specific match is found, it returns the default tool for the language.
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
	case Go:
		return Bpf
	case Rust:
		return CargoFlame
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

// GetProfilingToolsByProgrammingLanguage maps each ProgrammingLanguage to its supported ProfilingTools.
// The first tool in each slice is considered the default for that language.
var GetProfilingToolsByProgrammingLanguage = map[ProgrammingLanguage][]ProfilingTool{
	Java:          {Jcmd, AsyncProfiler},
	Python:        {Pyspy},
	Go:            {Bpf, CargoFlame},
	Node:          {Bpf, Perf, NodeDummy, CargoFlame},
	Clang:         {Bpf, Perf, CargoFlame},
	ClangPlusPlus: {Bpf, Perf, CargoFlame},
	Ruby:          {Rbspy},
	Rust:          {CargoFlame, Bpf, Perf},
	FakeLang:      {FakeTool},
}

// AvailableProfilingToolsString returns a JSON string representation of all profiling tools by programming language.
func AvailableProfilingToolsString() string {
	out, _ := json.Marshal(GetProfilingToolsByProgrammingLanguage)
	return string(out)
}

// IsValidProfilingTool checks if the given ProfilingTool is valid for the specified ProgrammingLanguage.
// It returns true if the tool is supported by the programming language.
func IsValidProfilingTool(tool ProfilingTool, language ProgrammingLanguage) bool {
	for _, current := range GetProfilingToolsByProgrammingLanguage[language] {
		if tool == current {
			return true
		}
	}
	return false
}
