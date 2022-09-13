package handlers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func TestDefaultHandler_HandleNotImplemented(t *testing.T) {
	h := &DefaultHandler{}
	r := h.Handle(context.TODO(), admission.Request{})
	assert.False(t, r.Allowed)
}

func TestDefaultHandler_InjectDecoder_FailsOnNil(t *testing.T) {
	h := &DefaultHandler{}
	assert.Error(t, h.InjectDecoder(nil))
}

func TestDefaultHandler_InjectDecoder(t *testing.T) {
	h := &DefaultHandler{}
	scheme := runtime.NewScheme()
	decoder, err := admission.NewDecoder(scheme)
	assert.NoError(t, err)
	assert.NoError(t, h.InjectDecoder(decoder))
}

func TestDefaultHandler_PatchResponse_ErrorsOnNil(t *testing.T) {
	h := &DefaultHandler{}
	resp := h.PatchResponse([]byte{}, "not json")
	assert.Equal(t, false, resp.Allowed)
}

func TestDefaultHandler_PatchResponse_OK(t *testing.T) {
	h := &DefaultHandler{}
	d := struct {
		Hello string
	}{
		Hello: "world",
	}
	resp := h.PatchResponse([]byte("{}"), &d)
	assert.Equal(t, true, resp.Allowed)
}
