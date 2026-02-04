package handlers

import (
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func TestJobHandler(t *testing.T) {
	t.Parallel()
	mutator := &MockMutator{}

	scheme := runtime.NewScheme()
	decoder := admission.NewDecoder(scheme)

	handler := NewJobHandler(decoder, mutator)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-job",
			Namespace: "test-ns",
		},
		Spec: batchv1.JobSpec{},
	}

	testHandler(t, job, mutator, handler)
}
