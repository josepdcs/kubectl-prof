package file

import (
	"bytes"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
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

func TestGetSize(t *testing.T) {
	type args struct {
		file string
	}
	tests := []struct {
		name  string
		given func() args
		when  func(args args) int64
		then  func(t *testing.T, result int64)
		after func(file string)
	}{
		{
			name: "Get size",
			given: func() args {
				file := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt")
				var b bytes.Buffer
				b.Write([]byte("test"))
				_ = os.WriteFile(file, b.Bytes(), 0644)
				return args{file: file}
			},
			when: func(args args) int64 {
				return GetSize(args.file)
			},
			then: func(t *testing.T, result int64) {
				assert.Equal(t, int64(4), result)
			},
			after: func(file string) {
				_ = Remove(file)
			},
		},
		{
			name: "Get size when error",
			given: func() args {
				log.SetPrintLogs(true)
				return args{file: "unknown"}
			},
			when: func(args args) int64 {
				return GetSize(args.file)
			},
			then: func(t *testing.T, result int64) {
				assert.Equal(t, int64(0), result)
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

			if tt.after != nil {
				tt.after(args.file)
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {
	type args struct {
		file string
	}
	tests := []struct {
		name  string
		given func() args
		when  func(args args) bool
		then  func(t *testing.T, result bool)
		after func(file string)
	}{
		{
			name: "Is not empty",
			given: func() args {
				file := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt")
				var b bytes.Buffer
				b.Write([]byte("test"))
				_ = os.WriteFile(file, b.Bytes(), 0644)
				return args{file: file}
			},
			when: func(args args) bool {
				return IsEmpty(args.file)
			},
			then: func(t *testing.T, result bool) {
				assert.False(t, result)
			},
			after: func(file string) {
				_ = Remove(file)
			},
		},
		{
			name: "Is empty",
			given: func() args {
				log.SetPrintLogs(true)
				return args{file: "unknown"}
			},
			when: func(args args) bool {
				return IsEmpty(args.file)
			},
			then: func(t *testing.T, result bool) {
				assert.True(t, result)
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

			if tt.after != nil {
				tt.after(args.file)
			}
		})
	}
}
