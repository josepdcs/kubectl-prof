package job

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProfilingJob_String(t *testing.T) {
	p := ProfilingJob{
		Duration:             10,
		Interval:             5,
		UID:                  "ID",
		ContainerRuntime:     "ContainerRuntime",
		ContainerRuntimePath: "ContainerRuntimePath",
		ContainerID:          "ContainerID",
		PodUID:               "PodUID",
		Language:             "Language",
		Event:                "Event",
		Compressor:           "Compressor",
		Tool:                 "Tool",
		OutputType:           "OutputType",
		FileName:             "FileName",
		AdditionalArguments: map[string]string{
			"maxsize":  "1024M",
			"settings": "custom",
		},
	}

	out := p.String()

	assert.NotEmpty(t, out)
	assert.Contains(t, out, "10")
	assert.Contains(t, out, "5")
	assert.Contains(t, out, "ID")
	assert.Contains(t, out, "ContainerRuntime")
	assert.Contains(t, out, "ContainerRuntimePath")
	assert.Contains(t, out, "ContainerID")
	assert.Contains(t, out, "PodUID")
	assert.Contains(t, out, "Language")
	assert.Contains(t, out, "Event")
	assert.Contains(t, out, "Compressor")
	assert.Contains(t, out, "Tool")
	assert.Contains(t, out, "FileName")
	assert.Contains(t, out, "\"maxsize\":\"1024M\"")
	assert.Contains(t, out, "\"settings\":\"custom\"")
}

func TestProfilingJob_ToMap(t *testing.T) {
	p := ProfilingJob{
		Duration:                 10,
		Interval:                 5,
		UID:                      "ID",
		ContainerRuntime:         "ContainerRuntime",
		ContainerRuntimePath:     "ContainerRuntimePath",
		ContainerID:              "ContainerID",
		PodUID:                   "PodUID",
		Language:                 "Language",
		Event:                    "Event",
		Compressor:               "Compressor",
		Tool:                     "Tool",
		OutputType:               "OutputType",
		FileName:                 "FileName",
		HeapDumpSplitInChunkSize: "100M",
		PID:                      "PID",
		Pgrep:                    "PGREP",
		AdditionalArguments: map[string]string{
			"maxsize":  "1024M",
			"settings": "custom",
		},
	}

	out := p.ToMap()

	assert.NotEmpty(t, out)
	assert.Len(t, out, 17)
	assert.Equal(t, float64(10), out["Duration"])
	assert.Equal(t, float64(5), out["Interval"])
	assert.Equal(t, "ID", out["UID"])
	assert.Equal(t, "ContainerRuntime", out["ContainerRuntime"])
	assert.Equal(t, "ContainerRuntimePath", out["ContainerRuntimePath"])
	assert.Equal(t, "ContainerID", out["ContainerID"])
	assert.Equal(t, "PodUID", out["PodUID"])
	assert.Equal(t, "Language", out["Language"])
	assert.Equal(t, "Event", out["Event"])
	assert.Equal(t, "Compressor", out["Compressor"])
	assert.Equal(t, "Tool", out["Tool"])
	assert.Equal(t, "OutputType", out["OutputType"])
	assert.Equal(t, "FileName", out["FileName"])
	assert.Equal(t, "100M", out["HeapDumpSplitInChunkSize"])
	assert.Equal(t, "PID", out["PID"])
	assert.Equal(t, "PGREP", out["Pgrep"])
	assert.Equal(t, map[string]interface{}{
		"maxsize":  "1024M",
		"settings": "custom",
	}, out["AdditionalArguments"])
}

func TestProfilingJob_GetWidthAdditionalArgument(t *testing.T) {
	type fields struct {
		job *ProfilingJob
	}
	tests := []struct {
		name  string
		given func() fields
		when  func(f fields) string
		then  func(t *testing.T, s string)
	}{
		{
			name: "should return width additional argument",
			given: func() fields {
				return fields{job: &ProfilingJob{AdditionalArguments: map[string]string{FlamegraphWidthInPixels: "1000"}}}
			},
			when: func(f fields) string {
				return f.job.GetWidthAdditionalArgument()
			},
			then: func(t *testing.T, s string) {
				assert.Equal(t, "1000", s)
			},
		},
		{
			name: "should return empty width additional argument when not numeric",
			given: func() fields {
				return fields{job: &ProfilingJob{AdditionalArguments: map[string]string{FlamegraphWidthInPixels: ""}}}
			},
			when: func(f fields) string {
				return f.job.GetWidthAdditionalArgument()
			},
			then: func(t *testing.T, s string) {
				assert.Empty(t, s)
			},
		},
		{
			name: "should return empty width additional argument",
			given: func() fields {
				return fields{job: &ProfilingJob{}}
			},
			when: func(f fields) string {
				return f.job.GetWidthAdditionalArgument()
			},
			then: func(t *testing.T, s string) {
				assert.Empty(t, s)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			fields := tt.given()

			// When
			result := tt.when(fields)

			// Then
			tt.then(t, result)
		})
	}
}

func TestProfilingJob_DeleteWidthAdditionalArgument(t *testing.T) {
	// Given
	job := &ProfilingJob{AdditionalArguments: map[string]string{FlamegraphWidthInPixels: "1000"}}

	// When
	job.DeleteWidthAdditionalArgument()

	// Then
	assert.Empty(t, job.AdditionalArguments)

}

func TestProfilingJob_GetWidthAdditionalArgumentAndDelete(t *testing.T) {
	// Given
	job := &ProfilingJob{AdditionalArguments: map[string]string{FlamegraphWidthInPixels: "1000"}}

	// When
	width := job.GetWidthAdditionalArgumentAndDelete()

	// Then
	assert.Equal(t, "1000", width)
	assert.Empty(t, job.AdditionalArguments)
}
