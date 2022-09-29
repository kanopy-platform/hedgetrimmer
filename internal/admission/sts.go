package admission

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type STSHandler struct {
	decoder *admission.Decoder
	ptm     PodTemplateSpecMutator
}

func NewSTSHandler(ptm PodTemplateSpecMutator) *STSHandler {
	return &STSHandler{ptm: ptm}
}

func (sts *STSHandler) Kind() string { return "StatefulSet" }

func (sts *STSHandler) InjectDecoder(d *admission.Decoder) error {
	sts.decoder = d
	return nil
}

func (sts *STSHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := log.FromContext(ctx)

	lrConfig := ctx.Value(LimitRangeContextTypeMemory)
	if lrConfig == nil {
		reason := fmt.Sprintf("failed to list LimitRanges in namespace: %s", req.Namespace)
		log.Error(err, reason)
		//If we cannot get LimitRanges due to an api error fail.
		return admission.Denied(reason)
	}

	in := &appsv1.StatefulSet{}
	err := sts.decoder.Decode(req, in)
	if err != nil {
		log.Error(err, fmt.Sprintf("failed to decode statefulset requests: %s", req.Name))
		return admission.Errored(http.StatusBadRequest, err)
	}

	var out appsv1.StatefulSet
	in.DeepCopyInto(&out)

	pts, err := sts.ptm.Mutate(out.Spec.Template, lrConfig)
	if err != nil {
		reason := fmt.Sprintf("Failed to mutate statefulset %s/%s: %s", in.Namespace, in.Name, err)
		log.Error(err, reason)
		return admission.Denied(reason)
	}

	out.Spec.Template = pts

	jout, err := json.Marshal(out)
	if err != nil {
		reason := fmt.Sprintf("Failed to marhsal statefulset %s/%s: %s", in.Namespace, in.Name, err)
		log.Error(err, reason)
		return admission.Denied(reason)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, jout)

}
