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
			name: "CargoFlame + Rust",
			given: args{
				tool:     CargoFlame,
				language: Rust,
			},
			then: true,
		},
		{
			name: "Bpf + Clang",
			given: args{
				tool:     Bpf,
				language: Clang,
			},
			then: true,
		},
		{
			name: "Bpf + ClangPlusPlus",
			given: args{
				tool:     Bpf,
				language: ClangPlusPlus,
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

func TestGetProfilingTool(t *testing.T) {
	type args struct {
		language   ProgrammingLanguage
		outputType OutputType
	}
	tests := []struct {
		name  string
		given args
		then  ProfilingTool
	}{
		// Java tests
		{
			name: "Java + Jfr",
			given: args{
				language:   Java,
				outputType: Jfr,
			},
			then: Jcmd,
		},
		{
			name: "Java + ThreadDump",
			given: args{
				language:   Java,
				outputType: ThreadDump,
			},
			then: Jcmd,
		},
		{
			name: "Java + HeapDump",
			given: args{
				language:   Java,
				outputType: HeapDump,
			},
			then: Jcmd,
		},
		{
			name: "Java + HeapHistogram",
			given: args{
				language:   Java,
				outputType: HeapHistogram,
			},
			then: Jcmd,
		},
		{
			name: "Java + FlameGraph",
			given: args{
				language:   Java,
				outputType: FlameGraph,
			},
			then: AsyncProfiler,
		},
		{
			name: "Java + Flat",
			given: args{
				language:   Java,
				outputType: Flat,
			},
			then: AsyncProfiler,
		},
		{
			name: "Java + Traces",
			given: args{
				language:   Java,
				outputType: Traces,
			},
			then: AsyncProfiler,
		},
		{
			name: "Java + Collapsed",
			given: args{
				language:   Java,
				outputType: Collapsed,
			},
			then: AsyncProfiler,
		},
		{
			name: "Java + Tree",
			given: args{
				language:   Java,
				outputType: Tree,
			},
			then: AsyncProfiler,
		},
		{
			name: "Java + Raw",
			given: args{
				language:   Java,
				outputType: Raw,
			},
			then: AsyncProfiler,
		},
		// Python tests
		{
			name: "Python + FlameGraph",
			given: args{
				language:   Python,
				outputType: FlameGraph,
			},
			then: Pyspy,
		},
		{
			name: "Python + Raw",
			given: args{
				language:   Python,
				outputType: Raw,
			},
			then: Pyspy,
		},
		// Go tests
		{
			name: "Go + FlameGraph",
			given: args{
				language:   Go,
				outputType: FlameGraph,
			},
			then: Bpf,
		},
		// Rust tests
		{
			name: "Rust + FlameGraph",
			given: args{
				language:   Rust,
				outputType: FlameGraph,
			},
			then: CargoFlame,
		},
		// Clang tests
		{
			name: "Clang + FlameGraph",
			given: args{
				language:   Clang,
				outputType: FlameGraph,
			},
			then: Bpf,
		},
		// ClangPlusPlus tests
		{
			name: "ClangPlusPlus + FlameGraph",
			given: args{
				language:   ClangPlusPlus,
				outputType: FlameGraph,
			},
			then: Bpf,
		},
		// Ruby tests
		{
			name: "Ruby + FlameGraph",
			given: args{
				language:   Ruby,
				outputType: FlameGraph,
			},
			then: Rbspy,
		},
		// Node tests
		{
			name: "Node + FlameGraph",
			given: args{
				language:   Node,
				outputType: FlameGraph,
			},
			then: Bpf,
		},
		{
			name: "Node + Raw",
			given: args{
				language:   Node,
				outputType: Raw,
			},
			then: Bpf,
		},
		{
			name: "Node + HeapSnapshot",
			given: args{
				language:   Node,
				outputType: HeapSnapshot,
			},
			then: NodeDummy,
		},
		{
			name: "Node + HeapDump",
			given: args{
				language:   Node,
				outputType: HeapDump,
			},
			then: NodeDummy,
		},
		// Default cases (when output type doesn't match any specific case)
		{
			name: "Java + Default (SpeedScope)",
			given: args{
				language:   Java,
				outputType: SpeedScope,
			},
			then: Jcmd, // default for Java
		},
		{
			name: "Node + Default (Tree)",
			given: args{
				language:   Node,
				outputType: Tree,
			},
			then: Bpf, // default for Node
		},
		{
			name: "Go + Default (Raw)",
			given: args{
				language:   Go,
				outputType: Raw,
			},
			then: Bpf, // default for Go
		},
		{
			name: "Python + Default (FlameGraph)",
			given: args{
				language:   Python,
				outputType: FlameGraph,
			},
			then: Pyspy, // default for Python
		},
		{
			name: "Ruby + Default (FlameGraph)",
			given: args{
				language:   Ruby,
				outputType: FlameGraph,
			},
			then: Rbspy, // default for Ruby
		},
		{
			name: "Rust + Default (FlameGraph)",
			given: args{
				language:   Rust,
				outputType: FlameGraph,
			},
			then: CargoFlame, // default for Rust
		},
		{
			name: "Clang + Default (Raw)",
			given: args{
				language:   Clang,
				outputType: Raw,
			},
			then: Bpf, // default for Clang
		},
		{
			name: "ClangPlusPlus + Default (Raw)",
			given: args{
				language:   ClangPlusPlus,
				outputType: Raw,
			},
			then: Bpf, // default for ClangPlusPlus
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.then, GetProfilingTool(tt.given.language, tt.given.outputType), "GetProfilingTool(%v, %v)", tt.given.language, tt.given.outputType)
		})
	}
}

func TestGetProfilingToolsByProgrammingLanguage(t *testing.T) {
	tests := []struct {
		name     string
		language ProgrammingLanguage
		expected []ProfilingTool
	}{
		{
			name:     "Java",
			language: Java,
			expected: []ProfilingTool{Jcmd, AsyncProfiler},
		},
		{
			name:     "Python",
			language: Python,
			expected: []ProfilingTool{Pyspy},
		},
		{
			name:     "Go",
			language: Go,
			expected: []ProfilingTool{Bpf, CargoFlame},
		},
		{
			name:     "Node",
			language: Node,
			expected: []ProfilingTool{Bpf, Perf, NodeDummy, CargoFlame},
		},
		{
			name:     "Clang",
			language: Clang,
			expected: []ProfilingTool{Bpf, Perf, CargoFlame},
		},
		{
			name:     "ClangPlusPlus",
			language: ClangPlusPlus,
			expected: []ProfilingTool{Bpf, Perf, CargoFlame},
		},
		{
			name:     "Ruby",
			language: Ruby,
			expected: []ProfilingTool{Rbspy},
		},
		{
			name:     "Rust",
			language: Rust,
			expected: []ProfilingTool{CargoFlame, Bpf, Perf},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, GetProfilingToolsByProgrammingLanguage[tt.language])
		})
	}
}
