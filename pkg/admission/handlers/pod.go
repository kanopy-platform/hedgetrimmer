package handlers

import (
	"context"
	"fmt"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/kanopy-platform/hedgetrimmer/pkg/admission"
	"github.com/kanopy-platform/hedgetrimmer/pkg/limitrange"
	kadmission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type PodHandler struct {
	DefaultDecoderInjector
	ptm admission.PodTemplateSpecMutator
}

func NewPodHandler(ptm admission.PodTemplateSpecMutator) *PodHandler {
	return &PodHandler{ptm: ptm}
}

func (p *PodHandler) Kind() string {
	return "Pod"
}

func (p *PodHandler) Handle(ctx context.Context, req kadmission.Request) kadmission.Response {
	log := log.FromContext(ctx)

	if req.Operation != admissionv1.Create {
		fmt.Println(req.Operation)
		return kadmission.Allowed("pod resources are immutable")
	}

	lrConfig, err := limitrange.MemoryConfigFromContext(ctx)
	if err != nil {
		reason := fmt.Sprintf("invalid LimitRange config for namespace: %s", req.Namespace)
		log.Error(err, reason)
		return kadmission.Denied(reason)
	}

	out := corev1.Pod{}
	if err := p.decoder.Decode(req, &out); err != nil {
		log.Error(err, "failed to decode request: %s", req.Name)
		return kadmission.Errored(http.StatusBadRequest, err)
	}

	//Shove the PodSpec into a PTS to leverage the existing mutator
	mout := corev1.PodTemplateSpec{
		Spec: out.Spec,
	}

	pts, err := p.ptm.Mutate(mout, lrConfig)
	if err != nil {
		reason := fmt.Sprintf("failed to mutate pod %s/%s: %s", out.Namespace, out.Name, err)
		log.Error(err, reason)
		return kadmission.Denied(reason)
	}

	//Pull the mutated spec off of the PTS and replace the Pod.Spec with it
	out.Spec = pts.Spec
	return PatchResponse(req.Object.Raw, &out)
}
