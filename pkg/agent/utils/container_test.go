package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNormalizeContainerID(t *testing.T) {
	tests := []string{
		"docker://b3f6972fb3a9f5d1eba91e43900b0839aad99f7428d0580ba1b4e501017ee949",
		"cri-o://b3f6972fb3a9f5d1eba91e43900b0839aad99f7428d0580ba1b4e501017ee949",
		"containerd://b3f6972fb3a9f5d1eba91e43900b0839aad99f7428d0580ba1b4e501017ee949",
	}
	for _, tc := range tests {
		result := NormalizeContainerID(tc)
		assert.Equal(t, "b3f6972fb3a9f5d1eba91e43900b0839aad99f7428d0580ba1b4e501017ee949", result)
	}
}
