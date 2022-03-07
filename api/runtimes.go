package api

type ContainerRuntime string

const (
	Docker     ContainerRuntime = "docker"
	Crio       ContainerRuntime = "crio"
	Containerd ContainerRuntime = "containerd"
)

var (
	containerRuntimes = []ContainerRuntime{Docker, Crio, Containerd}
)

var GetContainerRuntimePath = map[ContainerRuntime]string{
	Docker:     "/var/lib/docker/",
	Crio:       "/var/lib/containers/storage/",
	Containerd: "/var/lib/containerd/",
}

func AvailableContainerRuntimes() []ContainerRuntime {
	return containerRuntimes
}

func IsSupportedContainerRuntime(runtime string) bool {
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
