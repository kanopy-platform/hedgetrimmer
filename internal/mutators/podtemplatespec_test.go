package mutators

import (
	"testing"

	"github.com/kanopy-platform/hedgetrimmer/internal/limitrange"
	"github.com/kanopy-platform/hedgetrimmer/internal/quantity"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestApplyResourceRequirements(t *testing.T) {
	t.Parallel()

	limitRangeItem := corev1.LimitRangeItem{
		Type:           corev1.LimitTypeContainer,
		DefaultRequest: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("50Mi")},
		Default:        corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("64Mi")},
	}

	tests := []struct {
		inputPts  corev1.PodTemplateSpec
		inputLri  corev1.LimitRangeItem
		want      corev1.PodTemplateSpec
		wantError bool
		msg       string
	}{
		{
			inputPts: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name:      "init-container-1",
							Resources: corev1.ResourceRequirements{}, // empty requests and limits
						},
					},
					Containers: []corev1.Container{
						{
							Name: "container-1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("50Gi")}, // no memory request
								Limits:   corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("500m")},     // no memory limit
							},
						},
						{
							Name: "container-2",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("2Gi")}, // has memory request
								Limits:   corev1.ResourceList{},                                                 // no memory limit
							},
						},
					},
				},
			},
			inputLri: limitRangeItem,
			want: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name: "init-container-1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("50Mi")},                        // populated with default
								Limits:   corev1.ResourceList{corev1.ResourceMemory: *quantity.Ptr(resource.MustParse("64Mi")).ToDec()}, // populated with default
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name: "container-1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceStorage: resource.MustParse("50Gi"),
									corev1.ResourceMemory:  resource.MustParse("50Mi"), // populated with default
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: *quantity.Ptr(resource.MustParse("64Mi")).ToDec(), // populated with default
								},
							},
						},
						{
							Name: "container-2",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("2Gi")},   // do not override user config
								Limits:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("2.2Gi")}, // use defaultLimitRequestMemoryRatio
							},
						},
					},
				},
			},
			wantError: false,
			msg:       "Successfully applies requests and limits",
		},
		{
			inputPts: corev1.PodTemplateSpec{},
			inputLri: corev1.LimitRangeItem{
				Type: corev1.LimitTypePod,
			},
			wantError: true,
			msg:       "Incorrect input LimitRangeItem of type Pod",
		},
	}

	for _, test := range tests {
		pts := NewPtsMutator()
		result, err := pts.ApplyResourceRequirements(test.inputPts, test.inputLri)
		if test.wantError {
			assert.Error(t, err, test.msg)
		} else {
			assert.NoError(t, err, test.msg)

			// round up to scale 0 to avoid Scale differences
			roundUpContainersResourceQuantityScale(test.want.Spec.InitContainers, 0)
			roundUpContainersResourceQuantityScale(test.want.Spec.Containers, 0)

			roundUpContainersResourceQuantityScale(result.Spec.InitContainers, 0)
			roundUpContainersResourceQuantityScale(result.Spec.Containers, 0)

			assert.Equal(t, test.want, result, test.msg)
		}
	}
}

func TestValidateMemoryRatio(t *testing.T) {
	t.Parallel()

	memoryConfig := limitrange.Config{
		HasMaxLimitRequestRatio: true,
		MaxLimitRequestRatio:    resource.MustParse("1.25"),
	}

	tests := []struct {
		requests  corev1.ResourceList
		limits    corev1.ResourceList
		mc        limitrange.Config
		wantError bool
		msg       string
	}{
		{
			requests:  corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1Gi")},
			limits:    corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1.25Gi")},
			mc:        memoryConfig,
			wantError: false,
			msg:       "Memory limit/request ratio equals max ratio, allow",
		},
		{
			requests:  corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1Gi")},
			limits:    corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1.26Gi")},
			mc:        memoryConfig,
			wantError: true,
			msg:       "Memory limit/request ratio exceeds max ratio, error",
		},
		{
			requests:  corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1Gi")},
			limits:    corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1.26Gi")},
			mc:        limitrange.Config{HasMaxLimitRequestRatio: false},
			wantError: false,
			msg:       "LimitRange does not specify max ratio, no ratio check",
		},
		{
			requests:  corev1.ResourceList{},
			limits:    corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1.26Gi")},
			mc:        memoryConfig,
			wantError: true,
			msg:       "Memory request does not exist, error",
		},
		{
			requests:  corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1Gi")},
			limits:    corev1.ResourceList{},
			mc:        memoryConfig,
			wantError: true,
			msg:       "Memory limit does not exist, error",
		},
	}

	for _, test := range tests {
		container := corev1.Container{
			Resources: corev1.ResourceRequirements{
				Requests: test.requests,
				Limits:   test.limits,
			},
		}

		err := validateMemoryRequirements(container, test.mc)
		if test.wantError {
			assert.Error(t, err, test.msg)
		} else {
			assert.NoError(t, err, test.msg)
		}
	}
}

