package utils

import (
	"bufio"
	"encoding/base64"
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

	encoded := base64.StdEncoding.EncodeToString(compressed)

	return api.PublishEvent(eventType, api.OutputData{EncodedData: encoded})
}
