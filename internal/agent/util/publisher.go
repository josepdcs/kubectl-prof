package util

import (
	"bufio"
	"fmt"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"io"
	"os"
	"time"
)

func Publish(compressorType compressor.Type, file string, eventType api.EventType) error {
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
		return fmt.Errorf("could not save compressed file %s, error: %w", resultFile, err)
	}

	return log.EventLn(
		api.Result,
		api.ResultData{
			Time:           time.Now(),
			ResultType:     eventType,
			File:           resultFile,
			CompressorType: string(compressorType),
		},
	)
}
