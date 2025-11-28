package api

import "github.com/samber/lo"

// ContainerRuntime represents a container runtime environment.
type ContainerRuntime string

const (
	Crio       ContainerRuntime = "crio"       // Crio represents the CRI-O container runtime.
	Containerd ContainerRuntime = "containerd" // Containerd represents the containerd container runtime.

	FakeContainer                                      ContainerRuntime = "fake"                                      // FakeContainer represents a fake container runtime for testing purposes.
	FakeContainerWithRootFileSystemLocationResultError ContainerRuntime = "fakeWithRootFileSystemLocationResultError" // FakeContainerWithRootFileSystemLocationResultError represents a fake container that simulates root filesystem location errors.
	FakeContainerWithPIDResultError                    ContainerRuntime = "fakeWithPIDResultError"                    // FakeContainerWithPIDResultError represents a fake container that simulates PID retrieval errors.
	FakeContainerWithCWDResultError                    ContainerRuntime = "fakeWithCWDResultError"                    // FakeContainerWithCWDResultError represents a fake container that simulates current working directory errors.
)

var (
	// containerRuntimes contains all supported container runtimes.
	containerRuntimes = []ContainerRuntime{Crio, Containerd}
)

// GetContainerRuntimeRootPath maps each ContainerRuntime to its root filesystem path.
var GetContainerRuntimeRootPath = map[ContainerRuntime]string{
	Crio:       "/var/lib/containers/storage",
	Containerd: "/run/containerd",
}

// AvailableContainerRuntimes returns the list of all supported container runtimes.
func AvailableContainerRuntimes() []ContainerRuntime {
	return containerRuntimes
}

// IsSupportedContainerRuntime checks if the given container runtime string is supported.
// It returns true if the runtime is in the list of available runtimes or is a fake runtime for testing.
func IsSupportedContainerRuntime(runtime string) bool {
	if runtime == string(FakeContainer) || runtime == string(FakeContainerWithRootFileSystemLocationResultError) ||
		runtime == string(FakeContainerWithPIDResultError) || runtime == string(FakeContainerWithCWDResultError) {
		return true
	}
	return lo.Contains(AvailableContainerRuntimes(), ContainerRuntime(runtime))
}
