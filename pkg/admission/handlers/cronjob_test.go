package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/kanopy-platform/hedgetrimmer/internal/admission"
	"github.com/kanopy-platform/hedgetrimmer/internal/limitrange"
	"github.com/stretchr/testify/assert"
	admissionv1 "k8s.io/api/admission/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	admissionruntime "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type mockMutator struct {
	spec corev1.PodTemplateSpec
	err  error
}

func (m *mockMutator) setSpec(spec corev1.PodTemplateSpec) {
	m.spec = spec
}

func (m *mockMutator) setErr(err error) {
	m.err = err
}

func (m *mockMutator) Mutate(input corev1.PodTemplateSpec, config *limitrange.Config) (corev1.PodTemplateSpec, error) {
	return m.spec, m.err
}

func TestCronjobHandler_Handle(t *testing.T) {
	t.Parallel()

	mutator := &mockMutator{}
	scheme := runtime.NewScheme()
	decoder, err := admissionruntime.NewDecoder(scheme)
	assert.NoError(t, err)

	handler := NewCronjobHandler(mutator)
	assert.NoError(t, handler.InjectDecoder(decoder))

	cronjob := batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cronjob",
			Namespace: "test-ns",
		},
		Spec: batchv1.CronJobSpec{},
	}

	cronjobBytes, err := json.Marshal(cronjob)
	assert.NoError(t, err)

	req := admissionruntime.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			Object: runtime.RawExtension{
				Raw: cronjobBytes,
			},
		},
	}

	tests := []struct {
		msg          string
		config       *limitrange.Config
		mutatedPts   corev1.PodTemplateSpec
		mutatorError error
		wantAllow    bool
	}{
		{
			msg:    "Success",
			config: &limitrange.Config{},
			mutatedPts: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "mutated",
				},
			},
			mutatorError: nil,
			wantAllow:    true,
		},
		{
			msg:          "Fail due to missing config",
			config:       nil,
			mutatorError: nil,
			wantAllow:    false,
		},
		{
			msg:          "Fail due to mutation failure",
			config:       &limitrange.Config{},
			mutatorError: fmt.Errorf("mutate error"),
			wantAllow:    false,
		},
	}

	for _, test := range tests {
		ctx := context.WithValue(context.TODO(), admission.LimitRangeContextTypeMemory, test.config)
		mutator.setSpec(test.mutatedPts)
		mutator.setErr(test.mutatorError)

		resp := handler.Handle(ctx, req)
		assert.Equal(t, test.wantAllow, resp.Allowed, test.msg)
		if test.wantAllow {
			assert.True(t, len(resp.Patches) > 0, test.msg)
		}
	}
}
