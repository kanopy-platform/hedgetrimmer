package handlers

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func TestStatefulSetHandler(t *testing.T) {
	t.Parallel()
	mm := MockMutator{}

	scheme := runtime.NewScheme()
	decoder := admission.NewDecoder(scheme)

	handler := NewStatefulSetHandler(decoder, &mm)

	sts := appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-sts",
			Namespace: "test-ns",
		},
		Spec: appsv1.StatefulSetSpec{},
	}

	testHandler(t, &sts, &mm, handler)
}
