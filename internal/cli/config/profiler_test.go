package config

import (
	"testing"

	"github.com/josepdcs/kubectl-prof/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProfilerConfig(t *testing.T) {
	type args struct {
		Target  *TargetConfig
		options []Option
	}
	tests := []struct {
		name  string
		given func() args
		when  func(args) (*ProfilerConfig, error)
		then  func(*testing.T, *ProfilerConfig, error)
	}{
		{
			name: "With Job",
			given: func() args {
				return args{
					Target: &TargetConfig{
						Namespace: "Namespace",
					},
					options: []Option{
						WithJob(&JobConfig{
							ContainerConfig: ContainerConfig{
								Privileged: true,
							},
							Namespace: "Namespace",
						}),
						WithLogLevel(api.InfoLevel),
					},
				}
			},
			when: func(args args) (*ProfilerConfig, error) {
				return NewProfilerConfig(args.Target, args.options...)
			},
			then: func(t *testing.T, config *ProfilerConfig, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, config)
				assert.NotEmpty(t, config.Target)
				assert.NotEmpty(t, config.Job)
				assert.Nil(t, config.EphemeralContainer)
			},
		},
		{
			name: "With EphemeralContainer",
			given: func() args {
				return args{
					Target: &TargetConfig{
						Namespace: "Namespace",
					},
					options: []Option{
						WithEphemeralContainer(&EphemeralContainerConfig{
							Privileged: true,
						}),
						WithLogLevel(api.InfoLevel),
					},
				}
			},
			when: func(args args) (*ProfilerConfig, error) {
				return NewProfilerConfig(args.Target, args.options...)
			},
			then: func(t *testing.T, config *ProfilerConfig, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, config)
				assert.NotEmpty(t, config.Target)
				assert.Nil(t, config.Job)
				assert.NotEmpty(t, config.EphemeralContainer)
			},
		},
		{
			name: "with none Job or EphemeralContainer should fail",
			given: func() args {
				return args{
					Target: &TargetConfig{
						Namespace: "Namespace",
					},
					options: []Option{WithLogLevel(api.InfoLevel)},
				}
			},
			when: func(args args) (*ProfilerConfig, error) {
				return NewProfilerConfig(args.Target, args.options...)
			},
			then: func(t *testing.T, config *ProfilerConfig, err error) {
				require.Error(t, err)
				assert.EqualError(t, err, "JobConfig and EphemeralContainerConfig are missing. One of both is mandatory")
			},
		},
		{
			name: "with both Job and EphemeralContainer should fail",
			given: func() args {
				return args{
					Target: &TargetConfig{
						Namespace: "Namespace",
					},
					options: []Option{
						WithJob(&JobConfig{
							ContainerConfig: ContainerConfig{
								Privileged: true,
							},
							Namespace: "Namespace",
						}),
						WithEphemeralContainer(&EphemeralContainerConfig{
							Privileged: true,
						}),
						WithLogLevel(api.InfoLevel),
					},
				}
			},
			when: func(args args) (*ProfilerConfig, error) {
				return NewProfilerConfig(args.Target, args.options...)
			},
			then: func(t *testing.T, config *ProfilerConfig, err error) {
				require.Error(t, err)
				assert.EqualError(t, err, "JobConfig and EphemeralContainerConfig cannot be defined at the same time")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			args := tt.given()

			// When
			result, err := tt.when(args)

			// Then
			tt.then(t, result, err)
		})
	}
}
