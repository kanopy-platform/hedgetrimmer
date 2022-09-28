package handlers

import (
	"context"
	"net/http"

	appsv1 "k8s.io/api/apps/v1"
	klog "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/kanopy-platform/hedgetrimmer/pkg/admission/handlers"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type DeploymentHandler struct {
	handlers.DefaultHandler
	mutator interface{}
}

func NewDeploymentHandler(mutator interface{}) *DeploymentHandler {
	return &DeploymentHandler{
		mutator: mutator,
	}
}

func (dh *DeploymentHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	d := &appsv1.Deployment{}
	if err := dh.Decoder.Decode(req, d); err != nil {
		klog.Log.Info("Failed to decode resource")
		return admission.Errored(http.StatusBadRequest, err)
	}

	//pts := d.Spec.Template
	// get from context the limit config.
	// pass to dh.mutator

	return dh.PatchResponse(req.Object.Raw, d)
}

func (d *DeploymentHandler) Kind() string {
	return "Deployment"
}
