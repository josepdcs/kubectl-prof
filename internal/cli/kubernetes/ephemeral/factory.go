package ephemeral

import (
	"errors"
	"fmt"

	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	command       = "/app/agent"
	baseImageName = "josepdcs/kubectl-prof"
	containerName = "kubectl-prof"
)

// Creator defines the method for creating the ephemeral container according the programming language.
type Creator interface {
	// Create creates a profiling job for the given target pod and configuration
	Create(targetPod *v1.Pod, cfg *config.ProfilerConfig) *v1.EphemeralContainer
}

// Get returns the Creator implementation according the programming language.
func Get(lang api.ProgrammingLanguage) (Creator, error) {
	switch lang {
	case api.Java:
		return &jvmCreator{}, nil
	case api.Go, api.Node:
		return &bpfCreator{}, nil
	case api.Python:
		return &pythonCreator{}, nil
	case api.FakeLang:
		// for unit tests purpose
		return &fakeCreator{}, nil
	}

	// Should not happen
	return nil, errors.New("got language without job creator")
}

// getImageName returns the container name: constant plus a random value.
func getContainerName() string {
	return fmt.Sprintf("%s-%s", containerName, rand.String(5))
}
