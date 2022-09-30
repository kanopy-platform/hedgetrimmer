package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kanopy-platform/hedgetrimmer/pkg/admission"
	"github.com/kanopy-platform/hedgetrimmer/pkg/limitrange"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
	kadmission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type STSHandler struct {
	DefaultDecoderInjector
	ptm admission.PodTemplateSpecMutator
}

func NewSTSHandler(ptm admission.PodTemplateSpecMutator) *STSHandler {
	return &STSHandler{ptm: ptm}
}

func (sts *STSHandler) Kind() string { return "StatefulSet" }

func (sts *STSHandler) Handle(ctx context.Context, req kadmission.Request) kadmission.Response {
	log := log.FromContext(ctx)

	lrConfig, ok := ctx.Value(limitrange.LimitRangeContextTypeMemory).(*limitrange.Config)
	if !ok || lrConfig == nil {
		reason := fmt.Sprintf("failed to list LimitRanges in namespace: %s", req.Namespace)
		log.Error(fmt.Errorf(reason), reason)
		//If we cannot get LimitRanges due to an api error fail.
		return kadmission.Denied(reason)
	}

	in := &appsv1.StatefulSet{}
	err := sts.decoder.Decode(req, in)
	if err != nil {
		log.Error(err, fmt.Sprintf("failed to decode statefulset requests: %s", req.Name))
		return kadmission.Errored(http.StatusBadRequest, err)
	}

	var out appsv1.StatefulSet
	in.DeepCopyInto(&out)

	pts, err := sts.ptm.Mutate(out.Spec.Template, lrConfig)
	if err != nil {
		reason := fmt.Sprintf("Failed to mutate statefulset %s/%s: %s", in.Namespace, in.Name, err)
		log.Error(err, reason)
		return kadmission.Denied(reason)
	}

	out.Spec.Template = pts

	return PatchResponse(req.Object.Raw, out)
}
