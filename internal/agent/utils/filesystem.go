package utils

import (
	"fmt"
	"io/ioutil"
)

const (
	dockerMountIdLocation          = "/var/lib/docker/image/overlay2/layerdb/mounts/%s/mount-id"
	dockerTargetFileSystemLocation = "/var/lib/docker/overlay2/%s/merged"

	crioConfigLocation = "/var/lib/containers/storage/overlay-containers/%s/userdata/config.json"
)

func GetTargetFileSystemLocation(containerId string) (string, error) {
	fileName := fmt.Sprintf(dockerMountIdLocation, containerId)
	mountId, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(dockerTargetFileSystemLocation, string(mountId)), nil
}
