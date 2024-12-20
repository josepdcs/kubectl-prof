package crio

import (
	"fmt"
	"os"
	"strconv"

	"github.com/agrison/go-commons-lang/stringUtils"
	jsoniter "github.com/json-iterator/go"
	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
)

const (
	containerIDMandaToryError = "container ID is mandatory"
	containerRuntimePathError = "container runtime path is mandatory"
)

type Crio struct {
}

func NewCrio() *Crio {
	return &Crio{}
}

var crioConfigFile = func(containerID string, containerRuntimePath string) string {
	return fmt.Sprintf("%s/overlay-containers/%s/userdata/config.json", containerRuntimePath, containerID)
}

var crioStateFile = func(containerID string, containerRuntimePath string) string {
	return fmt.Sprintf("%s/overlay-containers/%s/userdata/state.json", containerRuntimePath, containerID)
}

// RootFileSystemLocation returns the root filesystem location of the container
func (c *Crio) RootFileSystemLocation(containerID string, containerRuntimePath string) (string, error) {
	if stringUtils.IsBlank(containerID) {
		return "", errors.New(containerIDMandaToryError)
	}
	if stringUtils.IsBlank(containerRuntimePath) {
		return "", errors.New(containerRuntimePathError)
	}

	spec, err := runtimeSpec(crioConfigFile(containerID, containerRuntimePath))
	if err != nil {
		return "", err
	}
	return spec.Root.Path, nil

}

// PID returns the PID of the container
func runtimeSpec(configFile string) (rspec.Spec, error) {
	file, err := os.ReadFile(configFile)
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

// PID returns the PID of the container
func (c *Crio) PID(containerID string, containerRuntimePath string) (string, error) {
	if stringUtils.IsBlank(containerID) {
		return "", errors.New(containerIDMandaToryError)
	}
	if stringUtils.IsBlank(containerRuntimePath) {
		return "", errors.New(containerRuntimePathError)
	}

	state, err := runtimeState(crioStateFile(containerID, containerRuntimePath))
	if err != nil {
		return "", err
	}
	return strconv.Itoa(state.Pid), nil

}

// runtimeState reads the runtime state from the container runtime
func runtimeState(stateFile string) (rspec.State, error) {
	file, err := os.ReadFile(stateFile)
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

// CWD returns the current working directory of the container
func (c *Crio) CWD(containerID string, containerRuntimePath string) (string, error) {
	if stringUtils.IsBlank(containerID) {
		return "", errors.New(containerIDMandaToryError)
	}
	if stringUtils.IsBlank(containerRuntimePath) {
		return "", errors.New(containerRuntimePathError)
	}

	spec, err := runtimeSpec(crioConfigFile(containerID, containerRuntimePath))
	if err != nil {
		return "", err
	}
	return spec.Process.Cwd, nil

}
