package handlers

import (
	"context"
	"fmt"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/kanopy-platform/hedgetrimmer/pkg/admission"
	"github.com/kanopy-platform/hedgetrimmer/pkg/limitrange"
	kadmission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type ReplicationControllerHandler struct {
	DefaultDecoderInjector
	AllVersionSupporter
	ptm admission.PodTemplateSpecMutator
}

func NewReplicationControllerHandler(ptm admission.PodTemplateSpecMutator) *ReplicationControllerHandler {
	return &ReplicationControllerHandler{ptm: ptm}
}

func (r *ReplicationControllerHandler) Kind() string {
	return "ReplicationController"
}

func (r *ReplicationControllerHandler) Handle(ctx context.Context, req kadmission.Request) kadmission.Response {
	log := log.FromContext(ctx)

	lrConfig, err := limitrange.MemoryConfigFromContext(ctx)
	if err != nil {
		reason := fmt.Sprintf("invalid LimitRange config for namespace: %s", req.Namespace)
		log.Error(err, reason)
		return kadmission.Denied(reason)
	}

	out := &corev1.ReplicationController{}
	if err := r.decoder.Decode(req, out); err != nil {
		log.Error(err, "failed to decode ReplicationController request: %s", req.Name)
		return kadmission.Errored(http.StatusBadRequest, err)
	}

	pts, mutated, err := r.ptm.Mutate(*out.Spec.Template, lrConfig)
	if err != nil {
		reason := fmt.Sprintf("failed to mutate ReplicationController %s/%s: %s", out.Namespace, out.Name, err)
		log.Error(err, reason)
		return kadmission.Denied(reason)
	}

	out.Spec.Template = &pts

	return PatchResponse(req.Object.Raw, mutated, out)
}
