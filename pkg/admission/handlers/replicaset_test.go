package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestReplicaSetHandler(t *testing.T) {
	t.Parallel()
	mutator := &MockMutator{}

	scheme := runtime.NewScheme()
	decoder := assertDecoder(t, scheme)

	handler := NewReplicaSetHandler(mutator)
	assert.NoError(t, handler.InjectDecoder(decoder))

	r := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-replicaset",
			Namespace: "test-ns",
		},
		Spec: appsv1.ReplicaSetSpec{},
	}

	testHandler(t, r, mutator, handler)
}
