package limitrange

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	corev1Listers "k8s.io/client-go/listers/core/v1"
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

//LimitRange provides an implementation of the LimitRanger interface defined in admission. This implementation is designed to provider a generic config sourcing tool for all mutation handlers.
type LimitRange struct {
	lister corev1Listers.LimitRangeLister
}

//LimitRangeConfig takes a namespace string and returns a Config for memory or a nil if no limit range of type Container is found in the namespace. It returns a non-nil error if there is an error sourcing data from the cluster api or the namespace name is empty
func (lr *LimitRange) LimitRangeConfig(namespace string) (*Config, error) {
	if namespace == "" {
		return nil, fmt.Errorf("Invalid namespace: %q", namespace)
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
