package api

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestAvailableProfilingTools(t *testing.T) {
	result := AvailableProfilingTools()

	assert.True(t, lo.Every(result, profilingTools))
}

func TestIsSupportedProfilingTool(t *testing.T) {
	tests := []struct {
		name  string
		given string
		then  bool
	}{
		{
			name:  "async-profiler",
			given: "async-profiler",
			then:  true,
		},
		{
			name:  "jcmd",
			given: "jcmd",
			then:  true,
		},
		{
			name:  "pyspy",
			given: "pyspy",
			then:  true,
		},
		{
			name:  "bpf",
			given: "bpf",
			then:  true,
		},
		{
			name:  "perf",
			given: "perf",
			then:  true,
		},
		{
			name:  "rbspy",
			given: "rbspy",
			then:  true,
		},
		{
			name:  "node-dummy",
			given: "node-dummy",
			then:  true,
		},
		{
			name:  "not found",
			given: "bpf2",
			then:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSupportedProfilingTool(tt.given); got != tt.then {
				t.Errorf("IsSupportedProfilingTool() = %v, then %v", got, tt.then)
			}
		})
	}
}

func TestAvailableProfilingToolsString(t *testing.T) {
	assert.NotEmpty(t, AvailableProfilingToolsString())
}

func TestIsValidProfilingTool(t *testing.T) {
	type args struct {
		tool     ProfilingTool
		language ProgrammingLanguage
	}
	tests := []struct {
		name  string
		given args
		then  bool
	}{
		{
			name: "AsyncProfiler + Java",
			given: args{
				tool:     AsyncProfiler,
				language: Java,
			},
			then: true,
		},
		{
			name: "Jcmd + Java",
			given: args{
				tool:     Jcmd,
				language: Java,
			},
			then: true,
		},
		{
			name: "Pyspy + Python",
			given: args{
				tool:     Pyspy,
				language: Python,
			},
			then: true,
		},
		{
			name: "Bpf + Go",
			given: args{
				tool:     Bpf,
				language: Go,
			},
			then: true,
		},
		{
			name: "Perf + Node",
			given: args{
				tool:     Perf,
				language: Node,
			},
			then: true,
		},
		{
			name: "Rbspy + Ruby",
			given: args{
				tool:     Rbspy,
				language: Ruby,
			},
			then: true,
		},
		{
			name: "NodeDummy + Node",
			given: args{
				tool:     NodeDummy,
				language: Node,
			},
			then: true,
		},
		{
			name: "Not valid",
			given: args{
				tool:     AsyncProfiler,
				language: Python,
			},
			then: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.then, IsValidProfilingTool(tt.given.tool, tt.given.language), "IsValidProfilingTool(%v, %v)", tt.given.tool, tt.given.language)
		})
	}
}
