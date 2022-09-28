package limitrange

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"
	corev1Listers "k8s.io/client-go/listers/core/v1"
)

type MockLimitRanger struct {
	nl corev1Listers.LimitRangeNamespaceLister
}

func (mlr *MockLimitRanger) List(selector labels.Selector) (ret []*corev1.LimitRange, err error) {
	return ret, err
}

func (mlr *MockLimitRanger) LimitRanges(namespace string) corev1Listers.LimitRangeNamespaceLister {
	return mlr.nl
}

type MockLimitRangeNamespaceLister struct {
	ranges []*corev1.LimitRange
	err    error
}

func (mlrnl *MockLimitRangeNamespaceLister) List(selector labels.Selector) (ret []*corev1.LimitRange, err error) {
	return mlrnl.ranges, mlrnl.err
}

func (mlrnl *MockLimitRangeNamespaceLister) Get(name string) (*corev1.LimitRange, error) {
	return nil, nil
}

func TestLimitRanger(t *testing.T) {
	t.Parallel()

	emptyNSL := MockLimitRangeNamespaceLister{
		ranges: []*corev1.LimitRange{},
	}
	empty := MockLimitRanger{
		nl: &emptyNSL,
	}
	mlrnl := MockLimitRangeNamespaceLister{
		ranges: []*corev1.LimitRange{
			{
				Spec: corev1.LimitRangeSpec{
					Limits: []corev1.LimitRangeItem{
						{Type: corev1.LimitTypeContainer},
					},
				},
			},
		},
	}
	lrl := MockLimitRanger{
		nl: &mlrnl,
	}

	tests := []struct {
		ns        string
		lr        *LimitRange
		want      *Config
		wantError bool
	}{
		{
			ns:        "",
			lr:        &LimitRange{},
			want:      nil,
			wantError: true,
		},
		{
			ns: "t",
			lr: &LimitRange{
				Lister: &lrl,
			},
			want: &Config{},
		},
		{
			ns: "t",
			lr: &LimitRange{
				Lister: &empty,
			},
			want: nil,
		},
	}

	for _, test := range tests {
		c, e := test.lr.LimitRangeConfig(test.ns)
		assert.Equal(t, test.wantError, e != nil)
		assert.Equal(t, test.want, c)
	}
}

func TestIsLimitRangeTypeContainer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		limitRange corev1.LimitRangeItem
		want       bool
		msg        string
	}{
		{
			limitRange: corev1.LimitRangeItem{Type: corev1.LimitTypeContainer},
			want:       true,
			msg:        "Container type",
		},
		{
			limitRange: corev1.LimitRangeItem{Type: corev1.LimitTypePod},
			want:       false,
			msg:        "Pod type",
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.want, IsLimitRangeTypeContainer(test.limitRange), test.msg)
	}
}

func TestLimitRangeConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		limitRange corev1.LimitRangeItem
		resource   corev1.ResourceName
		want       Config
		msg        string
	}{
		{
			// extract out memory resource
			limitRange: corev1.LimitRangeItem{
				DefaultRequest: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("1Gi"),
					corev1.ResourceCPU:    resource.MustParse("500m"),
				},
				Default: corev1.ResourceList{
					corev1.ResourceMemory:  resource.MustParse("2Gi"),
					corev1.ResourceStorage: resource.MustParse("5Gi"),
				},
				MaxLimitRequestRatio: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("1.5"),
				},
			},
			resource: corev1.ResourceMemory,
			want: Config{
				HasDefaultRequest:       true,
				HasDefaultLimit:         true,
				HasMaxLimitRequestRatio: true,
				DefaultRequest:          resource.MustParse("1Gi"),
				DefaultLimit:            resource.MustParse("2Gi"),
				MaxLimitRequestRatio:    resource.MustParse("1.5"),
			},
			msg: "extract out memory resource",
		},
		{
			// extract out CPU resource
			limitRange: corev1.LimitRangeItem{
				DefaultRequest: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
				Default: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("5Gi"),
				},
				MaxLimitRequestRatio: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("1.5"),
				},
			},
			resource: corev1.ResourceCPU,
			want: Config{
				HasDefaultRequest:       true,
				HasDefaultLimit:         false,
				HasMaxLimitRequestRatio: true,
				DefaultRequest:          resource.MustParse("500m"),
				MaxLimitRequestRatio:    resource.MustParse("1.5"),
			},
			msg: "extract out CPU resource",
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.want, NewConfig(test.limitRange, test.resource), test.msg)
	}
}
