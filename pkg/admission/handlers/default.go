package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type DefaultDecoderInjector struct {
	decoder *admission.Decoder
}

type AllVersionSupporter struct{}

func (s *AllVersionSupporter) VersionSupported(v string) bool {
	return true
}

func (d *DefaultDecoderInjector) InjectDecoder(decoder *admission.Decoder) error {
	if decoder == nil {
		return fmt.Errorf("decoder cannot be nil")
	}
	d.decoder = decoder
	return nil
}

func PatchResponse(raw []byte, v interface{}) admission.Response {
	pjson, err := json.Marshal(v)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	return admission.PatchResponseFromRaw(raw, pjson)
}
