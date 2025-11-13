package api

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestAvailableEvents(t *testing.T) {
	result := AvailableEvents()

	assert.True(t, lo.Every(result, supportedEvents))
}

func TestIsSupportedEvent(t *testing.T) {
	tests := []struct {
		name  string
		given string
		then  bool
	}{
		{
			name:  "cpu",
			given: "cpu",
			then:  true,
		},
		{
			name:  "alloc",
			given: "alloc",
			then:  true,
		},
		{
			name:  "lock",
			given: "lock",
			then:  true,
		},
		{
			name:  "cache-misses",
			given: "cache-misses",
			then:  true,
		},
		{
			name:  "wall",
			given: "wall",
			then:  true,
		},
		{
			name:  "itimer",
			given: "itimer",
			then:  true,
		},
		{
			name:  "ctimer",
			given: "ctimer",
			then:  true,
		},
		{
			name:  "not found",
			given: "itimer2",
			then:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSupportedEvent(tt.given); got != tt.then {
				t.Errorf("IsSupportedEvent() = %v, then %v", got, tt.then)
			}
		})
	}
}
