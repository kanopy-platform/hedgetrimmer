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

type StatefulSetHandler struct {
	AllVersionSupporter
	decoder kadmission.Decoder
	ptm     admission.PodTemplateSpecMutator
}

func NewStatefulSetHandler(decoder kadmission.Decoder, ptm admission.PodTemplateSpecMutator) *StatefulSetHandler {
	return &StatefulSetHandler{decoder: decoder, ptm: ptm}
}

func (sts *StatefulSetHandler) Kind() string { return "StatefulSet" }

func (sts *StatefulSetHandler) Handle(ctx context.Context, req kadmission.Request) kadmission.Response {
	log := log.FromContext(ctx)

	lrConfig, err := limitrange.MemoryConfigFromContext(ctx)
	if err != nil {
		reason := fmt.Sprintf("invalid LimitRange config for namespace: %s", req.Namespace)
		log.Error(err, reason)
		return kadmission.Denied(reason)
	}

	out := &appsv1.StatefulSet{}
	err = sts.decoder.Decode(req, out)
	if err != nil {
		log.Error(err, fmt.Sprintf("failed to decode statefulset requests: %s", req.Name))
		return kadmission.Errored(http.StatusBadRequest, err)
	}

	pts, err := sts.ptm.Mutate(ctx, out.Spec.Template, lrConfig)
	if err != nil {
		reason := fmt.Sprintf("failed to mutate statefulset %s/%s: %s", out.Namespace, out.Name, err)
		log.Error(err, reason)
		return kadmission.Denied(reason)
	}

	out.Spec.Template = pts

	return PatchResponse(req.Object.Raw, out)
}
