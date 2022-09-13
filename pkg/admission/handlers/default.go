package handlers

import (
	"context"
	"encoding/json"
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

func (dh *DefaultHandler) PatchResponse(raw []byte, v interface{}) admission.Response {
	pjson, err := json.Marshal(v)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	return admission.PatchResponseFromRaw(raw, pjson)
}
