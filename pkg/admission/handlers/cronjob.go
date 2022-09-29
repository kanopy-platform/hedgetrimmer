package handlers

import (
	"context"
	"net/http"

	batchv1 "k8s.io/api/batch/v1"
	klog "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/kanopy-platform/hedgetrimmer/internal/admission"
	admissionruntime "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type CronjobHandler struct {
	DefaultHandler
	mutator     admission.PodTemplateSpecMutator
	limitranger admission.LimitRanger
}

func NewCronjobHandler(mutator admission.PodTemplateSpecMutator, limitranger admission.LimitRanger) *CronjobHandler {
	return &CronjobHandler{
		mutator:     mutator,
		limitranger: limitranger,
	}
}

func (ch *CronjobHandler) Handle(ctx context.Context, req admissionruntime.Request) admissionruntime.Response {
	cronjob := &batchv1.CronJob{}

	if err := ch.Decoder.Decode(req, cronjob); err != nil {
		klog.Log.Error(err, "failed to decode resource")
		return admissionruntime.Errored(http.StatusBadRequest, err)
	}

	config, err := ch.limitranger.LimitRangeConfig(cronjob.Namespace)
	if err != nil {
		klog.Log.Error(err, "failed to get LimitRange config for namespace %s", cronjob.Namespace)
		return admissionruntime.Errored(http.StatusBadRequest, err)
	}

	mutatedSpec, err := ch.mutator.Mutate(cronjob.Spec.JobTemplate.Spec.Template, config)
	if err != nil {
		klog.Log.Error(err, "failed to mutate CronJob PodTemplateSpec")
		return admissionruntime.Errored(http.StatusBadRequest, err)
	}

	cronjob.Spec.JobTemplate.Spec.Template = mutatedSpec

	return ch.PatchResponse(req.Object.Raw, cronjob)
}

func (ch *CronjobHandler) Kind() string {
	return "CronJob"
}
