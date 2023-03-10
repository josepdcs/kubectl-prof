package api

import (
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAvailableLogLevels(t *testing.T) {
	result := AvailableLogLevels()

	assert.True(t, lo.Every(result, logLevels))
}

func TestIsSupportedLogLevel(t *testing.T) {
	tests := []struct {
		name  string
		given string
		then  bool
	}{
		{
			name:  "info",
			given: "info",
			then:  true,
		},
		{
			name:  "warn",
			given: "warn",
			then:  true,
		},
		{
			name:  "debug",
			given: "debug",
			then:  true,
		},
		{
			name:  "error",
			given: "error",
			then:  true,
		},
		{
			name:  "not found",
			given: "info2",
			then:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSupportedLogLevel(tt.given); got != tt.then {
				t.Errorf("IsSupportedLogLevel() = %v, then %v", got, tt.then)
			}
		})
	}
}
