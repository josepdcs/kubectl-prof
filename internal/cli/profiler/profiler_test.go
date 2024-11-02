package profiler

import (
	"testing"

	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/profiler/api/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobProfiler_Profile(t *testing.T) {
	type fields struct {
		*Profiler
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
			name: "should fail when when no target is provided",
			given: func() (fields, args) {
				return fields{
						Profiler: NewJobProfiler(
							fake.NewPodApi().WithReturnsEmpty(),
							fake.NewProfilingJobApi(),
							fake.NewProfilingContainerApi(),
						),
					},
					args{
						cfg: &config.ProfilerConfig{
							Target: &config.TargetConfig{
								Namespace:     "Namespace",
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
				assert.EqualError(t, err, "no target specified")
			},
		},
		{
			name: "should profile one pod",
			given: func() (fields, args) {
				return fields{
						Profiler: NewJobProfiler(
							fake.NewPodApi(),
							fake.NewProfilingJobApi(),
							fake.NewProfilingContainerApi(),
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
			name: "should profile a bunch of pods",
			given: func() (fields, args) {
				return fields{
						Profiler: NewJobProfiler(
							fake.NewPodApi(),
							fake.NewProfilingJobApi(),
							fake.NewProfilingContainerApi(),
						),
					},
					args{
						cfg: &config.ProfilerConfig{
							Target: &config.TargetConfig{
								Namespace:     "Namespace",
								ContainerID:   "ContainerID",
								DryRun:        false,
								LabelSelector: "app=app",
							},
							Job:      &config.JobConfig{},
							LogLevel: api.InfoLevel,
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
						Profiler: NewJobProfiler(
							fake.NewPodApi(),
							fake.NewProfilingJobApi(),
							fake.NewProfilingContainerApi(),
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
						Profiler: NewJobProfiler(
							fake.NewPodApi().WithReturnsError(),
							fake.NewProfilingJobApi(),
							fake.NewProfilingContainerApi(),
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
						Profiler: NewJobProfiler(
							fake.NewPodApi().WithReturnsEmpty(),
							fake.NewProfilingJobApi(),
							fake.NewProfilingContainerApi(),
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
			name: "should fail when get bunch of pods fail",
			given: func() (fields, args) {
				return fields{
						Profiler: NewJobProfiler(
							fake.NewPodApi().WithReturnsError(),
							fake.NewProfilingJobApi(),
							fake.NewProfilingContainerApi(),
						),
					},
					args{
						cfg: &config.ProfilerConfig{
							Target: &config.TargetConfig{
								Namespace:     "Namespace",
								ContainerName: "ContainerName",
								ContainerID:   "ContainerID",
								DryRun:        false,
								LabelSelector: "app=app",
							},
						},
					}
			},
			when: func(f fields, args args) error {
				return f.Profile(args.cfg)
			},
			then: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.EqualError(t, err, "error getting pods")
			},
		},
		{
			name: "should fail when get bunch of pods returns empty",
			given: func() (fields, args) {
				return fields{
						Profiler: NewJobProfiler(
							fake.NewPodApi().WithReturnsEmpty(),
							fake.NewProfilingJobApi(),
							fake.NewProfilingContainerApi(),
						),
					},
					args{
						cfg: &config.ProfilerConfig{
							Target: &config.TargetConfig{
								Namespace:     "Namespace",
								ContainerName: "ContainerName",
								ContainerID:   "ContainerID",
								DryRun:        false,
								LabelSelector: "app=app",
							},
						},
					}
			},
			when: func(f fields, args args) error {
				return f.Profile(args.cfg)
			},
			then: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.EqualError(t, err, "No pods found in namespace Namespace with label selector app=app")
			},
		},
		{
			name: "should fail when create profiling job fail",
			given: func() (fields, args) {
				return fields{
						Profiler: NewJobProfiler(
							fake.NewPodApi(),
							fake.NewProfilingJobApi().WithCreateProfilingJobReturnsError(),
							fake.NewProfilingContainerApi(),
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
						Profiler: NewJobProfiler(
							fake.NewPodApi(),
							fake.NewProfilingJobApi().WithGetProfilingPodReturnsError(),
							fake.NewProfilingContainerApi(),
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
						Profiler: NewJobProfiler(
							fake.NewPodApi(),
							fake.NewProfilingJobApi(),
							fake.NewProfilingContainerApi().WithHandleProfilingContainerLogsReturnsError(),
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
						Profiler: NewJobProfiler(
							fake.NewPodApi(),
							fake.NewProfilingJobApi(),
							fake.NewProfilingContainerApi().WithGetRemoteFileReturnsError(),
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
