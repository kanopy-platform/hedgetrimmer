package mutators

import (
	"context"
	"testing"

	"github.com/kanopy-platform/hedgetrimmer/pkg/limitrange"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestMutate(t *testing.T) {
	t.Parallel()

	pts := NewPodTemplateSpec(WithDefaultMemoryLimitRequestRatio(1.1))

	limitRangeMemory := &limitrange.Config{
		HasDefaultRequest:       true,
		HasDefaultLimit:         true,
		HasMaxLimitRequestRatio: false,
		DefaultRequest:          resource.MustParse("50Mi"),
		DefaultLimit:            resource.MustParse("64Mi"),
	}

	tests := []struct {
		msg        string
		containers []corev1.Container
		config     *limitrange.Config
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
						Limits:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("64Mi")},
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
						Limits:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("5632Mi")},
					},
				},
			},
			wantError: false,
		},
	}

	for _, test := range tests {
		// test both InitContainers and Containers using the given input and want
		inputs := []corev1.PodTemplateSpec{
			{Spec: corev1.PodSpec{InitContainers: test.containers}},
			{Spec: corev1.PodSpec{Containers: test.containers}},
		}

		for idx := range inputs {
			input := inputs[idx]

			result, err := pts.Mutate(context.Background(), input, test.config)
			if test.wantError {
				assert.Error(t, err, test.msg)
			} else {
				assert.NoError(t, err, test.msg)

				for idx, container := range result.Spec.InitContainers {
					wantRequest := test.want[idx].Resources.Requests.Memory()
					wantLimit := test.want[idx].Resources.Limits.Memory()
					assert.True(t, container.Resources.Requests.Memory().Equal(*wantRequest), test.msg)
					assert.True(t, container.Resources.Limits.Memory().Equal(*wantLimit), test.msg)
				}

				for idx, container := range result.Spec.Containers {
					wantRequest := test.want[idx].Resources.Requests.Memory()
					wantLimit := test.want[idx].Resources.Limits.Memory()
					assert.True(t, container.Resources.Requests.Memory().Equal(*wantRequest), test.msg)
					assert.True(t, container.Resources.Limits.Memory().Equal(*wantLimit), test.msg)
				}
			}
		}
	}
}

func TestMutateDryRun(t *testing.T) {
	t.Parallel()

	pts := NewPodTemplateSpec(WithDryRun(true))

	limitRangeMemory := &limitrange.Config{
		HasDefaultRequest:       true,
		HasDefaultLimit:         true,
		HasMaxLimitRequestRatio: false,
		DefaultRequest:          resource.MustParse("50Mi"),
		DefaultLimit:            resource.MustParse("64Mi"),
	}

	tests := []struct {
		msg        string
		containers []corev1.Container
		config     *limitrange.Config
		want       []corev1.Container
		wantError  bool
	}{
		{
			msg: "No request or limit specified, normally would apply defaults, but dry-run mode does not modify",
			containers: []corev1.Container{
				{
					Resources: corev1.ResourceRequirements{},
				},
			},
			config: limitRangeMemory,
			want: []corev1.Container{
				{
					Resources: corev1.ResourceRequirements{},
				},
			},
			wantError: false,
		},
		{
			msg: "Config is nil, normally would be an error, but dry-run mode does not error",
			containers: []corev1.Container{
				{
					Resources: corev1.ResourceRequirements{},
				},
			},
			config: nil,
			want: []corev1.Container{
				{
					Resources: corev1.ResourceRequirements{},
				},
			},
			wantError: false,
		},
	}

	for _, test := range tests {
		// test both InitContainers and Containers using the given input and want
		inputs := []corev1.PodTemplateSpec{
			{Spec: corev1.PodSpec{InitContainers: test.containers}},
			{Spec: corev1.PodSpec{Containers: test.containers}},
		}

		for idx := range inputs {
			input := inputs[idx]

			result, err := pts.Mutate(context.Background(), input, test.config)
			if test.wantError {
				assert.Error(t, err, test.msg)
			} else {
				assert.NoError(t, err, test.msg)
				assert.Equal(t, input, result, test.msg)
			}
		}
	}
}

func TestValidateMemoryRequirements(t *testing.T) {
	t.Parallel()

	pts := NewPodTemplateSpec(WithDefaultMemoryLimitRequestRatio(1.1))

	memoryConfig := &limitrange.Config{
		HasMaxLimitRequestRatio: true,
		MaxLimitRequestRatio:    resource.MustParse("1.25"),
	}

	tests := []struct {
		requests  corev1.ResourceList
		limits    corev1.ResourceList
		mc        *limitrange.Config
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
			mc:        &limitrange.Config{HasMaxLimitRequestRatio: false},
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

		err := pts.validateMemoryRequirements(context.Background(), container, test.mc)
		assert.Equal(t, test.wantError, err != nil, test.msg)
	}
}

