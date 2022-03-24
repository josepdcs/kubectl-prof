package utils

import (
	"bufio"
	"encoding/base64"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"io/ioutil"
	"os"
)

func PublishFlameGraph(c api.Compressor, flameFile string) error {
	file, err := os.Open(flameFile)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(file)
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

	encoded := base64.StdEncoding.EncodeToString(compressed)
	fgData := api.FlameGraphData{EncodedFile: encoded}

	return api.PublishEvent(api.FlameGraph, fgData)
}
