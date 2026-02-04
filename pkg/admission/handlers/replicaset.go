package handlers

import (
	"context"
	"fmt"
	"net/http"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/kanopy-platform/hedgetrimmer/pkg/admission"
	"github.com/kanopy-platform/hedgetrimmer/pkg/limitrange"
	kadmission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type ReplicaSetHandler struct {
	AllVersionSupporter
	decoder kadmission.Decoder
	ptm     admission.PodTemplateSpecMutator
}

func NewReplicaSetHandler(decoder kadmission.Decoder, ptm admission.PodTemplateSpecMutator) *ReplicaSetHandler {
	return &ReplicaSetHandler{decoder: decoder, ptm: ptm}
}

func (r *ReplicaSetHandler) Kind() string {
	return "ReplicaSet"
}

func (r *ReplicaSetHandler) Handle(ctx context.Context, req kadmission.Request) kadmission.Response {
	log := log.FromContext(ctx)

	lrConfig, err := limitrange.MemoryConfigFromContext(ctx)
	if err != nil {
		reason := fmt.Sprintf("invalid LimitRange config for namespace: %s", req.Namespace)
		log.Error(err, reason)
		return kadmission.Denied(reason)
	}

	out := &appsv1.ReplicaSet{}
	if err := r.decoder.Decode(req, out); err != nil {
		log.Error(err, "failed to decode ReplicaSet request: %s", req.Name)
		return kadmission.Errored(http.StatusBadRequest, err)
	}

	pts, err := r.ptm.Mutate(ctx, out.Spec.Template, lrConfig)
	if err != nil {
		reason := fmt.Sprintf("failed to mutate ReplicaSet %s/%s: %s", out.Namespace, out.Name, err)
		log.Error(err, reason)
		return kadmission.Denied(reason)
	}

	out.Spec.Template = pts

	return PatchResponse(req.Object.Raw, out)
}
