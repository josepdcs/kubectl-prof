package api

import (
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAvailableContainerRuntimes(t *testing.T) {
	result := AvailableContainerRuntimes()

	assert.True(t, lo.Every(result, containerRuntimes))
}

func TestIsSupportedContainerRuntime(t *testing.T) {
	tests := []struct {
		name  string
		given string
		then  bool
	}{
		{
			name:  "crio",
			given: "crio",
			then:  true,
		},
		{
			name:  "containerd",
			given: "containerd",
			then:  true,
		},
		{
			name:  "fake",
			given: "fake",
			then:  true,
		},
		{
			name:  "fakeWithRootFileSystemLocationResultError",
			given: "fakeWithRootFileSystemLocationResultError",
			then:  true,
		},
		{
			name:  "fakeWithPIDResultError",
			given: "fakeWithPIDResultError",
			then:  true,
		},
		{
			name:  "not found",
			given: "containerd2",
			then:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSupportedContainerRuntime(tt.given); got != tt.then {
				t.Errorf("IsSupportedContainerRuntime() = %v, then %v", got, tt.then)
			}
		})
	}
}
