package mutators

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
)

type OptionsFunc func(*PodTemplateSpec)

func WithDefaultMemoryLimitRequestRatio(ratio float64) OptionsFunc {
	return func(pts *PodTemplateSpec) {
		pts.defaultMemoryLimitRequestRatio = resource.MustParse(fmt.Sprintf("%v", ratio))
	}
}
