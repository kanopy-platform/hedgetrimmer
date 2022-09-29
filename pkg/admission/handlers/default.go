package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kanopy-platform/hedgetrimmer/pkg/admission/handlers/common"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type DefaultHandler struct {
	common.Decode
	common.Patch
}

func (dh *DefaultHandler) Kind() string {
	return "default"
}

func (dh *DefaultHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	return admission.Errored(http.StatusBadRequest, fmt.Errorf("resource %s not implemented", dh.Kind()))
}
