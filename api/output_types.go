package api

import jsoniter "github.com/json-iterator/go"

// GetOutputTypesByProfilingTool Gets the list of EventType related to the ProfilingTool that they will be considered as output types.
// The first one is considered the default
var GetOutputTypesByProfilingTool = map[ProfilingTool][]EventType{
	AsyncProfiler: {FlameGraph, Jfr, Flat, Traces, Collapsed, Tree},
	Jcmd:          {Jfr, ThreadDump, HeapDump, HeapHistogram},
	Pyspy:         {FlameGraph, SpeedScope, ThreadDump},
	Bpf:           {FlameGraph},
	Perf:          {FlameGraph},
	Rbspy:         {FlameGraph},
	FakeTool:      {FlameGraph},
}

func AvailableOutputTypesString() string {
	out, _ := jsoniter.Marshal(GetOutputTypesByProfilingTool)
	return string(out)
}

var (
	supportedOutputTypes = []EventType{
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
	}
)

func AvailableOutputTypes() []EventType {
	return supportedOutputTypes
}

func IsSupportedOutputType(outputType string) bool {
	return containsOutputType(EventType(outputType), AvailableOutputTypes())
}

func containsOutputType(eventType EventType, eventTypes []EventType) bool {
	for _, current := range eventTypes {
		if eventType == current {
			return true
		}
	}
	return false
}

// IsValidOutputType Identifies if given EventType is valid for the also given ProfilingTool
func IsValidOutputType(eventType EventType, profilingTool ProfilingTool) bool {
	for _, current := range GetOutputTypesByProfilingTool[profilingTool] {
		if eventType == current {
			return true
		}
	}
	return false
}
