package limitrange

import (
	corev1 "k8s.io/api/core/v1"
)

func IsLimitRangeTypeContainer(lri corev1.LimitRangeItem) bool {
	return lri.Type == corev1.LimitTypeContainer
}

func NewConfig(lri corev1.LimitRangeItem, resource corev1.ResourceName) Config {
	l := Config{}

	l.DefaultRequest, l.HasDefaultRequest = lri.DefaultRequest[resource]
	l.DefaultLimit, l.HasDefaultLimit = lri.Default[resource]
	l.MaxLimitRequestRatio, l.HasMaxLimitRequestRatio = lri.MaxLimitRequestRatio[resource]

	return l
}
