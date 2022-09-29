package common

import (
	"encoding/json"
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type Patch struct{}

func (p *Patch) PatchResponse(raw []byte, v interface{}) admission.Response {
	pjson, err := json.Marshal(v)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	return admission.PatchResponseFromRaw(raw, pjson)
}
