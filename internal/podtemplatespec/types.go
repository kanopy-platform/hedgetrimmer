package podtemplatespec

import (
	corev1 "k8s.io/api/core/v1"
)

type PodTemplateSpec struct {
	pts corev1.PodTemplateSpec
}
