// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	return fmt.Sprintf("/var/lib/docker/overlay2/%s/merged", mountID)
}

var crioConfigFile = func(containerID string) string {
	return fmt.Sprintf("/var/lib/containers/storage/overlay-containers/%s/userdata/config.json", containerID)
}

var crioStateFile = func(containerID string) string {
	return fmt.Sprintf("/var/lib/containers/storage/overlay-containers/%s/userdata/state.json", containerID)
}

// deprecated
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
