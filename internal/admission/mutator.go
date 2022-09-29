package admission

import (
	"github.com/kanopy-platform/hedgetrimmer/internal/limitrange"
	corev1 "k8s.io/api/core/v1"
)

type PodTemplateSpecMutator interface {
	Mutate(inputPts corev1.PodTemplateSpec, limitRangeMemory *limitrange.Config) (corev1.PodTemplateSpec, error)
}
