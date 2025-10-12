package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiv1 "k8s.io/api/core/v1"
)

func TestJobConfig_ParseTolerations(t *testing.T) {
	tests := []struct {
		name           string
		tolerationsRaw []string
		want           []apiv1.Toleration
		wantErr        bool
	}{
		{
			name:           "empty tolerations",
			tolerationsRaw: []string{},
			want:           nil,
			wantErr:        false,
		},
		{
			name:           "key=value:effect format",
			tolerationsRaw: []string{"node.kubernetes.io/disk-pressure=true:NoSchedule"},
			want: []apiv1.Toleration{
				{
					Key:      "node.kubernetes.io/disk-pressure",
					Operator: apiv1.TolerationOpEqual,
					Value:    "true",
					Effect:   apiv1.TaintEffectNoSchedule,
				},
			},
			wantErr: false,
		},
		{
			name:           "key:effect format",
			tolerationsRaw: []string{"node.kubernetes.io/memory-pressure:NoExecute"},
			want: []apiv1.Toleration{
				{
					Key:      "node.kubernetes.io/memory-pressure",
					Operator: apiv1.TolerationOpExists,
					Effect:   apiv1.TaintEffectNoExecute,
				},
			},
			wantErr: false,
		},
		{
			name:           "key only format (defaults to NoSchedule)",
			tolerationsRaw: []string{"node.kubernetes.io/unreachable"},
			want: []apiv1.Toleration{
				{
					Key:      "node.kubernetes.io/unreachable",
					Operator: apiv1.TolerationOpExists,
					Effect:   apiv1.TaintEffectNoSchedule,
				},
			},
			wantErr: false,
		},
		{
			name: "multiple tolerations",
			tolerationsRaw: []string{
				"node.kubernetes.io/disk-pressure=true:NoSchedule",
				"node.kubernetes.io/memory-pressure:NoExecute",
				"key1=value1:PreferNoSchedule",
			},
			want: []apiv1.Toleration{
				{
					Key:      "node.kubernetes.io/disk-pressure",
					Operator: apiv1.TolerationOpEqual,
					Value:    "true",
					Effect:   apiv1.TaintEffectNoSchedule,
				},
				{
					Key:      "node.kubernetes.io/memory-pressure",
					Operator: apiv1.TolerationOpExists,
					Effect:   apiv1.TaintEffectNoExecute,
				},
				{
					Key:      "key1",
					Operator: apiv1.TolerationOpEqual,
					Value:    "value1",
					Effect:   apiv1.TaintEffectPreferNoSchedule,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JobConfig{
				TolerationsRaw: tt.tolerationsRaw,
			}
			err := j.ParseTolerations()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, j.Tolerations)
			}
		})
	}
}

func TestJobConfig_DeepCopy(t *testing.T) {
	original := &JobConfig{
		ContainerConfig: ContainerConfig{
			Privileged: true,
		},
		Namespace: "test-namespace",
		Tolerations: []apiv1.Toleration{
			{
				Key:      "key1",
				Operator: apiv1.TolerationOpEqual,
				Value:    "value1",
				Effect:   apiv1.TaintEffectNoSchedule,
			},
		},
		TolerationsRaw: []string{"key1=value1:NoSchedule"},
	}

	copied := original.DeepCopy()

	// Verify the copy is equal
	assert.Equal(t, original.Namespace, copied.Namespace)
	assert.Equal(t, original.Privileged, copied.Privileged)
	assert.Equal(t, original.Tolerations, copied.Tolerations)
	assert.Equal(t, original.TolerationsRaw, copied.TolerationsRaw)

	// Verify modifying the copy doesn't affect the original
	copied.Namespace = "modified-namespace"
	copied.Tolerations[0].Key = "modified-key"
	copied.TolerationsRaw[0] = "modified"

	assert.NotEqual(t, original.Namespace, copied.Namespace)
	assert.NotEqual(t, original.Tolerations[0].Key, copied.Tolerations[0].Key)
	assert.NotEqual(t, original.TolerationsRaw[0], copied.TolerationsRaw[0])
}
