package utils

import (
	"bufio"
	"fmt"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"io/ioutil"
	"os"
)

func Publish(c api.Compressor, file string, eventType api.EventType) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(f)
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	comp, err := compressor.Get(c)
	if err != nil {
		return err
	}
	compressed, err := comp.Encode(content)
	if err != nil {
		return err
	}

	resultFile := file + api.GetExtensionFileByCompressor[c]
	err = ioutil.WriteFile(resultFile, compressed, 0644)
	if err != nil {
		return fmt.Errorf("could not save compressed file %s, error: %w", resultFile, err)
	}

	return api.PublishEvent(
		api.Result,
		api.ResultData{
			ResultType:     eventType,
			File:           resultFile,
			CompressorType: string(c),
		},
	)
}