func TestSetMemoryRequest(t *testing.T) {
	t.Parallel()

	memoryConfig := limitrange.Config{
		HasDefaultRequest: true,
		DefaultRequest:    resource.MustParse("5Gi"),
	}

	tests := []struct {
		requests     corev1.ResourceList
		mc           limitrange.Config
		wantRequests corev1.ResourceList
		msg          string
	}{
		{
			requests:     corev1.ResourceList{},
			mc:           memoryConfig,
			wantRequests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("5Gi")},
			msg:          "Memory request does not exist, set to default",
		},
		{
			requests:     corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1Gi")},
			mc:           memoryConfig,
			wantRequests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1Gi")},
			msg:          "Memory request already exists, do not set",
		},
		{
			requests: corev1.ResourceList{},
			mc: limitrange.Config{
				HasDefaultRequest: false,
			},
			wantRequests: corev1.ResourceList{},
			msg:          "Memory request does not exist but LimitRange does not have memory default, do not set",
		},
	}

	for _, test := range tests {
		container := &corev1.Container{
			Resources: corev1.ResourceRequirements{
				Requests: test.requests,
			},
		}

		wantContainer := &corev1.Container{
			Resources: corev1.ResourceRequirements{
				Requests: test.wantRequests,
			},
		}

		err := setMemoryRequest(container, test.mc)
		assert.NoError(t, err, test.msg)
		assert.Equal(t, wantContainer, container, test.msg)
	}
}

func TestSetMemoryLimit(t *testing.T) {
	t.Parallel()

	memoryConfigWithoutMaxRatio := limitrange.Config{
		HasDefaultLimit:         true,
		HasMaxLimitRequestRatio: false,
		DefaultLimit:            resource.MustParse("50Mi"),
	}

	memoryConfigWithMaxRatio := limitrange.Config{
		HasDefaultLimit:         true,
		HasMaxLimitRequestRatio: true,
		DefaultLimit:            resource.MustParse("50Mi"),
		MaxLimitRequestRatio:    resource.MustParse("1.05"),
	}

	tests := []struct {
		requests   corev1.ResourceList
		limits     corev1.ResourceList
		mc         limitrange.Config
		wantLimits corev1.ResourceList
		msg        string
	}{
		{
			requests:   corev1.ResourceList{},
			limits:     corev1.ResourceList{},
			mc:         memoryConfigWithMaxRatio,
			wantLimits: corev1.ResourceList{corev1.ResourceMemory: *quantity.Ptr(resource.MustParse("50Mi")).ToDec()},
			msg:        "Memory request and limit not set, set to default",
		},
		{
			requests:   corev1.ResourceList{},
			limits:     corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("10Mi")},
			mc:         memoryConfigWithMaxRatio,
			wantLimits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("10Mi")},
			msg:        "Memory limit already exists, do not set",
		},
		{
			requests:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("10Mi")},
			limits:     corev1.ResourceList{},
			mc:         memoryConfigWithoutMaxRatio,
			wantLimits: corev1.ResourceList{corev1.ResourceMemory: *quantity.Ptr(resource.MustParse("50Mi")).ToDec()},
			msg:        "No maxLimitRequestMemoryRatio set, use defaultMemoryLimit which is higher than defaultLimitRequestMemoryRatio",
		},
		{
			requests:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("49Mi")},
			limits:     corev1.ResourceList{},
			mc:         memoryConfigWithoutMaxRatio,
			wantLimits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("53.9Mi")},
			msg:        "No maxLimitRequestMemoryRatio set, use defaultLimitRequestMemoryRatio which is higher than defaultMemoryLimit",
		},
		{
			requests:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("45Mi")},
			limits:     corev1.ResourceList{},
			mc:         memoryConfigWithMaxRatio,
			wantLimits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("47.25Mi")},
			msg:        "MaxLimitRequestMemoryRatio set, defaultMemoryLimit and defaultLimitRequestMemoryRatio both larger, use MaxLimitRequestMemoryRatio",
		},
		{
			requests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("50Gi")},
			limits:   corev1.ResourceList{},
			mc: limitrange.Config{
				HasDefaultLimit:         true,
				HasMaxLimitRequestRatio: true,
				DefaultLimit:            resource.MustParse("50Mi"),
				MaxLimitRequestRatio:    resource.MustParse("1.3"),
			},
			wantLimits: corev1.ResourceList{corev1.ResourceMemory: *quantity.Ptr(resource.MustParse("65Gi")).ToDec()},
			msg:        "MaxLimitRequestMemoryRatio set, defaultMemoryLimit and defaultLimitRequestMemoryRatio both smaller, use MaxLimitRequestMemoryRatio",
		},
	}

	for _, test := range tests {
		container := &corev1.Container{
			Resources: corev1.ResourceRequirements{
				Requests: test.requests,
				Limits:   test.limits,
			},
		}

		wantContainer := &corev1.Container{
			Resources: corev1.ResourceRequirements{
				Requests: test.requests,
				Limits:   test.wantLimits,
			},
		}

		err := setMemoryLimit(container, test.mc)
		assert.NoError(t, err, test.msg)

		// round up to scale 0 to avoid Scale differences
		roundUpResourceQuantityScale(container, 0)
		roundUpResourceQuantityScale(wantContainer, 0)

		assert.Equal(t, wantContainer, container, test.msg)
	}
}

// convenience function for testing only
func roundUpResourceQuantityScale(container *corev1.Container, scale resource.Scale) {
	if request, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
		container.Resources.Requests[corev1.ResourceMemory] = quantity.RoundUp(request, 0)
	}
	if limit, ok := container.Resources.Limits[corev1.ResourceMemory]; ok {
		container.Resources.Limits[corev1.ResourceMemory] = quantity.RoundUp(limit, 0)
	}
}

// convenience function for testing only
func roundUpContainersResourceQuantityScale(containers []corev1.Container, scale resource.Scale) {
	for idx := range containers {
		roundUpResourceQuantityScale(&containers[idx], 0)
	}
}
