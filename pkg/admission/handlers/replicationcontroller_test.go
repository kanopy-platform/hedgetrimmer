package handlers

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func TestReplicationControllerHandler(t *testing.T) {
	t.Parallel()
	mutator := &MockMutator{}

	scheme := runtime.NewScheme()
	decoder := admission.NewDecoder(scheme)

	handler := NewReplicationControllerHandler(decoder, mutator)

	rc := &corev1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-replicationcontroller",
			Namespace: "test-ns",
		},
		Spec: corev1.ReplicationControllerSpec{
			Template: &corev1.PodTemplateSpec{},
		},
	}

	testHandler(t, rc, mutator, handler)
}
