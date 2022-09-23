package limitrange

import (
	corev1 "k8s.io/api/core/v1"
)

func IsLimitRangeTypeContainer(lri corev1.LimitRangeItem) bool {
	return lri.Type == corev1.LimitTypeContainer
}

func GetMemoryConfig(lri corev1.LimitRangeItem) MemoryConfig {
	l := MemoryConfig{}

	l.DefaultMemoryRequest, l.HasDefaultMemoryRequest = lri.DefaultRequest[corev1.ResourceMemory]
	l.DefaultMemoryLimit, l.HasDefaultMemoryLimit = lri.Default[corev1.ResourceMemory]
	l.MaxLimitRequestMemoryRatio, l.HasMaxLimitRequestMemoryRatio = lri.MaxLimitRequestRatio[corev1.ResourceMemory]

	return l
}
