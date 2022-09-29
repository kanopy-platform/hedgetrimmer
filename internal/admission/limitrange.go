package admission

import (
	"github.com/kanopy-platform/hedgetrimmer/internal/limitrange"
)

type LimitRanger interface {
	LimitRangeConfig(namespace string) (*limitrange.Config, error)
}
