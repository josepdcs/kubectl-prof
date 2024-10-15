package lists

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrimSpace(t *testing.T) {
	list := []string{"  this  ", "is  ", "a", "    list"}

	assert.Equal(t, []string{"this", "is", "a", "list"}, TrimSpace(list))
}
