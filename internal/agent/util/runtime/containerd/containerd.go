package containerd

import (
	"fmt"
	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/pkg/errors"
	"os"
)

type Containerd struct {
}

func NewContainerd() *Containerd {
	return &Containerd{}
}

var pidFile = func(containerID string, containerRuntimePath string) string {
	return fmt.Sprintf("%s/io.containerd.runtime.v2.task/k8s.io/%s/init.pid", containerRuntimePath, containerID)
}

var rootFS = func(containerID string, containerRuntimePath string) string {
	return fmt.Sprintf("%s/io.containerd.runtime.v2.task/k8s.io/%s/rootfs", containerRuntimePath, containerID)
}

func (c *Containerd) RootFileSystemLocation(containerID string, containerRuntimePath string) (string, error) {
	if stringUtils.IsBlank(containerID) {
		return "", errors.New("container ID is mandatory")
	}
	if stringUtils.IsBlank(containerRuntimePath) {
		return "", errors.New("container runtime path is mandatory")
	}

	return rootFS(containerID, containerRuntimePath), nil
}

func (c *Containerd) PID(containerID string, containerRuntimePath string) (string, error) {
	if stringUtils.IsBlank(containerID) {
		return "", errors.New("container ID is mandatory")
	}
	if stringUtils.IsBlank(containerRuntimePath) {
		return "", errors.New("container runtime path is mandatory")
	}

	file := pidFile(containerID, containerRuntimePath)
	PID, err := os.ReadFile(file)
	if err != nil {
		return "", errors.Wrapf(err, "read file failed: %s", file)
	}

	return string(PID), nil
}
