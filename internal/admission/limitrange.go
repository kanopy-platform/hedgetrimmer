package admission

import (
	"github.com/kanopy-platform/hedgetrimmer/internal/limitrange"
	corev1Listers "k8s.io/client-go/listers/core/v1"
)

type LimitRanger interface {
	corev1Listers.LimitRangeLister
	NewConfig(namespace string) limitrange.Config
}
