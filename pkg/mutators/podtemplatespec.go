package mutators

import (
	"fmt"

	"github.com/kanopy-platform/hedgetrimmer/pkg/limitrange"
	"github.com/kanopy-platform/hedgetrimmer/pkg/quantity"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type PodTemplateSpec struct {
	defaultMemoryLimitRequestRatio resource.Quantity
}

func NewPodTemplateSpec(opts ...OptionsFunc) *PodTemplateSpec {
	pts := &PodTemplateSpec{
		defaultMemoryLimitRequestRatio: resource.MustParse("1.1"),
	}

	for _, opt := range opts {
		opt(pts)
	}

	return pts
}

func (p *PodTemplateSpec) Mutate(inputPts corev1.PodTemplateSpec, limitRangeMemory *limitrange.Config) (corev1.PodTemplateSpec, error) {

	pts := *inputPts.DeepCopy()
	if limitRangeMemory == nil {
		return pts, fmt.Errorf("invalid limit range config")
	}

	if err := p.setAndValidateResourceRequirements(pts.Spec.InitContainers, limitRangeMemory); err != nil {
		return pts, err
	}

	if err := p.setAndValidateResourceRequirements(pts.Spec.Containers, limitRangeMemory); err != nil {
		return pts, err
	}

	return pts, nil
}

func (p *PodTemplateSpec) setAndValidateResourceRequirements(containers []corev1.Container, limitRangeMemory *limitrange.Config) error {
	for idx := range containers {
		container := &containers[idx]

		p.setMemoryRequest(container, limitRangeMemory)
		p.setMemoryLimit(container, limitRangeMemory)

		if err := p.validateMemoryRequirements(*container, limitRangeMemory); err != nil {
			return err
		}
	}

	return nil
}

func (p *PodTemplateSpec) validateMemoryRequirements(container corev1.Container, limitRangeMemory *limitrange.Config) error {
	memoryRequest := container.Resources.Requests.Memory()
	memoryLimit := container.Resources.Limits.Memory()

	if memoryRequest.IsZero() || memoryLimit.IsZero() {
		return fmt.Errorf("memory request (%s) and limit (%s) must be set", memoryRequest.String(), memoryLimit.String())
	}

	if memoryLimit.Cmp(*memoryRequest) == -1 {
		return fmt.Errorf("memory limit (%s) must be greater than request (%s)", memoryLimit.String(), memoryRequest.String())
	}

	if limitRangeMemory.HasMaxLimitRequestRatio {
		ratio := quantity.Div(*memoryLimit, *memoryRequest, infScaleMicro)
		if ratio.Cmp(limitRangeMemory.MaxLimitRequestRatio) == 1 {
			return fmt.Errorf("memory limit (%s) to request (%s) ratio (%s) exceeds MaxLimitRequestRatio (%s)",
				memoryLimit.String(), memoryRequest.String(), ratio.String(), limitRangeMemory.MaxLimitRequestRatio.String())
		}
	}

	return nil
}

func (p *PodTemplateSpec) setMemoryRequest(container *corev1.Container, limitRangeMemory *limitrange.Config) {
	memoryRequest := container.Resources.Requests.Memory()
	memoryLimit := container.Resources.Limits.Memory()

	if !memoryRequest.IsZero() {
		return
	}

	if container.Resources.Requests == nil {
		container.Resources.Requests = corev1.ResourceList{}
	}

	if !memoryLimit.IsZero() {
		container.Resources.Requests[corev1.ResourceMemory] = *memoryLimit
	} else if limitRangeMemory.HasDefaultRequest {
		container.Resources.Requests[corev1.ResourceMemory] = limitRangeMemory.DefaultRequest
	} else if limitRangeMemory.HasDefaultLimit {
		container.Resources.Requests[corev1.ResourceMemory] = limitRangeMemory.DefaultLimit
	}
}

func (p *PodTemplateSpec) setMemoryLimit(container *corev1.Container, limitRangeMemory *limitrange.Config) {
	memoryRequest := container.Resources.Requests.Memory()
	memoryLimit := container.Resources.Limits.Memory()

	if !memoryLimit.IsZero() {
		return
	}

	if container.Resources.Limits == nil {
		container.Resources.Limits = corev1.ResourceList{}
	}

	var calculatedLimit resource.Quantity

	if limitRangeMemory.HasMaxLimitRequestRatio && !memoryRequest.IsZero() {
		calculatedLimit = quantity.RoundUpBinarySI(quantity.Mul(*memoryRequest, limitRangeMemory.MaxLimitRequestRatio))
	} else {
		ratioMemoryLimit := quantity.RoundUpBinarySI(quantity.Mul(*memoryRequest, p.defaultMemoryLimitRequestRatio))
		calculatedLimit = quantity.Max(limitRangeMemory.DefaultLimit, ratioMemoryLimit)
	}

	if !calculatedLimit.IsZero() {
		container.Resources.Limits[corev1.ResourceMemory] = calculatedLimit
	}
}
