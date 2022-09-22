package podtemplatespec

import (
	"testing"

	"github.com/kanopy-platform/hedgetrimmer/internal/quantity"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestValidateMemoryRatio(t *testing.T) {
	t.Parallel()

	testConfig := limitRangeConfig{
		hasMaxLimitRequestMemoryRatio: true,
		maxLimitRequestMemoryRatio:    resource.MustParse("1.25"),
	}

	tests := []struct {
		requests  corev1.ResourceList
		limits    corev1.ResourceList
		lrc       limitRangeConfig
		wantError bool
		msg       string
	}{
		{
			requests:  corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1Gi")},
			limits:    corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1.25Gi")},
			lrc:       testConfig,
			wantError: false,
			msg:       "Container memory limit/request ratio equals max ratio, allow",
		},
		{
			requests:  corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1Gi")},
			limits:    corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1.26Gi")},
			lrc:       testConfig,
			wantError: true,
			msg:       "Container memory limit/request ratio exceeds max ratio, error",
		},
		{
			requests:  corev1.ResourceList{},
			limits:    corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1.26Gi")},
			lrc:       testConfig,
			wantError: false,
			msg:       "Container memory request does not exist, no ratio check",
		},
		{
			requests:  corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1Gi")},
			limits:    corev1.ResourceList{},
			lrc:       testConfig,
			wantError: false,
			msg:       "Container memory limit does not exist, no ratio check",
		},
		{
			requests:  corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1Gi")},
			limits:    corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1.26Gi")},
			lrc:       limitRangeConfig{hasMaxLimitRequestMemoryRatio: false},
			wantError: false,
			msg:       "LimitRange does not specify max ratio, no ratio check",
		},
	}

	for _, test := range tests {
		container := corev1.Container{
			Resources: corev1.ResourceRequirements{
				Requests: test.requests,
				Limits:   test.limits,
			},
		}

		err := validateMemoryRatio(container, test.lrc)
		if test.wantError {
			assert.Error(t, err, test.msg)
		} else {
			assert.NoError(t, err, test.msg)
		}
	}
}

func TestSetMemoryRequest(t *testing.T) {
	t.Parallel()

	testConfig := limitRangeConfig{
		hasDefaultMemoryRequest: true,
		defaultMemoryRequest:    resource.MustParse("5Gi"),
	}

	tests := []struct {
		requests     corev1.ResourceList
		lrc          limitRangeConfig
		wantRequests corev1.ResourceList
		msg          string
	}{
		{
			requests:     corev1.ResourceList{},
			lrc:          testConfig,
			wantRequests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("5Gi")},
			msg:          "Container memory request does not exist, set to default",
		},
		{
			requests:     corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1Gi")},
			lrc:          testConfig,
			wantRequests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1Gi")},
			msg:          "Container memory request already exists, do not set",
		},
		{
			requests: corev1.ResourceList{},
			lrc: limitRangeConfig{
				hasDefaultMemoryRequest: false,
			},
			wantRequests: corev1.ResourceList{},
			msg:          "Container memory request does not exist but LimitRange does not have memory default, do not set",
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

		err := setMemoryRequest(container, test.lrc)
		assert.NoError(t, err, test.msg)
		assert.Equal(t, wantContainer, container, test.msg)
	}
}

func TestSetMemoryLimit(t *testing.T) {
	t.Parallel()

	testConfig := limitRangeConfig{
		hasDefaultMemoryLimit: true,
		defaultMemoryLimit:    resource.MustParse("50Mi"),
	}

	tests := []struct {
		requests   corev1.ResourceList
		limits     corev1.ResourceList
		lrc        limitRangeConfig
		wantLimits corev1.ResourceList
		msg        string
	}{
		{
			requests:   corev1.ResourceList{},
			limits:     corev1.ResourceList{},
			lrc:        testConfig,
			wantLimits: corev1.ResourceList{corev1.ResourceMemory: *quantity.Ptr(resource.MustParse("50Mi")).ToDec()},
			msg:        "Container memory request and limit not set, set to default",
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

		err := setMemoryLimit(container, test.lrc)
		assert.NoError(t, err, test.msg)
		assert.Equal(t, wantContainer, container, test.msg)
	}
}
