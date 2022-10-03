package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestCronjobHandler(t *testing.T) {
	t.Parallel()
	mutator := &MockMutator{}

	scheme := runtime.NewScheme()
	decoder := assertDecoder(t, scheme)

	handler := NewCronjobHandler(mutator)
	assert.NoError(t, handler.InjectDecoder(decoder))

	cronjob := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cronjob",
			Namespace: "test-ns",
		},
		Spec: batchv1.CronJobSpec{},
	}

	testHandler(t, cronjob, mutator, handler)
}
