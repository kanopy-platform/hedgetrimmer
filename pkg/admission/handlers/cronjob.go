package handlers

import (
	"context"
	"fmt"
	"net/http"

	batchv1 "k8s.io/api/batch/v1"
	klog "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/kanopy-platform/hedgetrimmer/internal/admission"
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

func (ch *CronjobHandler) Handle(ctx context.Context, req admissionruntime.Request) admissionruntime.Response {
	cronjob := &batchv1.CronJob{}

	if err := ch.Decoder.Decode(req, cronjob); err != nil {
		klog.Log.Info("failed to decode resource")
		return admissionruntime.Errored(http.StatusBadRequest, err)
	}

	return ch.PatchResponse(req.Object.Raw, cronjob)
}

func (ch *CronjobHandler) Kind() string {
	return "CronJob"
}

func (ch *CronjobHandler) InjectDecoder(d *admissionruntime.Decoder) error {
	if d == nil {
		return fmt.Errorf("decoder cannot be nil")
	}
	ch.Decoder = d
	return nil
}
