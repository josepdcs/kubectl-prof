package api

import jsoniter "github.com/json-iterator/go"

type ProfilingTool string

const (
	asyncProfiler ProfilingTool = "async-profiler"
	jcmd          ProfilingTool = "jcmd"
	pyspy         ProfilingTool = "pyspy"
	bpf           ProfilingTool = "bpf"
	perf          ProfilingTool = "perf"
	rbspy         ProfilingTool = "rbspy"
)

var (
	ProfilingTools = []ProfilingTool{asyncProfiler, jcmd, pyspy, bpf, perf, rbspy}
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

//GetProfilingToolsByProgrammingLanguage Gets profiling tool related to the programming language.
//The first one is considered the default
var GetProfilingToolsByProgrammingLanguage = map[ProgrammingLanguage][]ProfilingTool{
	Java:   {asyncProfiler, jcmd},
	Python: {pyspy},
	Go:     {bpf},
	Node:   {bpf, perf},
	Ruby:   {rbspy},
}

func AvailableProfilingToolsString() string {
	out, _ := jsoniter.Marshal(GetProfilingToolsByProgrammingLanguage)
	return string(out)
}

//IsValidProfilingTool Identifies if given ProfilingTool is valid for the also given ProgrammingLanguage
func IsValidProfilingTool(tool ProfilingTool, language ProgrammingLanguage) bool {
	for _, current := range GetProfilingToolsByProgrammingLanguage[language] {
		if tool == current {
			return true
		}
	}
	return false
}
