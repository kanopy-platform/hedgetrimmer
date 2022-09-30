package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestDeploymentHandler(t *testing.T) {
	t.Parallel()
	mm := MockMutator{}

	scheme := runtime.NewScheme()
	decoder := assertDecoder(t, scheme)

	handler := NewDeploymentHandler(&mm)

	assert.NoError(t, handler.InjectDecoder(decoder))

	d := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-d",
			Namespace: "test-ns",
		},
		Spec: appsv1.DeploymentSpec{},
	}

	testHandler(t, &d, &mm, handler)
}
