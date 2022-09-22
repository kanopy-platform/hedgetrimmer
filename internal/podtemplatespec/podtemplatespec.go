package podtemplatespec

import (
	"fmt"

	"github.com/kanopy-platform/hedgetrimmer/internal/quantity"
	corev1 "k8s.io/api/core/v1"
)

const (
	defaultLimitRequestMemoryRatio float64 = 1.1
)

func ApplyLimits(inputPts corev1.PodTemplateSpec, lri corev1.LimitRangeItem) (corev1.PodTemplateSpec, error) {
	pts := *inputPts.DeepCopy()

	if !isLimitRangeTypeContainer(lri) {
		return pts, fmt.Errorf("expected LimitRangeItem type %q, got %q instead", corev1.LimitTypeContainer, lri.Type)
	}

	limitRangeConfig := getLimitRangeConfig(lri)

	if err := validateAndSetResourceRequirements(pts.Spec.InitContainers, limitRangeConfig); err != nil {
		return pts, err
	}

	if err := validateAndSetResourceRequirements(pts.Spec.Containers, limitRangeConfig); err != nil {
		return pts, err
	}

	return pts, nil
}

func validateAndSetResourceRequirements(containers []corev1.Container, lrc limitRangeConfig) error {
	for _, container := range containers {
		if err := validateMemoryRatio(container, lrc); err != nil {
			return err
		}

		if err := setMemoryRequest(&container, lrc); err != nil {
			return err
		}

		if err := setMemoryLimit(&container, lrc); err != nil {
			return err
		}
	}

	return nil
}

func validateMemoryRatio(container corev1.Container, lrc limitRangeConfig) error {
	memoryRequest := container.Resources.Requests.Memory()
	memoryLimit := container.Resources.Limits.Memory()

	if memoryRequest.IsZero() || memoryLimit.IsZero() || !lrc.hasMaxLimitRequestMemoryRatio {
		return nil
	}

	ratio := quantity.Div(*memoryLimit, *memoryRequest)
	if ratio.Cmp(lrc.maxLimitRequestMemoryRatio) == 1 {
		return fmt.Errorf("memory limit (%s) to request (%s) ratio (%s) exceeds MaxLimitRequestRatio (%s)",
			memoryLimit.String(), memoryRequest.String(), ratio.String(), lrc.maxLimitRequestMemoryRatio.String())
	}

	return nil
}

func setMemoryRequest(container *corev1.Container, lrc limitRangeConfig) error {
	memoryRequest := container.Resources.Requests.Memory()
	if memoryRequest.IsZero() {
		if !lrc.hasDefaultMemoryRequest {
			return nil
		}

		if container.Resources.Requests == nil {
			container.Resources.Requests = corev1.ResourceList{}
		}
		container.Resources.Requests[corev1.ResourceMemory] = lrc.defaultMemoryRequest
	}

	return nil
}

func setMemoryLimit(container *corev1.Container, lrc limitRangeConfig) error {
	memoryRequest := container.Resources.Requests.Memory()
	memoryLimit := container.Resources.Limits.Memory()

	if memoryLimit.IsZero() {
		if container.Resources.Limits == nil {
			container.Resources.Limits = corev1.ResourceList{}
		}

		if lrc.hasMaxLimitRequestMemoryRatio {
			container.Resources.Limits[corev1.ResourceMemory] = quantity.Min(
				lrc.defaultMemoryLimit, quantity.Mul(*memoryRequest, lrc.maxLimitRequestMemoryRatio))
		} else {
			ratioMemoryLimit, err := quantity.MulFloat64(*memoryRequest, defaultLimitRequestMemoryRatio)
			if err != nil {
				return err
			}

			container.Resources.Limits[corev1.ResourceMemory] = quantity.Max(
				lrc.defaultMemoryLimit, ratioMemoryLimit)
		}
	}

	return nil
}
