package utils

import (
	"fmt"
	"github.com/josepdcs/kubectl-profile/api"
	jsoniter "github.com/json-iterator/go"
	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"io/ioutil"
)

var dockerMountIdLocation = func(containerID string) string {
	return fmt.Sprintf("/var/lib/docker/image/overlay2/layerdb/mounts/%s/mount-id", containerID)
}

var dockerTargetFileSystemLocation = func(mountID string) string {
	return fmt.Sprintf("/var/lib/docker/overlay2/%s/merged", string(mountID))
}

var crioConfigFile = func(containerID string) string {
	return fmt.Sprintf("/var/lib/containers/storage/overlay-containers/%s/userdata/config.json", containerID)
}

func TargetFileSystemLocation(runtime api.ContainerRuntime, containerID string) (string, error) {
	if runtime == "" || containerID == "" {
		return "", errors.New("container runtime and container ID are mandatory")
	}
	switch runtime {
	case api.Crio:
		spec, err := runtimeSpec(crioConfigFile(containerID))
		if err != nil {
			return "", err
		}
		return spec.Root.Path, nil
	case api.Containerd:
		return "", errors.New("containerd at not supported yet, coming soon...")
	case api.Docker:
		fileName := dockerMountIdLocation(containerID)
		mountID, err := ioutil.ReadFile(fileName)
		if err != nil {
			return "", errors.Wrapf(err, "read file failed: %s", fileName)
		}
		return dockerTargetFileSystemLocation(string(mountID)), nil
	default:
		return "", errors.Errorf("unsupported container runtime: %s", runtime)
	}
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
