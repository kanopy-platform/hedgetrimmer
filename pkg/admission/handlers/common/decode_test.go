package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func TestInjectDecoder_FailsOnNil(t *testing.T) {
	t.Parallel()
	h := &Decode{}
	assert.Error(t, h.InjectDecoder(nil))
}

func TestInjectDecoder(t *testing.T) {
	t.Parallel()
	h := &Decode{}
	scheme := runtime.NewScheme()
	decoder, err := admission.NewDecoder(scheme)
	assert.NoError(t, err)
	assert.NoError(t, h.InjectDecoder(decoder))
}
