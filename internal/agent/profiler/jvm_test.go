package profiler

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewJvmProfiler(t *testing.T) {
	p := NewJvmProfiler()
	assert.IsType(t, p, &JvmProfiler{})
}
