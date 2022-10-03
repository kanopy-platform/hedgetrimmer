package handlers

import (
	"context"
	"fmt"
	"net/http"

	batchv1 "k8s.io/api/batch/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/kanopy-platform/hedgetrimmer/pkg/admission"
	"github.com/kanopy-platform/hedgetrimmer/pkg/limitrange"
	kadmission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type CronjobHandler struct {
	DefaultDecoderInjector
	ptm admission.PodTemplateSpecMutator
}

func NewCronjobHandler(ptm admission.PodTemplateSpecMutator) *CronjobHandler {
	return &CronjobHandler{
		ptm: ptm,
	}
}

func (c *CronjobHandler) Kind() string {
	return "CronJob"
}

func (c *CronjobHandler) Handle(ctx context.Context, req kadmission.Request) kadmission.Response {
	log := log.FromContext(ctx)

	lrConfig, err := limitrange.MemoryConfigFromContext(ctx)
	if err != nil {
		reason := fmt.Sprintf("invalid LimitRange config for namespace: %s", req.Namespace)
		log.Error(err, reason)
		return kadmission.Denied(reason)
	}

	out := &batchv1.CronJob{}
	if err := c.decoder.Decode(req, out); err != nil {
		log.Error(err, "failed to decode CronJob request: %s", req.Name)
		return kadmission.Errored(http.StatusBadRequest, err)
	}

	pts, err := c.ptm.Mutate(out.Spec.JobTemplate.Spec.Template, lrConfig)
	if err != nil {
		reason := fmt.Sprintf("failed to mutate CronJob %s/%s: %s", out.Namespace, out.Name, err)
		log.Error(err, reason)
		return kadmission.Denied(reason)
	}

	out.Spec.JobTemplate.Spec.Template = pts

	return PatchResponse(req.Object.Raw, out)
}
