package compressor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipFileCompressor_Encode(t *testing.T) {
	originalFile, err := os.Open(filepath.Join("testdata", "10mb-examplefile-com.txt"))
	require.NoError(t, err)
	defer func(originalFile *os.File) {
		err := originalFile.Close()
		if err != nil {
			t.Fatal(err)
		}
	}(originalFile)

	gzippedFile, err := os.Create(filepath.Join(common.TmpDir(), "10mb-examplefile-com.txt.gz"))
	require.NoError(t, err)
	defer func(gzippedFile *os.File) {
		err := gzippedFile.Close()
		if err != nil {
			t.Fatal(err)
		}
	}(gzippedFile)

	err = NewGzipCompressor().Encode(gzippedFile, originalFile)
	require.NoError(t, err)
	assert.True(t, file.Exists(filepath.Join(common.TmpDir(), "10mb-examplefile-com.txt.gz")))

	file.RemoveAll(common.TmpDir(), "10mb-examplefile-com.txt.gz*")
}

func TestGzipFileCompressor_Decode(t *testing.T) {
	gzippedFile, err := os.Open(filepath.Join("testdata", "10mb-examplefile-com.txt.gz"))
	require.NoError(t, err)
	defer func(gzippedFile *os.File) {
		err := gzippedFile.Close()
		if err != nil {
			t.Fatal(err)
		}
	}(gzippedFile)

	originalFile, err := os.Create(filepath.Join(common.TmpDir(), "10mb-examplefile-com.txt"))
	require.NoError(t, err)
	defer func(originalFile *os.File) {
		err := originalFile.Close()
		if err != nil {
			t.Fatal(err)
		}
	}(originalFile)

	err = NewGzipCompressor().Decode(originalFile, gzippedFile)
	require.NoError(t, err)
	assert.True(t, file.Exists(filepath.Join(common.TmpDir(), "10mb-examplefile-com.txt")))

	file.RemoveAll(common.TmpDir(), "10mb-examplefile-com.txt*")
}
