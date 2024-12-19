package api

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/samber/lo"
)

type OutputType string

const (
	FlameGraph    OutputType = "flamegraph"
	SpeedScope    OutputType = "speedscope"
	Jfr           OutputType = "jfr"
	ThreadDump    OutputType = "threaddump"
	HeapDump      OutputType = "heapdump"
	HeapHistogram OutputType = "heaphistogram"
	Flat          OutputType = "flat"
	Traces        OutputType = "traces"
	Collapsed     OutputType = "collapsed"
	Tree          OutputType = "tree"
	Callgrind     OutputType = "callgrind"
	Raw           OutputType = "raw"
	Pprof         OutputType = "pprof"
	Summary       OutputType = "summary"
	SummaryByLine OutputType = "summary-by-line"
	HeapSnapshot  OutputType = "heapsnapshot"
)

// GetOutputTypesByProfilingTool Gets the list of OutputType related to the ProfilingTool that they will be considered as output types.
// The first one is considered the default
var GetOutputTypesByProfilingTool = map[ProfilingTool][]OutputType{
	AsyncProfiler: {FlameGraph, Jfr, Flat, Traces, Collapsed, Tree, Raw},
	Jcmd:          {Jfr, ThreadDump, HeapDump, HeapHistogram},
	Pyspy:         {FlameGraph, SpeedScope, ThreadDump, Raw},
	Bpf:           {FlameGraph, Raw},
	Perf:          {FlameGraph, Raw},
	Rbspy:         {FlameGraph, SpeedScope, Callgrind, Summary, SummaryByLine},
	NodeDummy:     {HeapSnapshot, HeapDump},
	FakeTool:      {FlameGraph},
}

func AvailableOutputTypesString() string {
	out, _ := jsoniter.Marshal(GetOutputTypesByProfilingTool)
	return string(out)
}

var (
	supportedOutputTypes = []OutputType{
		FlameGraph,
		SpeedScope,
		Jfr,
		ThreadDump,
		HeapDump,
		HeapHistogram,
		Flat,
		Traces,
		Collapsed,
		Tree,
		Callgrind,
		Raw,
		Summary,
		SummaryByLine,
		HeapSnapshot,
	}
)

func AvailableOutputTypes() []OutputType {
	return supportedOutputTypes
}

func IsSupportedOutputType(outputType string) bool {
	return lo.Contains(AvailableOutputTypes(), OutputType(outputType))
}

// IsValidOutputType Identifies if given OutputType is valid for the also given ProfilingTool
func IsValidOutputType(OutputType OutputType, profilingTool ProfilingTool) bool {
	for _, current := range GetOutputTypesByProfilingTool[profilingTool] {
		if OutputType == current {
			return true
		}
	}
	return false
}
