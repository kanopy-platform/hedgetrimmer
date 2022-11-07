package admission

import (
	"context"

	"github.com/kanopy-platform/hedgetrimmer/pkg/limitrange"
	corev1 "k8s.io/api/core/v1"
)

type PodTemplateSpecMutator interface {
	Mutate(ctx context.Context, inputPts corev1.PodTemplateSpec, limitRangeMemory *limitrange.Config) (corev1.PodTemplateSpec, error)
}
