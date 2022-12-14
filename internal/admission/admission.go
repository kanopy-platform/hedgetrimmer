package admission

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kanopy-platform/hedgetrimmer/pkg/limitrange"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type AdmissionHandler interface {
	admission.Handler
	Kind() string
	VersionSupported(v string) bool
	InjectDecoder(*admission.Decoder) error
}

type OptionsFunc func(*Router) error

func WithAdmissionHandlers(handlers ...AdmissionHandler) OptionsFunc {
	return func(r *Router) error {
		for _, h := range handlers {
			handlers := r.handlers[h.Kind()]
			r.handlers[h.Kind()] = append(handlers, h)
		}
		return nil
	}
}

type Router struct {
	handlers    map[string][]AdmissionHandler
	limitRanger LimitRanger
}

func NewRouter(lr LimitRanger, opts ...OptionsFunc) (*Router, error) {
	r := &Router{
		handlers:    map[string][]AdmissionHandler{},
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
	for _, hs := range r.handlers {
		for _, h := range hs {
			if err := h.InjectDecoder(d); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Router) Handle(ctx context.Context, req admission.Request) admission.Response {

	kind := req.RequestKind
	if kind == nil {
		kind = &req.Kind
	}

	handlers, ok := r.handlers[kind.Kind]
	if !ok {
		return admission.Allowed(fmt.Sprintf("no handlers for kind: %s", kind.Kind))
	}

	var handler AdmissionHandler
	for _, h := range handlers {
		if h.VersionSupported(kind.Version) {
			handler = h
			break
		}
	}

	if handler == nil {
		return admission.Denied(fmt.Sprintf("no handlers for %s version %s", kind.Kind, kind.Version))
	}

	logr := log.FromContext(ctx,
		"resource", req.Resource,
		"namespace", req.Namespace,
		"name", req.Name,
		"operation", req.Operation,
	)
	ctx = log.IntoContext(ctx, logr)

	cfg, err := r.limitRanger.LimitRangeConfig(req.Namespace)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed to retrieve limit range information from namespace %s: %s", req.Namespace, err.Error()))
	}

	if cfg == nil {
		return admission.Allowed(fmt.Sprintf("No container limit range in namespace: %s", req.Namespace))
	}

	ctx = limitrange.WithMemoryConfig(ctx, cfg)
	return handler.Handle(ctx, req)
}
