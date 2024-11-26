package publish

import (
	"bytes"
	"os"
	"time"

	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/exec"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	fileutils "github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"github.com/pkg/errors"
)

// Publisher is the interface that wraps the basic Do method in order to publish the profiling result
type Publisher interface {
	// Do compress the file and publishes the result
	Do(compressorType compressor.Type, file string, eventType api.OutputType) error
	// DoWithNativeGzipAndSplit compress the file with gzip and split the result file in chunks
	DoWithNativeGzipAndSplit(file, chunkSize string, eventType api.OutputType) error
}

type publisher struct {
}

// NewPublisher returns a new publisher
func NewPublisher() Publisher {
	return &publisher{}
}

var newPublisher = NewPublisher()

// Do compress the file and publishes the result
func (p publisher) Do(compressorType compressor.Type, filePath string, eventType api.OutputType) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	resultFile := filePath + compressor.GetExtensionFileByCompressor[compressorType]
	compressedFile, err := os.Create(resultFile)
	if err != nil {
		return errors.Wrapf(err, "could not create result file %s", resultFile)
	}

	comp, err := compressor.Get(compressorType)
	if err != nil {
		return err
	}

	err = comp.Encode(compressedFile, file)
	if err != nil {
		return errors.Wrapf(err, "could not compress file %s", resultFile)
	}

	// get the size of the result file from stat command
	var outStat bytes.Buffer
	cmd := exec.Command("stat", "-c%s", resultFile)
	cmd.Stdout = &outStat
	_ = cmd.Run()

	return log.EventLn(
		api.Result,
		api.ResultData{
			Time:            time.Now(),
			ResultType:      eventType,
			File:            resultFile,
			FileSizeInBytes: fileutils.GetSize(resultFile),
			Checksum:        fileutils.GetChecksum(resultFile),
			CompressorType:  string(compressorType),
		},
	)
}

// DoWithNativeGzipAndSplit compress the file with gzip and split the result file in chunks
func (p publisher) DoWithNativeGzipAndSplit(file, chunkSize string, eventType api.OutputType) error {
	if !fileutils.Exists(file) {
		return errors.Errorf("file %s does not exist", file)
	}
	if stringUtils.IsBlank(chunkSize) {
		return errors.Errorf("chunk size is mandatory")
	}

	// compresses the file with gzip
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("gzip", "-3", file)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "gzip failed on file %s; detail: %s", file, stderr.String())
	}

	// split the result file from gzip command with split command
	cmd = exec.Command("split", "-b", chunkSize, "-e", "--numeric-suffixes", file+".gz", file+".gz.")
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "split failed on file %s; detail: %s", file+".gz", stderr.String())
	}

	// get the size of the result file
	fileSizeInBytes := fileutils.GetSize(file + ".gz")

	// try to remove the result file from gzip command since is not needed anymore
	_ = fileutils.Remove(file + ".gz")

	// get the list of chunks and fill the data structure
	chunkFiles := fileutils.List(file + ".gz.*")
	chunkFilesData := make([]api.ChunkData, 0, len(chunkFiles))
	for _, chunkFile := range chunkFiles {
		chunkFilesData = append(
			chunkFilesData,
			api.ChunkData{
				File:            chunkFile,
				FileSizeInBytes: fileutils.GetSize(chunkFile),
				Checksum:        fileutils.GetChecksum(chunkFile),
			})
	}

	return log.EventLn(
		api.Result,
		api.ResultData{
			Time:            time.Now(),
			ResultType:      eventType,
			File:            file + ".gz",
			FileSizeInBytes: fileSizeInBytes,
			CompressorType:  compressor.Gzip,
			Chunks:          chunkFilesData,
		},
	)
}

// Do compress the file and publishes the result
func Do(compressorType compressor.Type, file string, eventType api.OutputType) error {
	return newPublisher.Do(compressorType, file, eventType)
}

// DoWithNativeGzipAndSplit compress the file with gzip and split the result file in chunks
func DoWithNativeGzipAndSplit(file, chunkSize string, eventType api.OutputType) error {
	return newPublisher.DoWithNativeGzipAndSplit(file, chunkSize, eventType)
}
