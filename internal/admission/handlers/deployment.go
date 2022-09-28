package handlers

import (
	"context"
	"fmt"
	"net/http"

	appsv1 "k8s.io/api/apps/v1"
	klog "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/kanopy-platform/hedgetrimmer/internal/admission"
	"github.com/kanopy-platform/hedgetrimmer/internal/limitrange"
	"github.com/kanopy-platform/hedgetrimmer/pkg/admission/handlers"
	admissionruntime "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type DeploymentHandler struct {
	handlers.DefaultHandler
	mutator admission.PodTemplateSpecMutator
}

func NewDeploymentHandler(mutator admission.PodTemplateSpecMutator) *DeploymentHandler {
	return &DeploymentHandler{
		mutator: mutator,
	}
}

func (dh *DeploymentHandler) Handle(ctx context.Context, req admissionruntime.Request) admissionruntime.Response {
	d := &appsv1.Deployment{}
	if err := dh.Decoder.Decode(req, d); err != nil {
		klog.Log.Info("Failed to decode resource")
		return admissionruntime.Errored(http.StatusBadRequest, err)
	}

	cfg := ctx.Value("LIMIT_RANGER").(limitrange.Config)

	mutatedSpec, err := dh.mutator.Mutate(d.Spec.Template, cfg) // OR this should be adjusted to accept a cfg pointer.
	if err != nil {
		return admissionruntime.Errored(http.StatusBadRequest, fmt.Errorf("resource failed to mutate: %v", err))
	}
	d.Spec.Template = mutatedSpec

	return dh.PatchResponse(req.Object.Raw, d)
}

func (d *DeploymentHandler) Kind() string {
	return "Deployment"
}
