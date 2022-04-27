package api

//GetOutputTypesByProfilingTool Gets the list of EventType related to the ProfilingTool that they will be considered as output types.
//The first one is considered the default
var GetOutputTypesByProfilingTool = map[ProfilingTool][]EventType{
	asyncProfiler: {FlameGraph, Jfr},
	jcmd:          {Jfr},
	pyspy:         {FlameGraph, Raw},
	bpf:           {FlameGraph},
	perf:          {FlameGraph},
	rbspy:         {FlameGraph},
}

var (
	supportedOutputTypes = []EventType{FlameGraph, Jfr, Raw, StackThread}
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

//IsValidOutputType Identifies if given EventType is valid for the also given ProfilingTool
func IsValidOutputType(eventType EventType, profilingTool ProfilingTool) bool {
	for _, current := range GetOutputTypesByProfilingTool[profilingTool] {
		if eventType == current {
			return true
		}
	}
	return false
}
