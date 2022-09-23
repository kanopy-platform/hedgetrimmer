package limitrange

import "k8s.io/apimachinery/pkg/api/resource"

type MemoryConfig struct {
	HasDefaultMemoryRequest       bool
	HasDefaultMemoryLimit         bool
	HasMaxLimitRequestMemoryRatio bool
	DefaultMemoryLimit            resource.Quantity
	DefaultMemoryRequest          resource.Quantity
	MaxLimitRequestMemoryRatio    resource.Quantity
}
