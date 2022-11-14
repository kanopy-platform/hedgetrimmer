package mutators

import (
	"context"
	"errors"
	"fmt"

	"github.com/kanopy-platform/hedgetrimmer/pkg/limitrange"
	"github.com/kanopy-platform/hedgetrimmer/pkg/quantity"
	"gopkg.in/inf.v0"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type PodTemplateSpec struct {
	dryRun                         bool
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

func (p *PodTemplateSpec) Mutate(ctx context.Context, inputPts corev1.PodTemplateSpec, limitRangeMemory *limitrange.Config) (corev1.PodTemplateSpec, error) {

	pts := *inputPts.DeepCopy()
	if limitRangeMemory == nil {
		return pts, p.errorIfNotDryRun(ctx, "invalid limit range config")
	}

	if err := p.setAndValidateResourceRequirements(ctx, pts.Spec.InitContainers, limitRangeMemory); err != nil {
		return pts, err
	}

	if err := p.setAndValidateResourceRequirements(ctx, pts.Spec.Containers, limitRangeMemory); err != nil {
		return pts, err
	}

	return pts, nil
}

func (p *PodTemplateSpec) setAndValidateResourceRequirements(ctx context.Context, containers []corev1.Container, limitRangeMemory *limitrange.Config) error {
	for idx := range containers {
		container := &containers[idx]
		if p.dryRun {
			// On dry-run use a copy to go through the motions, do not modify original
			container = container.DeepCopy()
		}

		p.setMemoryRequest(ctx, container, limitRangeMemory)
		p.setMemoryLimit(ctx, container, limitRangeMemory)

		if err := p.validateMemoryRequirements(ctx, *container, limitRangeMemory); err != nil {
			return p.errorIfNotDryRun(ctx, err.Error())
		}
	}

	return nil
}

func (p *PodTemplateSpec) errorIfNotDryRun(ctx context.Context, err string) error {
	log := log.FromContext(ctx)

	if p.dryRun {
		log.Info(fmt.Sprintf("[dry-run] %s", err))
		return nil
	}

	return errors.New(err)
}

func (p *PodTemplateSpec) validateMemoryRequirements(ctx context.Context, container corev1.Container, limitRangeMemory *limitrange.Config) error {
	memoryRequest := container.Resources.Requests.Memory()
	memoryLimit := container.Resources.Limits.Memory()

	if memoryRequest.IsZero() || memoryLimit.IsZero() {
		return fmt.Errorf("memory request (%s) and limit (%s) must be set", memoryRequest.String(), memoryLimit.String())
	}

	if memoryLimit.Cmp(*memoryRequest) == -1 {
		return fmt.Errorf("memory limit (%s) must be greater than request (%s)", memoryLimit.String(), memoryRequest.String())
	}

	if limitRangeMemory.HasMaxLimitRequestRatio {
		ratio := quantity.Div(*memoryLimit, *memoryRequest, infScaleMicro, inf.RoundUp)
		if ratio.Cmp(limitRangeMemory.MaxLimitRequestRatio) == 1 {
			return fmt.Errorf("memory limit (%s) to request (%s) ratio (%s) exceeds MaxLimitRequestRatio (%s)",
				memoryLimit.String(), memoryRequest.String(), ratio.String(), limitRangeMemory.MaxLimitRequestRatio.String())
		}
	}

	return nil
}

func (p *PodTemplateSpec) setMemoryRequest(ctx context.Context, container *corev1.Container, limitRangeMemory *limitrange.Config) {
	log := log.FromContext(ctx)
	memoryRequest := container.Resources.Requests.Memory()
	memoryLimit := container.Resources.Limits.Memory()

	if !memoryRequest.IsZero() {
		return
	}

	if container.Resources.Requests == nil {
		container.Resources.Requests = corev1.ResourceList{}
	}

	var calculatedRequest resource.Quantity

	if !memoryLimit.IsZero() {
		calculatedRequest = *memoryLimit
	} else if limitRangeMemory.HasDefaultRequest {
		calculatedRequest = limitRangeMemory.DefaultRequest
	} else if limitRangeMemory.HasDefaultLimit {
		calculatedRequest = limitRangeMemory.DefaultLimit
	}

	if !calculatedRequest.IsZero() {
		if p.dryRun {
			log.Info(fmt.Sprintf("[dry-run] setting memory request to %s", calculatedRequest.String()))
		} else {
			log.Info(fmt.Sprintf("setting memory request to %s", calculatedRequest.String()))
		}
		container.Resources.Requests[corev1.ResourceMemory] = calculatedRequest
	}
}

func (p *PodTemplateSpec) setMemoryLimit(ctx context.Context, container *corev1.Container, limitRangeMemory *limitrange.Config) {
	log := log.FromContext(ctx)
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
		calculatedLimit = quantity.RoundBinarySI(quantity.Mul(*memoryRequest, limitRangeMemory.MaxLimitRequestRatio), inf.RoundDown)
	} else {
		ratioMemoryLimit := quantity.RoundBinarySI(quantity.Mul(*memoryRequest, p.defaultMemoryLimitRequestRatio), inf.RoundDown)
		calculatedLimit = quantity.Max(limitRangeMemory.DefaultLimit, ratioMemoryLimit)
	}

	if !calculatedLimit.IsZero() {
		if p.dryRun {
			log.Info(fmt.Sprintf("[dry-run] setting memory limit to %s", calculatedLimit.String()))
		} else {
			log.Info(fmt.Sprintf("setting memory limit to %s", calculatedLimit.String()))
		}
		container.Resources.Limits[corev1.ResourceMemory] = calculatedLimit
	}
}
