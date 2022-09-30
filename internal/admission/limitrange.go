package admission

import (
	"github.com/kanopy-platform/hedgetrimmer/pkg/limitrange"
)

type LimitRanger interface {
	LimitRangeConfig(namespace string) (*limitrange.Config, error)
}
