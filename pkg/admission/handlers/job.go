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

type JobHandler struct {
	DefaultDecoderInjector
	AllVersionSupporter
	ptm admission.PodTemplateSpecMutator
}

func NewJobHandler(ptm admission.PodTemplateSpecMutator) *JobHandler {
	return &JobHandler{ptm: ptm}
}

func (j *JobHandler) Kind() string {
	return "Job"
}

func (j *JobHandler) Handle(ctx context.Context, req kadmission.Request) kadmission.Response {
	log := log.FromContext(ctx)

	lrConfig, err := limitrange.MemoryConfigFromContext(ctx)
	if err != nil {
		reason := fmt.Sprintf("invalid LimitRange config for namespace: %s", req.Namespace)
		log.Error(err, reason)
		return kadmission.Denied(reason)
	}

	out := &batchv1.Job{}
	if err := j.decoder.Decode(req, out); err != nil {
		log.Error(err, "failed to decode job request: %s", req.Name)
		return kadmission.Errored(http.StatusBadRequest, err)
	}

	pts, err := j.ptm.Mutate(out.Spec.Template, lrConfig)
	if err != nil {
		reason := fmt.Sprintf("failed to mutate job %s/%s: %s", out.Namespace, out.Name, err)
		log.Error(err, reason)
		return kadmission.Denied(reason)
	}

	out.Spec.Template = pts

	return PatchResponse(req.Object.Raw, out)
}
