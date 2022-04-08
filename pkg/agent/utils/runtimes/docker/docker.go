package docker

import (
	"fmt"
	"github.com/josepdcs/kubectl-prof/pkg/agent/config"
	"github.com/pkg/errors"
	"io/ioutil"
)

type Docker struct {
}

func NewDocker() *Docker {
	return &Docker{}
}

var dockerMountIdLocation = func(containerID string) string {
	return fmt.Sprintf("/var/lib/docker/image/overlay2/layerdb/mounts/%s/mount-id", containerID)
}

var dockerTargetFileSystemLocation = func(mountID string) string {
	return fmt.Sprintf("/var/lib/docker/overlay2/%s/merged", mountID)
}

func (d *Docker) RootFileSystemLocation(containerID string) (string, error) {
	fileName := dockerMountIdLocation(containerID)
	mountID, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", errors.Wrapf(err, "read file failed: %s", fileName)
	}
	return dockerTargetFileSystemLocation(string(mountID)), nil

}

func (d *Docker) PID(job *config.ProfilingJob) (string, error) {
	return FindProcessId(job)
}

func (d *Docker) PPID(job *config.ProfilingJob) (string, error) {
	return FindRootProcessId(job)
}
