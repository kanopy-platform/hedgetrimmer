package podtemplatespec

import corev1 "k8s.io/api/core/v1"

type Stub struct {
	pts corev1.PodTemplateSpec
}

func NewStub(p corev1.PodTemplateSpec) *Stub {
	stub := &Stub{
		pts: *p.DeepCopy(),
	}
	return stub
}

func (s *Stub) ApplyResourceRequirements(lri corev1.LimitRangeItem) (corev1.PodTemplateSpec, error) {
	return s.pts, nil
}
