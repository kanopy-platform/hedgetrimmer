package podtemplatespec

import (
	"fmt"

	"github.com/kanopy-platform/hedgetrimmer/internal/limitrange"
	"github.com/kanopy-platform/hedgetrimmer/internal/quantity"
	corev1 "k8s.io/api/core/v1"
)

const (
	defaultLimitRequestMemoryRatio float64 = 1.1
)

func New(p corev1.PodTemplateSpec) *PodTemplateSpec {
	pts := &PodTemplateSpec{
		pts: *p.DeepCopy(),
	}
	return pts
}

func (p *PodTemplateSpec) ApplyResourceRequirements(lri corev1.LimitRangeItem) (corev1.PodTemplateSpec, error) {
	pts := *p.pts.DeepCopy()

	if !limitrange.IsLimitRangeTypeContainer(lri) {
		return p.pts, fmt.Errorf("expected LimitRangeItem type Container, got %q instead", lri.Type)
	}

	limitRangeConfig := limitrange.GetConfig(lri, corev1.ResourceMemory)

	if err := setAndValidateResourceRequirements(pts.Spec.InitContainers, limitRangeConfig); err != nil {
		return p.pts, err
	}

	if err := setAndValidateResourceRequirements(pts.Spec.Containers, limitRangeConfig); err != nil {
		return p.pts, err
	}

	p.pts = pts
	return pts, nil
}

func setAndValidateResourceRequirements(containers []corev1.Container, mc limitrange.Config) error {
	for idx := range containers {
		container := &containers[idx]
		if err := setMemoryRequest(container, mc); err != nil {
			return err
		}

		if err := setMemoryLimit(container, mc); err != nil {
			return err
		}

		if err := validateMemoryRequirements(*container, mc); err != nil {
			return err
		}
	}

	return nil
}

func validateMemoryRequirements(container corev1.Container, mc limitrange.Config) error {
	memoryRequest := container.Resources.Requests.Memory()
	memoryLimit := container.Resources.Limits.Memory()

	if memoryRequest.IsZero() || memoryLimit.IsZero() {
		return fmt.Errorf("memory request (%s) and limit (%s) must be set", memoryRequest.String(), memoryLimit.String())
	}

	if mc.HasMaxLimitRequestRatio {
		ratio := quantity.Div(*memoryLimit, *memoryRequest)
		if ratio.Cmp(mc.MaxLimitRequestRatio) == 1 {
			return fmt.Errorf("memory limit (%s) to request (%s) ratio (%s) exceeds MaxLimitRequestRatio (%s)",
				memoryLimit.String(), memoryRequest.String(), ratio.String(), mc.MaxLimitRequestRatio.String())
		}
	}

	return nil
}

func setMemoryRequest(container *corev1.Container, mc limitrange.Config) error {
	memoryRequest := container.Resources.Requests.Memory()
	if memoryRequest.IsZero() {
		if !mc.HasDefaultRequest {
			return nil
		}

		if container.Resources.Requests == nil {
			container.Resources.Requests = corev1.ResourceList{}
		}
		container.Resources.Requests[corev1.ResourceMemory] = mc.DefaultRequest
	}
	return nil
}

func setMemoryLimit(container *corev1.Container, mc limitrange.Config) error {
	memoryRequest := container.Resources.Requests.Memory()
	memoryLimit := container.Resources.Limits.Memory()

	if memoryLimit.IsZero() {
		if container.Resources.Limits == nil {
			container.Resources.Limits = corev1.ResourceList{}
		}

		if mc.HasMaxLimitRequestRatio && !memoryRequest.IsZero() {
			container.Resources.Limits[corev1.ResourceMemory] = quantity.Mul(*memoryRequest, mc.MaxLimitRequestRatio)
		} else {
			ratioMemoryLimit, err := quantity.MulFloat64(*memoryRequest, defaultLimitRequestMemoryRatio)
			if err != nil {
				return err
			}

			container.Resources.Limits[corev1.ResourceMemory] = quantity.Max(
				mc.DefaultLimit, ratioMemoryLimit)
		}
	}

	return nil
}
