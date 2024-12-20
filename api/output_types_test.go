package api

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestAvailableOutputTypesString(t *testing.T) {
	assert.NotEmpty(t, AvailableOutputTypesString())
}

func TestAvailableOutputTypes(t *testing.T) {
	result := AvailableOutputTypes()

	assert.True(t, lo.Every(result, supportedOutputTypes))
}

func TestIsSupportedOutputType(t *testing.T) {
	tests := []struct {
		name  string
		given string
		then  bool
	}{
		{
			name:  "flamegraph",
			given: "flamegraph",
			then:  true,
		},
		{
			name:  "speedscope",
			given: "speedscope",
			then:  true,
		},
		{
			name:  "jfr",
			given: "jfr",
			then:  true,
		},
		{
			name:  "threaddump",
			given: "threaddump",
			then:  true,
		},
		{
			name:  "heapdump",
			given: "heapdump",
			then:  true,
		},
		{
			name:  "heaphistogram",
			given: "heaphistogram",
			then:  true,
		},
		{
			name:  "flat",
			given: "flat",
			then:  true,
		},
		{
			name:  "traces",
			given: "traces",
			then:  true,
		},
		{
			name:  "collapsed",
			given: "collapsed",
			then:  true,
		},
		{
			name:  "tree",
			given: "tree",
			then:  true,
		},
		{
			name:  "raw",
			given: "raw",
			then:  true,
		},
		{
			name:  "callgrind",
			given: "callgrind",
			then:  true,
		},
		{
			name:  "summary",
			given: "summary",
			then:  true,
		},
		{
			name:  "summary-by-line",
			given: "summary-by-line",
			then:  true,
		},
		{
			name:  "heapsnapshot",
			given: "heapsnapshot",
			then:  true,
		},
		{
			name:  "not found",
			given: "raw2",
			then:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSupportedOutputType(tt.given); got != tt.then {
				t.Errorf("IsSupportedOutputType() = %v, then %v", got, tt.then)
			}
		})
	}
}

func TestIsValidOutputType(t *testing.T) {
	type args struct {
		eventType     OutputType
		profilingTool ProfilingTool
	}
	tests := []struct {
		name  string
		given args
		then  bool
	}{
		{
			name: "Flamegraph + AsyncProfiler",
			given: args{
				eventType:     FlameGraph,
				profilingTool: AsyncProfiler,
			},
			then: true,
		},
		{
			name: "Not valid",
			given: args{
				eventType:     FlameGraph,
				profilingTool: Jcmd,
			},
			then: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.then, IsValidOutputType(tt.given.eventType, tt.given.profilingTool), "IsValidOutputType(%v, %v)", tt.given.eventType, tt.given.profilingTool)
		})
	}
}
