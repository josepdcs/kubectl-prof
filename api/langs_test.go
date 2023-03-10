package api

import (
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAvailableLanguages(t *testing.T) {
	// Given & When
	result := AvailableLanguages()

	// Then
	assert.True(t, lo.Every(result, supportedLangs))
}

func TestIsSupportedLanguage(t *testing.T) {
	tests := []struct {
		name  string
		given string
		then  bool
	}{
		{
			name:  "java",
			given: "java",
			then:  true,
		},
		{
			name:  "go",
			given: "go",
			then:  true,
		},
		{
			name:  "python",
			given: "python",
			then:  true,
		},
		{
			name:  "node",
			given: "node",
			then:  true,
		},
		{
			name:  "fake",
			given: "fake",
			then:  true,
		},
		{
			name:  "not found",
			given: "java2",
			then:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSupportedLanguage(tt.given); got != tt.then {
				t.Errorf("IsSupportedLanguage() = %v, then %v", got, tt.then)
			}
		})
	}
}
