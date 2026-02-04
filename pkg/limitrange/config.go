package limitrange

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"
	corev1Listers "k8s.io/client-go/listers/core/v1"
)

type LimitRangeContextType string

const LimitRangeContextTypeMemory LimitRangeContextType = "memory"

type Config struct {
	HasDefaultRequest       bool
	HasDefaultLimit         bool
	HasMaxLimitRequestRatio bool
	DefaultLimit            resource.Quantity
	DefaultRequest          resource.Quantity
	MaxLimitRequestRatio    resource.Quantity
}

func NewConfig(lri corev1.LimitRangeItem, resource corev1.ResourceName) Config {
	l := Config{}

	l.DefaultRequest, l.HasDefaultRequest = lri.DefaultRequest[resource]
	l.DefaultLimit, l.HasDefaultLimit = lri.Default[resource]
	l.MaxLimitRequestRatio, l.HasMaxLimitRequestRatio = lri.MaxLimitRequestRatio[resource]

	return l
}

func MemoryConfigFromContext(ctx context.Context) (*Config, error) {
	return configFromContext(ctx, LimitRangeContextTypeMemory)
}

func WithMemoryConfig(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, LimitRangeContextTypeMemory, cfg)
}

func configFromContext(ctx context.Context, key LimitRangeContextType) (*Config, error) {
	lrc, ok := ctx.Value(key).(*Config)
	if !ok || lrc == nil {
		return nil, fmt.Errorf("invalid limitrange Config")
	}

	return lrc, nil
}

// LimitRange provides an implementation of the LimitRanger interface defined in admission. This implementation is designed to provider a generic config sourcing tool for all mutation handlers.
type LimitRange struct {
	lister corev1Listers.LimitRangeLister
}

// NewLimitRanger take a limitrangelister and returns a pointer to a configured LimitRange. This satisfies the LimitRanger interface
func NewLimitRanger(lister corev1Listers.LimitRangeLister) *LimitRange {
	return &LimitRange{lister: lister}
}

// LimitRangeConfig takes a namespace string and returns a Config for memory or a nil if no limit range of type Container is found in the namespace. It returns a non-nil error if there is an error sourcing data from the cluster api or the namespace name is empty
func (lr *LimitRange) LimitRangeConfig(namespace string) (*Config, error) {
	if namespace == "" {
		return nil, fmt.Errorf("invalid namespace: %q", namespace)
	}
	ranges, err := lr.lister.LimitRanges(namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	for _, lr := range ranges {
		for _, item := range lr.Spec.Limits {
			if item.Type == corev1.LimitTypeContainer {
				config := NewConfig(item, corev1.ResourceMemory)
				return &config, nil
			}
		}
	}
	return nil, nil
}
