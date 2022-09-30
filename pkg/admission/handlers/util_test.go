package handlers

import (
	"github.com/kanopy-platform/hedgetrimmer/pkg/limitrange"
	corev1 "k8s.io/api/core/v1"
)

type MockMutator struct {
	spec corev1.PodTemplateSpec
	err  error
}

func (mm *MockMutator) SetSpec(spec corev1.PodTemplateSpec) {
	mm.spec = spec
}

func (mm *MockMutator) SetErr(err error) {
	mm.err = err
}

func (mm *MockMutator) Mutate(inputs corev1.PodTemplateSpec, config *limitrange.Config) (corev1.PodTemplateSpec, error) {
	return mm.spec, mm.err
}
