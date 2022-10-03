package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestJobHandler(t *testing.T) {
	t.Parallel()
	mutator := &MockMutator{}

	scheme := runtime.NewScheme()
	decoder := assertDecoder(t, scheme)

	handler := NewJobHandler(mutator)
	assert.NoError(t, handler.InjectDecoder(decoder))

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-job",
			Namespace: "test-ns",
		},
		Spec: batchv1.JobSpec{},
	}

	testHandler(t, job, mutator, handler)
}
