package mutators

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestWithDefaultMemoryLimitRequestRatio(t *testing.T) {
	t.Parallel()

	pts := NewPodTemplateSpec(WithDefaultMemoryLimitRequestRatio(12.3456))
	assert.Equal(t, resource.MustParse("12.3456"), pts.defaultMemoryLimitRequestRatio)
}
