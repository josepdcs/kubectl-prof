package crio

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"io/ioutil"
	"strconv"
)

type Crio struct {
}

func NewCrio() *Crio {
	return &Crio{}
}

var crioConfigFile = func(containerID string) string {
	return fmt.Sprintf("/var/lib/containers/storage/overlay-containers/%s/userdata/config.json", containerID)
}

var crioStateFile = func(containerID string) string {
	return fmt.Sprintf("/var/lib/containers/storage/overlay-containers/%s/userdata/state.json", containerID)
}

func (c *Crio) RootFileSystemLocation(containerID string) (string, error) {
	if containerID == "" {
		return "", errors.New("container ID is mandatory")
	}

	spec, err := runtimeSpec(crioConfigFile(containerID))
	if err != nil {
		return "", err
	}
	return spec.Root.Path, nil

}

func runtimeSpec(configFile string) (rspec.Spec, error) {
	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		return rspec.Spec{}, errors.Wrapf(err, "read file failed: %s", configFile)
	}

	var result rspec.Spec
	err = jsoniter.Unmarshal(file, &result)
	if err != nil {
		return rspec.Spec{}, errors.Wrapf(err, "unmarshal file failed: %s", configFile)
	}

	return result, nil
}

func (c *Crio) PID(containerID string) (string, error) {
	if containerID == "" {
		return "", errors.New("container ID is mandatory")
	}

	state, err := runtimeState(crioStateFile(containerID))
	if err != nil {
		return "", err
	}
	return strconv.Itoa(state.Pid), nil

}

func runtimeState(stateFile string) (rspec.State, error) {
	file, err := ioutil.ReadFile(stateFile)
	if err != nil {
		return rspec.State{}, errors.Wrapf(err, "read file failed: %s", stateFile)
	}

	var result rspec.State
	err = jsoniter.Unmarshal(file, &result)
	if err != nil {
		return rspec.State{}, errors.Wrapf(err, "unmarshal file failed: %s", stateFile)
	}

	return result, nil
}
