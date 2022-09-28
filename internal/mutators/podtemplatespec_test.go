package mutators

import (
	"os"
	"testing"

	"github.com/kanopy-platform/hedgetrimmer/internal/limitrange"
	"github.com/kanopy-platform/hedgetrimmer/internal/quantity"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var testPts PodTemplateSpec

const testDefaultMaxLimitRequestRatio float64 = 1.1

func TestMain(m *testing.M) {
	testPts = NewPodTemplateSpec(testDefaultMaxLimitRequestRatio)

	os.Exit(m.Run())
}

func TestMutate(t *testing.T) {
	t.Parallel()

	limitRangeMemory := limitrange.Config{
		HasDefaultRequest:       true,
		HasDefaultLimit:         true,
		HasMaxLimitRequestRatio: false,
		DefaultRequest:          resource.MustParse("50Mi"),
		DefaultLimit:            resource.MustParse("64Mi"),
	}

	tests := []struct {
		msg        string
		containers []corev1.Container
		config     limitrange.Config
		want       []corev1.Container
		wantError  bool
	}{
		{
			msg: "No request or limit specified, apply default",
			containers: []corev1.Container{
				{
					Resources: corev1.ResourceRequirements{},
				},
			},
			config: limitRangeMemory,
			want: []corev1.Container{
				{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("50Mi")},
						Limits:   corev1.ResourceList{corev1.ResourceMemory: *quantity.Ptr(resource.MustParse("64Mi")).ToDec()},
					},
				},
			},
			wantError: false,
		},
		{
			msg: "No request but has limit, apply DefaultRequest",
			containers: []corev1.Container{
				{
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("100Mi")},
					},
				},
			},
			config: limitRangeMemory,
			want: []corev1.Container{
				{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("50Mi")},
						Limits:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("100Mi")},
					},
				},
			},
			wantError: false,
		},
		{
			msg: "Has request but no limit specified, apply defaultMaxLimitRequestRatio which exceeds DefaultLimit",
			containers: []corev1.Container{
				{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("5Gi")},
					},
				},
			},
			config: limitRangeMemory,
			want: []corev1.Container{
				{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("5Gi")},
						Limits:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("5.5Gi")},
					},
				},
			},
			wantError: false,
		},
		{
			msg: "No request but has limit set below DefaultRequest, error",
			containers: []corev1.Container{
				{
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("10Mi")},
					},
				},
			},
			config:    limitRangeMemory,
			wantError: true,
		},
	}

	for _, test := range tests {
		// test both InitContainers and Containers using the given input and want
		inputs := []corev1.PodTemplateSpec{
			{Spec: corev1.PodSpec{InitContainers: test.containers}},
			{Spec: corev1.PodSpec{Containers: test.containers}},
		}

		wants := []corev1.PodTemplateSpec{
			{Spec: corev1.PodSpec{InitContainers: test.want}},
			{Spec: corev1.PodSpec{Containers: test.want}},
		}

		for idx := range inputs {
			input := inputs[idx]
			want := wants[idx]

			result, err := testPts.Mutate(input, test.config)
			if test.wantError {
				assert.Error(t, err, test.msg)
			} else {
				assert.NoError(t, err, test.msg)

				// round up to scale 0 to avoid Scale differences
				if len(input.Spec.InitContainers) > 0 {
					roundUpContainersQuantityScale(want.Spec.InitContainers, corev1.ResourceMemory, 0)
					roundUpContainersQuantityScale(result.Spec.InitContainers, corev1.ResourceMemory, 0)
				}
				if len(input.Spec.Containers) > 0 {
					roundUpContainersQuantityScale(want.Spec.Containers, corev1.ResourceMemory, 0)
					roundUpContainersQuantityScale(result.Spec.Containers, corev1.ResourceMemory, 0)
				}

				assert.Equal(t, want, result, test.msg)
			}
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
			msg:       "Memory limit/request ratio equals max ratio, allow",
			requests:  corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1Gi")},
			limits:    corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1.25Gi")},
			mc:        memoryConfig,
			wantError: false,
		},
		{
			msg:       "Memory limit/request ratio exceeds max ratio, error",
			requests:  corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1Gi")},
			limits:    corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1.25000001Gi")},
			mc:        memoryConfig,
			wantError: true,
		},
		{
			msg:       "LimitRange does not specify max ratio, no ratio check",
			requests:  corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1Gi")},
			limits:    corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1.26Gi")},
			mc:        limitrange.Config{HasMaxLimitRequestRatio: false},
			wantError: false,
		},
		{
			msg:       "Memory request does not exist, error",
			requests:  corev1.ResourceList{},
			limits:    corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1.26Gi")},
			mc:        memoryConfig,
			wantError: true,
		},
		{
			msg:       "Memory limit does not exist, error",
			requests:  corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1Gi")},
			limits:    corev1.ResourceList{},
			mc:        memoryConfig,
			wantError: true,
		},
		{
			msg:       "Memory limit is smaller than request, error",
			requests:  corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("50Mi")},
			limits:    corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("49.99Mi")},
			mc:        memoryConfig,
			wantError: true,
		},
	}

	for _, test := range tests {
		container := corev1.Container{
			Resources: corev1.ResourceRequirements{
				Requests: test.requests,
				Limits:   test.limits,
			},
		}

		err := testPts.validateMemoryRequirements(container, test.mc)
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
			msg:          "Memory request does not exist, set to default",
			requests:     corev1.ResourceList{},
			mc:           memoryConfig,
			wantRequests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("5Gi")},
		},
		{
			msg:          "Memory request already exists, do not set",
			requests:     corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1Gi")},
			mc:           memoryConfig,
			wantRequests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1Gi")},
		},
		{
			msg:      "Memory request does not exist but LimitRange does not have memory default, do not set",
			requests: corev1.ResourceList{},
			mc: limitrange.Config{
				HasDefaultRequest: false,
			},
			wantRequests: corev1.ResourceList{},
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

		testPts.setMemoryRequest(container, test.mc)
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
			msg:        "Memory request and limit not set, set to default",
			requests:   corev1.ResourceList{},
			limits:     corev1.ResourceList{},
			mc:         memoryConfigWithMaxRatio,
			wantLimits: corev1.ResourceList{corev1.ResourceMemory: *quantity.Ptr(resource.MustParse("50Mi")).ToDec()},
		},
		{
			msg:        "Memory request and limit not set, config does not have defaults, do not set",
			requests:   corev1.ResourceList{},
			limits:     corev1.ResourceList{},
			mc:         limitrange.Config{},
			wantLimits: corev1.ResourceList{},
		},
		{
			msg:        "Memory limit already exists, do not set",
			requests:   corev1.ResourceList{},
			limits:     corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("10Mi")},
			mc:         memoryConfigWithMaxRatio,
			wantLimits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("10Mi")},
		},
		{
			msg:        "No maxLimitRequestMemoryRatio set, use defaultMemoryLimit which is higher than defaultLimitRequestMemoryRatio",
			requests:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("10Mi")},
			limits:     corev1.ResourceList{},
			mc:         memoryConfigWithoutMaxRatio,
			wantLimits: corev1.ResourceList{corev1.ResourceMemory: *quantity.Ptr(resource.MustParse("50Mi")).ToDec()},
		},
		{
			msg:        "No maxLimitRequestMemoryRatio set, use defaultLimitRequestMemoryRatio which is higher than defaultMemoryLimit",
			requests:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("49Mi")},
			limits:     corev1.ResourceList{},
			mc:         memoryConfigWithoutMaxRatio,
			wantLimits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("53.9Mi")},
		},
		{
			msg:        "MaxLimitRequestMemoryRatio set, defaultMemoryLimit and defaultLimitRequestMemoryRatio both larger, use MaxLimitRequestMemoryRatio",
			requests:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("45Mi")},
			limits:     corev1.ResourceList{},
			mc:         memoryConfigWithMaxRatio,
			wantLimits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("47.25Mi")},
		},
		{
			msg:      "MaxLimitRequestMemoryRatio set, defaultMemoryLimit and defaultLimitRequestMemoryRatio both smaller, use MaxLimitRequestMemoryRatio",
			requests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("50Gi")},
			limits:   corev1.ResourceList{},
			mc: limitrange.Config{
				HasDefaultLimit:         true,
				HasMaxLimitRequestRatio: true,
				DefaultLimit:            resource.MustParse("50Mi"),
				MaxLimitRequestRatio:    resource.MustParse("1.3"),
			},
			wantLimits: corev1.ResourceList{corev1.ResourceMemory: *quantity.Ptr(resource.MustParse("65Gi")).ToDec()},
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

		err := testPts.setMemoryLimit(container, test.mc)
		assert.NoError(t, err, test.msg)

		// round up to scale 0 to avoid Scale differences
		roundUpContainerQuantityToScale(container, corev1.ResourceMemory, 0)
		roundUpContainerQuantityToScale(wantContainer, corev1.ResourceMemory, 0)

		assert.Equal(t, wantContainer, container, test.msg)
	}
}

func roundUpQuantityToScale(list corev1.ResourceList, name corev1.ResourceName, scale resource.Scale) {
	if request, ok := list[name]; ok {
		request.RoundUp(scale)
		list[name] = request
	}
}

func roundUpContainerQuantityToScale(container *corev1.Container, name corev1.ResourceName, scale resource.Scale) {
	roundUpQuantityToScale(container.Resources.Requests, name, scale)
	roundUpQuantityToScale(container.Resources.Limits, name, scale)
}

func roundUpContainersQuantityScale(containers []corev1.Container, name corev1.ResourceName, scale resource.Scale) {
	for idx := range containers {
		roundUpContainerQuantityToScale(&containers[idx], name, scale)
	}
}
