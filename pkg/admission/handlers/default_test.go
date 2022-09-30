package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func TestDefaultDecoderInjector(t *testing.T) {
	d := &DefaultDecoderInjector{}
	scheme := runtime.NewScheme()
	decoder, err := admission.NewDecoder(scheme)
	assert.NoError(t, err)
	assert.NoError(t, d.InjectDecoder(decoder))
}

func TestPatchResponse_ErrorsOnNil(t *testing.T) {
	resp := PatchResponse([]byte{}, "not json")
	assert.Equal(t, false, resp.Allowed)
}

func TestPatchResponse_OK(t *testing.T) {
	d := struct {
		Hello string
	}{
		Hello: "world",
	}
	resp := PatchResponse([]byte("{}"), &d)
	assert.Equal(t, true, resp.Allowed)
}
