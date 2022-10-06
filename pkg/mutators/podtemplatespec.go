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

func (p *PodTemplateSpec) Mutate(inputPts corev1.PodTemplateSpec, limitRangeMemory *limitrange.Config) (corev1.PodTemplateSpec, bool, error) {

	var mutated bool
	pts := *inputPts.DeepCopy()
	if limitRangeMemory == nil {
		return pts, mutated, fmt.Errorf("invalid limit range config")
	}

	m, err := p.setAndValidateResourceRequirements(pts.Spec.InitContainers, limitRangeMemory)
	if err != nil {
		return pts, m, err
	}
	mutated = m

	m, err = p.setAndValidateResourceRequirements(pts.Spec.Containers, limitRangeMemory)
	if err != nil {
		return pts, m || mutated, err
	}

	mutated = m || mutated

	return pts, mutated, nil
}

func (p *PodTemplateSpec) setAndValidateResourceRequirements(containers []corev1.Container, limitRangeMemory *limitrange.Config) (bool, error) {
	var mutated bool
	for idx := range containers {
		container := &containers[idx]

		mutated = p.setMemoryRequest(container, limitRangeMemory)
		mutated = p.setMemoryLimit(container, limitRangeMemory) || mutated

		if err := p.validateMemoryRequirements(*container, limitRangeMemory); err != nil {
			return mutated, err
		}
	}

	return mutated, nil
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

func (p *PodTemplateSpec) setMemoryRequest(container *corev1.Container, limitRangeMemory *limitrange.Config) bool {
	memoryRequest := container.Resources.Requests.Memory()
	memoryLimit := container.Resources.Limits.Memory()

	if !memoryRequest.IsZero() {
		return false
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

	return true
}

func (p *PodTemplateSpec) setMemoryLimit(container *corev1.Container, limitRangeMemory *limitrange.Config) bool {
	memoryRequest := container.Resources.Requests.Memory()
	memoryLimit := container.Resources.Limits.Memory()

	if !memoryLimit.IsZero() {
		return false
	}

	if container.Resources.Limits == nil {
		container.Resources.Limits = corev1.ResourceList{}
	}

	var calculatedLimit resource.Quantity

	if limitRangeMemory.HasMaxLimitRequestRatio && !memoryRequest.IsZero() {
		calculatedLimit = quantity.Mul(*memoryRequest, limitRangeMemory.MaxLimitRequestRatio)
	} else {
		ratioMemoryLimit := quantity.Mul(*memoryRequest, p.defaultMemoryLimitRequestRatio)
		calculatedLimit = quantity.Max(limitRangeMemory.DefaultLimit, ratioMemoryLimit)
	}

	if !calculatedLimit.IsZero() {
		container.Resources.Limits[corev1.ResourceMemory] = calculatedLimit
	}
	return true
}
