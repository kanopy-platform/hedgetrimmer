package handlers

import (
	"github.com/kanopy-platform/hedgetrimmer/internal/admission"
)

type CronjobHandler struct {
	DefaultHandler
	mutator admission.PodTemplateSpecMutator
}
