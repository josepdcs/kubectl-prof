package profiler

import (
	"github.com/josepdcs/kubectl-prof/internal/cli/adapter/fake"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestJobProfiler_Profile(t *testing.T) {
	type fields struct {
		JobProfiler
	}
	type args struct {
		cfg *config.ProfilerConfig
	}
	tests := []struct {
		name  string
		given func() (fields, args)
		when  func(fields, args) error
		then  func(t *testing.T, err error)
	}{
		{
			name: "should profile",
			given: func() (fields, args) {
				return fields{
						JobProfiler: NewJobProfiler(
							fake.NewPodAdapter(),
							fake.NewProfilingJobAdapter(),
							fake.NewProfilingContainerAdapter(),
						),
					},
					args{
						cfg: &config.ProfilerConfig{
							Target: &config.TargetConfig{
								Namespace:     "Namespace",
								PodName:       "PodName",
								ContainerName: "ContainerName",
								ContainerID:   "ContainerID",
								DryRun:        false,
							},
						},
					}
			},
			when: func(f fields, args args) error {
				return f.Profile(args.cfg)
			},
			then: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "should skip when dry run",
			given: func() (fields, args) {
				return fields{
						JobProfiler: NewJobProfiler(
							fake.NewPodAdapter(),
							fake.NewProfilingJobAdapter(),
							fake.NewProfilingContainerAdapter(),
						),
					},
					args{
						cfg: &config.ProfilerConfig{
							Target: &config.TargetConfig{
								Namespace:     "Namespace",
								PodName:       "PodName",
								ContainerName: "ContainerName",
								ContainerID:   "ContainerID",
								DryRun:        true,
							},
						},
					}
			},
			when: func(f fields, args args) error {
				return f.Profile(args.cfg)
			},
			then: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "should fail when get pod fail",
			given: func() (fields, args) {
				return fields{
						JobProfiler: NewJobProfiler(
							fake.NewPodAdapter().WithGetPodReturnsError(),
							fake.NewProfilingJobAdapter(),
							fake.NewProfilingContainerAdapter(),
						),
					},
					args{
						cfg: &config.ProfilerConfig{
							Target: &config.TargetConfig{
								Namespace:     "Namespace",
								PodName:       "PodName",
								ContainerName: "ContainerName",
								ContainerID:   "ContainerID",
								DryRun:        false,
							},
						},
					}
			},
			when: func(f fields, args args) error {
				return f.Profile(args.cfg)
			},
			then: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.EqualError(t, err, "error getting pod")
			},
		},
		{
			name: "should fail when get an invalid pod",
			given: func() (fields, args) {
				return fields{
						JobProfiler: NewJobProfiler(
							fake.NewPodAdapter().WithGetPodReturnsAnInvalidPod(),
							fake.NewProfilingJobAdapter(),
							fake.NewProfilingContainerAdapter(),
						),
					},
					args{
						cfg: &config.ProfilerConfig{
							Target: &config.TargetConfig{
								Namespace:     "Namespace",
								PodName:       "PodName",
								ContainerName: "ContainerName",
								ContainerID:   "ContainerID",
								DryRun:        false,
							},
						},
					}
			},
			when: func(f fields, args args) error {
				return f.Profile(args.cfg)
			},
			then: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.EqualError(t, err, "Could not find pod PodName in Namespace Namespace")
			},
		},
		{
			name: "should fail when create profiling job fail",
			given: func() (fields, args) {
				return fields{
						JobProfiler: NewJobProfiler(
							fake.NewPodAdapter(),
							fake.NewProfilingJobAdapter().WithCreateProfilingJobReturnsError(),
							fake.NewProfilingContainerAdapter(),
						),
					},
					args{
						cfg: &config.ProfilerConfig{
							Target: &config.TargetConfig{
								Namespace:     "Namespace",
								PodName:       "PodName",
								ContainerName: "ContainerName",
								ContainerID:   "ContainerID",
								DryRun:        false,
							},
						},
					}
			},
			when: func(f fields, args args) error {
				return f.Profile(args.cfg)
			},
			then: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.EqualError(t, err, "error creating profiling job")
			},
		},
		{
			name: "should fail when get profiling pod fail",
			given: func() (fields, args) {
				return fields{
						JobProfiler: NewJobProfiler(
							fake.NewPodAdapter(),
							fake.NewProfilingJobAdapter().WithGetProfilingPodReturnsError(),
							fake.NewProfilingContainerAdapter(),
						),
					},
					args{
						cfg: &config.ProfilerConfig{
							Target: &config.TargetConfig{
								Namespace:     "Namespace",
								PodName:       "PodName",
								ContainerName: "ContainerName",
								ContainerID:   "ContainerID",
								DryRun:        false,
							},
						},
					}
			},
			when: func(f fields, args args) error {
				return f.Profile(args.cfg)
			},
			then: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.EqualError(t, err, "error getting profiling pod")
			},
		},
		{
			name: "should fail when handle profiling container logs fail",
			given: func() (fields, args) {
				return fields{
						JobProfiler: NewJobProfiler(
							fake.NewPodAdapter(),
							fake.NewProfilingJobAdapter(),
							fake.NewProfilingContainerAdapter().WithHandleProfilingContainerLogsReturnsError(),
						),
					},
					args{
						cfg: &config.ProfilerConfig{
							Target: &config.TargetConfig{
								Namespace:     "Namespace",
								PodName:       "PodName",
								ContainerName: "ContainerName",
								ContainerID:   "ContainerID",
								DryRun:        false,
							},
						},
					}
			},
			when: func(f fields, args args) error {
				return f.Profile(args.cfg)
			},
			then: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.EqualError(t, err, "error handling profiling container logs")
			},
		},
		{
			name: "should terminate when get remote file fail",
			given: func() (fields, args) {
				return fields{
						JobProfiler: NewJobProfiler(
							fake.NewPodAdapter(),
							fake.NewProfilingJobAdapter(),
							fake.NewProfilingContainerAdapter().WithGetRemoteFileReturnsError(),
						),
					},
					args{
						cfg: &config.ProfilerConfig{
							Target: &config.TargetConfig{
								Namespace:     "Namespace",
								PodName:       "PodName",
								ContainerName: "ContainerName",
								ContainerID:   "ContainerID",
								DryRun:        false,
							},
						},
					}
			},
			when: func(f fields, args args) error {
				return f.Profile(args.cfg)
			},
			then: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			fields, args := tt.given()

			// When
			err := tt.when(fields, args)

			// Then
			tt.then(t, err)
		})
	}
}
