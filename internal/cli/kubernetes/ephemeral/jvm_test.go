package ephemeral

import (
	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_jvmCreator_Create(t *testing.T) {
	// Given
	targetPod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID: "UID",
		},
		Spec: v1.PodSpec{
			NodeName: "NodeName",
		},
	}
	cfg := &config.ProfilerConfig{
		Target: &config.TargetConfig{
			Namespace:            "Namespace",
			PodName:              "PodName",
			ContainerName:        "ContainerName",
			ContainerID:          "ContainerID",
			Event:                "Event",
			Duration:             100,
			Id:                   "ID",
			LocalPath:            "LocalPath",
			DryRun:               false,
			Image:                "Image",
			ContainerRuntime:     "ContainerRuntime",
			ContainerRuntimePath: "ContainerRuntimePath",
			Language:             "Language",
			Compressor:           "Compressor",
			ImagePullSecret:      "ImagePullSecret",
			ServiceAccountName:   "ServiceAccountName",
			ImagePullPolicy:      v1.PullAlways,
		},
		EphemeralContainer: &config.EphemeralContainerConfig{Privileged: true},
	}
	b := &jvmCreator{}

	// When
	result := b.Create(targetPod, cfg)

	// Then
	assert.True(t, stringUtils.StartsWith(result.Name, containerName))
	expected := &v1.EphemeralContainer{
		EphemeralContainerCommon: v1.EphemeralContainerCommon{
			Name:    result.Name,
			Image:   "Image",
			Command: []string{"/app/agent"},
			Args:    []string{"--target-container-runtime", "ContainerRuntime", "--target-pod-uid", "UID", "--target-container-id", "ContainerID", "--lang", "Language", "--event-type", "Event", "--compressor-type", "Compressor", "--profiling-tool", "", "--output-type", "", "--grace-period-ending", "0s", "--duration", "100ns"},
			SecurityContext: &v1.SecurityContext{
				Privileged: &cfg.EphemeralContainer.Privileged,
			},
			ImagePullPolicy: v1.PullAlways,
		},
		TargetContainerName: "ContainerName",
	}
	assert.Equal(t, expected, result)
}
