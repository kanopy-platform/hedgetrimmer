package handlers

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func TestDaemonSetHandler(t *testing.T) {
	t.Parallel()
	mutator := &MockMutator{}

	scheme := runtime.NewScheme()
	decoder := admission.NewDecoder(scheme)

	handler := NewDaemonSetHandler(decoder, mutator)

	d := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-daemonset",
			Namespace: "test-ns",
		},
		Spec: appsv1.DaemonSetSpec{},
	}

	testHandler(t, d, mutator, handler)
}
