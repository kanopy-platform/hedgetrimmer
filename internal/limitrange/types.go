package limitrange

import "k8s.io/apimachinery/pkg/api/resource"

type Config struct {
	HasDefaultRequest       bool
	HasDefaultLimit         bool
	HasMaxLimitRequestRatio bool
	DefaultLimit            resource.Quantity
	DefaultRequest          resource.Quantity
	MaxLimitRequestRatio    resource.Quantity
}
