package util

import (
	"bufio"
	"bytes"
	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/exec"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	fileutils "github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"github.com/pkg/errors"
	"io"
	"os"
	"time"
)

func Publish(compressorType compressor.Type, file string, eventType api.OutputType) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(f)
	content, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	comp, err := compressor.Get(compressorType)
	if err != nil {
		return err
	}
	compressed, err := comp.Encode(content)
	if err != nil {
		return err
	}

	resultFile := file + compressor.GetExtensionFileByCompressor[compressorType]
	err = os.WriteFile(resultFile, compressed, 0644)
	if err != nil {
		return errors.Wrapf(err, "could not save compressed file %s", resultFile)
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

func PublishWithNativeGzipAndSplit(file, chunkSize string, eventType api.OutputType) error {
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
