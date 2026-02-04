package handlers

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func TestDeploymentHandler(t *testing.T) {
	t.Parallel()
	mm := MockMutator{}

	scheme := runtime.NewScheme()
	decoder := admission.NewDecoder(scheme)

	handler := NewDeploymentHandler(decoder, &mm)

	d := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-d",
			Namespace: "test-ns",
		},
		Spec: appsv1.DeploymentSpec{},
	}

	testHandler(t, &d, &mm, handler)
}
