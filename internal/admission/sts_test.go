package admission

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/kanopy-platform/hedgetrimmer/internal/limitrange"
	"github.com/stretchr/testify/assert"
	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type MockMutator struct {
	spec corev1.PodTemplateSpec
	err  error
}

func (mm *MockMutator) SetSpec(spec corev1.PodTemplateSpec) {
	mm.spec = spec
}

func (mm *MockMutator) SetErr(err error) {
	mm.err = err
}

func (mm *MockMutator) Mutate(inputs corev1.PodTemplateSpec, config *limitrange.Config) (corev1.PodTemplateSpec, error) {
	return mm.spec, mm.err
}

func TestSTSHandler(t *testing.T) {
	t.Parallel()
	mm := MockMutator{}
	mlr := MockLimitRanger{}

	scheme := runtime.NewScheme()
	decoder, err := admission.NewDecoder(scheme)
	assert.NoError(t, err)

	handler := NewSTSHandler(&mm, &mlr)

	assert.NoError(t, handler.InjectDecoder(decoder))

	sts := appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-sts",
			Namespace: "test-ns",
		},
		Spec: appsv1.StatefulSetSpec{},
	}

	stsBytes, err := json.Marshal(sts)
	assert.NoError(t, err)

	ar := admissionv1.AdmissionRequest{
		Object: runtime.RawExtension{
			Raw: stsBytes,
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
			lrerr:  fmt.Errorf("Fail"),
			reject: true,
			msg:    "Deny for lr error",
		},
		{
			msg: "Allow for a namespace with no limitranges",
		},
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
		mlr.SetConfig(test.config)
		mlr.SetErr(test.lrerr)
		mm.SetSpec(test.pts)
		mm.SetErr(test.merr)

		resp := handler.Handle(context.TODO(), admission.Request{AdmissionRequest: ar})
		assert.Equal(t, test.reject, !resp.Allowed, test.msg)
		assert.Equal(t, test.mutated, len(resp.Patches) > 0)
	}
}
