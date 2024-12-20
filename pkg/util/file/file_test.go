package file

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	files := []string{
		"/tmp/exists.txt",
		"/tmp/exists2-PID.txt",
		"/tmp/exists3.txt",
	}
	fileName, ok := lo.Find(files, func(f string) bool { return strings.Contains(f, "PID") })
	assert.True(t, ok)
	assert.Equal(t, "/tmp/exists2-PID.txt", fileName)
}

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
			name: "Remove all files",
			given: func() args {
				f := filepath.Join("/tmp", config.ProfilingPrefix+"file.txt")
				_, _ = os.Create(f)
				return args{
					targetTmpDir: "/tmp",
				}
			},
			when: func(args args) {
				RemoveAll(args.targetTmpDir, "file.txt")
			},
			then: func(t *testing.T) {
				f := filepath.Join("/tmp", config.ProfilingPrefix+"file.txt")
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
				return Size(args.file)
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
				return Size(args.file)
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

func TestWrite(t *testing.T) {
	type args struct {
		file    string
		content string
	}
	tests := []struct {
		name  string
		given func() args
		when  func(args args)
		then  func(t *testing.T, file string)
		after func(file string)
	}{
		{
			name: "write file",
			given: func() args {
				return args{file: filepath.Join(common.TmpDir(), "test.txt"), content: "content"}
			},
			when: func(args args) {
				Write(args.file, args.content)
			},
			then: func(t *testing.T, file string) {
				assert.True(t, Exists(file))
				assert.Equal(t, int64(len("content")), Size(file))
			},
			after: func(file string) {
				_ = Remove(file)
			},
		},
		{
			name: "not write file",
			given: func() args {
				return args{file: filepath.Join("/", "test.txt"), content: "content"}
			},
			when: func(args args) {
				log.SetPrintLogs(true)
				Write(args.file, args.content)
			},
			then: func(t *testing.T, file string) {
				assert.False(t, Exists(file))
			},
			after: func(file string) {
				_ = Remove(file)
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
			tt.then(t, args.file)

			if tt.after != nil {
				tt.after(args.file)
			}
		})
	}
}

func TestRead(t *testing.T) {
	type args struct {
		file string
	}
	tests := []struct {
		name  string
		given func() args
		when  func(args args) string
		then  func(t *testing.T, content string)
		after func(file string)
	}{
		{
			name: "read file",
			given: func() args {
				file := filepath.Join(common.TmpDir(), "test.txt")
				Write(file, "content")
				return args{file: file}
			},
			when: func(args args) string {
				return Read(args.file)
			},
			then: func(t *testing.T, content string) {
				assert.Equal(t, "content", content)
			},
			after: func(file string) {
				_ = Remove(file)
			},
		},
		{
			name: "no read file",
			given: func() args {
				return args{file: "other_file.txt"}
			},
			when: func(args args) string {
				log.SetPrintLogs(true)
				return Read(args.file)
			},
			then: func(t *testing.T, content string) {
				assert.Empty(t, content)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			args := tt.given()

			// When
			content := tt.when(args)

			// Then
			tt.then(t, content)

			if tt.after != nil {
				tt.after(args.file)
			}
		})
	}
}

func TestGetChecksum(t *testing.T) {
	type args struct {
		file string
	}
	tests := []struct {
		name  string
		given func() args
		when  func(args args) string
		then  func(t *testing.T, content string)
		after func(file string)
	}{
		{
			name: "get checksum",
			given: func() args {
				file := filepath.Join(common.TmpDir(), "test.txt")
				Write(file, "content")
				return args{file: file}
			},
			when: func(args args) string {
				return Checksum(args.file)
			},
			then: func(t *testing.T, content string) {
				assert.Equal(t, "9a0364b9e99bb480dd25e1f0284c8555", content)
			},
			after: func(file string) {
				_ = Remove(file)
			},
		},
		{
			name: "no read file",
			given: func() args {
				return args{file: "other_file.txt"}
			},
			when: func(args args) string {
				log.SetPrintLogs(true)
				return Checksum(args.file)
			},
			then: func(t *testing.T, content string) {
				assert.Empty(t, content)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			args := tt.given()

			// When
			content := tt.when(args)

			// Then
			tt.then(t, content)

			if tt.after != nil {
				tt.after(args.file)
			}
		})
	}
}

func TestMergeFiles(t *testing.T) {
	type args struct {
		outputPath string
		inputPaths []string
	}
	tests := []struct {
		name  string
		given func() args
		when  func(args args)
		then  func(t *testing.T)
		after func(args args)
	}{
		{
			name: "merge files",
			given: func() args {
				file1 := filepath.Join(common.TmpDir(), "test1.txt")
				Write(file1, "content1")
				file2 := filepath.Join(common.TmpDir(), "test2.txt")
				Write(file2, "content2")
				file3 := filepath.Join(common.TmpDir(), "test3.txt")
				Write(file3, "content3")
				return args{
					outputPath: filepath.Join(common.TmpDir(), "test.txt"),
					inputPaths: []string{file1, file2, file3},
				}
			},
			when: func(args args) {
				MergeFiles(args.outputPath, args.inputPaths)
			},
			then: func(t *testing.T) {
				content := Read(filepath.Join(common.TmpDir(), "test.txt"))
				assert.Equal(t, "content1content2content3", content)
			},
			after: func(args args) {
				for _, f := range args.inputPaths {
					_ = Remove(f)
				}
				_ = Remove(args.outputPath)
			},
		},
		{
			name: "output file does not exist",
			given: func() args {
				file1 := filepath.Join(common.TmpDir(), "test1.txt")
				Write(file1, "content1")
				file2 := filepath.Join(common.TmpDir(), "test2.txt")
				Write(file2, "content2")
				file3 := filepath.Join(common.TmpDir(), "test3.txt")
				Write(file3, "content3")
				return args{
					outputPath: "/other_file.txt",
					inputPaths: []string{file1, file2, file3},
				}
			},
			when: func(args args) {
				MergeFiles(args.outputPath, args.inputPaths)
			},
			then: func(t *testing.T) {
				content := Read(filepath.Join(common.TmpDir(), "test.txt"))
				assert.Empty(t, content)
			},
			after: func(args args) {
				for _, f := range args.inputPaths {
					_ = Remove(f)
				}
			},
		},
		{
			name: "input files do not exist",
			given: func() args {
				file1 := filepath.Join(common.TmpDir(), "test1.txt")
				Write(file1, "content1")
				file2 := filepath.Join(common.TmpDir(), "test2.txt")
				Write(file2, "content2")
				file3 := filepath.Join(common.TmpDir(), "test3.txt")
				return args{
					outputPath: filepath.Join(common.TmpDir(), "test.txt"),
					inputPaths: []string{file1, file2, file3},
				}
			},
			when: func(args args) {
				MergeFiles(args.outputPath, args.inputPaths)
			},
			then: func(t *testing.T) {
				content := Read(filepath.Join(common.TmpDir(), "test.txt"))
				assert.Equal(t, "content1content2", content)
			},
			after: func(args args) {
				for _, f := range args.inputPaths {
					_ = Remove(f)
				}
				_ = Remove(args.outputPath)
			},
		},
	}
	for _, tt := range tests {
		log.SetPrintLogs(true)
		t.Run(tt.name, func(t *testing.T) {
			// Given
			args := tt.given()

			// When
			tt.when(args)

			// Then
			tt.then(t)

			if tt.after != nil {
				tt.after(args)
			}
		})
	}
}

func TestCopy(t *testing.T) {
	type args struct {
		src string
		dst string
	}
	tests := []struct {
		name  string
		given func() args
		when  func(args args) (int64, error)
		then  func(t *testing.T, nBytes int64, err error)
		after func(args args)
	}{
		{
			name: "copy file src to dst",
			given: func() args {
				src := filepath.Join(common.TmpDir(), "src.txt")
				Write(src, "content")
				dst := filepath.Join(common.TmpDir(), "dst.txt")
				return args{
					src: src,
					dst: dst,
				}
			},
			when: func(args args) (int64, error) {
				return Copy(args.src, args.dst)
			},
			then: func(t *testing.T, nBytes int64, err error) {
				assert.Nil(t, err)
				assert.Equal(t, int64(7), nBytes)
				content := Read(filepath.Join(common.TmpDir(), "dst.txt"))
				assert.Equal(t, "content", content)

			},
			after: func(args args) {
				_ = Remove(args.src)
				_ = Remove(args.dst)
			},
		},
	}
	for _, tt := range tests {
		log.SetPrintLogs(true)
		t.Run(tt.name, func(t *testing.T) {
			// Given
			args := tt.given()

			// When
			nBytes, err := tt.when(args)

			// Then
			tt.then(t, nBytes, err)

			if tt.after != nil {
				tt.after(args)
			}
		})
	}
}

func TestMove(t *testing.T) {
	type args struct {
		src string
		dst string
	}
	tests := []struct {
		name  string
		given func() args
		when  func(args args) error
		then  func(t *testing.T, err error)
		after func(args args)
	}{
		{
			name: "move file src to dst",
			given: func() args {
				src := filepath.Join(common.TmpDir(), "src.txt")
				Write(src, "content")
				dst := filepath.Join(common.TmpDir(), "dst.txt")
				return args{
					src: src,
					dst: dst,
				}
			},
			when: func(args args) error {
				return Move(args.src, args.dst)
			},
			then: func(t *testing.T, err error) {
				assert.Nil(t, err)
				content := Read(filepath.Join(common.TmpDir(), "dst.txt"))
				assert.Equal(t, "content", content)
				assert.False(t, Exists(filepath.Join(common.TmpDir(), "src.txt")))

			},
			after: func(args args) {
				_ = Remove(args.dst)
			},
		},
	}
	for _, tt := range tests {
		log.SetPrintLogs(true)
		t.Run(tt.name, func(t *testing.T) {
			// Given
			args := tt.given()

			// When
			err := tt.when(args)

			// Then
			tt.then(t, err)

			if tt.after != nil {
				tt.after(args)
			}
		})
	}
}