func TestValidateMemoryRequirementsDryRun(t *testing.T) {
	t.Parallel()

	pts := NewPodTemplateSpec(WithDryRun(true))

	limitrangeConfig := &limitrange.Config{
		HasMaxLimitRequestRatio: true,
		MaxLimitRequestRatio:    resource.MustParse("1.25"),
	}

	tests := []struct {
		requests corev1.ResourceList
		limits   corev1.ResourceList
		mc       *limitrange.Config
		msg      string
	}{

		{
			msg:      "Memory request does not exist, error, no error on dry-run",
			requests: corev1.ResourceList{},
			limits:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1.26Gi")},
			mc:       limitrangeConfig,
		},
		{
			msg:      "Memory limit does not exist, error, no error on dry-run",
			requests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1Gi")},
			limits:   corev1.ResourceList{},
			mc:       limitrangeConfig,
		},
		{
			msg:      "Memory limit is smaller than request, no error on dry-run",
			requests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("50Mi")},
			limits:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("40Mi")},
			mc:       limitrangeConfig,
		},
		{
			msg:      "Memory limit/request ratio exceeds max ratio, no error on dry-run",
			requests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1Gi")},
			limits:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("2.5Gi")},
			mc:       limitrangeConfig,
		},
	}

	for _, test := range tests {
		container := corev1.Container{
			Resources: corev1.ResourceRequirements{
				Requests: test.requests,
				Limits:   test.limits,
			},
		}

		err := pts.validateMemoryRequirements(context.Background(), container, test.mc)
		assert.NoError(t, err, test.msg)
	}
}

func TestSetMemoryRequest(t *testing.T) {
	t.Parallel()

	pts := NewPodTemplateSpec(WithDefaultMemoryLimitRequestRatio(1.1))

	memoryConfig := &limitrange.Config{
		HasDefaultRequest: true,
		HasDefaultLimit:   true,
		DefaultRequest:    resource.MustParse("5Gi"),
		DefaultLimit:      resource.MustParse("6Gi"),
	}

	tests := []struct {
		requests     corev1.ResourceList
		limits       corev1.ResourceList
		mc           *limitrange.Config
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
			msg:          "Memory request does not exist, but limit is set, set request = limit",
			requests:     nil,
			limits:       corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("100Mi")},
			mc:           memoryConfig,
			wantRequests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("100Mi")},
		},
		{
			msg:      "Memory request and limit does not exist, but default limit is set, set request = defaultLimit",
			requests: nil,
			limits:   nil,
			mc: &limitrange.Config{
				HasDefaultLimit: true,
				DefaultLimit:    resource.MustParse("100Mi"),
			},
			wantRequests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("100Mi")},
		},
		{
			msg:      "Memory request and limit does not exist and LimitRange does not have default request or limit, do not set",
			requests: corev1.ResourceList{},
			mc: &limitrange.Config{
				HasDefaultRequest: false,
				HasDefaultLimit:   false,
			},
			wantRequests: corev1.ResourceList{},
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
				Requests: test.wantRequests,
				Limits:   test.limits,
			},
		}

		pts.setMemoryRequest(context.Background(), container, test.mc)
		assert.Equal(t, wantContainer, container, test.msg)
	}
}

func TestSetMemoryLimit(t *testing.T) {
	t.Parallel()

	pts := NewPodTemplateSpec(WithDefaultMemoryLimitRequestRatio(1.1))

	memoryConfigWithoutMaxRatio := &limitrange.Config{
		HasDefaultLimit:         true,
		HasMaxLimitRequestRatio: false,
		DefaultLimit:            resource.MustParse("50Mi"),
	}

	memoryConfigWithMaxRatio := &limitrange.Config{
		HasDefaultLimit:         true,
		HasMaxLimitRequestRatio: true,
		DefaultLimit:            resource.MustParse("50Mi"),
		MaxLimitRequestRatio:    resource.MustParse("1.05"),
	}

	tests := []struct {
		requests   corev1.ResourceList
		limits     corev1.ResourceList
		mc         *limitrange.Config
		wantLimits corev1.ResourceList
		msg        string
	}{
		{
			msg:        "Memory request and limit not set, set to default",
			requests:   corev1.ResourceList{},
			limits:     corev1.ResourceList{},
			mc:         memoryConfigWithMaxRatio,
			wantLimits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("50Mi")},
		},
		{
			msg:        "Memory request and limit not set, config does not have defaults, do not set",
			requests:   corev1.ResourceList{},
			limits:     corev1.ResourceList{},
			mc:         &limitrange.Config{},
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
			wantLimits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("50Mi")},
		},
		{
			msg:        "No maxLimitRequestMemoryRatio set, use defaultLimitRequestMemoryRatio which is higher than defaultMemoryLimit",
			requests:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("49Mi")},
			limits:     corev1.ResourceList{},
			mc:         memoryConfigWithoutMaxRatio,
			wantLimits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("54Mi")},
		},
		{
			msg:        "MaxLimitRequestMemoryRatio set, defaultMemoryLimit and defaultLimitRequestMemoryRatio both larger, use MaxLimitRequestMemoryRatio",
			requests:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("45Mi")},
			limits:     corev1.ResourceList{},
			mc:         memoryConfigWithMaxRatio,
			wantLimits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("48Mi")},
		},
		{
			msg:      "MaxLimitRequestMemoryRatio set, defaultMemoryLimit and defaultLimitRequestMemoryRatio both smaller, use MaxLimitRequestMemoryRatio",
			requests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("50Gi")},
			limits:   corev1.ResourceList{},
			mc: &limitrange.Config{
				HasDefaultLimit:         true,
				HasMaxLimitRequestRatio: true,
				DefaultLimit:            resource.MustParse("50Mi"),
				MaxLimitRequestRatio:    resource.MustParse("1.3"),
			},
			wantLimits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("65Gi")},
		},
	}

	for _, test := range tests {
		container := corev1.Container{
			Resources: corev1.ResourceRequirements{
				Requests: test.requests,
				Limits:   test.limits,
			},
		}

		pts.setMemoryLimit(context.Background(), &container, test.mc)

		assert.True(t, test.requests.Memory().Equal(*container.Resources.Requests.Memory()), test.msg)
		assert.True(t, test.wantLimits.Memory().Equal(*container.Resources.Limits.Memory()), test.msg)
	}
}
