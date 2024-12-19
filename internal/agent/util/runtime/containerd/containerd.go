package containerd

import (
	"fmt"
	"os"

	"github.com/agrison/go-commons-lang/stringUtils"
	jsoniter "github.com/json-iterator/go"
	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
)

type Containerd struct {
}

func NewContainerd() *Containerd {
	return &Containerd{}
}

var pidFile = func(containerID string, containerRuntimePath string) string {
	return fmt.Sprintf("%s/io.containerd.runtime.v2.task/k8s.io/%s/init.pid", containerRuntimePath, containerID)
}

var pidContainerIDFile = func(containerID string, containerRuntimePath string) string {
	return fmt.Sprintf("%s/io.containerd.runtime.v2.task/k8s.io/%s/%s.pid", containerRuntimePath, containerID, containerID)
}

var rootFS = func(containerID string, containerRuntimePath string) string {
	return fmt.Sprintf("%s/io.containerd.runtime.v2.task/k8s.io/%s/rootfs", containerRuntimePath, containerID)
}

var configFile = func(containerID string, containerRuntimePath string) string {
	return fmt.Sprintf("%s/io.containerd.runtime.v2.task/k8s.io/%s/config.json", containerRuntimePath, containerID)
}

// RootFileSystemLocation returns the root filesystem location of the container
func (c *Containerd) RootFileSystemLocation(containerID string, containerRuntimePath string) (string, error) {
	if stringUtils.IsBlank(containerID) {
		return "", errors.New("container ID is mandatory")
	}
	if stringUtils.IsBlank(containerRuntimePath) {
		return "", errors.New("container runtime path is mandatory")
	}

	return rootFS(containerID, containerRuntimePath), nil
}

// PID returns the PID of the container
func (c *Containerd) PID(containerID string, containerRuntimePath string) (string, error) {
	if stringUtils.IsBlank(containerID) {
		return "", errors.New("container ID is mandatory")
	}
	if stringUtils.IsBlank(containerRuntimePath) {
		return "", errors.New("container runtime path is mandatory")
	}

	file := pidFile(containerID, containerRuntimePath)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		file = pidContainerIDFile(containerID, containerRuntimePath)
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return "", errors.Wrapf(err, "pid file not found: %s", file)
		}
	}
	PID, err := os.ReadFile(file)
	if err != nil {
		return "", errors.Wrapf(err, "read file failed: %s", file)
	}

	return string(PID), nil
}

// runtimeSpec reads the runtime spec from the container runtime
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

// CWD returns the current working directory of the container
func (c *Containerd) CWD(containerID string, containerRuntimePath string) (string, error) {
	if stringUtils.IsBlank(containerID) {
		return "", errors.New("container ID is mandatory")
	}
	if stringUtils.IsBlank(containerRuntimePath) {
		return "", errors.New("container runtime path is mandatory")
	}

	spec, err := runtimeSpec(configFile(containerID, containerRuntimePath))
	if err != nil {
		return "", err
	}
	return spec.Process.Cwd, nil

}
