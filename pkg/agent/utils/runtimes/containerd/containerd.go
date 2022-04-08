package containerd

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
)

type Containerd struct {
}

func NewContainerd() *Containerd {
	return &Containerd{}
}

var pidFile = func(containerID string) string {
	return fmt.Sprintf("/run/containerd/io.containerd.runtime.v2.task/k8s.io/%s/init.pid", containerID)
}

var rootFS = func(containerID string) string {
	return fmt.Sprintf("/run/containerd/io.containerd.runtime.v2.task/k8s.io/%s/rootfs", containerID)
}

func (c *Containerd) RootFileSystemLocation(containerID string) (string, error) {
	if containerID == "" {
		return "", errors.New("container ID is mandatory")
	}

	return rootFS(containerID), nil
}

func (c *Containerd) PID(containerID string) (string, error) {
	if containerID == "" {
		return "", errors.New("container ID is mandatory")
	}

	file := pidFile(containerID)
	PID, err := ioutil.ReadFile(file)
	if err != nil {
		return "", errors.Wrapf(err, "read file failed: %s", file)
	}

	return string(PID), nil
}
