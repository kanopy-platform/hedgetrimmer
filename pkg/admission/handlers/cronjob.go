package handlers

import (
	"context"
	"fmt"
	"net/http"

	batchv1 "k8s.io/api/batch/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/kanopy-platform/hedgetrimmer/internal/admission"
	"github.com/kanopy-platform/hedgetrimmer/internal/limitrange"
	admissionruntime "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type CronjobHandler struct {
	DefaultHandler
	mutator admission.PodTemplateSpecMutator
}

func NewCronjobHandler(mutator admission.PodTemplateSpecMutator) *CronjobHandler {
	return &CronjobHandler{
		mutator: mutator,
	}
}

func (ch *CronjobHandler) Kind() string {
	return "CronJob"
}

func (ch *CronjobHandler) Handle(ctx context.Context, req admissionruntime.Request) admissionruntime.Response {
	log := log.FromContext(ctx)

	lrConfig, ok := ctx.Value(admission.LimitRangeContextTypeMemory).(*limitrange.Config)
	if !ok {
		err := fmt.Errorf("failed to get LimitRange config from context")
		log.Error(err, "failed to get value for %s from context", admission.LimitRangeContextTypeMemory)
		return admissionruntime.Errored(http.StatusBadRequest, err)
	}
	if lrConfig == nil {
		err := fmt.Errorf("got nil LimitRange config from context")
		log.Error(err, "LimitRange config for %s is nil", admission.LimitRangeContextTypeMemory)
		return admissionruntime.Errored(http.StatusBadRequest, err)
	}

	cronjob := &batchv1.CronJob{}
	if err := ch.Decoder.Decode(req, cronjob); err != nil {
		log.Error(err, "failed to decode CronJob request: %s", req.Name)
		return admissionruntime.Errored(http.StatusBadRequest, err)
	}

	mutatedSpec, err := ch.mutator.Mutate(cronjob.Spec.JobTemplate.Spec.Template, lrConfig)
	if err != nil {
		log.Error(err, "failed to mutate CronJob request: %s", req.Name)
		return admissionruntime.Errored(http.StatusBadRequest, err)
	}

	cronjob.Spec.JobTemplate.Spec.Template = mutatedSpec

	return ch.PatchResponse(req.Object.Raw, cronjob)
}
