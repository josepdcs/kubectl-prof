package api

import "github.com/samber/lo"

type ContainerRuntime string

const (
	Crio                                               ContainerRuntime = "crio"
	Containerd                                         ContainerRuntime = "containerd"
	FakeContainer                                      ContainerRuntime = "fake"
	FakeContainerWithRootFileSystemLocationResultError ContainerRuntime = "fakeWithRootFileSystemLocationResultError"
	FakeContainerWithPIDResultError                    ContainerRuntime = "fakeWithPIDResultError"
)

var (
	containerRuntimes = []ContainerRuntime{Crio, Containerd}
)

var GetContainerRuntimeRootPath = map[ContainerRuntime]string{
	Crio:       "/var/lib/containers/storage",
	Containerd: "/run/containerd",
}

func AvailableContainerRuntimes() []ContainerRuntime {
	return containerRuntimes
}

func IsSupportedContainerRuntime(runtime string) bool {
	if runtime == string(FakeContainer) || runtime == string(FakeContainerWithRootFileSystemLocationResultError) ||
		runtime == string(FakeContainerWithPIDResultError) {
		return true
	}
	return lo.Contains(AvailableContainerRuntimes(), ContainerRuntime(runtime))
}
