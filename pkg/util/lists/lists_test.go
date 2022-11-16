package lists

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTrimSpace(t *testing.T) {
	list := []string{"  this  ", "is  ", "a", "    list"}

	assert.Equal(t, []string{"this", "is", "a", "list"}, TrimSpace(list))
}
