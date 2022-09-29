package admission

import (
	"context"
	"fmt"
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type LimitRangeContextType string

const LimitRangeContextTypeMemory LimitRangeContextType = "memory"

type AdmissionHandler interface {
	admission.Handler
	Kind() string
	InjectDecoder(*admission.Decoder) error
}

type OptionsFunc func(*Router) error

func WithAdmissionHandlers(handlers ...AdmissionHandler) OptionsFunc {
	return func(r *Router) error {
		for _, h := range handlers {
			if _, ok := r.handlers[h.Kind()]; ok {
				return fmt.Errorf("duplicate handler %s registered", h.Kind())
			}
			r.handlers[h.Kind()] = h
		}
		return nil
	}
}

type Router struct {
	handlers    map[string]AdmissionHandler
	limitRanger LimitRanger
}

func NewRouter(lr LimitRanger, opts ...OptionsFunc) (*Router, error) {
	r := &Router{
		handlers:    map[string]AdmissionHandler{},
		limitRanger: lr,
	}

	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (r *Router) SetupWithManager(m manager.Manager) {
	m.GetWebhookServer().Register("/mutate", &webhook.Admission{Handler: r})
}

func (r *Router) InjectDecoder(d *admission.Decoder) error {
	if d == nil {
		return fmt.Errorf("decoder cannot be nil")
	}
	for _, h := range r.handlers {
		if err := h.InjectDecoder(d); err != nil {
			return err
		}
	}
	return nil
}

func (r *Router) Handle(ctx context.Context, req admission.Request) admission.Response {
	h, ok := r.handlers[req.RequestKind.Kind]
	if !ok {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("resource %s not implemented", req.RequestKind.Kind))
	}

	cfg, err := r.limitRanger.LimitRangeConfig(req.Namespace)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed to retrieve limit range information from namespace %s: %s", req.Namespace, err.Error()))
	}

	if cfg == nil {
		return admission.Allowed(fmt.Sprintf("No container limit range in namespace: %s", req.Namespace))
	}

	ctx = context.WithValue(ctx, LimitRangeContextTypeMemory, cfg)
	return h.Handle(ctx, req)
}
