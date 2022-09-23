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

	limitRangeConfig := limitrange.GetMemoryConfig(lri)

	if err := validateAndSetResourceRequirements(pts.Spec.InitContainers, limitRangeConfig); err != nil {
		return p.pts, err
	}

	if err := validateAndSetResourceRequirements(pts.Spec.Containers, limitRangeConfig); err != nil {
		return p.pts, err
	}

	p.pts = pts
	return pts, nil
}

func validateAndSetResourceRequirements(containers []corev1.Container, mc limitrange.MemoryConfig) error {
	for idx := range containers {
		container := &containers[idx]
		if err := setMemoryRequest(container, mc); err != nil {
			return err
		}

		if err := setMemoryLimit(container, mc); err != nil {
			return err
		}

		if err := validateMemoryRatio(*container, mc); err != nil {
			return err
		}
	}

	return nil
}

func validateMemoryRatio(container corev1.Container, mc limitrange.MemoryConfig) error {
	memoryRequest := container.Resources.Requests.Memory()
	memoryLimit := container.Resources.Limits.Memory()

	if memoryRequest.IsZero() || memoryLimit.IsZero() || !mc.HasMaxLimitRequestMemoryRatio {
		return nil
	}

	ratio := quantity.Div(*memoryLimit, *memoryRequest)
	if ratio.Cmp(mc.MaxLimitRequestMemoryRatio) == 1 {
		return fmt.Errorf("memory limit (%s) to request (%s) ratio (%s) exceeds MaxLimitRequestRatio (%s)",
			memoryLimit.String(), memoryRequest.String(), ratio.String(), mc.MaxLimitRequestMemoryRatio.String())
	}

	return nil
}

func setMemoryRequest(container *corev1.Container, mc limitrange.MemoryConfig) error {
	memoryRequest := container.Resources.Requests.Memory()
	if memoryRequest.IsZero() {
		if !mc.HasDefaultMemoryRequest {
			return nil
		}

		if container.Resources.Requests == nil {
			container.Resources.Requests = corev1.ResourceList{}
		}
		container.Resources.Requests[corev1.ResourceMemory] = mc.DefaultMemoryRequest
	}
	return nil
}

func setMemoryLimit(container *corev1.Container, mc limitrange.MemoryConfig) error {
	memoryRequest := container.Resources.Requests.Memory()
	memoryLimit := container.Resources.Limits.Memory()

	if memoryLimit.IsZero() {
		if container.Resources.Limits == nil {
			container.Resources.Limits = corev1.ResourceList{}
		}

		if mc.HasMaxLimitRequestMemoryRatio && !memoryRequest.IsZero() {
			container.Resources.Limits[corev1.ResourceMemory] = quantity.Mul(*memoryRequest, mc.MaxLimitRequestMemoryRatio)
		} else {
			ratioMemoryLimit, err := quantity.MulFloat64(*memoryRequest, defaultLimitRequestMemoryRatio)
			if err != nil {
				return err
			}

			container.Resources.Limits[corev1.ResourceMemory] = quantity.Max(
				mc.DefaultMemoryLimit, ratioMemoryLimit)
		}
	}

	return nil
}
