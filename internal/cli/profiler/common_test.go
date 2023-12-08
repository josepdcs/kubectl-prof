package profiler

import (
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_validatePodAndRetrieveContainerInfo(t *testing.T) {
	type args struct {
		pod *v1.Pod
		cfg *config.ProfilerConfig
	}
	tests := []struct {
		name  string
		given func() args
		when  func(args) error
		then  func(t *testing.T, cfg *config.ProfilerConfig, err error)
	}{
		{
			name: "should validate and set container info when one container",
			given: func() args {
				return args{
					pod: &v1.Pod{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name: "ContainerName",
								},
							},
						},
						Status: v1.PodStatus{
							ContainerStatuses: []v1.ContainerStatus{
								{
									Name:        "ContainerName",
									ContainerID: "ContainerID",
								},
							},
						},
					},
					cfg: &config.ProfilerConfig{
						Target: &config.TargetConfig{
							ContainerName: "ContainerName",
							PodName:       "PodName",
							Namespace:     "Namespace",
						},
						Job:                nil,
						EphemeralContainer: nil,
						LogLevel:           "",
					},
				}
			},
			when: func(args args) error {
				return validatePodAndRetrieveContainerInfo(args.pod, args.cfg)
			},
			then: func(t *testing.T, cfg *config.ProfilerConfig, err error) {
				require.NoError(t, err)
				assert.Equal(t, cfg.Target.ContainerName, "ContainerName")
				assert.Equal(t, cfg.Target.ContainerID, "ContainerID")
			},
		},
		{
			name: "should validate and set container info when more one container",
			given: func() args {
				return args{
					pod: &v1.Pod{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name: "ContainerName",
								},
								{
									Name: "ContainerName2",
								},
							},
						},
						Status: v1.PodStatus{
							ContainerStatuses: []v1.ContainerStatus{
								{
									Name:        "ContainerName",
									ContainerID: "ContainerID",
								},
								{
									Name:        "ContainerName2",
									ContainerID: "ContainerID2",
								},
							},
						},
					},
					cfg: &config.ProfilerConfig{
						Target: &config.TargetConfig{
							ContainerName: "ContainerName",
							PodName:       "PodName",
							Namespace:     "Namespace",
						},
						Job:                nil,
						EphemeralContainer: nil,
						LogLevel:           "",
					},
				}
			},
			when: func(args args) error {
				return validatePodAndRetrieveContainerInfo(args.pod, args.cfg)
			},
			then: func(t *testing.T, cfg *config.ProfilerConfig, err error) {
				require.NoError(t, err)
				assert.Equal(t, cfg.Target.ContainerName, "ContainerName")
				assert.Equal(t, cfg.Target.ContainerID, "ContainerID")
			},
		},
		{
			name: "should validate fail when not pod found",
			given: func() args {
				return args{
					pod: nil,
					cfg: &config.ProfilerConfig{
						Target: &config.TargetConfig{
							ContainerName: "ContainerName",
							PodName:       "PodName",
							Namespace:     "Namespace",
						},
						Job:                nil,
						EphemeralContainer: nil,
						LogLevel:           "",
					},
				}
			},
			when: func(args args) error {
				return validatePodAndRetrieveContainerInfo(args.pod, args.cfg)
			},
			then: func(t *testing.T, cfg *config.ProfilerConfig, err error) {
				require.Error(t, err)
				assert.Empty(t, cfg.Target.ContainerID)
				assert.EqualError(t, err, "Could not find pod PodName in Namespace Namespace")
			},
		},
		{
			name: "should validate fail when not match container status",
			given: func() args {
				return args{
					pod: &v1.Pod{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name: "ContainerName",
								},
								{
									Name: "ContainerName2",
								},
							},
						},
						Status: v1.PodStatus{
							ContainerStatuses: []v1.ContainerStatus{
								{
									Name:        "ContainerNameA",
									ContainerID: "ContainerID",
								},
								{
									Name:        "ContainerNameB",
									ContainerID: "ContainerID2",
								},
							},
						},
					},
					cfg: &config.ProfilerConfig{
						Target: &config.TargetConfig{
							ContainerName: "ContainerName",
							PodName:       "PodName",
							Namespace:     "Namespace",
						},
						Job:                nil,
						EphemeralContainer: nil,
						LogLevel:           "",
					},
				}
			},
			when: func(args args) error {
				return validatePodAndRetrieveContainerInfo(args.pod, args.cfg)
			},
			then: func(t *testing.T, cfg *config.ProfilerConfig, err error) {
				require.Error(t, err)
				assert.Empty(t, cfg.Target.ContainerID)
				assert.EqualError(t, err, "Could not find container id for ContainerName")
			},
		},
		{
			name: "should validate fail when not match container names",
			given: func() args {
				return args{
					pod: &v1.Pod{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name: "ContainerName1",
								},
								{
									Name: "ContainerName2",
								},
							},
						},
						Status: v1.PodStatus{
							ContainerStatuses: []v1.ContainerStatus{
								{
									Name:        "ContainerName1",
									ContainerID: "ContainerID",
								},
								{
									Name:        "ContainerName2",
									ContainerID: "ContainerID2",
								},
							},
						},
					},
					cfg: &config.ProfilerConfig{
						Target: &config.TargetConfig{
							ContainerName: "ContainerName",
							PodName:       "PodName",
							Namespace:     "Namespace",
						},
						Job:                nil,
						EphemeralContainer: nil,
						LogLevel:           "",
					},
				}
			},
			when: func(args args) error {
				return validatePodAndRetrieveContainerInfo(args.pod, args.cfg)
			},
			then: func(t *testing.T, cfg *config.ProfilerConfig, err error) {
				require.Error(t, err)
				assert.Empty(t, cfg.Target.ContainerID)
				assert.EqualError(t, err, "Could not determine container. please specify one of [ContainerName1 ContainerName2]")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			args := tt.given()

			// When
			err := tt.when(args)

			// Then
			tt.then(t, args.cfg, err)
		})
	}
}
