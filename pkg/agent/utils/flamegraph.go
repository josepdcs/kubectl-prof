package utils

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"github.com/josepdcs/kubectl-profile/api"
	"io/ioutil"
	"os"
	"time"
)

func PublishFlameGraph(flameFile string) error {
	file, err := os.Open(flameFile)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(file)
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(content)
	fgData := api.FlameGraphData{EncodedFile: encoded}

	return api.PublishEvent(api.FlameGraph, fgData)
}

func PublishLogEvent(level string, msg string) {
	if len(msg) > 0 {
		_ = api.PublishEvent(
			api.Log,
			&api.LogData{
				Time:  time.Now(),
				Level: level,
				Msg:   fmt.Sprint(msg)},
		)
	}
}
