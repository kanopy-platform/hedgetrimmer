package handlers

import (
	"context"
	"fmt"
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type DefaultHandler struct {
	Decoder *admission.Decoder
}

func (dh *DefaultHandler) Kind() string {
	return "default"
}

func (dh *DefaultHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	return admission.Errored(http.StatusBadRequest, fmt.Errorf("resource %s not implemented", dh.Kind()))
}

func (dh *DefaultHandler) InjectDecoder(d *admission.Decoder) error {
	if d == nil {
		return fmt.Errorf("decoder cannot be nil")
	}
	dh.Decoder = d
	return nil
}
