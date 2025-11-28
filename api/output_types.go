package api

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/samber/lo"
)

// OutputType represents the format of the profiling output.
type OutputType string

const (
	FlameGraph    OutputType = "flamegraph"      // FlameGraph represents a flame graph visualization format.
	SpeedScope    OutputType = "speedscope"      // SpeedScope represents a speedscope visualization format.
	Jfr           OutputType = "jfr"             // Jfr represents Java Flight Recorder format.
	ThreadDump    OutputType = "threaddump"      // ThreadDump represents a thread dump output.
	HeapDump      OutputType = "heapdump"        // HeapDump represents a heap dump output.
	HeapHistogram OutputType = "heaphistogram"   // HeapHistogram represents a heap histogram output.
	Flat          OutputType = "flat"            // Flat represents a flat profile format.
	Traces        OutputType = "traces"          // Traces represents trace data format.
	Collapsed     OutputType = "collapsed"       // Collapsed represents collapsed stack traces format.
	Tree          OutputType = "tree"            // Tree represents a tree-based profile format.
	Callgrind     OutputType = "callgrind"       // Callgrind represents callgrind format for use with kcachegrind.
	Raw           OutputType = "raw"             // Raw represents raw profiling data.
	Pprof         OutputType = "pprof"           // Pprof represents Go pprof format.
	Summary       OutputType = "summary"         // Summary represents a summary report.
	SummaryByLine OutputType = "summary-by-line" // SummaryByLine represents a line-by-line summary report.
	HeapSnapshot  OutputType = "heapsnapshot"    // HeapSnapshot represents a heap snapshot for Node.js applications.
)

// GetOutputTypesByProfilingTool maps each ProfilingTool to its supported output types.
// The first output type in each slice is considered the default for that tool.
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

// AvailableOutputTypesString returns a JSON string representation of all output types by profiling tool.
func AvailableOutputTypesString() string {
	out, _ := jsoniter.Marshal(GetOutputTypesByProfilingTool)
	return string(out)
}

var (
	// supportedOutputTypes contains all supported output types for profiling.
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

// AvailableOutputTypes returns the list of all supported output types.
func AvailableOutputTypes() []OutputType {
	return supportedOutputTypes
}

// IsSupportedOutputType checks if the given output type string is a supported output type.
// It returns true if the output type is in the list of available output types.
func IsSupportedOutputType(outputType string) bool {
	return lo.Contains(AvailableOutputTypes(), OutputType(outputType))
}

// IsValidOutputType checks if the given OutputType is valid for the specified ProfilingTool.
// It returns true if the output type is supported by the profiling tool.
func IsValidOutputType(OutputType OutputType, profilingTool ProfilingTool) bool {
	for _, current := range GetOutputTypesByProfilingTool[profilingTool] {
		if OutputType == current {
			return true
		}
	}
	return false
}
