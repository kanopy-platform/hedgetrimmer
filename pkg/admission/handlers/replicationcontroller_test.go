package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestReplicationControllerHandler(t *testing.T) {
	t.Parallel()
	mutator := &MockMutator{}

	scheme := runtime.NewScheme()
	decoder := assertDecoder(t, scheme)

	handler := NewReplicationControllerHandler(mutator)
	assert.NoError(t, handler.InjectDecoder(decoder))

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
