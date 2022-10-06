package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/kanopy-platform/hedgetrimmer/pkg/limitrange"
	"github.com/stretchr/testify/assert"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type MockMutator struct {
	spec    corev1.PodTemplateSpec
	err     error
	mutated bool
}

func (mm *MockMutator) SetSpec(spec corev1.PodTemplateSpec) {
	mm.spec = spec
}

func (mm *MockMutator) SetErr(err error) {
	mm.err = err
}

func (mm *MockMutator) SetMutated(mutated bool) {
	mm.mutated = mutated
}

func (mm *MockMutator) Mutate(inputs corev1.PodTemplateSpec, config *limitrange.Config) (corev1.PodTemplateSpec, bool, error) {
	return mm.spec, mm.mutated, mm.err
}

func assertDecoder(t *testing.T, s *runtime.Scheme) *admission.Decoder {
	decoder, err := admission.NewDecoder(s)
	assert.NoError(t, err)
	return decoder
}

func testHandler(t *testing.T, in runtime.Object, mm *MockMutator, handler admission.Handler) {
	bytes, err := json.Marshal(in)
	assert.NoError(t, err)

	ar := admissionv1.AdmissionRequest{
		Object: runtime.RawExtension{
			Raw: bytes,
		},
	}

	tests := []struct {
		config  *limitrange.Config
		lrerr   error
		pts     corev1.PodTemplateSpec
		merr    error
		reject  bool
		msg     string
		mutated bool
	}{
		{
			reject: true,
			config: &limitrange.Config{},
			merr:   fmt.Errorf("Fail"),
			msg:    "Reject for failed mutation",
		},
		{
			config: &limitrange.Config{},
			pts: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "mutated",
				},
			},
			msg:     "Allow for a namespace with no limitranges",
			mutated: true,
		},
	}

	for _, test := range tests {
		mm.SetSpec(test.pts)
		mm.SetErr(test.merr)

		ctx := context.WithValue(context.Background(), limitrange.LimitRangeContextTypeMemory, test.config)

		resp := handler.Handle(ctx, admission.Request{AdmissionRequest: ar})
		assert.Equal(t, test.reject, !resp.Allowed, test.msg)
		assert.Equal(t, test.mutated, len(resp.Patches) > 0)
	}
}
