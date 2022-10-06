package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const HedgetrimmerMutationAnnotation = "kanopy-platform/hedgetrimmer-mutated"

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

func PatchResponse(raw []byte, mutated bool, v interface{}) admission.Response {

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(v)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	unstruct := unstructured.Unstructured{}
	unstruct.SetUnstructuredContent(obj)

	if mutated {
		annotations := unstruct.GetAnnotations()
		if annotations == nil {
			annotations = map[string]string{}
		}
		annotations[HedgetrimmerMutationAnnotation] = "true"
		unstruct.SetAnnotations(annotations)
	}

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstruct.UnstructuredContent(), v)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	pjson, err := json.Marshal(v)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	return admission.PatchResponseFromRaw(raw, pjson)
}
