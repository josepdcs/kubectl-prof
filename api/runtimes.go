package api

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
	Crio:       "/var/lib/containers/storage/",
	Containerd: "/run/containerd/",
}

func AvailableContainerRuntimes() []ContainerRuntime {
	return containerRuntimes
}

func IsSupportedContainerRuntime(runtime string) bool {
	if runtime == string(FakeContainer) || runtime == string(FakeContainerWithRootFileSystemLocationResultError) ||
		runtime == string(FakeContainerWithPIDResultError) {
		return true
	}
	return containsContainerRuntime(ContainerRuntime(runtime), AvailableContainerRuntimes())
}

func containsContainerRuntime(cl ContainerRuntime, runtimes []ContainerRuntime) bool {
	for _, current := range runtimes {
		if cl == current {
			return true
		}
	}

	return false
}
