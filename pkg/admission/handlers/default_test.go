package handlers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func TestDefaultHandler_HandleNotImplemented(t *testing.T) {
	h := &DefaultHandler{}
	r := h.Handle(context.TODO(), admission.Request{})
	assert.False(t, r.Allowed)
}
