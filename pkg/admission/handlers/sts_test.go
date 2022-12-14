package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestStatefulSetHandler(t *testing.T) {
	t.Parallel()
	mm := MockMutator{}

	scheme := runtime.NewScheme()
	decoder := assertDecoder(t, scheme)

	handler := NewStatefulSetHandler(&mm)

	assert.NoError(t, handler.InjectDecoder(decoder))

	sts := appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-sts",
			Namespace: "test-ns",
		},
		Spec: appsv1.StatefulSetSpec{},
	}

	testHandler(t, &sts, &mm, handler)
}
