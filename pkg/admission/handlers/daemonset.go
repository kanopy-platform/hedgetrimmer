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

type DaemonSetHandler struct {
	DefaultDecoderInjector
	AllVersionSupporter
	ptm admission.PodTemplateSpecMutator
}

func NewDaemonSetHandler(ptm admission.PodTemplateSpecMutator) *DaemonSetHandler {
	return &DaemonSetHandler{ptm: ptm}
}

func (d *DaemonSetHandler) Kind() string {
	return "DaemonSet"
}

func (d *DaemonSetHandler) Handle(ctx context.Context, req kadmission.Request) kadmission.Response {
	log := log.FromContext(ctx)

	lrConfig, err := limitrange.MemoryConfigFromContext(ctx)
	if err != nil {
		reason := fmt.Sprintf("invalid LimitRange config for namespace: %s", req.Namespace)
		log.Error(err, reason)
		return kadmission.Denied(reason)
	}

	out := &appsv1.DaemonSet{}
	if err := d.decoder.Decode(req, out); err != nil {
		log.Error(err, "failed to decode DaemonSet request: %s", req.Name)
		return kadmission.Errored(http.StatusBadRequest, err)
	}

	pts, err := d.ptm.Mutate(ctx, out.Spec.Template, lrConfig)
	if err != nil {
		reason := fmt.Sprintf("failed to mutate DaemonSet %s/%s: %s", out.Namespace, out.Name, err)
		log.Error(err, reason)
		return kadmission.Denied(reason)
	}

	out.Spec.Template = pts

	return PatchResponse(req.Object.Raw, out)
}
