package podtemplatespec

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type limitRangeConfig struct {
	hasDefaultMemoryRequest       bool
	hasDefaultMemoryLimit         bool
	hasMaxLimitRequestMemoryRatio bool
	defaultMemoryLimit            resource.Quantity
	defaultMemoryRequest          resource.Quantity
	maxLimitRequestMemoryRatio    resource.Quantity
}

func isLimitRangeTypeContainer(lri corev1.LimitRangeItem) bool {
	return lri.Type == corev1.LimitTypeContainer
}

func getLimitRangeConfig(lri corev1.LimitRangeItem) limitRangeConfig {
	l := limitRangeConfig{}

	l.defaultMemoryRequest, l.hasDefaultMemoryRequest = lri.DefaultRequest[corev1.ResourceMemory]
	l.defaultMemoryLimit, l.hasDefaultMemoryLimit = lri.Default[corev1.ResourceMemory]
	l.maxLimitRequestMemoryRatio, l.hasMaxLimitRequestMemoryRatio = lri.MaxLimitRequestRatio[corev1.ResourceMemory]

	return l
}
