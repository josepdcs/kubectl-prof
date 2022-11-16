package file

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestExists(t *testing.T) {
	type args struct {
		file string
	}
	tests := []struct {
		name  string
		given func() args
		when  func(args) bool
		then  func(t *testing.T, whenResult bool)
		after func(args)
	}{
		{
			name: "File exists",
			given: func() args {
				file := filepath.Join("/tmp", "exists.txt")
				_, _ = os.Create(file)
				return args{
					file: file,
				}
			},
			when: func(args args) bool {
				return Exists(args.file)
			},
			then: func(t *testing.T, whenResult bool) {
				assert.True(t, whenResult)
			},
			after: func(args args) {
				_ = os.Remove(args.file)
			},
		},
		{
			name: "File not exists",
			given: func() args {
				file := filepath.Join("/tmp", "exists.txt")
				return args{
					file: file,
				}
			},
			when: func(args args) bool {
				return Exists(args.file)
			},
			then: func(t *testing.T, whenResult bool) {
				assert.False(t, whenResult)
			},
			after: func(args args) {
				// nothing
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			args := tt.given()

			// When
			result := tt.when(args)

			// Then
			tt.then(t, result)

			// after each test
			tt.after(args)
		})
	}
}

func TestList(t *testing.T) {
	type args struct {
		pattern string
	}
	tests := []struct {
		name  string
		given func() args
		when  func(args) []string
		then  func(t *testing.T, whenResult []string)
	}{
		{
			name: "List files",
			given: func() args {
				return args{
					pattern: "/tmp/*",
				}
			},
			when: func(args args) []string {
				return List(args.pattern)
			},
			then: func(t *testing.T, whenResult []string) {
				assert.NotEmpty(t, whenResult)
			},
		},
		{
			name: "Not list files",
			given: func() args {
				return args{
					pattern: "/tmp/*xfdsfdf",
				}
			},
			when: func(args args) []string {
				return List(args.pattern)
			},
			then: func(t *testing.T, whenResult []string) {
				assert.Empty(t, whenResult)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			args := tt.given()

			// When
			result := tt.when(args)

			// Then
			tt.then(t, result)
		})
	}
}

func Test_RemoveAll(t *testing.T) {
	type args struct {
		targetTmpDir string
	}
	tests := []struct {
		name  string
		given func() args
		when  func(args args)
		then  func(t *testing.T)
	}{
		{
			name: "should remove profiling results",
			given: func() args {
				f := filepath.Join("/tmp", "file.txt")
				_, _ = os.Create(f)
				return args{
					targetTmpDir: "/tmp",
				}
			},
			when: func(args args) {
				RemoveAll(args.targetTmpDir, "file.txt")
			},
			then: func(t *testing.T) {
				f := filepath.Join("/tmp", "file.txt")
				assert.False(t, Exists(f))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			args := tt.given()

			// When
			tt.when(args)

			// Then
			tt.then(t)
		})
	}
}
