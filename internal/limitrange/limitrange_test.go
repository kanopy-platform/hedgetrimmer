package limitrange

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestIsLimitRangeTypeContainer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		limitRange corev1.LimitRangeItem
		want       bool
		msg        string
	}{
		{
			limitRange: corev1.LimitRangeItem{Type: corev1.LimitTypeContainer},
			want:       true,
			msg:        "Container type",
		},
		{
			limitRange: corev1.LimitRangeItem{Type: corev1.LimitTypePod},
			want:       false,
			msg:        "Pod type",
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.want, IsLimitRangeTypeContainer(test.limitRange), test.msg)
	}
}

func TestGetMemoryConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		limitRange corev1.LimitRangeItem
		want       MemoryConfig
		msg        string
	}{
		{
			limitRange: corev1.LimitRangeItem{
				DefaultRequest: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
				Default: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("2Gi"),
				},
				MaxLimitRequestRatio: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("1.5"),
				},
			},
			want: MemoryConfig{
				HasDefaultMemoryRequest:       true,
				HasDefaultMemoryLimit:         true,
				HasMaxLimitRequestMemoryRatio: true,
				DefaultMemoryRequest:          resource.MustParse("1Gi"),
				DefaultMemoryLimit:            resource.MustParse("2Gi"),
				MaxLimitRequestMemoryRatio:    resource.MustParse("1.5"),
			},
			msg: "Memory resource request, limit, and ratio exists",
		},
		{
			limitRange: corev1.LimitRangeItem{
				DefaultRequest: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("500m"),
				},
				Default: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("5Gi"),
				},
				MaxLimitRequestRatio: corev1.ResourceList{},
			},
			want: MemoryConfig{
				HasDefaultMemoryRequest:       false,
				HasDefaultMemoryLimit:         false,
				HasMaxLimitRequestMemoryRatio: false,
			},
			msg: "Memory resource request, limit, and ratio does not exist",
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.want, GetMemoryConfig(test.limitRange), test.msg)
	}
}
