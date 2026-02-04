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

func TestPodHandler(t *testing.T) {
	t.Parallel()
	mutator := &MockMutator{}

	scheme := runtime.NewScheme()
	decoder := admission.NewDecoder(scheme)

	handler := NewPodHandler(decoder, mutator)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cronjob",
			Namespace: "test-ns",
		},
		Spec: corev1.PodSpec{},
	}
	bytes, err := json.Marshal(pod)
	assert.NoError(t, err)

	ar := admissionv1.AdmissionRequest{
		Object: runtime.RawExtension{
			Raw: bytes,
		},
		Operation: admissionv1.Create,
	}

	tests := []struct {
		config  *limitrange.Config
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
				Spec: corev1.PodSpec{
					NodeName: "mutated",
				},
			},
			msg:     "Allow for a namespace with no limitranges",
			mutated: true,
		},
	}

	for _, test := range tests {
		mutator.SetSpec(test.pts)
		mutator.SetErr(test.merr)
		ctx := context.WithValue(context.Background(), limitrange.LimitRangeContextTypeMemory, test.config)

		resp := handler.Handle(ctx, admission.Request{AdmissionRequest: ar})
		assert.Equal(t, test.reject, !resp.Allowed, test.msg)
		assert.Equal(t, test.mutated, len(resp.Patches) > 0)
	}
}
